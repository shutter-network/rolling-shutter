package gnosis

import (
	"bytes"
	"context"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gethLog "github.com/ethereum/go-ethereum/log"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	"golang.org/x/exp/slog"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

var ErrParseKeyperSet = errors.New("cannot parse KeyperSet")

// maximum age of a tx pointer in blocks before it is considered outdated
const maxTxPointerAge = 2

type Keyper struct {
	core            *keyper.KeyperCore
	config          *Config
	dbpool          *pgxpool.Pool
	chainSyncClient *chainsync.Client

	sequencerSyncer          *SequencerSyncer
	decryptionTriggerChannel chan<- *broker.Event[*epochkghandler.DecryptionTrigger]
}

func New(c *Config) *Keyper {
	return &Keyper{
		config: c,
	}
}

func (kpr *Keyper) Start(ctx context.Context, runner service.Runner) error {
	var err error

	decryptionTriggerChannel := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])
	kpr.decryptionTriggerChannel = decryptionTriggerChannel
	runner.Defer(func() { close(kpr.decryptionTriggerChannel) })

	kpr.dbpool, err = db.Connect(ctx, runner, kpr.config.DatabaseURL, database.Definition.Name())
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	kpr.core, err = keyper.New(
		&kprconfig.Config{
			InstanceID:        kpr.config.InstanceID,
			DatabaseURL:       kpr.config.DatabaseURL,
			HTTPEnabled:       kpr.config.HTTPEnabled,
			HTTPListenAddress: kpr.config.HTTPListenAddress,
			P2P:               kpr.config.P2P,
			Ethereum:          kpr.config.Gnosis,
			Shuttermint:       kpr.config.Shuttermint,
			Metrics:           kpr.config.Metrics,
		},
		decryptionTriggerChannel,
		keyper.WithDBPool(kpr.dbpool),
		keyper.NoBroadcastEonPublicKey(),
		keyper.WithEonPublicKeyHandler(kpr.newEonPublicKey),
	)
	if err != nil {
		return errors.Wrap(err, "can't instantiate keyper core")
	}

	kpr.chainSyncClient, err = chainsync.NewClient(
		ctx,
		chainsync.WithClientURL(kpr.config.Gnosis.EthereumURL),
		chainsync.WithKeyperSetManager(kpr.config.GnosisContracts.KeyperSetManager),
		chainsync.WithKeyBroadcastContract(kpr.config.GnosisContracts.KeyBroadcastContract),
		chainsync.WithSyncNewBlock(kpr.newBlock),
		chainsync.WithSyncNewKeyperSet(kpr.newKeyperSet),
		chainsync.WithPrivateKey(kpr.config.Gnosis.PrivateKey.Key),
		chainsync.WithLogger(gethLog.NewLogger(slog.Default().Handler())),
	)
	if err != nil {
		return err
	}

	err = kpr.initSequencerSyncer(ctx)
	if err != nil {
		return err
	}

	return runner.StartService(kpr.core, kpr.chainSyncClient)
}

// initSequencerSycer initializes the sequencer syncer if the keyper is known to be a member of a
// keyper set. Otherwise, the syncer will only be initialized once such a keyper set is observed to
// be added, as only then we will know which eon(s) we are responsible for.
func (kpr *Keyper) initSequencerSyncer(ctx context.Context) error {
	obskeyperdb := obskeyper.New(kpr.dbpool)
	keyperSets, err := obskeyperdb.GetKeyperSets(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to query keyper sets from db")
	}

	keyperSetFound := false
	minEon := uint64(0)
	for _, keyperSet := range keyperSets {
		for _, m := range keyperSet.Keypers {
			mAddress := common.HexToAddress(m)
			if mAddress.Cmp(kpr.config.GetAddress()) == 0 {
				keyperSetFound = true
				if minEon > uint64(keyperSet.KeyperConfigIndex) {
					minEon = uint64(keyperSet.KeyperConfigIndex)
				}
				break
			}
		}
	}

	if keyperSetFound {
		err := kpr.ensureSequencerSyncing(ctx, minEon)
		if err != nil {
			return err
		}
	}
	return nil
}

func (kpr *Keyper) ensureSequencerSyncing(ctx context.Context, eon uint64) error {
	if kpr.sequencerSyncer == nil {
		log.Info().
			Uint64("eon", eon).
			Str("contract-address", kpr.config.GnosisContracts.KeyperSetManager.Hex()).
			Msg("initializing sequencer syncer")
		client, err := ethclient.DialContext(ctx, kpr.config.Gnosis.ContractsURL)
		if err != nil {
			return err
		}
		contract, err := sequencerBindings.NewSequencer(kpr.config.GnosisContracts.Sequencer, client)
		if err != nil {
			return err
		}
		kpr.sequencerSyncer = &SequencerSyncer{
			Contract: contract,
			DBPool:   kpr.dbpool,
			StartEon: eon,
		}

		// TODO: perform an initial sync without blocking and/or set start block
	}

	if eon < kpr.sequencerSyncer.StartEon {
		log.Info().
			Uint64("old-start-eon", kpr.sequencerSyncer.StartEon).
			Uint64("new-start-eon", eon).
			Msg("decreasing sequencer syncing start eon")
		kpr.sequencerSyncer.StartEon = eon
	}
	return nil
}

func (kpr *Keyper) newBlock(ctx context.Context, ev *syncevent.LatestBlock) error {
	if kpr.sequencerSyncer != nil {
		if err := kpr.sequencerSyncer.Sync(ctx, ev.Number.Uint64()); err != nil {
			return err
		}
	}

	queries := obskeyper.New(kpr.dbpool)
	keyperSet, err := queries.GetKeyperSet(ctx, ev.Number.Int64())
	if err == pgx.ErrNoRows {
		log.Debug().Uint64("block", ev.Number.Uint64()).Msg("ignoring block as no keyper set has been found for it")
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to query keyper set for block %d", ev.Number)
	}
	for _, m := range keyperSet.Keypers {
		if m == shdb.EncodeAddress(kpr.config.GetAddress()) {
			return kpr.triggerDecryption(ctx, ev, &keyperSet)
		}
	}
	log.Debug().Uint64("block", ev.Number.Uint64()).Msg("ignoring block as not part of keyper set")
	return nil
}

func (kpr *Keyper) newKeyperSet(ctx context.Context, ev *syncevent.KeyperSet) error {
	isMember := false
	for _, m := range ev.Members {
		if m.Cmp(kpr.config.GetAddress()) == 0 {
			isMember = true
			break
		}
	}
	log.Info().
		Uint64("activation-block", ev.ActivationBlock).
		Uint64("eon", ev.Eon).
		Bool("is-member", isMember).
		Msg("new keyper set added")

	if isMember {
		if err := kpr.ensureSequencerSyncing(ctx, ev.Eon); err != nil {
			return err
		}
	}

	return kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		obskeyperdb := obskeyper.New(tx)

		keyperConfigIndex, err := medley.Uint64ToInt64Safe(ev.Eon)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		activationBlockNumber, err := medley.Uint64ToInt64Safe(ev.ActivationBlock)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		threshold, err := medley.Uint64ToInt64Safe(ev.Threshold)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}

		return obskeyperdb.InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
			KeyperConfigIndex:     keyperConfigIndex,
			ActivationBlockNumber: activationBlockNumber,
			Keypers:               shdb.EncodeAddresses(ev.Members),
			Threshold:             int32(threshold),
		})
	})
}

func (kpr *Keyper) newEonPublicKey(_ context.Context, pubKey keyper.EonPublicKey) error {
	log.Info().
		Uint64("eon", pubKey.Eon).
		Uint64("activation-block", pubKey.ActivationBlock).
		Msg("new eon pk")
	return nil
}

func (kpr *Keyper) triggerDecryption(ctx context.Context, ev *syncevent.LatestBlock, keyperSet *obskeyper.KeyperSet) error {
	queries := database.New(kpr.dbpool)
	eon := keyperSet.KeyperConfigIndex
	var txPointer int64
	var txPointerAge int64
	txPointerDB, err := queries.GetTxPointer(ctx, eon)
	if err == pgx.ErrNoRows {
		txPointer = 0
		txPointerAge = ev.Number.Int64() - keyperSet.ActivationBlockNumber + 1
	} else if err != nil {
		return errors.Wrap(err, "failed to query tx pointer from db")
	} else {
		txPointerAge = ev.Number.Int64() - txPointerDB.Block
		txPointer = txPointerDB.Value
	}
	if txPointerAge == 0 {
		// A pointer of age 0 means we already received the pointer from a DecryptionKeys message
		// even though we haven't sent our shares yet. In that case, sending our shares is
		// unnecessary.
		log.Warn().
			Int64("block-number", ev.Number.Int64()).
			Int64("eon", eon).
			Int64("tx-pointer", txPointer).
			Int64("tx-pointer-age", txPointerAge).
			Msg("ignoring new block as tx pointer age is 0")
		return nil
	}
	if txPointerAge > maxTxPointerAge {
		// If the tx pointer is outdated, the system has failed to generate decryption keys (or at
		// least we haven't received them). This either means not enough keypers are online or they
		// don't agree on the current value of the tx pointer. In order to recover, we choose the
		// current length of the transaction queue as the new tx pointer, as this is a value
		// everyone can agree on.
		log.Warn().
			Int64("block-number", ev.Number.Int64()).
			Int64("eon", eon).
			Int64("tx-pointer", txPointer).
			Int64("tx-pointer-age", txPointerAge).
			Msg("outdated tx pointer")
		txPointer, err = queries.GetTransactionSubmittedEventCount(ctx, keyperSet.KeyperConfigIndex)
		if err == pgx.ErrNoRows {
			txPointer = 0
		} else if err != nil {
			return errors.Wrap(err, "failed to query transaction submitted event count from db")
		}
	}

	identityPreimages, err := kpr.getDecryptionIdentityPreimages(ctx, ev, keyperSet.KeyperConfigIndex, txPointer)
	if err != nil {
		return err
	}
	trigger := epochkghandler.DecryptionTrigger{
		BlockNumber:       uint64(ev.Number.Int64() + 1),
		IdentityPreimages: identityPreimages,
	}
	event := broker.NewEvent(&trigger)
	log.Debug().
		Int64("block-number", int64(trigger.BlockNumber)).
		Int("num-identities", len(trigger.IdentityPreimages)).
		Int64("tx-pointer", txPointer).
		Int64("tx-pointer-age", txPointerAge).
		Msg("sending decryption trigger")
	kpr.decryptionTriggerChannel <- event

	return nil
}

func (kpr *Keyper) getDecryptionIdentityPreimages(
	ctx context.Context, ev *syncevent.LatestBlock, eon int64, txPointer int64,
) ([]identitypreimage.IdentityPreimage, error) {
	identityPreimages := []identitypreimage.IdentityPreimage{}

	queries := database.New(kpr.dbpool)
	limitUint64 := kpr.config.EncryptedGasLimit/kpr.config.MinGasPerTransaction + 1
	if limitUint64 > math.MaxInt32 {
		return identityPreimages, errors.New("gas limit too big")
	}
	limit := int32(limitUint64)

	events, err := queries.GetTransactionSubmittedEvents(ctx, database.GetTransactionSubmittedEventsParams{
		Eon:   eon,
		Index: txPointer,
		Limit: limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query transaction submitted events from index %d", txPointer)
	}

	identityPreimages = []identitypreimage.IdentityPreimage{
		makeBlockIdentityPreimage(ev),
	}
	gas := uint64(0)
	for _, event := range events {
		gas += uint64(event.GasLimit)
		if gas > kpr.config.EncryptedGasLimit {
			break
		}
		identityPreimage, err := transactionSubmittedEventToIdentityPreimage(event)
		if err != nil {
			return []identitypreimage.IdentityPreimage{}, err
		}
		identityPreimages = append(identityPreimages, identityPreimage)
	}
	return identityPreimages, nil
}

func transactionSubmittedEventToIdentityPreimage(event database.TransactionSubmittedEvent) (identitypreimage.IdentityPreimage, error) {
	sender, err := shdb.DecodeAddress(event.Sender)
	if err != nil {
		return identitypreimage.IdentityPreimage{}, errors.Wrap(err, "failed to decode sender address of transaction submitted event from db")
	}

	var buf bytes.Buffer
	buf.Write(event.IdentityPrefix)
	buf.Write(sender.Bytes())

	return identitypreimage.IdentityPreimage(buf.Bytes()), nil
}

func makeBlockIdentityPreimage(ev *syncevent.LatestBlock) identitypreimage.IdentityPreimage {
	return identitypreimage.IdentityPreimage(ev.Number.Bytes())
}

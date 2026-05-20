package gnosis

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func (kpr *Keyper) processNewKeyperSet(ctx context.Context, ev *syncevent.KeyperSet) error {
	ownAddress := kpr.config.GetAddress()
	isMember := false
	ownIndex := -1
	for i, m := range ev.Members {
		if m.Cmp(ownAddress) == 0 {
			isMember = true
			ownIndex = i
			break
		}
	}
	log.Info().
		Uint64("activation-block", ev.ActivationBlock).
		Uint64("eon", ev.Eon).
		Int("num-members", len(ev.Members)).
		Uint64("threshold", ev.Threshold).
		Bool("is-member", isMember).
		Msg("new keyper set added")

	err := kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
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
	if err != nil {
		return err
	}

	if isMember {
		if err := kpr.maybeRegisterECIESKey(ctx, ev.Eon, ownIndex, ownAddress); err != nil {
			return errors.Wrap(err, "failed to register ECIES key")
		}
	}
	return nil
}

// maybeRegisterECIESKey submits a `registerKey` transaction if the local
// `ecies_keys` cache has no entry for the keyper's own address. The cache
// is the source of truth for prior registrations because it is populated
// both by the initial poll and live `KeyRegistered` events.
func (kpr *Keyper) maybeRegisterECIESKey(
	ctx context.Context,
	keyperSetIndex uint64,
	keyperIndex int,
	ownAddress common.Address,
) error {
	exists, err := corekeyperdb.New(kpr.dbpool).ExistsECIESKey(ctx, shdb.EncodeAddress(ownAddress))
	if err != nil {
		return errors.Wrap(err, "query ecies_keys for own address")
	}
	if exists {
		log.Debug().
			Str("address", ownAddress.Hex()).
			Msg("ECIES key already registered, skipping")
		return nil
	}

	pubKey := ethcrypto.FromECDSAPub(&kpr.config.Shuttermint.EncryptionKey.Key.PublicKey)
	chainID, err := kpr.chainSyncClient.ChainID(ctx)
	if err != nil {
		return errors.Wrap(err, "get chain id")
	}
	opts, err := bind.NewKeyedTransactorWithChainID(kpr.config.Gnosis.Node.PrivateKey.Key, chainID)
	if err != nil {
		return errors.Wrap(err, "construct signer transaction opts")
	}
	opts.Context = ctx

	tx, err := kpr.chainSyncClient.ECIESKeyRegistry.RegisterKey(
		opts,
		keyperSetIndex,
		uint64(keyperIndex),
		pubKey,
	)
	if err != nil {
		return errors.Wrap(err, "submit registerKey transaction")
	}
	log.Info().
		Uint64("keyper-set-index", keyperSetIndex).
		Int("keyper-index", keyperIndex).
		Str("tx-hash", tx.Hash().Hex()).
		Msg("submitted ECIES key registration")
	return nil
}

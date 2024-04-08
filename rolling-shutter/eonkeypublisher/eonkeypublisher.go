package eonkeypublisher

import (
	"context"
	"crypto/ecdsa"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/shop-contracts/bindings"

	obskeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

const (
	eonKeyChannelSize = 32
	retryInterval     = time.Second * 12
)

// EonKeyPublisher is a service that publishes eon keys via a eon key publisher contract.
type EonKeyPublisher struct {
	dbpool     *pgxpool.Pool
	client     *ethclient.Client
	contract   *bindings.EonKeyPublish
	privateKey *ecdsa.PrivateKey

	keys chan keyper.EonPublicKey
}

func NewEonKeyPublisher(
	dbpool *pgxpool.Pool,
	client *ethclient.Client,
	eonKeyPublishAddress common.Address,
	privateKey *ecdsa.PrivateKey,
) (*EonKeyPublisher, error) {
	contract, err := bindings.NewEonKeyPublish(eonKeyPublishAddress, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate eon key publisher contract")
	}
	return &EonKeyPublisher{
		dbpool:     dbpool,
		client:     client,
		contract:   contract,
		privateKey: privateKey,

		keys: make(chan keyper.EonPublicKey, eonKeyChannelSize),
	}, nil
}

func (p *EonKeyPublisher) Start(ctx context.Context, runner service.Runner) error { //nolint: unparam
	log.Info().Msg("starting eon key publisher")
	runner.Go(func() error {
		p.publishOldKeys(ctx)
		for {
			select {
			case key := <-p.keys:
				p.publishIfResponsible(ctx, key)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
	return nil
}

// Publish schedules a eon key to be published.
func (p *EonKeyPublisher) Publish(key keyper.EonPublicKey) {
	p.keys <- key
}

// publishIfResponsible publishes a eon key if the keyper is part of the corresponding keyper
// set, unless the key is already confirmed or the keyper has already voted on it.
func (p *EonKeyPublisher) publishIfResponsible(ctx context.Context, key keyper.EonPublicKey) {
	db := obskeyperdb.New(p.dbpool)
	keyperSet, err := db.GetKeyperSetByKeyperConfigIndex(ctx, int64(key.Eon))
	if err != nil {
		log.Error().
			Err(err).
			Uint64("keyper-set-index", key.KeyperConfigIndex).
			Hex("key", key.PublicKey).
			Msg("failed to check if eon key should be published")
		return
	}
	keyperAddress := ethcrypto.PubkeyToAddress(p.privateKey.PublicKey)
	keyperIndex, err := keyperSet.GetIndex(keyperAddress)
	if err != nil {
		log.Info().
			Uint64("keyper-set-index", key.KeyperConfigIndex).
			Str("keyper-address", keyperAddress.Hex()).
			Hex("key", key.PublicKey).
			Msg("not publishing eon key as keyper is not part of corresponding keyper set")
		return
	}
	p.publish(ctx, key.PublicKey, key.KeyperConfigIndex, keyperIndex)
}

// publishOldKeys publishes all eon keys that are already in the database, unless they're already
// confirmed or the keyper has already voted on them.
func (p *EonKeyPublisher) publishOldKeys(ctx context.Context) {
	db := corekeyperdb.New(p.dbpool)
	dkgResultsDB, err := db.GetAllDKGResults(ctx)
	if err != nil {
		err := errors.Wrap(err, "failed to query DKG results from db")
		log.Error().Err(err).Msg("failed to publish old eon keys")
		return
	}
	for _, dkgResultDB := range dkgResultsDB {
		if !dkgResultDB.Success {
			continue
		}
		dkgResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
		if err != nil {
			err := errors.Wrapf(err, "failed to decode DKG result of eon %d", dkgResultDB.Eon)
			log.Error().Err(err).Msg("failed to publish old eon keys")
			continue
		}
		p.publish(ctx, dkgResult.PublicKey.Marshal(), dkgResult.Eon, dkgResult.Keyper)
	}
}

// publish publishes an eon key, unless it's already confirmed or the keyper has already voted on
// it. On errors, publishing will be retried a few times and eventually aborted.
func (p *EonKeyPublisher) publish(ctx context.Context, key []byte, keyperSetIndex uint64, keyperIndex uint64) {
	_, err := retry.FunctionCall[struct{}](ctx, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, p.tryPublish(ctx, key, keyperSetIndex, keyperIndex)
	}, retry.Interval(retryInterval))
	if err != nil {
		log.Error().
			Err(err).
			Uint64("keyper-set-index", keyperSetIndex).
			Hex("key", key).
			Msg("failed to publish eon key")
	}
}

func (p *EonKeyPublisher) tryPublish(ctx context.Context, key []byte, keyperSetIndex uint64, keyperIndex uint64) error {
	keyperAddress := ethcrypto.PubkeyToAddress(p.privateKey.PublicKey)
	hasAlreadyVoted, err := p.contract.HasKeyperVoted(&bind.CallOpts{}, keyperAddress)
	if err != nil {
		return errors.Wrap(err, "failed to query eon key publisher contract if keyper has already voted")
	}
	if hasAlreadyVoted {
		log.Info().
			Uint64("keyper-set-index", keyperSetIndex).
			Str("keyper-address", keyperAddress.Hex()).
			Hex("key", key).
			Msg("not publishing eon key as keyper has already voted")
		return nil
	}
	isAlreadyConfirmed, err := p.contract.EonKeyConfirmed(&bind.CallOpts{}, key)
	if err != nil {
		return errors.Wrap(err, "failed to query eon key publisher contract if eon key is confirmed")
	}
	if isAlreadyConfirmed {
		log.Info().
			Uint64("keyper-set-index", keyperSetIndex).
			Hex("key", key).
			Msg("not publishing eon key as it is already confirmed")
		return nil
	}

	chainID, err := p.client.ChainID(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get chain ID")
	}
	opts, err := bind.NewKeyedTransactorWithChainID(p.privateKey, chainID)
	if err != nil {
		return errors.Wrap(err, "failed to construct tx opts")
	}
	tx, err := p.contract.PublishEonKey(opts, key, keyperIndex)
	if err != nil {
		return errors.Wrap(err, "failed to send publish eon key tx")
	}
	log.Info().
		Uint64("keyper-set-index", keyperSetIndex).
		Hex("key", key).
		Hex("tx-hash", tx.Hash().Bytes()).
		Msg("eon key publish tx sent")
	receipt, err := bind.WaitMined(ctx, p.client, tx)
	if err != nil {
		log.Error().Err(err).Msg("error waiting for eon key publish tx to be mined")
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Error().
			Hex("tx-hash", tx.Hash().Bytes()).
			Interface("receipt", receipt).
			Msg("eon key publish tx failed")
		return errors.New("eon key publish tx failed")
	}
	log.Info().
		Uint64("keyper-set-index", keyperSetIndex).
		Hex("key", key).
		Hex("tx-hash", tx.Hash().Bytes()).
		Msg("successfully published eon key")
	return nil
}

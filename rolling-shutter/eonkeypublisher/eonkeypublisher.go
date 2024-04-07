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
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
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
	runner.Go(func() error {
		for {
			select {
			case key := <-p.keys:
				p.publish(ctx, key)
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

func (p *EonKeyPublisher) publish(ctx context.Context, key keyper.EonPublicKey) {
	_, err := retry.FunctionCall[struct{}](ctx, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, p.tryPublish(ctx, key)
	}, retry.Interval(retryInterval))
	if err != nil {
		log.Error().
			Err(err).
			Uint64("keyper-set-index", key.KeyperConfigIndex).
			Hex("key", key.PublicKey).
			Msg("failed to publish eon key")
	}
}

func (p *EonKeyPublisher) tryPublish(ctx context.Context, key keyper.EonPublicKey) error {
	db := obskeyperdb.New(p.dbpool)
	keyperSet, err := db.GetKeyperSetByKeyperConfigIndex(ctx, int64(key.Eon))
	if err != nil {
		return errors.Wrapf(err, "failed to query keyper set %d by index from db", key.KeyperConfigIndex)
	}
	keyperAddress := ethcrypto.PubkeyToAddress(p.privateKey.PublicKey)
	keyperIndex, err := keyperSet.GetIndex(keyperAddress)
	if err != nil {
		log.Info().
			Uint64("keyper-set-index", key.KeyperConfigIndex).
			Str("keyper-address", keyperAddress.Hex()).
			Msg("not publishing eon key as keyper is not part of corresponding keyper set")
		return nil
	}

	hasAlreadyVoted, err := p.contract.HasKeyperVoted(&bind.CallOpts{}, keyperAddress)
	if err != nil {
		return errors.Wrap(err, "failed to query eon key publisher contract if keyper has already voted")
	}
	if hasAlreadyVoted {
		log.Info().
			Uint64("keyper-set-index", key.KeyperConfigIndex).
			Str("keyper-address", keyperAddress.Hex()).
			Msg("not publishing eon key as keyper has already voted")
		return nil
	}
	isAlreadyConfirmed, err := p.contract.EonKeyConfirmed(&bind.CallOpts{}, key.PublicKey)
	if err != nil {
		return errors.Wrap(err, "failed to query eon key publisher contract if eon key is confirmed")
	}
	if isAlreadyConfirmed {
		log.Info().
			Uint64("keyper-set-index", key.KeyperConfigIndex).
			Hex("key", key.PublicKey).
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
	tx, err := p.contract.PublishEonKey(opts, key.PublicKey, keyperIndex)
	if err != nil {
		return errors.Wrap(err, "failed to send publish eon key tx")
	}
	log.Info().
		Uint64("keyper-set-index", key.KeyperConfigIndex).
		Hex("key", key.PublicKey).
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
	log.Info().Msg("successfully published eon key")
	return nil
}

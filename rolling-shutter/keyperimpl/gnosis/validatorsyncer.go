package gnosis

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	validatorRegistryBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/validatorregistry"
	blst "github.com/supranational/blst/bindings/go"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/beaconapiclient"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/validatorregistry"
)

const (
	maxRequestBlockRange = 10_000
)

type ValidatorSyncer struct {
	Contract             *validatorRegistryBindings.Validatorregistry
	DBPool               *pgxpool.Pool
	BeaconAPIClient      *beaconapiclient.Client
	ExecutionClient      *ethclient.Client
	ChainID              uint64
	SyncStartBlockNumber uint64
}

func (v *ValidatorSyncer) Sync(ctx context.Context, header *types.Header) error {
	db := database.New(v.DBPool)
	syncedUntil, err := db.GetValidatorRegistrationsSyncedUntil(ctx)
	if err != nil && err != pgx.ErrNoRows {
		return errors.Wrap(err, "failed to query validator registration sync status")
	}
	var start uint64
	if err == pgx.ErrNoRows {
		start = v.SyncStartBlockNumber
	} else {
		start = uint64(syncedUntil.BlockNumber + 1)
	}
	endBlock := header.Number.Uint64()
	log.Debug().
		Uint64("start-block", start).
		Uint64("end-block", endBlock).
		Msg("syncing validator registry")

	syncRanges := medley.GetSyncRanges(start, endBlock, maxRequestBlockRange)
	for _, r := range syncRanges {
		err = v.syncRange(ctx, r[0], r[1])
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *ValidatorSyncer) syncRange(ctx context.Context, start, end uint64) error {
	db := database.New(v.DBPool)
	events, err := v.fetchEvents(ctx, start, end)
	if err != nil {
		return err
	}
	filteredEvents, err := v.filterEvents(ctx, events)
	if err != nil {
		return err
	}
	header, err := v.ExecutionClient.HeaderByNumber(ctx, new(big.Int).SetUint64(end))
	if err != nil {
		return errors.Wrap(err, "failed to get execution block header by number")
	}
	err = v.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		err = v.insertEvents(ctx, tx, filteredEvents)
		if err != nil {
			return err
		}
		err = db.SetValidatorRegistrationsSyncedUntil(ctx, database.SetValidatorRegistrationsSyncedUntilParams{
			BlockNumber: int64(end),
			BlockHash:   header.Hash().Bytes(),
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	numRegistrations, err := db.GetNumValidatorRegistrations(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get number of validator registrations")
	}
	log.Info().
		Uint64("start-block", start).
		Uint64("end-block", end).
		Int("num-inserted-events", len(filteredEvents)).
		Int("num-discarded-events", len(events)-len(filteredEvents)).
		Int64("num-registrations", numRegistrations).
		Msg("synced validator registry")
	metricsNumValidatorRegistrations.Set(float64(numRegistrations))
	metricsValidatorRegistrationsSyncedUntil.Set(float64(end))
	return nil
}

func (v *ValidatorSyncer) fetchEvents(
	ctx context.Context,
	start,
	end uint64,
) ([]*validatorRegistryBindings.ValidatorregistryUpdated, error) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     &end,
		Context: ctx,
	}
	it, err := v.Contract.ValidatorregistryFilterer.FilterUpdated(&opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query validator registry update events")
	}
	events := []*validatorRegistryBindings.ValidatorregistryUpdated{}
	for it.Next() {
		events = append(events, it.Event)
	}
	if it.Error() != nil {
		return nil, errors.Wrap(it.Error(), "failed to iterate validator registry update events")
	}
	return events, nil
}

func (v *ValidatorSyncer) filterEvents(
	ctx context.Context,
	events []*validatorRegistryBindings.ValidatorregistryUpdated,
) ([]*validatorRegistryBindings.ValidatorregistryUpdated, error) {
	db := database.New(v.DBPool)
	filteredEvents := []*validatorRegistryBindings.ValidatorregistryUpdated{}
	for _, event := range events {
		evLog := log.With().
			Hex("block-hash", event.Raw.BlockHash.Bytes()).
			Uint64("block-number", event.Raw.BlockNumber).
			Uint("tx-index", event.Raw.TxIndex).
			Uint("log-index", event.Raw.Index).
			Logger()

		msg := new(validatorregistry.AggregateRegistrationMessage)
		err := msg.Unmarshal(event.Message)
		if err != nil {
			evLog.Warn().
				Err(err).
				Msg("failed to unmarshal registration message")
			continue
		}

		if !checkStaticRegistrationMessageFields(msg, v.ChainID, event.Raw.Address, evLog) {
			continue
		}

		pubKeys := make([]*blst.P1Affine, 0)
		for _, validatorIndex := range msg.ValidatorIndices() {
			evLog = evLog.With().Int64("validator-index", validatorIndex).Logger()
			latestNonce, err := db.GetValidatorRegistrationNonceBefore(ctx, database.GetValidatorRegistrationNonceBeforeParams{
				ValidatorIndex: validatorIndex,
				BlockNumber:    int64(event.Raw.BlockNumber),
				TxIndex:        int64(event.Raw.TxIndex),
				LogIndex:       int64(event.Raw.Index),
			})
			if err != nil && err != pgx.ErrNoRows {
				return nil, errors.Wrapf(err, "failed to query latest nonce for validator %d", msg.ValidatorIndex)
			}
			if err == pgx.ErrNoRows {
				latestNonce = -1
			}

			if msg.Nonce > math.MaxInt32 || int64(msg.Nonce) <= latestNonce {
				evLog.Warn().
					Uint32("nonce", msg.Nonce).
					Int64("latest-nonce", latestNonce).
					Msg("ignoring registration message with invalid nonce")
				continue
			}

			validator, err := v.BeaconAPIClient.GetValidatorByIndex(ctx, "head", uint64(validatorIndex))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get validator %d", msg.ValidatorIndex)
			}
			if validator == nil {
				evLog.Warn().Msg("ignoring registration message for unknown validator")
				continue
			}
			pubkey, err := validator.Data.Validator.GetPubkey()
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get pubkey of validator %d", msg.ValidatorIndex)
			}
			pubKeys = append(pubKeys, pubkey)
		}

		sig := new(blst.P2Affine).Uncompress(event.Signature)
		if sig == nil {
			evLog.Warn().Msg("ignoring registration message with undecodable signature")
			continue
		}

		if msg.Version == validatorregistry.LegacyValidatorRegistrationMessageVersion {
			msg := new(validatorregistry.LegacyRegistrationMessage)
			err := msg.Unmarshal(event.Message)
			if err != nil {
				evLog.Warn().
					Err(err).
					Msg("failed to unmarshal registration message")
				continue
			}
			if validSignature := validatorregistry.VerifySignature(sig, pubKeys[0], msg); !validSignature {
				evLog.Warn().Msg("ignoring registration message with invalid signature")
				continue
			}
		} else {
			validSignature := validatorregistry.VerifyAggregateSignature(sig, pubKeys, msg)
			if !validSignature {
				evLog.Warn().Msg("ignoring registration message with invalid signature")
				continue
			}
		}

		filteredEvents = append(filteredEvents, event)
	}
	return filteredEvents, nil
}

func (v *ValidatorSyncer) insertEvents(ctx context.Context, tx pgx.Tx, events []*validatorRegistryBindings.ValidatorregistryUpdated) error {
	db := database.New(tx)
	for _, event := range events {
		msg := new(validatorregistry.AggregateRegistrationMessage)
		err := msg.Unmarshal(event.Message)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal registration message")
		}
		for _, validatorIndex := range msg.ValidatorIndices() {
			err = db.InsertValidatorRegistration(ctx, database.InsertValidatorRegistrationParams{
				BlockNumber:    int64(event.Raw.BlockNumber),
				BlockHash:      event.Raw.BlockHash.Bytes(),
				TxIndex:        int64(event.Raw.TxIndex),
				LogIndex:       int64(event.Raw.Index),
				ValidatorIndex: validatorIndex,
				Nonce:          int64(msg.Nonce),
				IsRegistration: msg.IsRegistration,
			})
			if err != nil {
				return errors.Wrap(err, "failed to insert validator registration into db")
			}
		}
	}
	return nil
}

func checkStaticRegistrationMessageFields(
	msg *validatorregistry.AggregateRegistrationMessage,
	chainID uint64,
	validatorRegistryAddress common.Address,
	logger zerolog.Logger,
) bool {
	if msg.Version != validatorregistry.AggregateValidatorRegistrationMessageVersion &&
		msg.Version != validatorregistry.LegacyValidatorRegistrationMessageVersion {
		logger.Warn().
			Uint8("version", msg.Version).
			Str("expected-version", fmt.Sprintf("%d or %d", validatorregistry.LegacyValidatorRegistrationMessageVersion,
				validatorregistry.AggregateValidatorRegistrationMessageVersion)).
			Uint64("validator-index", msg.ValidatorIndex).
			Msg("ignoring registration message with invalid version")
		return false
	}
	if msg.ChainID != chainID {
		logger.Warn().
			Uint64("chain-id", msg.ChainID).
			Uint64("expected-chain-id", chainID).
			Uint64("validator-index", msg.ValidatorIndex).
			Msg("ignoring registration message with invalid chain ID")
		return false
	}
	if msg.ValidatorRegistryAddress != validatorRegistryAddress {
		logger.Warn().
			Hex("validator-registry-address", msg.ValidatorRegistryAddress.Bytes()).
			Hex("expected-validator-registry-address", validatorRegistryAddress.Bytes()).
			Uint64("validator-index", msg.ValidatorIndex).
			Msg("ignoring registration message with invalid validator registry address")
		return false
	}
	if msg.ValidatorIndex > math.MaxInt64 {
		logger.Warn().
			Uint64("validator-index", msg.ValidatorIndex).
			Msg("ignoring registration message with invalid validator index")
		return false
	}
	return true
}

package gnosis

import (
	"context"
	"math"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	registryBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/validatorregistry"
	blst "github.com/supranational/blst/bindings/go"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/beaconapiclient"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/validatorregistry"
)

const (
	ValidatorRegistrationMessageVersion = 0
)

type ValidatorSyncer struct {
	Contract        *registryBindings.Validatorregistry
	DBPool          *pgxpool.Pool
	BeaconAPIClient *beaconapiclient.Client
	ChainID         uint64
}

func (v *ValidatorSyncer) Sync(ctx context.Context, header *types.Header) error {
	db := database.New(v.DBPool)
	syncedUntil, err := db.GetValidatorRegistrationsSyncedUntil(ctx)
	if err != nil && err != pgx.ErrNoRows {
		return errors.Wrap(err, "failed to query validator registration sync status")
	}
	var start uint64
	if err == pgx.ErrNoRows {
		start = 0
	} else {
		start = uint64(syncedUntil.BlockNumber + 1)
	}

	log.Debug().
		Uint64("start-block", start).
		Uint64("end-block", header.Number.Uint64()).
		Msg("syncing validator registry")

	endBlock := header.Number.Uint64()
	opts := bind.FilterOpts{
		Start:   start,
		End:     &endBlock,
		Context: ctx,
	}
	it, err := v.Contract.ValidatorregistryFilterer.FilterUpdated(&opts)
	if err != nil {
		return errors.Wrap(err, "failed to query validator registry update events")
	}
	events := []*registryBindings.ValidatorregistryUpdated{}
	for it.Next() {
		events = append(events, it.Event)
	}
	if it.Error() != nil {
		return errors.Wrap(it.Error(), "failed to iterate validator registry update events")
	}
	if len(events) == 0 {
		log.Debug().
			Uint64("start-block", start).
			Uint64("end-block", endBlock).
			Msg("no validator registry update events found")
	}

	filteredEvents, err := v.filterEvents(ctx, events)
	if err != nil {
		return err
	}
	return v.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		err = v.insertEvents(ctx, tx, filteredEvents)
		if err != nil {
			return err
		}
		err = db.SetValidatorRegistrationsSyncedUntil(ctx, database.SetValidatorRegistrationsSyncedUntilParams{
			BlockNumber: int64(endBlock),
			BlockHash:   header.Hash().Bytes(),
		})
		if err != nil {
			return err
		}
		return nil
	})
}

func (v *ValidatorSyncer) filterEvents(
	ctx context.Context,
	events []*registryBindings.ValidatorregistryUpdated,
) ([]*registryBindings.ValidatorregistryUpdated, error) {
	db := database.New(v.DBPool)
	filteredEvents := []*registryBindings.ValidatorregistryUpdated{}
	for _, event := range events {
		evLog := log.With().
			Hex("block-hash", event.Raw.BlockHash.Bytes()).
			Uint64("block-number", event.Raw.BlockNumber).
			Uint("tx-index", event.Raw.TxIndex).
			Uint("log-index", event.Raw.Index).
			Logger()

		msg := new(validatorregistry.RegistrationMessage)
		err := msg.Unmarshal(event.Message)
		if err != nil {
			evLog.Warn().
				Err(err).
				Msg("failed to unmarshal registration message")
			continue
		}
		evLog = evLog.With().Uint64("validator-index", msg.ValidatorIndex).Logger()

		if !checkStaticRegistrationMessageFields(msg, v.ChainID, event.Raw.Address, evLog) {
			continue
		}

		latestNonce, err := db.GetValidatorRegistrationNonceBefore(ctx, database.GetValidatorRegistrationNonceBeforeParams{
			ValidatorIndex: int64(msg.ValidatorIndex),
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
		if msg.Nonce <= uint64(latestNonce) || msg.Nonce > math.MaxInt64 {
			evLog.Warn().
				Uint64("nonce", msg.Nonce).
				Int64("latest-nonce", latestNonce).
				Msg("ignoring registration message with invalid nonce")
			continue
		}

		validator, err := v.BeaconAPIClient.GetValidatorByIndex(ctx, "head", msg.ValidatorIndex)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get validator %d", msg.ValidatorIndex)
		}
		if validator == nil {
			evLog.Warn().Msg("ignoring registration message for unknown validator")
			continue
		}
		pubkey := &validator.Data.Validator.Pubkey
		sig := new(blst.P2Affine).Deserialize(event.Signature)
		if sig == nil {
			evLog.Warn().Msg("ignoring registration message with invalid signature")
			continue
		}
		validSignature := validatorregistry.VerifySignature(sig, pubkey, msg)
		if !validSignature {
			evLog.Warn().Msg("ignoring registration message with invalid signature")
			continue
		}

		filteredEvents = append(filteredEvents, event)
	}
	return filteredEvents, nil
}

func (v *ValidatorSyncer) insertEvents(ctx context.Context, tx pgx.Tx, events []*registryBindings.ValidatorregistryUpdated) error {
	db := database.New(tx)
	for _, event := range events {
		msg := new(validatorregistry.RegistrationMessage)
		err := msg.Unmarshal(event.Message)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal registration message")
		}
		err = db.InsertValidatorRegistration(ctx, database.InsertValidatorRegistrationParams{
			BlockNumber:    int64(event.Raw.BlockNumber),
			BlockHash:      event.Raw.BlockHash.Bytes(),
			TxIndex:        int64(event.Raw.TxIndex),
			LogIndex:       int64(event.Raw.Index),
			ValidatorIndex: int64(msg.ValidatorIndex),
			Nonce:          int64(msg.Nonce),
			IsRegistration: msg.IsRegistration,
		})
		if err != nil {
			return errors.Wrap(err, "failed to insert validator registration into db")
		}
	}
	return nil
}

func checkStaticRegistrationMessageFields(msg *validatorregistry.RegistrationMessage, chainID uint64, validatorRegistryAddress common.Address, log zerolog.Logger) bool {
	if msg.Version != ValidatorRegistrationMessageVersion {
		log.Warn().
			Uint8("version", msg.Version).
			Uint8("expected-version", ValidatorRegistrationMessageVersion).
			Uint64("validator-index", msg.ValidatorIndex).
			Msg("ignoring registration message with invalid version")
		return false
	}
	if msg.ChainID != chainID {
		log.Warn().
			Uint64("chain-id", msg.ChainID).
			Uint64("expected-chain-id", chainID).
			Uint64("validator-index", msg.ValidatorIndex).
			Msg("ignoring registration message with invalid chain ID")
		return false
	}
	if msg.ValidatorRegistryAddress != validatorRegistryAddress {
		log.Warn().
			Hex("validator-registry-address", msg.ValidatorRegistryAddress.Bytes()).
			Hex("expected-validator-registry-address", validatorRegistryAddress.Bytes()).
			Uint64("validator-index", msg.ValidatorIndex).
			Msg("ignoring registration message with invalid validator registry address")
		return false
	}
	if msg.ValidatorIndex > math.MaxInt64 {
		log.Warn().
			Uint64("validator-index", msg.ValidatorIndex).
			Msg("ignoring registration message with invalid validator index")
		return false
	}
	return true
}

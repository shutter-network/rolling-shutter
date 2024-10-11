package synchandler

import (
	"context"
	"math"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	bindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/validatorregistry"
	blst "github.com/supranational/blst/bindings/go"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/metrics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/beaconapiclient"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/validatorregistry"
)

const ValidatorRegistrationMessageVersion = 0

func init() {
	var err error
	ValidatorRegistryContractABI, err = bindings.ValidatorregistryMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

var ValidatorRegistryContractABI *abi.ABI

func NewValidatorUpdated(
	dbPool *pgxpool.Pool,
	ethClient *ethclient.Client,
	beaconClient *beaconapiclient.Client,
	contractAddress common.Address,
	chainID uint64,
) (syncer.ContractEventHandler, error) {
	contract, err := bindings.NewValidatorregistry(contractAddress, ethClient)
	if err != nil {
		return nil, err
	}
	return syncer.WrapHandler(
		&ValidatorUpdated{
			evABI:               ValidatorRegistryContractABI,
			address:             contractAddress,
			valRegistryContract: contract,
			dbPool:              dbPool,
			beaconClient:        beaconClient,
			chainID:             chainID,
		})
}

type ValidatorUpdated struct {
	evABI   *abi.ABI
	address common.Address

	valRegistryContract *bindings.Validatorregistry
	dbPool              *pgxpool.Pool
	beaconClient        *beaconapiclient.Client

	chainID uint64
}

func (vu *ValidatorUpdated) Address() common.Address {
	return vu.address
}

func (*ValidatorUpdated) Event() string {
	return "Updated"
}

func (vu *ValidatorUpdated) ABI() abi.ABI {
	return *vu.evABI
}

func (vu *ValidatorUpdated) Accept(
	_ context.Context,
	_ types.Header,
	_ bindings.ValidatorregistryUpdated,
) (bool, error) {
	return true, nil
}

func (vu *ValidatorUpdated) Handle(
	ctx context.Context,
	update syncer.ChainUpdateContext,
	events []bindings.ValidatorregistryUpdated,
) error {
	db := database.New(vu.dbPool)
	filteredEvents, err := vu.filterEvents(ctx, events)
	if err != nil {
		return err
	}
	err = vu.dbPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		dtbs := database.New(tx)
		for _, event := range filteredEvents {
			msg := new(validatorregistry.RegistrationMessage)
			err := msg.Unmarshal(event.Message)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal registration message")
			}
			err = dtbs.InsertValidatorRegistration(ctx, database.InsertValidatorRegistrationParams{
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
		err = db.SetValidatorRegistrationsSyncedUntil(ctx, database.SetValidatorRegistrationsSyncedUntilParams{
			//TODO: check int64 overflow
			BlockNumber: update.Append.Latest().Number.Int64(),
			BlockHash:   update.Append.Latest().Hash().Bytes(),
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
		Int64("start-block", update.Append.Earliest().Number.Int64()).
		Int64("end-block", update.Append.Latest().Number.Int64()).
		Int("num-inserted-events", len(filteredEvents)).
		Int("num-discarded-events", len(events)-len(filteredEvents)).
		Int64("num-registrations", numRegistrations).
		Msg("synced validator registry")

	metrics.NumValidatorRegistrations.Set(float64(numRegistrations))
	metrics.ValidatorRegistrationsSyncedUntil.Set(float64(update.Append.Latest().Number.Uint64()))
	return nil
}

func (vu *ValidatorUpdated) filterEvents(
	ctx context.Context,
	events []bindings.ValidatorregistryUpdated,
) ([]bindings.ValidatorregistryUpdated, error) {
	db := database.New(vu.dbPool)
	filteredEvents := []bindings.ValidatorregistryUpdated{}
	for _, event := range events {
		logger := log.With().
			Hex("block-hash", event.Raw.BlockHash.Bytes()).
			Uint64("block-number", event.Raw.BlockNumber).
			Uint("tx-index", event.Raw.TxIndex).
			Uint("log-index", event.Raw.Index).
			Logger()

		msg := new(validatorregistry.RegistrationMessage)
		err := msg.Unmarshal(event.Message)
		if err != nil {
			logger.Warn().
				Err(err).
				Msg("failed to unmarshal registration message")
			continue
		}

		if !vu.checkStaticRegistrationMessageFields(msg, event.Raw.Address, logger) {
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
		if msg.Nonce > math.MaxInt64 || int64(msg.Nonce) <= latestNonce {
			logger.Warn().
				Uint64("nonce", msg.Nonce).
				Int64("latest-nonce", latestNonce).
				Msg("ignoring registration message with invalid nonce")
			continue
		}

		validator, err := vu.beaconClient.GetValidatorByIndex(ctx, "head", msg.ValidatorIndex)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get validator %d", msg.ValidatorIndex)
		}
		if validator == nil {
			logger.Warn().Msg("ignoring registration message for unknown validator")
			continue
		}
		pubkey, err := validator.Data.Validator.GetPubkey()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get pubkey of validator %d", msg.ValidatorIndex)
		}
		sig := new(blst.P2Affine).Uncompress(event.Signature)
		if sig == nil {
			logger.Warn().Msg("ignoring registration message with undecodable signature")
			continue
		}
		validSignature := validatorregistry.VerifySignature(sig, pubkey, msg)
		if !validSignature {
			logger.Warn().Msg("ignoring registration message with invalid signature")
			continue
		}

		filteredEvents = append(filteredEvents, event)
	}
	return filteredEvents, nil
}

func (vu *ValidatorUpdated) checkStaticRegistrationMessageFields(
	msg *validatorregistry.RegistrationMessage,
	validatorRegistryAddress common.Address,
	logger zerolog.Logger,
) bool {
	logger = logger.With().Uint64("validator-index", msg.ValidatorIndex).Logger()
	if msg.Version != ValidatorRegistrationMessageVersion {
		logger.Warn().
			Uint8("version", msg.Version).
			Uint8("expected-version", ValidatorRegistrationMessageVersion).
			Msg("ignoring registration message with invalid version")
		return false
	}
	if msg.ChainID != vu.chainID {
		logger.Warn().
			Uint64("chain-id", msg.ChainID).
			Uint64("expected-chain-id", vu.chainID).
			Uint64("validator-index", msg.ValidatorIndex).
			Msg("ignoring registration message with invalid chain ID")
		return false
	}
	if msg.ValidatorRegistryAddress != validatorRegistryAddress {
		logger.Warn().
			Hex("validator-registry-address", msg.ValidatorRegistryAddress.Bytes()).
			Hex("expected-validator-registry-address", validatorRegistryAddress.Bytes()).
			Msg("ignoring registration message with invalid validator registry address")
		return false
	}
	if msg.ValidatorIndex > math.MaxInt64 {
		logger.Warn().
			Msg("ignoring registration message with invalid validator index")
		return false
	}
	return true
}

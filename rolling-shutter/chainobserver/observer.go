package chainobserver

import (
	"context"
	"log"
	"math"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/shutter/shuttermint/commondb"
	"github.com/shutter-network/shutter/shuttermint/contract"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/decryptor/blsregistry"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/medley/eventsyncer"
	"github.com/shutter-network/shutter/shuttermint/shdb"
)

const finalityOffset = 3

func retryGetAddrs(ctx context.Context, addrsSeq *contract.AddrsSeq, n uint64) ([]common.Address, error) {
	callOpts := &bind.CallOpts{
		Pending: false,
		// We call for the current height instead of the height at which the event was emitted,
		// because the sets cannot change retroactively and we won't need an archive node.
		BlockNumber: nil,
		Context:     ctx,
	}
	addrsUntyped, err := medley.Retry(ctx, func() (interface{}, error) {
		return addrsSeq.GetAddrs(callOpts, n)
	})
	if err != nil {
		return []common.Address{}, errors.Wrapf(err, "failed to query decryptor addrs set from contract")
	}
	addrs := addrsUntyped.([]common.Address)
	return addrs, nil
}

type ChainObserver struct {
	contracts *deployment.Contracts
	dbpool    *pgxpool.Pool
}

func New(contracts *deployment.Contracts, dbpool *pgxpool.Pool) *ChainObserver {
	return &ChainObserver{contracts: contracts, dbpool: dbpool}
}

func (chainobs *ChainObserver) Observe(ctx context.Context, eventTypes []*eventsyncer.EventType) error {
	db := commondb.New(chainobs.dbpool)
	eventSyncProgress, err := db.GetEventSyncProgress(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get last synced event from db")
	}
	fromBlock := uint64(eventSyncProgress.NextBlockNumber)
	fromLogIndex := uint64(eventSyncProgress.NextLogIndex)

	log.Printf("starting event syncing from block %d log %d", fromBlock, fromLogIndex)
	syncer := eventsyncer.New(chainobs.contracts.Client, finalityOffset, eventTypes, fromBlock, fromLogIndex)

	errorgroup, errorctx := errgroup.WithContext(ctx)
	errorgroup.Go(func() error {
		return syncer.Run(errorctx)
	})
	errorgroup.Go(func() error {
		for {
			eventSyncUpdate, err := syncer.Next(errorctx)
			if err != nil {
				return err
			}
			if err := chainobs.handleEventSyncUpdate(errorctx, eventSyncUpdate); err != nil {
				return err
			}
		}
	})
	return errorgroup.Wait()
}

// handleEventSyncUpdate handles events and advances the sync state, but rolls back any db updates
// on failure.
func (chainobs *ChainObserver) handleEventSyncUpdate(
	ctx context.Context, eventSyncUpdate eventsyncer.EventSyncUpdate,
) error {
	return chainobs.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := commondb.New(tx)

		if eventSyncUpdate.Event != nil {
			if err := chainobs.handleEvent(ctx, db, eventSyncUpdate.Event); err != nil {
				return err
			}
		}

		var nextBlockNumber uint64
		var nextLogIndex uint64
		if eventSyncUpdate.Event == nil {
			nextBlockNumber = eventSyncUpdate.BlockNumber + 1
			nextLogIndex = 0
		} else {
			nextBlockNumber = eventSyncUpdate.BlockNumber
			nextLogIndex = eventSyncUpdate.LogIndex + 1
		}
		if err := db.UpdateEventSyncProgress(ctx, commondb.UpdateEventSyncProgressParams{
			NextBlockNumber: int32(nextBlockNumber),
			NextLogIndex:    int32(nextLogIndex),
		}); err != nil {
			return errors.Wrap(err, "failed to update last synced event")
		}
		return nil
	})
}

func (chainobs *ChainObserver) handleEvent(
	ctx context.Context, db *commondb.Queries, event interface{},
) error {
	var err error
	switch event := event.(type) {
	case contract.KeypersConfigsListNewConfig:
		err = chainobs.handleKeypersConfigsListNewConfigEvent(ctx, db, event)
	case contract.DecryptorsConfigsListNewConfig:
		err = chainobs.handleDecryptorsConfigsListNewConfigEvent(ctx, db, event)
	case contract.RegistryRegistered:
		err = chainobs.handleBLSRegistryRegistered(ctx, db, event)
	case contract.CollatorConfigsListNewConfig:
		err = chainobs.handleCollatorConfigsListNewConfigEvent(ctx, db, event)
	default:
		log.Printf("ignoring unknown event %+v %T", event, event)
	}
	return err
}

func (chainobs *ChainObserver) handleKeypersConfigsListNewConfigEvent(
	ctx context.Context, db *commondb.Queries, event contract.KeypersConfigsListNewConfig,
) error {
	log.Printf(
		"handling NewConfig event from keypers config contract in block %d (index %d, activation block number %d)",
		event.Raw.BlockNumber, event.Index, event.ActivationBlockNumber,
	)
	addrs, err := retryGetAddrs(ctx, chainobs.contracts.Keypers, event.Index)
	if err != nil {
		return err
	}

	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf(
			"activation block number %d from config contract would overflow int64",
			event.ActivationBlockNumber)
	}
	err = db.InsertKeyperSet(ctx, commondb.InsertKeyperSetParams{
		ActivationBlockNumber: int64(event.ActivationBlockNumber),
		Keypers:               shdb.EncodeAddresses(addrs),
		Threshold:             int32(event.Threshold),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to insert keyper set into db")
	}
	return nil
}

func (chainobs *ChainObserver) handleDecryptorsConfigsListNewConfigEvent(
	ctx context.Context, db *commondb.Queries, event contract.DecryptorsConfigsListNewConfig,
) error {
	log.Printf(
		"handling NewConfig event from decryptors config contract in block %d (index %d, activation block number %d)",
		event.Raw.BlockNumber, event.Index, event.ActivationBlockNumber,
	)
	addrs, err := retryGetAddrs(ctx, chainobs.contracts.Decryptors, event.Index)
	if err != nil {
		return err
	}
	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf(
			"activation block number %d from config contract would overflow int64",
			event.ActivationBlockNumber)
	}
	for i, addr := range addrs {
		encodedAddress := shdb.EncodeAddress(addr)
		err = db.InsertDecryptorSetMember(ctx, commondb.InsertDecryptorSetMemberParams{
			ActivationBlockNumber: int64(event.ActivationBlockNumber),
			Index:                 int32(i),
			Address:               encodedAddress,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to insert decryptor set member into db")
		}
	}
	return nil
}

func (chainobs *ChainObserver) handleBLSRegistryRegistered(
	ctx context.Context, db *commondb.Queries, event contract.RegistryRegistered,
) error {
	log.Printf(
		"handling BLS Registry event in block %d for decryptor %s",
		event.Raw.BlockNumber, event.A,
	)
	encodedAddress := shdb.EncodeAddress(event.A)
	rawIdentity := blsregistry.RawIdentity{
		KeyAndSignature: event.Data,
		EthereumAddress: event.A,
	}
	key, signature, err := rawIdentity.GetKeyAndSignature()
	if err != nil {
		log.Printf("Decryptor %s registered invalid BLS key and signature: %s", event.A, err)
		err = db.InsertDecryptorIdentity(ctx, commondb.InsertDecryptorIdentityParams{
			Address:        encodedAddress,
			BlsPublicKey:   []byte{},
			BlsSignature:   []byte{},
			SignatureValid: false,
		})
	} else {
		log.Printf("Decryptor %s successfully registered their BLS key", event.A)
		err = db.InsertDecryptorIdentity(ctx, commondb.InsertDecryptorIdentityParams{
			Address:        encodedAddress,
			BlsPublicKey:   shdb.EncodeBLSPublicKey(key),
			BlsSignature:   shdb.EncodeBLSSignature(signature),
			SignatureValid: true,
		})
	}
	if err != nil {
		return errors.Wrapf(
			err, "failed to insert identity of decryptor %s into db", event.A,
		)
	}
	return nil
}

func (chainobs *ChainObserver) handleCollatorConfigsListNewConfigEvent(
	ctx context.Context, db *commondb.Queries, event contract.CollatorConfigsListNewConfig,
) error {
	log.Printf(
		"handling NewConfig event from collator config contract in block %d (index %d, activation block number %d)",
		event.Raw.BlockNumber, event.Index, event.ActivationBlockNumber,
	)
	addrs, err := retryGetAddrs(ctx, chainobs.contracts.Collators, event.Index)
	if err != nil {
		return err
	}
	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf(
			"activation block number %d from config contract would overflow int64",
			event.ActivationBlockNumber,
		)
	}
	if len(addrs) > 1 {
		return errors.Errorf("got multiple collators from collator addrs set contract: %s", addrs)
	} else if len(addrs) == 1 {
		err = db.InsertChainCollator(ctx, commondb.InsertChainCollatorParams{
			ActivationBlockNumber: int64(event.ActivationBlockNumber),
			Collator:              shdb.EncodeAddress(addrs[0]),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to insert collator into db")
		}
	}
	return nil
}

package mocksequencer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/httphandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// func (proc *Sequencer) setupP2PHandler() {
// 	proc.p2p.AddMessageHandler(
// 		&eonPublicKeyHandler{config: c.Config, dbpool: c.dbpool},
// 	)
//
// 	c.p2p.AddGossipTopic(cltrtopics.DecryptionTrigger)
// }

type Eon struct {
	PublicKey             []byte
	ActivationBlockNumber uint64
	KeyperConfigIndex     uint64
	Eon                   uint64
}

type Collator struct {
	ActivationBlockNumber uint64
	Address               *common.Address
}

func (proc *Sequencer) getServices() []service.Service {
	rpcConfig := &httphandler.Config{
		L2BackendURL:       proc.Config.L2BackendURL,
		HTTPListenAddress:  proc.Config.HTTPListenAddress,
		EnableAdminService: proc.Config.Admin,
	}
	rpcServer := httphandler.NewRPCService(proc, rpcConfig)
	services := []service.Service{
		// service.ServiceFn{Fn: proc.handleContractEvents},
		service.ServiceFn{Fn: proc.handleOnChainChanges},
		service.ServiceFn{Fn: proc.pollTransactionReceipts},
		rpcServer,
	}
	return services
}

func (proc *Sequencer) handleContractEvents(ctx context.Context) error {
	events := []*eventsyncer.EventType{
		proc.contracts.KeypersConfigsListNewConfig,
		proc.contracts.CollatorConfigsListNewConfig,
	}
	return chainobserver.New(proc.contracts, proc.dbpool).Observe(ctx, events)
}

func (proc *Sequencer) pollTransactionReceipts(ctx context.Context) error {
	pollInterval := time.Duration(proc.Config.EthereumPollInterval) * time.Second
	ticker := time.NewTicker(pollInterval)
	client, err := ethclient.DialContext(ctx, proc.Config.L2BackendURL.String())
	if err != nil {
		return errors.Wrap(err, "error connecting to layer 1 JSON RPC endpoint")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			deleteHashes := []common.Hash{}
			proc.sentTransactionsLock.RLock()
			for hsh := range proc.sentTransactions {
				rcpt, err := client.TransactionReceipt(ctx, hsh)
				if err != nil {
					log.Warn().Err(err).Msg("error while polling transaction receipt")
				}
				if rcpt != nil {
					rcptJSON, _ := json.Marshal(rcpt)
					log.Info().Str("transaction-hash", hsh.Hex()).RawJSON("receipt", rcptJSON).Msg("got transaction receipt")
					deleteHashes = append(deleteHashes, hsh)
				}
			}
			proc.sentTransactionsLock.RUnlock()
			proc.sentTransactionsLock.Lock()
			for _, hsh := range deleteHashes {
				delete(proc.sentTransactions, hsh)
			}
			proc.sentTransactionsLock.Unlock()
		}
	}
}

func (proc *Sequencer) handleOnChainChanges(ctx context.Context) error {
	l1PollInterval := time.Duration(proc.Config.EthereumPollInterval) * time.Second
	ticker := time.NewTicker(l1PollInterval)
	l1Client, err := ethclient.DialContext(ctx, proc.Config.L2BackendURL.String())
	if err != nil {
		return errors.Wrap(err, "error connecting to layer 1 JSON RPC endpoint")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			newBlockNumber, err := l1Client.BlockNumber(ctx)
			if err != nil {
				log.Warn().Err(err).Msg("error retrieving block-number from layer 1 RPC")
				continue
			}
			proc.Mux.Lock()
			if newBlockNumber > proc.l1BlockNumber {
				proc.l1BlockNumber = newBlockNumber
				log.Debug().Int("l1-block-number", int(newBlockNumber)).Msg("updated state from layer 1 node")
			} else if newBlockNumber < proc.l1BlockNumber {
				log.Warn().Int("new-l1-block-number", int(newBlockNumber)).
					Int("cached-l1-block-number", int(proc.l1BlockNumber)).
					Msg("l1 cache inconsistency")
			}
			eon, err := proc.getEonForBlock(ctx, newBlockNumber)
			if err != nil {
				log.Warn().Err(err).Msg("error handling onchain collator")
			}
			// TODO We don't seem to use the collator contract currently
			// collator, err := proc.getChainCollatorForBlock(ctx, newBlockNumber)
			// if err != nil {
			// 	log.Warn().Err(err).Msg("error handling onchain collator")
			// }
			if eon == nil {
				log.Info().
					Interface("eon", eon).
					// Interface("collator", collator).
					Uint64("block-number", newBlockNumber).
					Msg("no EonPublicKey for this block")
				proc.Mux.Unlock()
				continue
			}
			if eon.ActivationBlockNumber < newBlockNumber &&
				!proc.active {
				err := proc.DisableAutomine(ctx)
				if err != nil {
					log.Error().Err(err).Msg("couldn't disable automine")
					proc.Mux.Unlock()
					continue
				}
				proc.active = true
				log.Info().Msg("disabled automining")
				// break out of the loop,
				// we don't need it anymore
				proc.Mux.Unlock()
				return nil
			}
		}
		proc.Mux.Unlock()
	}
}

func (proc *Sequencer) getEonForBlock(ctx context.Context, blockNumber uint64) (*Eon, error) {
	var eon *Eon
	err := proc.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		cdb := cltrdb.New(tx)
		b, err := medley.Uint64ToInt64Safe(blockNumber)
		if err != nil {
			return err
		}
		eonCandidate, err := cdb.FindEonPublicKeyForBlock(ctx, b)
		if err == pgx.ErrNoRows {
			return nil
		} else if err != nil {
			return err
		}
		activationBlockNumber, err := medley.Int64ToUint64Safe(eonCandidate.ActivationBlockNumber)
		if err != nil {
			return err
		}
		keyperConfigIndex, err := medley.Int64ToUint64Safe(eonCandidate.KeyperConfigIndex)
		if err != nil {
			return err
		}
		eonNumber, err := medley.Int64ToUint64Safe(eonCandidate.Eon)
		if err != nil {
			return err
		}
		eon = &Eon{
			PublicKey:             eonCandidate.EonPublicKey,
			ActivationBlockNumber: activationBlockNumber,
			KeyperConfigIndex:     keyperConfigIndex,
			Eon:                   eonNumber,
		}
		return nil
	})
	return eon, err
}

func (proc *Sequencer) getChainCollatorForBlock(
	ctx context.Context,
	blockNumber uint64,
) (*Collator, error) {
	var collator *Collator
	err := proc.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		cq := chainobsdb.New(tx)
		blockNum, err := medley.Uint64ToInt64Safe(blockNumber)
		if err != nil {
			return err
		}
		chainCollator, err := cq.GetChainCollator(ctx, blockNum)
		if err == pgx.ErrNoRows {
			return nil
		} else if err != nil {
			return err
		}
		collAddress, err := shdb.DecodeAddress(chainCollator.Collator)
		if err != nil {
			return err
		}
		collActivationBlock, err := medley.Int64ToUint64Safe(chainCollator.ActivationBlockNumber)
		if err != nil {
			return err
		}
		collator = &Collator{
			ActivationBlockNumber: collActivationBlock,
			Address:               &collAddress,
		}
		return nil
	})
	return collator, err
}

func (proc *Sequencer) handleAutomine(ctx context.Context) error {
	l2AutoBatchInterval := 10 * time.Second
	l2Ticker := time.NewTicker(l2AutoBatchInterval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-l2Ticker.C:
			proc.Mux.Lock()
			log.Info().Msg("will start automining block")
			proc.Mux.Unlock()
		}
	}
}

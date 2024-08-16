package syncer

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/chainsegment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

const MaxRequestBlockRange = 1000
const MaxSyncedBlockCacheSize = 1000

var ErrServerStateInconsistent = errors.New("server's chain-state differs from assumed local state")

type ChainSyncEthClient interface {
	ethereum.LogFilterer
	ethereum.ChainReader
}

type Fetcher struct {
	ethClient  ChainSyncEthClient
	chainCache ChainCache
	syncMux    sync.RWMutex

	chainUpdate *chainsegment.ChainSegment

	contractEventHandlers []ContractEventHandler
	topics                [][]common.Hash
	addresses             []common.Address

	chainUpdateHandlers []ChainUpdateHandler

	inChan         chan *types.Header
	processingTrig chan struct{}
}

func NewFetcher(c ChainSyncEthClient, chainCache ChainCache) *Fetcher {
	return &Fetcher{
		chainCache:            chainCache,
		ethClient:             c,
		syncMux:               sync.RWMutex{},
		chainUpdate:           &chainsegment.ChainSegment{},
		contractEventHandlers: []ContractEventHandler{},
		chainUpdateHandlers:   []ChainUpdateHandler{},
		topics:                [][]common.Hash{},
		addresses:             []common.Address{},
		inChan:                make(chan *types.Header),
		processingTrig:        make(chan struct{}, 1),
	}
}
func (f *Fetcher) GetHeaderByHash(ctx context.Context, h common.Hash) (*types.Header, error) {
		log.Error().Err(err).Msg("failed to query header from chain-cache")
	if err != nil {
		log.Error("failed to query header from chain-cache", "error", err)
		err = nil
		header, err = f.ethClient.HeaderByHash(ctx, h)
	if header == nil {
		header, err = f.client.HeaderByHash(ctx, h)
		if err != nil {
			err = fmt.Errorf("failed to query header from RPC client: %w", err)
		}
	}
	return header, err
}

func (f *Fetcher) Start(ctx context.Context, runner service.Runner) error {
	var err error
	for _, h := range f.contractEventHandlers {
		f.addresses = append(f.addresses, h.Address())
	}
	f.topics, err = topics(f.contractEventHandlers)
	if err != nil {
		return fmt.Errorf("can't construct topics for handler: %w", err)
	}

	// TODO: retry
	latest, err := f.client.HeaderByNumber(ctx, big.NewInt(-2))
	if err != nil {
		return fmt.Errorf("can't get header by number: %w", err)
	}

	f.chainUpdate = chainsegment.NewChainSegment(latest)

	subs, err := f.ethClient.SubscribeNewHead(ctx, f.inChan)
	if err != nil {
		return fmt.Errorf("can't subscribe to new head: %w", err)
	}
	runner.Defer(subs.Unsubscribe)
	runner.Defer(func() {
		close(f.inChan)
		close(f.processingTrig)
	})
	runner.Go(func() error {
		err := f.loop(ctx)
		if err != nil {
			return fmt.Errorf("fetcher loop errored: %w", err)
		}
		return nil
	})
	return nil
}

// This method has to be called before starting the Fetcher.
func (f *Fetcher) RegisterContractEventHandler(h ContractEventHandler) {
	f.syncMux.Lock()
	defer f.syncMux.Unlock()

	f.contractEventHandlers = append(f.contractEventHandlers, h)
}

// This method has to be called before starting the Fetcher.
func (f *Fetcher) RegisterChainUpdateHandler(h ChainUpdateHandler) {
	f.syncMux.Lock()
	defer f.syncMux.Unlock()

	f.chainUpdateHandlers = append(f.chainUpdateHandlers, h)
}

func (f *Fetcher) processChainUpdateHandler(ctx context.Context, update ChainUpdateContext, h ChainUpdateHandler) error {
	return h.Handle(ctx, update)
}

func (f *Fetcher) processContractEventHandler(
	ctx context.Context,
	update ChainUpdateContext,
	h ContractEventHandler,
	logs []types.Log,
) error {
	var result error
	events := []any{}
	for _, l := range logs {
		// don't process logs from a different contract
		if h.Address().Cmp(l.Address) != 0 {
			continue
		}
		// don't process logs with non-matching topics
		topicMatch := false
		for _, t := range l.Topics {
			if h.Topic().Cmp(t) == 0 {
				topicMatch = true
				break
			}
		}
		if !topicMatch {
			continue
		}

		a, err := h.Parse(l)
		// error here means we skip processing for this handler
		if err != nil {
			// TODO: we could log some errors here if they are not "wrong topic"
			continue
		}
		header := update.Append.GetHeaderByHash(l.BlockHash)
		if header == nil {
			log.Error().Err(ErrServerStateInconsistent).Str("log-block-hash", l.BlockHash.String())
			result = multierror.Append(result, err)
			continue
		}
		accept, err := h.Accept(ctx, *header, a)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		if accept {
			events = append(events, a)
		}
	}
	if errors.Is(result, ErrCritical) {
	if errors.Is(result, errs.ErrCritical) {
	}
	return h.Handle(ctx, update, events)
}

func (f *Fetcher) FetchAndHandle(ctx context.Context, update ChainUpdateContext) error {
	query := ethereum.FilterQuery{
		Addresses: f.addresses,
		Topics:    f.topics,
		//FIXME: oes this work when from and to are the same?
		FromBlock: update.Append.Earliest().Number,
		ToBlock:   update.Append.Latest().Number,
	}

	logs, err := f.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return err
	}
	for _, l := range logs {
		if update.Append.GetHeaderByHash(l.BlockHash) == nil {
			// The API only allows filtering by blocknumber.
			// If the retrieved log's block-hash is not present in the
			// update query-context,
			// this means the server is operating on a different
			// chain-state (e.g. reorged).
			return ErrServerStateInconsistent
		}
	}

	wg := sync.WaitGroup{}
	var result error
	f.syncMux.RLock()
	for _, h := range f.contractEventHandlers {
		// TODO: copy all the logs?
		handler := h
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := f.processContractEventHandler(ctx, update, handler, logs)
			if err != nil {
				err = fmt.Errorf("contract-event-handler error: %w", err)
				log.Error().Err(err).Msg("handler processing errored")
				result = multierror.Append(result, err)
			}
		}()
	}
	f.syncMux.RUnlock()
	// run the chain-update handlers after the contract event handlers did run.
	for _, h := range f.chainUpdateHandlers {
	wg.Wait()
	for i, h := range f.chainUpdateHandlers {
		handler := h
		f.log.Info("spawning chain update handler", "num", i)
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := f.processChainUpdateHandler(ctx, update, handler)
			if err != nil {
				err = fmt.Errorf("chain-update-handler error: %w", err)
				log.Error().Err(err).Msg("handler processing errored")
				result = multierror.Append(result, err)
			}
		}()
	}
	f.syncMux.RUnlock()
	wg.Wait()
	return result
}

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
	"github.com/ethereum/go-ethereum/log"
	"github.com/hashicorp/go-multierror"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/chainsegment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

const MaxRequestBlockRange = 1000
const MaxSyncedBlockCacheSize = 1000

var ErrServerStateInconsistent = errors.New("server's chain-state differs from assumed local state")

// TODO: naming
type ChainClient interface {
	ethereum.LogFilterer
	ethereum.ChainReader
}

type Fetcher struct {
	client     ChainClient
	chainCache ChainCache
	log        log.Logger
	syncMux    sync.RWMutex

	chainUpdate *chainsegment.ChainSegment

	contractEventHandlers []ContractEventHandler
	topics                [][]common.Hash
	addresses             []common.Address

	chainUpdateHandlers []ChainUpdateHandler

	inChan         chan *types.Header
	processingTrig chan struct{}
}

func NewFetcher(c ChainClient, chainCache ChainCache, log log.Logger) *Fetcher {
	return &Fetcher{
		chainCache:            chainCache,
		client:                c,
		log:                   log,
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
	f.log.Info("current latest head", "latest", latest)
	f.chainUpdate = chainsegment.NewChainSegment(latest)

	subs, err := f.client.SubscribeNewHead(ctx, f.inChan)
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
	f.contractEventHandlers = append(f.contractEventHandlers, h)
}

// This method has to be called before starting the Fetcher.
func (f *Fetcher) RegisterChainUpdateHandler(h ChainUpdateHandler) {
	f.chainUpdateHandlers = append(f.chainUpdateHandlers, h)
}

func (f *Fetcher) processChainUpdateHandler(ctx context.Context, qCtx QueryContext, h ChainUpdateHandler) error {
	return h.Handle(ctx, qCtx)
}

func (f *Fetcher) processContractEventHandler(ctx context.Context, qCtx QueryContext, h ContractEventHandler, logs []types.Log) error {
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

		// XXX: err ineffassign
		// TODO: maybe remove the bool ok?
		a, ok, err := h.Parse(l)
		if !ok {
			continue
		}
		header := qCtx.Update.GetHeaderByHash(l.BlockHash)
		if header == nil {
			f.log.Error(ErrServerStateInconsistent.Error(), "log-block-hash", l.BlockHash)
			result = multierror.Append(result, err)
			continue
		}
		accept, err := h.Accept(ctx, *header, a)
		if err != nil {
			f.log.Error("accept handler errored", "error", err)
			result = multierror.Append(result, err)
			continue
		}
		if accept {
			events = append(events, a)
		}
	}
	if errors.Is(result, ErrCritical) {
		return result
	}
	return h.Handle(ctx, qCtx, events)
}

func (f *Fetcher) FetchAndHandle(ctx context.Context, qCtx QueryContext) error {
	query := ethereum.FilterQuery{
		Addresses: f.addresses,
		Topics:    f.topics,
		//FIXME: does this work when from and to are the same?
		FromBlock: qCtx.Update.Earliest().Number,
		ToBlock:   qCtx.Update.Latest().Number,
	}

	logs, err := f.client.FilterLogs(ctx, query)
	if err != nil {
		return err
	}
	for _, l := range logs {
		if qCtx.Update.GetHeaderByHash(l.BlockHash) == nil {
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
	for _, h := range f.contractEventHandlers {
		// TODO: copy all the logs?
		handler := h
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := f.processContractEventHandler(ctx, qCtx, handler, logs)
			if err != nil {
				err = fmt.Errorf("contract-event-handler error: %w", err)
				f.log.Error("handler processing errored", "error", err)
				result = multierror.Append(result, err)
			}
		}()
	}
	//XXX: this runs the chain update handler in parallel
	// with the event handlers. Is this good or would it
	// be better to e.g. run them AFTER the contract
	// event handlers did run?
	for i, h := range f.chainUpdateHandlers {
		handler := h
		f.log.Info("spawing chain update handler", "num", i)
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := f.processChainUpdateHandler(ctx, qCtx, handler)
			if err != nil {
				err = fmt.Errorf("chain-update-handler error: %w", err)
				f.log.Error("handler processing errored", "error", err)
				result = multierror.Append(result, err)
			}
		}()
	}
	wg.Wait()
	return result
}

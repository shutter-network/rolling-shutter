package eventsyncer

import (
	"context"
	"math/big"
	"reflect"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

const (
	outputChannelCapacity = 32              // number of log entries we put on the (internal) log channel
	pageSizeBlocks        = 3               // number of blocks over that one filter query spans
	blockPollInterval     = 2 * time.Second // time to wait before checking for new blocks
)

var (
	ErrAlreadyRunning = errors.New("event syncer already running")
	ErrNotRunning     = errors.New("event syncer not running")
)

type database interface {
	db.Definition
	HandleEvent(context.Context, pgx.Tx, any) error
}

type (
	EventHandlerFunc               func(context.Context, pgx.Tx, any) error
	EventHandlerFuncGeneric[T any] func(context.Context, pgx.Tx, T) error
)

// MakeHandler converts a handler function with specific type to a handler
// with type any. The handler is wrapped with a type assertion on the specific type.
func MakeHandler[T any](handler EventHandlerFuncGeneric[T]) EventHandlerFunc {
	anyHandler := func(ctx context.Context, tx pgx.Tx, anyEvent any) error {
		event, ok := anyEvent.(T)
		if !ok {
			return errors.New("event does not match the type for the target handler function")
		}
		return handler(ctx, tx, event)
	}
	return anyHandler
}

// EventType defines a single event type to filter for.
type EventType struct {
	Contract        *bind.BoundContract
	Address         common.Address
	FromBlockNumber uint64
	ABI             abi.ABI
	Name            string
	Type            reflect.Type
	Handler         EventHandlerFunc
}

// logChannelItem is what is put on the (internal) channel of found logs. It can either contain a
// log (with the number of the block in which it was found and its event type), or only a block
// number (with nil log and event type). The latter communicates that no further logs have been
// found up until the given block.
type logChannelItem struct {
	log         *types.Log
	blockNumber uint64
	eventType   *EventType
}

type EventSyncUpdate struct {
	Event       interface{}
	BlockNumber uint64
	LogIndex    uint64
}

// dispatchKey identifies an EventType by its contract address and the first
// topic of the matching log.
type dispatchKey struct {
	address common.Address
	topic   common.Hash
}

// EventSyncer watches the blockchain for events of given types and yields them in order.
type EventSyncer struct {
	Client         *ethclient.Client
	FinalityOffset uint64

	Events       []*EventType
	FromBlock    uint64
	FromLogIndex uint64

	addresses []common.Address
	topics    []common.Hash
	dispatch  map[dispatchKey]*EventType

	started    bool
	logChannel chan logChannelItem
}

// New creates a new event syncer. It will look for events starting at a certain block number and
// log index. The types of events to filter for are specified as a set of EventTypes. The finality
// offset is the number of blocks we trail behind the current block to be safe from reorgs.
func New(client *ethclient.Client, finalityOffset uint64, events []*EventType, fromBlock uint64, fromLogIndex uint64) *EventSyncer {
	addressSet := map[common.Address]struct{}{}
	topicSet := map[common.Hash]struct{}{}
	dispatch := map[dispatchKey]*EventType{}
	for _, ev := range events {
		topic := ev.ABI.Events[ev.Name].ID
		addressSet[ev.Address] = struct{}{}
		topicSet[topic] = struct{}{}
		dispatch[dispatchKey{address: ev.Address, topic: topic}] = ev
	}
	addresses := make([]common.Address, 0, len(addressSet))
	for a := range addressSet {
		addresses = append(addresses, a)
	}
	topics := make([]common.Hash, 0, len(topicSet))
	for t := range topicSet {
		topics = append(topics, t)
	}

	return &EventSyncer{
		Client:         client,
		FinalityOffset: finalityOffset,

		Events:       events,
		FromBlock:    fromBlock,
		FromLogIndex: fromLogIndex,

		addresses: addresses,
		topics:    topics,
		dispatch:  dispatch,

		started:    false,
		logChannel: make(chan logChannelItem, outputChannelCapacity),
	}
}

// Next returns the next found event, if any. The first return value is of type `EventType.Type`
// (depending on which event was found). The second return value contains the block number in
// which the event was found. The function can also return a nil event, communicating that no
// further events were found up until the block number in the second return value. The function
// may take up to the poll interval to return. It must only be called after the syncer was started
// with `Run`.
func (s *EventSyncer) Next(ctx context.Context) (EventSyncUpdate, error) {
	select {
	case item := <-s.logChannel:
		if item.log == nil {
			return EventSyncUpdate{
				Event:       nil,
				BlockNumber: item.blockNumber,
				LogIndex:    0,
			}, nil
		}

		event := reflect.New(item.eventType.Type)
		err := item.eventType.Contract.UnpackLog(event.Interface(), item.eventType.Name, *item.log)
		if err != nil {
			return EventSyncUpdate{}, errors.Wrapf(
				err,
				"failed to unpack log of %s event", item.eventType.Name,
			)
		}
		reflect.Indirect(event).FieldByName("Raw").Set(reflect.ValueOf(*item.log))

		return EventSyncUpdate{
			Event:       reflect.Indirect(event).Interface(),
			BlockNumber: item.blockNumber,
			LogIndex:    uint64(item.log.Index),
		}, nil
	case <-ctx.Done():
		return EventSyncUpdate{}, ctx.Err()
	}
}

// Run the syncer.
func (s *EventSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.started {
		return ErrAlreadyRunning
	}
	s.started = true

	runner.Go(
		func() error {
			return s.sync(ctx)
		})
	return nil
}

// sync continuously searches for events. It prefers a new-head subscription
// (one eth_subscribe at startup) over the periodic eth_blockNumber poll and
// falls back to polling when the transport does not support subscriptions
// (typically HTTP-only ethclients).
func (s *EventSyncer) sync(ctx context.Context) error {
	fromBlock := s.FromBlock

	headCh := make(chan *types.Header, 1)
	sub, err := s.Client.SubscribeNewHead(ctx, headCh)
	if err != nil {
		log.Debug().Err(err).Msg("SubscribeNewHead unavailable, falling back to polling")
		return s.syncByPolling(ctx, fromBlock)
	}
	defer sub.Unsubscribe()

	return s.syncBySubscription(ctx, fromBlock, headCh, sub)
}

// syncBySubscription drives the sync loop off a SubscribeNewHead
// subscription instead of polling BlockNumber. On startup it also queries
// HeaderByNumber(nil) once so the initial historical range can be drained
// before the first live head arrives; from then on the subscription is the
// only signal for new work.
func (s *EventSyncer) syncBySubscription(
	ctx context.Context,
	fromBlock uint64,
	headCh <-chan *types.Header,
	sub ethereum.Subscription,
) error {
	// Kick off with the current tip so we cover the historical gap between
	// FromBlock and the head, without waiting for a new block to arrive.
	tip, err := retry.FunctionCall(ctx, func(ctx context.Context) (*types.Header, error) {
		return s.Client.HeaderByNumber(ctx, nil)
	})
	if err != nil {
		return errors.Wrap(err, "failed to query current head")
	}
	fromBlock, err = s.advanceUpTo(ctx, fromBlock, tip.Number.Uint64())
	if err != nil {
		return err
	}

	for {
		select {
		case head, ok := <-headCh:
			if !ok {
				return errors.New("event syncer head channel closed")
			}
			fromBlock, err = s.advanceUpTo(ctx, fromBlock, head.Number.Uint64())
			if err != nil {
				return err
			}
		case err := <-sub.Err():
			return errors.Wrap(err, "event syncer new-head subscription failed")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// syncByPolling is the pre-subscription loop, retained for transports that
// cannot subscribe (HTTP ethclient).
func (s *EventSyncer) syncByPolling(ctx context.Context, fromBlock uint64) error {
	for {
		currentBlock, err := retry.FunctionCall(ctx, s.Client.BlockNumber)
		if err != nil {
			return errors.Wrap(err, "failed to query current block number")
		}

		fromBlock, err = s.advanceUpTo(ctx, fromBlock, currentBlock)
		if err != nil {
			return err
		}

		select {
		case <-time.After(blockPollInterval):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// advanceUpTo advances the syncer up to (currentBlock - FinalityOffset),
// consuming as many pageSizeBlocks-wide ranges as fit. Returns the new
// fromBlock. When the current chain is not yet ahead of the finality offset,
// nothing is fetched.
func (s *EventSyncer) advanceUpTo(ctx context.Context, fromBlock, currentBlock uint64) (uint64, error) {
	var maxToBlock uint64
	if currentBlock >= s.FinalityOffset {
		maxToBlock = currentBlock - s.FinalityOffset
	}
	for fromBlock <= maxToBlock {
		toBlock := fromBlock + pageSizeBlocks - 1
		if toBlock > maxToBlock {
			toBlock = maxToBlock
		}
		logItems, err := s.syncAllInRange(ctx, fromBlock, toBlock)
		if err != nil {
			return fromBlock, err
		}
		if err := s.sendLogItemsToChannel(ctx, logItems, toBlock); err != nil {
			return fromBlock, err
		}
		fromBlock = toBlock + 1
	}
	return fromBlock, nil
}

// syncAllInRange returns all events found in the given block range. It issues
// a single FilterLogs query covering every registered address and topic, then
// dispatches each returned log to its matching EventType.
func (s *EventSyncer) syncAllInRange(ctx context.Context, fromBlock uint64, toBlock uint64) ([]logChannelItem, error) {
	if len(s.Events) == 0 {
		return nil, nil
	}
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: s.addresses,
		Topics:    [][]common.Hash{s.topics},
	}

	logs, err := retry.FunctionCall(ctx, func(ctx context.Context) ([]types.Log, error) {
		return s.Client.FilterLogs(ctx, query)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to filter event logs")
	}

	items := make([]logChannelItem, 0, len(logs))
	for i := range logs {
		l := &logs[i]
		if len(l.Topics) == 0 {
			continue
		}
		ev, ok := s.dispatch[dispatchKey{address: l.Address, topic: l.Topics[0]}]
		if !ok {
			// The (address, topic) combination has no registered handler. This
			// can happen when an address listed for one event type also emits a
			// topic that belongs to another address; we simply skip it.
			continue
		}
		items = append(items, logChannelItem{
			log:         l,
			blockNumber: l.BlockNumber,
			eventType:   ev,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		bi := items[i].log.BlockNumber
		bj := items[j].log.BlockNumber
		if bi != bj {
			return bi < bj
		}
		return items[i].log.Index < items[j].log.Index
	})

	return items, nil
}

// sendLogItemsToChannel puts the given log channel items to the internal logChannel and finishes
// with an empty log item with block number `syncedUntil`.
func (s *EventSyncer) sendLogItemsToChannel(ctx context.Context, items []logChannelItem, syncedUntil uint64) error {
	for _, item := range items {
		// ignore logs older than (s.FromBlock, s.FromLogIndex)
		if item.log.BlockNumber < s.FromBlock {
			continue
		}
		if item.log.BlockNumber == s.FromBlock && uint64(item.log.Index) < s.FromLogIndex {
			continue
		}

		select {
		case s.logChannel <- item:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// signal that all logs up until syncedUntil have been synced
	endItem := logChannelItem{
		log:         nil,
		blockNumber: syncedUntil,
	}
	select {
	case s.logChannel <- endItem:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

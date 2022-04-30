package eventsyncer

import (
	"context"
	"math/big"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
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

// EventType defines a single event type to filter for.
type EventType struct {
	Contract *bind.BoundContract
	Address  common.Address
	ABI      abi.ABI
	Name     string
	Type     reflect.Type
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

// EventSyncer watches the blockchain for events of given types and yields them in order.
type EventSyncer struct {
	Client         *ethclient.Client
	FinalityOffset uint64

	Events       []*EventType
	FromBlock    uint64
	FromLogIndex uint64

	started    bool
	logChannel chan logChannelItem
}

// New creates a new event syncer. It will look for events starting at a certain block number and
// log index. The types of events to filter for are specified as a set of EventTypes. The finality
// offset is the number of blocks we trail behind the current block to be safe from reorgs.
func New(client *ethclient.Client, finalityOffset uint64, events []*EventType, fromBlock uint64, fromLogIndex uint64) *EventSyncer {
	return &EventSyncer{
		Client:         client,
		FinalityOffset: finalityOffset,

		Events:       events,
		FromBlock:    fromBlock,
		FromLogIndex: fromLogIndex,

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
func (s *EventSyncer) Run(ctx context.Context) error {
	if s.started {
		return ErrAlreadyRunning
	}
	s.started = true

	return s.sync(ctx)
}

// sync continuously searches for events.
func (s *EventSyncer) sync(ctx context.Context) error {
	fromBlock := s.FromBlock
	for {
		currentBlock, err := medley.Retry(ctx, func() (uint64, error) {
			return s.Client.BlockNumber(ctx)
		})
		if err != nil {
			return errors.Wrap(err, "failed to query current block number")
		}

		toBlock := fromBlock + pageSizeBlocks - 1
		var maxToBlock uint64
		if currentBlock >= s.FinalityOffset {
			maxToBlock = currentBlock - s.FinalityOffset
		} else {
			maxToBlock = 0
		}
		if toBlock > maxToBlock {
			toBlock = maxToBlock
		}

		// if there's no new blocks, wait some time and try again
		if toBlock < fromBlock {
			select {
			case <-time.After(blockPollInterval):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		logItems, err := s.syncAllInRange(ctx, fromBlock, toBlock)
		if err != nil {
			return err
		}
		err = s.sendLogItemsToChannel(ctx, logItems, toBlock)
		if err != nil {
			return err
		}

		fromBlock = toBlock + 1
	}
}

// syncAllInRange returns all events found in the given block range.
func (s *EventSyncer) syncAllInRange(ctx context.Context, fromBlock uint64, toBlock uint64) ([]logChannelItem, error) {
	logs := []logChannelItem{}
	mu := sync.Mutex{}

	errorgroup, errorctx := errgroup.WithContext(ctx)
	for _, event := range s.Events {
		ev := event
		errorgroup.Go(func() error {
			logsSingle, err := s.syncSingleInRange(errorctx, ev, fromBlock, toBlock)
			if err != nil {
				return err
			}

			mu.Lock()
			defer mu.Unlock()
			logs = append(logs, logsSingle...)
			return nil
		})
	}
	if err := errorgroup.Wait(); err != nil {
		return nil, err
	}

	sort.Slice(logs, func(i, j int) bool {
		bi := logs[i].log.BlockNumber
		bj := logs[j].log.BlockNumber
		if bi < bj {
			return true
		}
		if bi == bj {
			li := logs[i].log.Index
			lj := logs[j].log.Index
			return li < lj
		}
		return false
	})

	return logs, nil
}

// syncSingleInRange returns the events matching the given type in the given block range.
func (s *EventSyncer) syncSingleInRange(ctx context.Context, event *EventType, fromBlock uint64, toBlock uint64) ([]logChannelItem, error) {
	topic := event.ABI.Events[event.Name].ID
	query := ethereum.FilterQuery{
		BlockHash: nil,
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{event.Address},
		Topics:    [][]common.Hash{{topic}},
	}

	logs, err := medley.Retry(ctx, func() ([]types.Log, error) {
		return s.Client.FilterLogs(ctx, query)
	})
	if err != nil {
		return nil, errors.New("failed to filter event logs")
	}

	items := []logChannelItem{}
	for i := range logs {
		items = append(items, logChannelItem{
			log:         &logs[i],
			blockNumber: logs[i].BlockNumber,
			eventType:   event,
		})
	}
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

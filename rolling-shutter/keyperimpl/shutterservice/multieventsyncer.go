package shutterservice

import (
	"bytes"
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
)

const (
	DefaultAssumedReorgDepth    = 10
	DefaultMaxRequestBlockRange = 10_000
)

type MultiEventSyncer struct {
	Processors           map[string]EventProcessor
	DBPool               *pgxpool.Pool
	ExecutionClient      *ethclient.Client
	SyncStartBlockNumber uint64
	AssumedReorgDepth    int
	MaxRequestBlockRange uint64
}

type SyncStatus struct {
	BlockNumber int64
	BlockHash   []byte
}

func NewMultiEventSyncer(
	dbPool *pgxpool.Pool,
	executionClient *ethclient.Client,
	syncStartBlockNumber uint64,
	processors []EventProcessor,
) (*MultiEventSyncer, error) {
	processorMap := make(map[string]EventProcessor)
	for _, processor := range processors {
		name := processor.GetProcessorName()
		if _, exists := processorMap[name]; exists {
			return nil, errors.Errorf("duplicate processor name: %s", name)
		}
		processorMap[name] = processor
	}

	return &MultiEventSyncer{
		Processors:           processorMap,
		DBPool:               dbPool,
		ExecutionClient:      executionClient,
		SyncStartBlockNumber: syncStartBlockNumber,
		AssumedReorgDepth:    DefaultAssumedReorgDepth,
		MaxRequestBlockRange: DefaultMaxRequestBlockRange,
	}, nil
}

func (s *MultiEventSyncer) Sync(ctx context.Context, header *types.Header) error {
	if err := s.handlePotentialReorg(ctx, header); err != nil {
		return errors.Wrap(err, "failed to handle potential reorg")
	}

	syncedUntil, err := s.getSyncedUntil(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to determine sync start point")
	}
	start := uint64(syncedUntil + 1)
	end := header.Number.Uint64()
	if start > end {
		log.Debug().
			Uint64("start-block", start).
			Uint64("end-block", end).
			Msg("already synced up to target block")
		return nil
	}

	syncRanges := medley.GetSyncRanges(start, end, s.MaxRequestBlockRange)
	log.Debug().
		Uint64("start-block", start).
		Uint64("end-block", end).
		Int("num-sync-ranges", len(syncRanges)).
		Msg("starting multi event sync")
	numEvents := 0
	for _, r := range syncRanges {
		numEventsInRange, err := s.syncRange(ctx, r[0], r[1])
		if err != nil {
			return errors.Wrapf(err, "failed to sync range [%d, %d]", r[0], r[1])
		}
		numEvents += numEventsInRange
	}

	log.Info().
		Uint64("start-block", start).
		Uint64("end-block", end).
		Int("num-events", numEvents).
		Msg("completed multi event sync")
	return nil
}

// processorScope holds the per-processor address/topic filter so we can route
// the unified FilterLogs response back to each processor without re-checking
// criteria.
type processorScope struct {
	processor EventProcessor
	addresses map[common.Address]struct{}
	topics0   map[common.Hash]struct{}
	anyTopic  bool
}

func (s *MultiEventSyncer) syncRange(ctx context.Context, start, end uint64) (int, error) {
	header, err := s.ExecutionClient.HeaderByNumber(ctx, new(big.Int).SetUint64(end))
	if err != nil {
		return 0, errors.Wrap(err, "failed to get execution block header")
	}

	scopes, addressUnion, topicUnion, anyTopic, err := s.gatherFilterCriteria(ctx, start, end)
	if err != nil {
		return 0, err
	}

	var logs []types.Log
	if len(addressUnion) > 0 {
		query := ethereum.FilterQuery{
			FromBlock: new(big.Int).SetUint64(start),
			ToBlock:   new(big.Int).SetUint64(end),
			Addresses: addressUnion,
		}
		if !anyTopic && len(topicUnion) > 0 {
			query.Topics = [][]common.Hash{topicUnion}
		}
		logs, err = s.ExecutionClient.FilterLogs(ctx, query)
		if err != nil {
			return 0, errors.Wrap(err, "failed to fetch logs for combined event query")
		}
	}

	allEvents := make(map[string][]Event)
	numEvents := 0
	for _, scope := range scopes {
		scoped := filterLogsForScope(logs, scope)
		events, err := scope.processor.ParseEvents(ctx, start, end, scoped)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to parse events for processor %s in range [%d, %d]",
				scope.processor.GetProcessorName(), start, end)
		}
		allEvents[scope.processor.GetProcessorName()] = events
		numEvents += len(events)
	}

	err = s.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		for _, scope := range scopes {
			events := allEvents[scope.processor.GetProcessorName()]
			if err := scope.processor.ProcessEvents(ctx, tx, events); err != nil {
				return errors.Wrapf(err, "failed to process events for processor %s",
					scope.processor.GetProcessorName())
			}
		}
		if err := s.setSyncStatus(ctx, tx, int64(end), header.Hash().Bytes()); err != nil {
			return errors.Wrap(err, "failed to update global sync status")
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return numEvents, nil
}

// gatherFilterCriteria asks each processor for its filter criteria and
// returns:
//   - per-processor scopes used to route logs back after the combined fetch,
//   - the deduplicated union of addresses for the FilterQuery,
//   - the deduplicated union of topic[0] hashes, and
//   - anyTopic=true if any processor declined to constrain topic[0] (in which
//     case the combined query must not filter on topics).
func (s *MultiEventSyncer) gatherFilterCriteria(
	ctx context.Context, start, end uint64,
) ([]processorScope, []common.Address, []common.Hash, bool, error) {
	scopes := make([]processorScope, 0, len(s.Processors))
	addressSet := map[common.Address]struct{}{}
	topicSet := map[common.Hash]struct{}{}
	anyTopic := false
	for name, processor := range s.Processors {
		crit, err := processor.FilterCriteria(ctx, start, end)
		if err != nil {
			return nil, nil, nil, false, errors.Wrapf(err,
				"failed to gather filter criteria for processor %s", name)
		}
		scope := processorScope{
			processor: processor,
			addresses: map[common.Address]struct{}{},
			topics0:   map[common.Hash]struct{}{},
			anyTopic:  len(crit.Topics0) == 0,
		}
		for _, addr := range crit.Addresses {
			scope.addresses[addr] = struct{}{}
			addressSet[addr] = struct{}{}
		}
		for _, topic := range crit.Topics0 {
			scope.topics0[topic] = struct{}{}
			topicSet[topic] = struct{}{}
		}
		if scope.anyTopic && len(crit.Addresses) > 0 {
			anyTopic = true
		}
		scopes = append(scopes, scope)
	}

	addresses := make([]common.Address, 0, len(addressSet))
	for a := range addressSet {
		addresses = append(addresses, a)
	}
	topics := make([]common.Hash, 0, len(topicSet))
	for t := range topicSet {
		topics = append(topics, t)
	}
	return scopes, addresses, topics, anyTopic, nil
}

// filterLogsForScope returns the subset of logs that fall within a processor's
// (address, topic[0]) criteria.
func filterLogsForScope(logs []types.Log, scope processorScope) []types.Log {
	if len(scope.addresses) == 0 {
		return nil
	}
	filtered := make([]types.Log, 0, len(logs))
	for i := range logs {
		l := &logs[i]
		if _, ok := scope.addresses[l.Address]; !ok {
			continue
		}
		if !scope.anyTopic {
			if len(l.Topics) == 0 {
				continue
			}
			if _, ok := scope.topics0[l.Topics[0]]; !ok {
				continue
			}
		}
		filtered = append(filtered, *l)
	}
	return filtered
}

func (s *MultiEventSyncer) getSyncStatus(ctx context.Context) (*SyncStatus, error) {
	queries := database.New(s.DBPool)
	status, err := queries.GetMultiEventSyncStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &SyncStatus{
		BlockNumber: status.BlockNumber,
		BlockHash:   status.BlockHash,
	}, nil
}

func (s *MultiEventSyncer) setSyncStatus(ctx context.Context, tx pgx.Tx, blockNumber int64, blockHash []byte) error {
	queries := database.New(tx)
	return queries.SetMultiEventSyncStatus(ctx, database.SetMultiEventSyncStatusParams{
		BlockNumber: blockNumber,
		BlockHash:   blockHash,
	})
}

func (s *MultiEventSyncer) getSyncedUntil(ctx context.Context) (int64, error) {
	status, err := s.getSyncStatus(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return int64(s.SyncStartBlockNumber), nil
		}
		return 0, err
	}
	return status.BlockNumber, nil
}

func calculateReorgDepth(status *SyncStatus, header *types.Header, assumedReorgDepth int) int {
	shouldBeParent := header.Number.Int64() == status.BlockNumber+1
	isParent := bytes.Equal(header.ParentHash.Bytes(), status.BlockHash)
	isReorg := shouldBeParent && !isParent
	if !isReorg {
		return 0
	}

	// To avoid finding the exact branch point, we just assume a fixed, conservative depth
	depth := assumedReorgDepth
	if status.BlockNumber < int64(depth) {
		return int(status.BlockNumber)
	}
	return depth
}

func (s *MultiEventSyncer) handlePotentialReorg(ctx context.Context, header *types.Header) error {
	status, err := s.getSyncStatus(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil // if nothing is synced yet, no reorg is necessary
		}
		return errors.Wrap(err, "failed to get sync status")
	}
	numReorgedBlocks := calculateReorgDepth(status, header, s.AssumedReorgDepth)
	if numReorgedBlocks == 0 {
		return nil
	}

	toBlock := status.BlockNumber - int64(numReorgedBlocks)
	log.Info().
		Int("reorg-depth", numReorgedBlocks).
		Int64("rollback-to-block-number", toBlock).
		Uint64("current-block-number", header.Number.Uint64()).
		Hex("current-block-hash", header.Hash().Bytes()).
		Msg("detected blockchain reorg, rolling back processors")
	return s.rollback(ctx, toBlock)
}

func (s *MultiEventSyncer) rollback(ctx context.Context, toBlock int64) error {
	status, err := s.getSyncStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get sync status during rollback")
	}

	if toBlock > status.BlockNumber {
		return errors.Errorf("invalid rollback target: toBlock (%d) is greater than current synced block (%d)",
			toBlock, status.BlockNumber)
	}

	return s.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		for name, processor := range s.Processors {
			err = processor.RollbackEvents(ctx, tx, toBlock)
			if err != nil {
				return errors.Wrapf(err, "failed to rollback events for processor %s", name)
			}
		}

		err = s.setSyncStatus(ctx, tx, toBlock, []byte{})
		if err != nil {
			return errors.Wrap(err, "failed to update sync status during rollback")
		}

		log.Info().
			Int64("previous-synced-until", status.BlockNumber).
			Int64("new-synced-until", toBlock).
			Msg("rolled back all processors due to reorg")
		return nil
	})
}

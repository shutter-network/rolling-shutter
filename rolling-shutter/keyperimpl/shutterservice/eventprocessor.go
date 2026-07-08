package shutterservice

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jackc/pgx/v4"
)

// FilterCriteria describes the set of contract addresses and event topic[0]
// hashes that an EventProcessor wants logs for. The MultiEventSyncer unions
// these across all processors and issues a single FilterLogs query per range.
//
// An empty Topics0 means the processor wants every log emitted by its
// Addresses, regardless of topic. The union query then drops topic filtering
// entirely. Avoid this unless strictly necessary.
type FilterCriteria struct {
	Addresses []common.Address
	Topics0   []common.Hash
}

// EventProcessor defines the interface that event processors for MultiEventSyncer must implement.
type EventProcessor interface {
	// GetProcessorName returns a unique name for this processor.
	GetProcessorName() string
	// FilterCriteria returns the addresses and topic[0] hashes this processor
	// wants logs for over the given block range. May be recomputed per range
	// since the criteria can depend on state (e.g. active triggers from the db).
	FilterCriteria(ctx context.Context, start, end uint64) (FilterCriteria, error)
	// ParseEvents inspects the pre-filtered logs and turns them into Events.
	// The provided logs are guaranteed to match this processor's FilterCriteria
	// from the same range; processors do not need to re-check (address,
	// topic[0]) but may apply additional predicates.
	ParseEvents(ctx context.Context, start, end uint64, logs []types.Log) ([]Event, error)
	// ProcessEvents processes the parsed events and stores them in the database.
	ProcessEvents(ctx context.Context, tx pgx.Tx, events []Event) error
	// RollbackEvents removes events with block numbers greater than the specified block number.
	RollbackEvents(ctx context.Context, tx pgx.Tx, toBlock int64) error
}

// Event represents a generic blockchain event that can be processed.
type Event interface{}

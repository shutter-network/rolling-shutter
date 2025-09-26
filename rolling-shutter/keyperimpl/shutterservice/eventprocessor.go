package shutterservice

import (
	"context"

	"github.com/jackc/pgx/v4"
)

// EventProcessor defines the interface that event processors for MultiEventSyncer must implement.
type EventProcessor interface {
	// GetProcessorName returns a unique name for this processor.
	GetProcessorName() string
	// FetchEvents retrieves events in the given block range (inclusive).
	FetchEvents(ctx context.Context, start, end uint64) ([]Event, error)
	// ProcessEvents processes the fetched events and stores them in the database.
	ProcessEvents(ctx context.Context, tx pgx.Tx, events []Event) error
	// RollbackEvents removes events with block numbers greater than the specified block number.
	RollbackEvents(ctx context.Context, tx pgx.Tx, toBlock int64) error
}

// Event represents a generic blockchain event that can be processed.
type Event interface{}

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package chainobsdb

import (
	"context"
)

type Querier interface {
	GetEventSyncProgress(ctx context.Context) (GetEventSyncProgressRow, error)
	GetNextBlockNumber(ctx context.Context) (int32, error)
	UpdateEventSyncProgress(ctx context.Context, arg UpdateEventSyncProgressParams) error
}

var _ Querier = (*Queries)(nil)
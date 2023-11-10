package db

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type (
	TxFunc            func(pgx.Tx) error
	TxFuncWithContext func(context.Context, pgx.Tx) error
)

// WrapContext is an adapter for using functions with a signature
// `func(context.Context, pgx.Tx) error` and adapt them for functions
// with a call signature of `func(pgx.Tx) error`.
// The `ctx` will be passed to the wrapped source function as a closure.
func WrapContext(ctx context.Context, f TxFuncWithContext) (context.Context, TxFunc) {
	wrapped := func(tx pgx.Tx) error {
		return f(ctx, tx)
	}
	return ctx, wrapped
}

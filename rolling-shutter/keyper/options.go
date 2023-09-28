package keyper

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Option func(*options) error

type options struct {
	dbpool             *pgxpool.Pool
	broadcastEonPubKey bool
	eonPubkeyHandler   EonPublicKeyHandlerFunc
}

func newDefaultOptions() *options {
	ops := &options{
		dbpool:             nil,
		broadcastEonPubKey: true,
		eonPubkeyHandler:   nil,
	}
	return ops
}

func validateOptions(o *options) error {
	if !o.broadcastEonPubKey && o.eonPubkeyHandler == nil {
		// TODO error message
		return errors.New("neither broadcasting EonPublicKey nor handler registerred")
	}
	return nil
}

func NoBroadcastEonPublicKey() Option {
	return func(o *options) error {
		o.broadcastEonPubKey = false
		return nil
	}
}

type EonPublicKeyHandlerFunc func(context.Context, EonPublicKey) error

func EonPublicKeyHandler(handler EonPublicKeyHandlerFunc) Option {
	return func(o *options) error {
		o.eonPubkeyHandler = handler
		return nil
	}
}

func DBPool(pool *pgxpool.Pool) Option {
	return func(o *options) error {
		o.dbpool = pool
		return nil
	}
}

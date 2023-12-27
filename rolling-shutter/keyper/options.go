package keyper

import (
	"context"
	"errors"
	"reflect"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

type Option func(*options) error

type options struct {
	dbpool             *pgxpool.Pool
	broadcastEonPubKey bool
	messaging          p2p.Messaging
	blockSyncClient    *ethclient.Client
	messageHandler     []p2p.MessageHandler
	eonPubkeyHandler   EonPublicKeyHandlerFunc
}

func newDefaultOptions() *options {
	return &options{
		dbpool:             nil,
		broadcastEonPubKey: true,
		blockSyncClient:    nil,
		messageHandler:     []p2p.MessageHandler{},
		eonPubkeyHandler:   nil,
	}
}

var keyperNewConfigType = reflect.TypeOf(contract.KeypersConfigsListNewConfig{})

func validateOptions(o *options) error {
	if !o.broadcastEonPubKey && o.eonPubkeyHandler == nil {
		return errors.New("no eon public key broadcast nor handler function provided. " +
			"newly negotiated eon public-keys would not be forwarded")
	}
	return nil
}

// NoBroadcastEonPublicKey deactivates the broadcasting of
// the keyper's newly negotiated DKG public-keys via the P2P network.
// If this option is given, an EonPublicKeyHandlerFunc MUST be
// provided via the WithEonPublicKeyHandler option.
func NoBroadcastEonPublicKey() Option {
	return func(o *options) error {
		o.broadcastEonPubKey = false
		return nil
	}
}

type EonPublicKeyHandlerFunc func(context.Context, EonPublicKey) error

// WithEonPublicKeyHandler registers a handler function that will
// be called whenever the keyper newly negotiated a DKG public key.
// If the NoBroadcastEonPublicKey() option is given, an
// EonPublicKeyHandlerFunc MUST be provided.
func WithEonPublicKeyHandler(handler EonPublicKeyHandlerFunc) Option {
	return func(o *options) error {
		o.eonPubkeyHandler = handler
		return nil
	}
}

// WithDBPool hands the pgxpool.Pool instance to the keyper
// that will we used for database connections.
// If this option is not given, a new pool will be instantiated.
func WithDBPool(pool *pgxpool.Pool) Option {
	return func(o *options) error {
		o.dbpool = pool
		return nil
	}
}

// WithMessageHandler adds additional P2P message handler implementations
// to the keypers. Multiple handlers for the same message type can be
// registered.
func WithMessageHandler(h p2p.MessageHandler) Option {
	return func(o *options) error {
		o.messageHandler = append(o.messageHandler, h)
		return nil
	}
}

// WithBlockSyncClient passes an Ethereum JSON-RPC client
// to the keyper. This client will be used to sync activation
// block numbers of e.g. the keyper set changes.
func WithBlockSyncClient(client *ethclient.Client) Option {
	return func(o *options) error {
		o.blockSyncClient = client
		return nil
	}
}

// WithMessaging passes the P2P messaging implementation
// to the keyper. It handles topic subscription and broadcasting
// of messages.
func WithMessaging(sender p2p.Messaging) Option {
	return func(o *options) error {
		o.messaging = sender
		return nil
	}
}

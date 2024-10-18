package keyper

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

type Option func(*options) error

type options struct {
	dbpool               *pgxpool.Pool
	broadcastEonPubKey   bool
	messaging            p2p.Messaging
	syncStartBlockNumber *big.Int
	blockSyncClient      *ethclient.Client
	messageHandler       []p2p.MessageHandler
	eonPubkeyHandler     EonPublicKeyHandlerFunc
	ethereumAddress      common.Address
	chainHandler         []syncer.ChainUpdateHandler
	eventHandler         []syncer.ContractEventHandler
}

func newDefaultOptions() *options {
	return &options{
		broadcastEonPubKey: true,
		messageHandler:     []p2p.MessageHandler{},
		chainHandler:       []syncer.ChainUpdateHandler{},
		eventHandler:       []syncer.ContractEventHandler{},
	}
}

func validateOptions(o *options) error {
	if !o.broadcastEonPubKey && o.eonPubkeyHandler == nil {
		return errors.New("no eon public key broadcast nor handler function provided. " +
			"newly negotiated eon public-keys would not be forwarded")
	}
	// TODO: check for non-nil contract addresses
	// TODO: ethereum address required
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

// TODO: docs.
func WithContractEventHandler(h syncer.ContractEventHandler) Option {
	return func(o *options) error {
		o.eventHandler = append(o.eventHandler, h)
		return nil
	}
}

// TODO: docs.
func WithChainUpdateHandler(h syncer.ChainUpdateHandler) Option {
	return func(o *options) error {
		o.chainHandler = append(o.chainHandler, h)
		return nil
	}
}

// TODO: docs.
func WithEthereumAddress(address common.Address) Option {
	return func(o *options) error {
		o.ethereumAddress = address
		return nil
	}
}

// TODO: use this e.g. with e.g. the Gnosis keyper and the gnosis config value "SyncStartBlockNumber".
func WithSyncStartBlockNumber(num big.Int) Option {
	return func(o *options) error {
		o.syncStartBlockNumber = &num
		return nil
	}
}

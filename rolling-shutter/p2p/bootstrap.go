package p2p

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
)

var (
	errInsufficientBootstrpConfigured error = errors.New("not enough bootstrap peers configured")
	errBootsrpsUnreachable            error = errors.New("could not connect to any bootstrap node")
)

func connectBootstrapNodes(ctx context.Context, h host.Host, peers []peer.AddrInfo) error {
	var (
		connectedNodes atomic.Uint32
		waitGroup      sync.WaitGroup
	)

	if len(peers) < 1 {
		return errInsufficientBootstrpConfigured
	}

	for _, pr := range peers {
		waitGroup.Add(1)

		// Add the bootstrap nodes to the peerstore as
		// permanent addr, because it is not expected to change
		h.Peerstore().AddAddrs(pr.ID, pr.Addrs, peerstore.PermanentAddrTTL)
		go func(ctx context.Context, a peer.AddrInfo) {
			defer waitGroup.Done()
			if err := h.Connect(ctx, a); err != nil {
				log.Debug().
					Err(err).
					Str("peer", a.String()).
					Msg("couldn't connect to boostrap node")
				return
			}
			connectedNodes.Add(1)
		}(ctx, pr)
	}

	waitGroup.Wait()
	if connectedNodes.Load() == 0 {
		return errBootsrpsUnreachable
	}
	return nil
}

func bootstrap(
	ctx context.Context,
	h host.Host,
	config p2pNodeConfig,
	hashTables ...*dht.IpfsDHT,
) error {
	for _, d := range hashTables {
		if d != nil {
			if err := d.Bootstrap(ctx); err != nil {
				return err
			}
		}
	}

	f := func(c context.Context) (bool, error) {
		if err := connectBootstrapNodes(c, h, config.BootstrapPeers); err != nil {
			return false, err
		}
		return true, nil
	}

	if config.IsBootstrapNode {
		if len(config.BootstrapPeers) > 0 {
			// A bootstrap node is not required to connect to other bootstrap nodes.
			// If however we did configure a list of bootstrap nodes,
			// we should try a long time to connect to at least one other bootstrapper first.
			backoffMult := float64(1.01)
			_, err := retry.FunctionCall(
				ctx,
				f,
				retry.MaxInterval(1*time.Minute),
				retry.StopOnErrors(errInsufficientBootstrpConfigured),
				retry.Interval(2*time.Second),
				retry.ExponentialBackoff(&backoffMult))
			if err != nil {
				log.Error().Err(err).
					Msg("failed to bootstrap, continuing without peer connections.")
			}
		}
	} else {
		_, err := retry.FunctionCall(
			ctx,
			f,
			retry.NumberOfRetries(5),
			retry.StopOnErrors(errInsufficientBootstrpConfigured),
			retry.Interval(2*time.Second))
		if err != nil {
			// For normal peers, after trying some time it is reasonable to halt.
			// If we don't get an initial connection to a bootsrap node,
			// we wil'l never participate.
			return err
		}
	}

	return nil
}

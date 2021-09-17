// Package keyper contains the keyper implementation
package keyper

import (
	"context"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"

	"github.com/shutter-network/shutter/shuttermint/p2p"
)

var gossipTopicNames = [2]string{"decryptionTrigger", "decryptionKey"}

func InitP2p(ctx context.Context, listenAddress multiaddr.Multiaddr, peerMultiaddrs []multiaddr.Multiaddr, p2pkey crypto.PrivKey) error {
	p2pConfig := p2p.Config{
		ListenAddr:     listenAddress,
		PeerMultiaddrs: peerMultiaddrs,
		PrivKey:        p2pkey,
	}
	p := p2p.NewP2P(p2pConfig)
	// FIXME: manage goroutine
	go p.Run(ctx, gossipTopicNames[:]) // nolint
	return nil
}

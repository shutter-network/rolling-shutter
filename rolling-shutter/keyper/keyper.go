// Package keyper contains the keyper implementation
package keyper

import (
	"context"

	"github.com/multiformats/go-multiaddr"

	"github.com/shutter-network/shutter/shuttermint/p2p"
)

var gossipTopicNames = [2]string{"decryptionTrigger", "decryptionKey"}

func InitP2p(ctx context.Context, listenAddress multiaddr.Multiaddr, peerMultiaddrs []multiaddr.Multiaddr) error {
	p := p2p.NewP2p()
	if err := p.CreateHost(ctx, listenAddress); err != nil {
		return err
	}
	if err := p.JoinTopics(ctx, gossipTopicNames[:]); err != nil {
		return err
	}
	if err := p.ConnectToPeers(ctx, peerMultiaddrs); err != nil {
		return err
	}
	return nil
}

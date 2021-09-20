package p2p

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	peerCheckInterval = 5 * time.Second
	minPeers          = 3
)

func (p *P2P) managePeers(ctx context.Context) error {
	if err := p.connectToConfiguredPeers(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-time.After(peerCheckInterval):
			n := len(p.host.Network().Peers())
			log.Printf("connected to %d peers, want at least %d", n, minPeers)
			if n < minPeers {
				if err := p.connectToConfiguredPeers(ctx); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *P2P) connectToConfiguredPeers(ctx context.Context) error {
	candidates := make(map[peer.ID]*peer.AddrInfo)

	// fill candidates from config file, disregarding those without a peer id (as we can't check
	// if we're already connected to them)
	for _, m := range p.Config.PeerMultiaddrs {
		addrInfo, err := peer.AddrInfoFromP2pAddr(m)
		if err != nil {
			log.Printf("ignoring invalid address from config %s: %s", m, err)
			continue
		}
		if len(addrInfo.Addrs) == 0 {
			log.Println("ignoring address from config without transport", m)
			continue
		}
		candidates[addrInfo.ID] = addrInfo
	}

	// remove candidates that we're already connected to
	for _, pid := range p.host.Network().Peers() {
		delete(candidates, pid)
	}

	if len(candidates) == 0 {
		log.Println("no peers to connect to")
	}

	// try to connect to all remaining candidates
	for _, addrInfo := range candidates {
		err := p.host.Connect(ctx, *addrInfo)
		if err != nil {
			log.Printf("error connecting to %s: %s", addrInfo, err)
		}
	}

	return nil
}

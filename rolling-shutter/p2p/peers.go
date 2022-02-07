package p2p

import (
	"context"
	"log"
	"math/rand"
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

	indices := []int{}
	for i := range p.Config.PeerMultiaddrs {
		indices = append(indices, i)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(indices), func(i, j int) { indices[i], indices[j] = indices[j], indices[i] })

	n := 20
	if n > len(indices) {
		n = len(indices)
	}

	// fill candidates from config file, disregarding those without a peer id (as we can't check
	// if we're already connected to them)
	for _, i := range indices[:n] {
		m := p.Config.PeerMultiaddrs[i]
		addrInfo, err := peer.AddrInfoFromP2pAddr(m)
		if err != nil {
			log.Printf("ignoring invalid address from config %s: %s", m, err)
			continue
		}
		if len(addrInfo.Addrs) == 0 {
			log.Println("ignoring address from config without transport", m)
			continue
		}
		// don't connect to yourself
		if addrInfo.ID == p.host.ID() {
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
			if len(addrInfo.Addrs) > 0 {
				log.Printf("error connecting to %s at %s: %s", addrInfo.ID, addrInfo.Addrs[0], err)
			} else {
				log.Printf("error connecting to %s without known multiaddr: %s", addrInfo.ID, err)
			}
		}
	}

	return nil
}

package p2p

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
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
			if n != p.logNumberOfPeers {
				log.Info().Int("want", minPeers).Int("have", n).Int("previous", p.logNumberOfPeers).Msg("number of peer connections")
				p.logNumberOfPeers = n
			}

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
			log.Info().Err(err).Str("address", m.String()).Str("origin", "config").Msg("ignoring invalid address")
			continue
		}
		if len(addrInfo.Addrs) == 0 {
			log.Info().Str("address", m.String()).Str("origin", "config").
				Msg("ignoring address without transport")
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
		log.Info().Msg("no peers to connect to")
	}

	// try to connect to all remaining candidates
	for _, addrInfo := range candidates {
		err := p.host.Connect(ctx, *addrInfo)
		if err != nil {
			if len(addrInfo.Addrs) > 0 {
				log.Info().Err(err).Str("peer-id", addrInfo.ID.String()).
					Str("peer-address", addrInfo.Addrs[0].String()).Msg("connection error")
			} else {
				log.Info().Err(err).Str("peer-id", addrInfo.ID.String()).
					Str("peer-address", "").Msg("connection error")
			}
		}
	}

	return nil
}

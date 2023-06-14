package address

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable"
)

func MustAddrInfo(s string) peer.AddrInfo {
	a, err := peer.AddrInfoFromString(s)
	if err != nil {
		panic(err)
	}
	return *a
}

type P2PIdentifier struct {
	peer.ID
}

func (a *P2PIdentifier) UnmarshalText(b []byte) error {
	id, err := peer.Decode(string(b))
	if err != nil {
		return err
	}
	a.ID = id
	return nil
}

func (a P2PIdentifier) MarshalText() ([]byte, error) {
	return []byte(a.ID.String()), nil
}

func (a P2PIdentifier) String() string {
	return encodeable.String(a)
}

func (a *P2PIdentifier) Equal(b *P2PIdentifier) bool {
	return a.ID == b.ID
}

func MustP2PAddress(s string) *P2PAddress {
	p := &P2PAddress{}
	err := p.UnmarshalText([]byte(s))
	if err != nil {
		panic(err)
	}
	return p
}

type P2PAddress struct {
	multiaddr.Multiaddr
}

func (a *P2PAddress) UnmarshalText(b []byte) error {
	na, err := multiaddr.NewMultiaddr(string(b))
	if err != nil {
		return err
	}
	a.Multiaddr = na
	return nil
}

func (a P2PAddress) MarshalText() ([]byte, error) {
	return []byte(a.Multiaddr.String()), nil
}

func (a *P2PAddress) String() string {
	return encodeable.String(a)
}

func (a P2PAddress) Identifier() (*P2PIdentifier, error) {
	pid := &P2PIdentifier{}
	ai, err := peer.AddrInfoFromP2pAddr(a.Multiaddr)
	if err != nil {
		return nil, err
	}
	pid.ID = ai.ID
	return pid, nil
}

func P2PAddressesToAddrInfos(addrs []P2PAddress) ([]peer.AddrInfo, error) {
	multiAddrs := P2PAddressesToMultiaddrs(addrs)
	return peer.AddrInfosFromP2pAddrs(multiAddrs...)
}

func P2PAddressesToMultiaddrs(addrs []P2PAddress) []multiaddr.Multiaddr {
	multiAddrs := make([]multiaddr.Multiaddr, len(addrs))
	for i, a := range addrs {
		multiAddrs[i] = a.Multiaddr
	}
	return multiAddrs
}

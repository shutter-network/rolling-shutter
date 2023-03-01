package p2p

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// MustMultiaddr converts the given string to a multiaddr.Multiaddr.
func MustMultiaddr(s string) multiaddr.Multiaddr {
	a, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		panic(err)
	}
	return a
}

func MustAddrInfo(s string) peer.AddrInfo {
	a, err := peer.AddrInfoFromString(s)
	if err != nil {
		panic(err)
	}
	return *a
}

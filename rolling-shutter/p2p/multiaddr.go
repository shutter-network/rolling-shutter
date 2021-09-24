package p2p

import "github.com/multiformats/go-multiaddr"

// MustMultiaddr converts the given string to a multiaddr.Multiaddr.
func MustMultiaddr(s string) multiaddr.Multiaddr {
	a, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		panic(err)
	}
	return a
}

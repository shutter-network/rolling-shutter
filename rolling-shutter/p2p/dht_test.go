package p2p

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"gotest.tools/v3/assert"
)

func TestRandomisePeers(t *testing.T) {
	peerAddr := make([]peer.AddrInfo, 0)

	for i := 0; i < 5; i++ {
		peerAddr = append(peerAddr, peer.AddrInfo{
			ID: peer.ID(fmt.Sprintf("%d", i)),
		})
	}

	randomisedPeers := randomizePeers(peerAddr)

	equal := reflect.DeepEqual(peerAddr, randomisedPeers)
	assert.Assert(t, !equal, "randomized unsuccessful")
}

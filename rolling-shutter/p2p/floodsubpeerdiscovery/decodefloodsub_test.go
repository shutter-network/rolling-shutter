package floodsubpeerdiscovery

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"
)

func TestDecodeConst(t *testing.T) {
	raw := "CiD53zJL+X7bj6vatGWpsYWmCBSsCTBi0jqBsHyc9yeM3xIIBKdjseMGWd0SCwSnY7HjkQJZ3c0D"
	msg, err := base64.StdEncoding.DecodeString(raw)
	assert.NilError(t, err, "could not b64 decode")
	p := Peer{}
	proto.Unmarshal(msg, &p)
	for _, addr := range p.Addrs {
		ma, err := multiaddr.NewMultiaddrBytes(addr)
		assert.NilError(t, err, "could not decode multiaddr")
		fmt.Println(ma.String())
	}
	assert.Check(t, len(p.PublicKey) > 0, "no pubkey decoded")
	assert.Check(t, len(p.Addrs) > 0, "no addresses decoded")
}

func TestEncode(t *testing.T) {
	input := Peer{}
	input.PublicKey = []byte("abcdef")
	input.Addrs = append(input.Addrs, []byte("multiaddr"))
	x, err := proto.Marshal(&input)
	assert.NilError(t, err, "couldn't marshal")
	fmt.Println(string(x))
	p := Peer{}
	proto.Unmarshal(x, &p)
	fmt.Println(p.Addrs, p.PublicKey)
	assert.Check(t, len(p.PublicKey) > 0, "no pubkey decoded")
	assert.Check(t, len(p.Addrs) > 0, "no addresses decoded")
	assert.Check(t, len(p.Addrs[0]) > 0, "no address decoded")
	for i := range input.PublicKey {
		assert.Check(t, input.PublicKey[i] == p.PublicKey[i], "pubkey doesn't match")
	}
}

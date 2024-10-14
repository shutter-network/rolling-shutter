package floodsubpeerdiscovery

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"
)

func (p *Peer) Print() error {
	for _, addr := range p.Addrs {
		ma, err := multiaddr.NewMultiaddrBytes(addr)
		if err != nil {
			return fmt.Errorf("could not parse multiaddr %v: %e", string(addr), err)
		}
		fmt.Printf("%v: %v\n", base64.RawStdEncoding.EncodeToString(p.PublicKey), ma.String())
	}
	return nil
}

func TestDecodeConst(t *testing.T) {
	raw := "CiD53zJL+X7bj6vatGWpsYWmCBSsCTBi0jqBsHyc9yeM3xIIBKdjseMGWd0SCwSnY7HjkQJZ3c0D"
	p, _, err := decodeFromB64(raw)
	assert.NilError(t, err, "could not decode")
	assert.Check(t, len(p.PublicKey) > 0, "no pubkey decoded")
	assert.Check(t, len(p.Addrs) > 0, "no addresses decoded")
	err = p.Print()
	assert.NilError(t, err, "could not print")
}

// extract msg data from nethermind logfile download (as $LOGFILE) via
//
// grep _peer-disc $LOGFILE|grep -v ihave|cut -b 119-|jq  -r '.publish.[].data' > /tmp/allmsgs.txt
//
// before running this test
func TestDecodeMany(t *testing.T) {
	file, err := os.Open("/tmp/allmsgs.txt")
	defer file.Close()
	assert.NilError(t, err, "could not open file")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		p, raw, err := decodeFromB64(scanner.Text())
		assert.NilError(t, err, "could not decode")
		err = p.Print()
		if err != nil {
			fmt.Printf("could not parse %v\n", raw)
		}
	}
}

func decodeFromB64(raw string) (Peer, string, error) {
	p := Peer{}
	msg, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return p, raw, err
	}
	proto.Unmarshal(msg, &p)
	return p, raw, nil
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

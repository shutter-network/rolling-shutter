package floodsubpeerdiscovery

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"
)

type DebugTranscoder multiaddr.Transcoder

var DebugProtocol1 = multiaddr.Protocol{
	Name:       "P_DEBUG",
	Code:       420,
	VCode:      []byte{},
	Size:       0,
	Path:       false,
	Transcoder: multiaddr.TranscoderIP4,
}

func (p *Peer) Print() error {
	knownIps := KnownIps()
	for _, addr := range p.Addrs {
		multiaddr.AddProtocol(DebugProtocol1)
		ma, err := multiaddr.NewMultiaddrBytes(addr)
		if err != nil {
			return fmt.Errorf("could not parse multiaddr %v: %e", string(addr), err)
		}
		ip, err := manet.ToIP(ma)
		if err != nil {
			return fmt.Errorf("could not extract ip %v: %e", string(addr), err)
		}
		name := knownIps["ip-to-name"][ip.String()]
		id := knownIps["name-to-id"][name]
		fmt.Println(name, id, base64.RawStdEncoding.EncodeToString(p.PublicKey), ma.String())
	}
	return nil
}

func KnownIps() map[string]map[string]string {
	file, err := os.Open("known_ips.json")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	dec := json.NewDecoder(file)
	var result map[string]map[string]string
	dec.Decode(&result)
	return result
}

func TestParse(t *testing.T) {
	KnownIps()
}

func TestDecodeConst(t *testing.T) {
	raw := "CiD53zJL+X7bj6vatGWpsYWmCBSsCTBi0jqBsHyc9yeM3xIIBKdjseMGWd0SCwSnY7HjkQJZ3c0D"
	p, _, _, err := decodeFromB64(raw)
	assert.NilError(t, err, "could not decode")
	assert.Check(t, len(p.PublicKey) > 0, "no pubkey decoded")
	assert.Check(t, len(p.Addrs) > 0, "no addresses decoded")
	err = p.Print()
	assert.NilError(t, err, "could not print")
}

// extract msg data from nethermind logfile download (as $LOGFILE) via
//
// grep _peer-disc $LOGFILE|cut -b 119-|jq -r 'select(.publish != null)|.publish.[] | select (.topic=="_peer-discovery._p2p._pubsub")|.data' > /tmp/allmsgs.txt
//
// before running this test
func TestDecodeMany(t *testing.T) {
	file, err := os.Open("/tmp/allmsgs.txt")
	defer file.Close()
	assert.NilError(t, err, "could not open file")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		p, raw, msg, err := decodeFromB64(scanner.Text())
		assert.NilError(t, err, "could not decode")
		err = p.Print()
		if err != nil {
			assert.Check(t, raw != "", "raw empty")
			assert.Check(t, msg != nil, "msg empty")
			fmt.Printf("could not parse %v: %v : %v : %v\n", raw, msg, p, err)
		}
	}
}

func decodeFromB64(raw string) (Peer, string, []byte, error) {
	p := Peer{}
	msg, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return p, raw, msg, err
	}
	proto.Unmarshal(msg, &p)
	return p, raw, msg, nil
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

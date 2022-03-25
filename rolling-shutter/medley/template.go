package medley

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/ethereum/go-ethereum/crypto"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"

	"github.com/shutter-network/shutter/shlib/shcrypto"
)

func p2pKeyPublic(privkey p2pcrypto.PrivKey) string {
	id, _ := peer.IDFromPublicKey(privkey.GetPublic())
	return id.Pretty()
}

func p2pKey(privkey p2pcrypto.PrivKey) string {
	d, _ := p2pcrypto.MarshalPrivateKey(privkey)
	return p2pcrypto.ConfigEncodeKey(d)
}

func eonPublicKey(pubkey *shcrypto.EonPublicKey) string {
	b := pubkey.Marshal()
	return hex.EncodeToString(b)
}

func quoteList(lst []multiaddr.Multiaddr) string {
	var strlist []string
	for _, x := range lst {
		// We use json.Marshal here, not sure if it's the right thing to do, since we're
		// writing TOML
		d, _ := json.Marshal(x.String())
		strlist = append(strlist, string(d))
	}

	return strings.Join(strlist, ", ")
}

func MustBuildTemplate(name, content string) *template.Template {
	t, err := template.New(name).Funcs(template.FuncMap{
		"FromECDSA":    crypto.FromECDSA,
		"QuoteList":    quoteList,
		"P2PKey":       p2pKey,
		"P2PKeyPublic": p2pKeyPublic,
		"EonPublicKey": eonPublicKey,
	}).Parse(content)
	if err != nil {
		panic(err)
	}
	return t
}

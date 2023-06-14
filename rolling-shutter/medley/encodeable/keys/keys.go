package keys

import (
	"encoding"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type Algorithm int

const (
	Ed25519 Algorithm = iota
	ECDSASecp256k1
	LibP2P
)

type key interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	fmt.Stringer
	Bytes() []byte
	Type() Algorithm
}

type Private interface {
	key
	Public() Public
	Sign([]byte) ([]byte, error)
}

type Public interface {
	key
	Verify([]byte, []byte) (bool, error)
}

var ErrUnknownKeyAlgorithm = errors.New("key algorithm not supported")

type RepeatedByteReader struct {
	Value uint8
}

func (c RepeatedByteReader) Read(p []byte) (n int, err error) { //nolint:unparam
	for i := range p {
		p[i] = c.Value
	}
	return len(p), nil
}

// GenerateKeyPairWithReader returns a keypair of the given type and bitsize.
func GenerateNew(algorithm Algorithm, src io.Reader) (Private, error) {
	switch algorithm {
	case Ed25519:
		return GenerateEd25519Key(src)
	case ECDSASecp256k1:
		return GenerateECDSAKey(src)
	case LibP2P:
		return GenerateLibp2pPrivate(src)
	default:
		return nil, ErrUnknownKeyAlgorithm
	}
}

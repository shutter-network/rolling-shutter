package keys

import (
	"crypto/ed25519"
	"io"

	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/hex"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/tee"
)

type nullReader struct{}

func (nullReader) Read(p []byte) (n int, err error) {
	return len(p), nil
}

var (
	nullReaderInst = nullReader{}
	privSignOpts   = &ed25519.Options{}
)

func GenerateEd25519Key(src io.Reader) (*Ed25519Private, error) {
	_, priv, err := ed25519.GenerateKey(src)
	if err != nil {
		return nil, err
	}
	return &Ed25519Private{Key: priv}, nil
}

var (
	errFailedInitialiseEd2551Private error = errors.New("failed to initialize Ed2551Private key")
	errFailedUnmarshalEd2551Private  error = errors.New("failed to unmarshal Ed2551Private key")
	errFailedMarshalEd2551Private    error = errors.New("failed to marshal Ed2551Private key")

	errFailedInitialiseEd2551Public error = errors.New("failed to initialize Ed2551Public key")
	errFailedUnmarshalEd2551Public  error = errors.New("failed to unmarshal Ed2551Public key")
	errFailedMarshalEd2551Public    error = errors.New("failed to marshal Ed2551Public key")
)

type Ed25519Private struct {
	// FIXME here we have the same problem with uninitialized keys as in the Libp2p-keys
	Key ed25519.PrivateKey
}

func (k *Ed25519Private) Public() Public {
	pub, ok := k.Key.Public().(ed25519.PublicKey)
	if !ok {
		panic(makeUpstreamAPIChangeMessage("crypto/ed25519's Public()"))
	}
	return &Ed25519Public{Key: pub}
}

func (k *Ed25519Private) Bytes() []byte {
	return k.Key
}

func (k *Ed25519Private) Sign(data []byte) ([]byte, error) {
	return k.Key.Sign(nullReaderInst, data, privSignOpts)
}

func (k *Ed25519Private) Type() Algorithm {
	return Ed25519
}

func (k *Ed25519Private) Equal(b *Ed25519Private) bool {
	return k.Key.Equal(b.Key)
}

func (k *Ed25519Private) UnmarshalText(b []byte) error {
	seed, err := tee.UnsealSecretFromHex(string(b))
	if err != nil {
		return err
	}
	if len(seed) != ed25519.SeedSize {
		return errors.Errorf(
			"invalid seed length %d (must be %d)",
			len(seed),
			ed25519.SeedSize,
		)
	}
	k.Key = ed25519.NewKeyFromSeed(seed)
	return nil
}

func (k *Ed25519Private) MarshalText() ([]byte, error) {
	str, err := tee.SealSecretAsHex(k.Key.Seed())
	return []byte(str), err
}

func (k *Ed25519Private) String() string {
	return encodeable.String(k)
}

type Ed25519Public struct {
	Key ed25519.PublicKey
}

func (k *Ed25519Public) Type() Algorithm {
	return Ed25519
}

func (k *Ed25519Public) Bytes() []byte {
	return k.Key
}

func (k *Ed25519Public) Equal(b *Ed25519Public) bool {
	return k.Key.Equal(b.Key)
}

func (k *Ed25519Public) Verify(data []byte, signature []byte) (bool, error) {
	return ed25519.Verify(k.Key, data, signature), nil
}

func (k *Ed25519Public) UnmarshalText(b []byte) error {
	p, err := hex.DecodeHex(b)
	if err != nil {
		return err
	}
	if len(p) != ed25519.PublicKeySize {
		return errors.New("badly formed ed25519 public key")
	}
	k.Key = ed25519.PublicKey(p)
	return nil
}

func (k *Ed25519Public) MarshalText() ([]byte, error) {
	return hex.EncodeHex(k.Key), nil
}

func (k *Ed25519Public) String() string {
	return encodeable.String(k)
}

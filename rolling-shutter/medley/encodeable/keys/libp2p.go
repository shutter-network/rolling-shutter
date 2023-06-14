package keys

import (
	"crypto/ed25519"
	"io"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/address"
)

func GenerateLibp2pPrivate(src io.Reader) (*Libp2pPrivate, error) {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := io.ReadFull(src, seed); err != nil {
		return nil, err
	}
	return NewLibp2pPrivate(ed25519.NewKeyFromSeed(seed))
}

var (
	errFailedInitialiseLibp2pPrivate error = errors.New("failed to initialize Libp2pPrivate key")
	errFailedUnmarshalLibp2pPrivate  error = errors.New("failed to unmarshal Libp2pPrivate key")
	errFailedMarshalLibp2pPrivate    error = errors.New("failed to marshal Libp2pPrivate key")

	errFailedInitialiseLibp2pPublic error = errors.New("failed to initialize Libp2pPublic key")
	errFailedUnmarshalLibp2pPublic  error = errors.New("failed to unmarshal Libp2pPublic key")
	errFailedMarshalLibp2pPublic    error = errors.New("failed to marshal Libp2pPublic key")
)

func NewLibp2pPrivate(b []byte) (*Libp2pPrivate, error) {
	// We have to do all of this only because libp2p did decide to make the
	// underlying key member private, not provide a simple constructor and
	// only return interfaces everywhere ...

	if len(b) != ed25519.PrivateKeySize {
		return nil, errors.Wrap(
			errFailedInitialiseLibp2pPrivate,
			"input bytes size does not match expected size",
		)
	}
	priv, err := crypto.UnmarshalEd25519PrivateKey(b)
	if err != nil {
		panic(errors.Wrap(errFailedUnmarshalLibp2pPrivate, err.Error()))
	}
	privEd, ok := priv.(*crypto.Ed25519PrivateKey)
	if !ok {
		panic(errors.Wrap(
			errFailedUnmarshalLibp2pPrivate,
			"libp2p's UnmarshalEd25519PrivateKey() did return unexpected type, "+
				"indicating an upstream API change",
		))
	}
	return &Libp2pPrivate{
		Key: *privEd,
	}, nil
}

type Libp2pPrivate struct {
	Key crypto.Ed25519PrivateKey
}

func (k *Libp2pPrivate) Public() Public {
	pub := k.Key.GetPublic()
	edPubK, ok := pub.(*crypto.Ed25519PublicKey)
	if !ok {
		panic(errors.New(
			"libp2p's GetPublic() did return unexpected type, " +
				"indicating an upstream API change",
		))
	}
	return &Libp2pPublic{Key: *edPubK}
}

func (k *Libp2pPrivate) Bytes() []byte {
	b, err := k.Key.Raw()
	if err != nil {
		return []byte{}
	}
	return b
}

func (k *Libp2pPrivate) PeerID() (*address.P2PIdentifier, error) {
	pid, err := peer.IDFromPrivateKey(&k.Key)
	if err != nil {
		return nil, err
	}
	return &address.P2PIdentifier{ID: pid}, nil
}

func (k *Libp2pPrivate) Sign(data []byte) ([]byte, error) {
	return k.Key.Sign(data)
}

func (k *Libp2pPrivate) Type() Algorithm {
	return LibP2P
}

func (k *Libp2pPrivate) Equal(b *Libp2pPrivate) bool {
	return crypto.KeyEqual(&k.Key, &b.Key)
}

func (k *Libp2pPrivate) UnmarshalText(b []byte) error {
	dec, err := crypto.ConfigDecodeKey(string(b))
	if err != nil {
		return errors.Wrap(errFailedUnmarshalLibp2pPrivate, err.Error())
	}
	privkey, err := crypto.UnmarshalPrivateKey(dec)
	if err != nil {
		return errors.Wrap(errFailedUnmarshalLibp2pPrivate, err.Error())
	}
	edPk, ok := privkey.(*crypto.Ed25519PrivateKey)
	if !ok {
		panic(errors.Wrap(
			errFailedUnmarshalLibp2pPrivate,
			"libp2p's UnmarshalPrivateKey() did return unexpected type, "+
				"indicating an upstream API change",
		))
	}
	k.Key = *edPk
	return nil
}

func (k *Libp2pPrivate) MarshalText() ([]byte, error) {
	b, err := crypto.MarshalPrivateKey(&k.Key)
	if err != nil {
		return []byte{}, errors.Wrap(errFailedMarshalLibp2pPrivate, err.Error())
	}
	enc := crypto.ConfigEncodeKey(b)
	return []byte(enc), nil
}

func (k *Libp2pPrivate) String() string {
	return encodeable.String(k)
}

type Libp2pPublic struct {
	Key crypto.Ed25519PublicKey
}

func NewLibp2pPublic(b []byte) (*Libp2pPublic, error) {
	// We have to do all of this only because libp2p did decide to make the
	// underlying key member private, not provide a simple constructor and
	// only return interfaces everywhere ...

	if len(b) != ed25519.PublicKeySize {
		return nil, errors.Wrap(
			errFailedInitialiseLibp2pPublic,
			"input bytes size does not match expected size",
		)
	}
	pub, err := crypto.UnmarshalEd25519PublicKey(b)
	if err != nil {
		panic(errors.Wrap(errFailedUnmarshalLibp2pPublic, err.Error()))
	}
	pubEd, ok := pub.(*crypto.Ed25519PublicKey)
	if !ok {
		panic(errors.Wrap(
			errFailedUnmarshalLibp2pPrivate,
			"libp2p's UnmarshalEd25519PublicKey() did return unexpected type, "+
				"indicating an upstream API change",
		))
	}
	return &Libp2pPublic{
		Key: *pubEd,
	}, nil
}

func (k *Libp2pPublic) Type() Algorithm {
	return LibP2P
}

func (k *Libp2pPublic) Bytes() []byte {
	b, err := k.Key.Raw()
	if err != nil {
		return []byte{}
	}
	return b
}

func (k *Libp2pPublic) Equal(b *Libp2pPublic) bool {
	return crypto.KeyEqual(&k.Key, &b.Key)
}

func (k *Libp2pPublic) PeerID() (address.P2PIdentifier, error) {
	pid, err := peer.IDFromPublicKey(&k.Key)
	if err != nil {
		return address.P2PIdentifier{}, err
	}
	return address.P2PIdentifier{ID: pid}, nil
}

func (k *Libp2pPublic) Verify(data []byte, signature []byte) (bool, error) {
	return k.Key.Verify(data, signature)
}

func (k *Libp2pPublic) UnmarshalText(b []byte) error {
	dec, err := crypto.ConfigDecodeKey(string(b))
	if err != nil {
		return errors.Wrap(errFailedUnmarshalLibp2pPublic, err.Error())
	}
	pubkey, err := crypto.UnmarshalPublicKey(dec)
	if err != nil {
		return errors.Wrap(errFailedUnmarshalLibp2pPublic, err.Error())
	}

	edPubK, ok := pubkey.(*crypto.Ed25519PublicKey)
	if !ok {
		panic(errors.Wrap(
			errFailedUnmarshalLibp2pPublic,
			"libp2p's UnmarshalPublicKey() did return unexpected type, "+
				"indicating an upstream API change",
		))
	}
	k.Key = *edPubK
	return nil
}

func (k *Libp2pPublic) MarshalText() ([]byte, error) {
	b, err := crypto.MarshalPublicKey(&k.Key)
	if err != nil {
		return []byte{}, errors.Wrap(errFailedMarshalLibp2pPublic, err.Error())
	}
	enc := crypto.ConfigEncodeKey(b)
	return []byte(enc), nil
}

func (k *Libp2pPublic) String() string {
	return encodeable.String(k)
}

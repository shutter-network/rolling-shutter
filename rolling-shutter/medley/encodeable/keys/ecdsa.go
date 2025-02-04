package keys

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"io"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/hex"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/tee"
)

func GenerateECDSAKey(src io.Reader) (*ECDSAPrivate, error) {
	// Ethereum ECDSA uses the secp256k1 elliptic curve
	priv, err := ecdsa.GenerateKey(ethcrypto.S256(), src)
	if err != nil {
		return nil, err
	}
	return &ECDSAPrivate{Key: priv}, nil
}

type ECDSAPrivate struct {
	Key *ecdsa.PrivateKey
}

func (k *ECDSAPrivate) EthereumAddress() common.Address {
	if k.Key == nil {
		return common.Address{}
	}
	pub := k.Key.PublicKey
	return ethcrypto.PubkeyToAddress(pub)
}

func (k *ECDSAPrivate) Public() Public {
	if k.Key == nil {
		return nil
	}
	pub := k.Key.PublicKey
	return &ECDSAPublic{Key: &pub}
}

func (k *ECDSAPrivate) Bytes() []byte {
	return ethcrypto.FromECDSA(k.Key)
}

func (k *ECDSAPrivate) Sign(data []byte) ([]byte, error) {
	var noHash crypto.Hash
	return k.Key.Sign(rand.Reader, data, noHash)
}

func (k *ECDSAPrivate) Type() Algorithm {
	return ECDSASecp256k1
}

func (k *ECDSAPrivate) Equal(b *ECDSAPrivate) bool {
	return k.Key.Equal(b.Key)
}

func (k *ECDSAPrivate) UnmarshalText(b []byte) error {
	dec, err := tee.UnsealSecretFromHex(string(b))
	if err != nil {
		return err
	}
	key, err := ethcrypto.ToECDSA(dec)
	if err != nil {
		return err
	}
	k.Key = key
	return nil
}

func (k *ECDSAPrivate) MarshalText() ([]byte, error) {
	str, err := tee.SealSecretAsHex(k.Bytes())
	return []byte(str), err
}

func (k *ECDSAPrivate) String() string {
	return encodeable.String(k)
}

type ECDSAPublic struct {
	Key *ecdsa.PublicKey
}

func (k *ECDSAPublic) Type() Algorithm {
	return ECDSASecp256k1
}

func (k *ECDSAPublic) Bytes() []byte {
	return ethcrypto.FromECDSAPub(k.Key)
}

func (k *ECDSAPublic) Verify(data []byte, signature []byte) (bool, error) {
	return ecdsa.VerifyASN1(k.Key, data, signature), nil
}

func (k *ECDSAPublic) Equal(b *ECDSAPublic) bool {
	return k.Key.Equal(b.Key)
}

func (k *ECDSAPublic) EthereumAddress() common.Address {
	return ethcrypto.PubkeyToAddress(*k.Key)
}

func (k *ECDSAPublic) UnmarshalText(b []byte) error {
	dec, err := hex.DecodeHex(b)
	if err != nil {
		return err
	}
	key, err := ethcrypto.UnmarshalPubkey(dec)
	if err != nil {
		return err
	}
	k.Key = key
	return nil
}

func (k *ECDSAPublic) MarshalText() ([]byte, error) {
	return hex.EncodeHex(k.Bytes()), nil
}

func (k *ECDSAPublic) String() string {
	return encodeable.String(k)
}

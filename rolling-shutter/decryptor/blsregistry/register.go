// package blsregistry contains functions to interact with the bls registry contracts.
package blsregistry

import (
	"bytes"
	"context"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/contract"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/medley/txbatch"
)

var keyAndSignatureABIArguments abi.Arguments

func init() {
	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		panic("unexpected error creating representation of type bytes")
	}
	keyAndSignatureABIArguments = abi.Arguments{
		{Type: bytesType},
		{Type: bytesType},
	}
}

type addrsSeqPos struct {
	N uint64
	I uint64
}

func findCoordinates(
	ctx context.Context,
	addrsSeq *contract.AddrsSeq,
	addr common.Address) ([]addrsSeqPos, error) {
	var res []addrsSeqPos
	opts := &bind.CallOpts{Context: ctx}
	count, err := addrsSeq.Count(opts)
	if err != nil {
		return nil, err
	}
	for n := uint64(0); n < count; n++ {
		countNth, err := addrsSeq.CountNth(opts, n)
		if err != nil {
			return nil, err
		}
		for i := uint64(0); i < countNth; i++ {
			a, err := addrsSeq.At(opts, n, i)
			if err != nil {
				return nil, err
			}
			if addr == a {
				res = append(res, addrsSeqPos{N: n, I: i})
			}
		}
	}
	return res, nil
}

func encodeKeyAndSignature(key *shbls.PublicKey, signature *shbls.Signature) []byte {
	encodedKey := key.Marshal()
	encodedSignature := signature.Marshal()

	values := []interface{}{
		encodedKey,
		encodedSignature,
	}
	encodedKeyAndSignature, err := keyAndSignatureABIArguments.PackValues(values)
	if err != nil {
		panic(errors.Wrapf(err, "unexpected error packing key and signature"))
	}
	return encodedKeyAndSignature
}

func decodeKeyAndSignature(d []byte) (*shbls.PublicKey, *shbls.Signature, error) {
	values, err := keyAndSignatureABIArguments.UnpackValues(d)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "invalid encoding of key and signature")
	}
	if len(values) != 2 {
		panic(errors.Errorf("expected key and signature to be two values, got %d", len(values)))
	}
	encodedKey := values[0].([]byte)
	encodedSignature := values[1].([]byte)

	key := new(shbls.PublicKey)
	err = key.Unmarshal(encodedKey)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "invalid BLS public key %X", encodedKey)
	}

	signature := new(shbls.Signature)
	err = signature.Unmarshal(encodedSignature)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "invalid BLS signature %X", encodedKey)
	}

	return key, signature, nil
}

type RawIdentity struct {
	KeyAndSignature []byte
	EthereumAddress common.Address
}

func getRawIdentity(opts *bind.CallOpts, contracts *deployment.Contracts, addr common.Address) (RawIdentity, error) {
	keyAndSignature, err := contracts.BLSRegistry.Get(opts, addr)
	if err != nil {
		return RawIdentity{}, err
	}

	return RawIdentity{
		KeyAndSignature: keyAndSignature,
		EthereumAddress: addr,
	}, nil
}

func registerRawIdentity(
	batch *txbatch.TXBatch,
	contracts *deployment.Contracts,
	identity RawIdentity) error {
	from := batch.TransactOpts.From
	ctx := batch.TransactOpts.Context
	callOpts := &bind.CallOpts{Context: ctx}

	currentIdentity, err := getRawIdentity(callOpts, contracts, from)
	if err != nil {
		return err
	}
	register := len(currentIdentity.KeyAndSignature) == 0
	if !register && !bytes.Equal(currentIdentity.KeyAndSignature, identity.KeyAndSignature) {
		return errors.Errorf(
			"Identity already registered: wrong key and/or signature: have %x, expected %x",
			currentIdentity.KeyAndSignature, identity.KeyAndSignature,
		)
	}

	if !register {
		log.Printf("Identity already registered.")
		return nil
	}

	pos, err := findCoordinates(ctx, contracts.Decryptors, from)
	if err != nil {
		return err
	}
	if len(pos) == 0 {
		log.Printf("Address %s not registered as decryptor.", from.Hex())
		return nil
	}
	n, i := pos[0].N, pos[0].I

	tx, err := contracts.BLSRegistry.Register(batch.TransactOpts, n, i, identity.KeyAndSignature)
	if err != nil {
		return err
	}
	batch.Add(tx)
	return nil
}

// Register registers the public key corresponding to the given private key on chain.
func Register(
	batch *txbatch.TXBatch,
	contracts *deployment.Contracts,
	key *shbls.SecretKey,
) error {
	// We sign the sender's address with the BLS key. That makes sure that the sender has
	// access to the private key belonging to the registered key.
	msg := batch.TransactOpts.From.Bytes()
	signature := shbls.Sign(msg, key)
	keyAndSignature := encodeKeyAndSignature(shbls.SecretToPublicKey(key), signature)
	raw := RawIdentity{
		KeyAndSignature: keyAndSignature,
		EthereumAddress: batch.TransactOpts.From,
	}

	return registerRawIdentity(batch, contracts, raw)
}

// GetKeyAndSignature returns the BLS public key and the signature after verifying that the
// signature is correct.
func (raw *RawIdentity) GetKeyAndSignature() (*shbls.PublicKey, *shbls.Signature, error) {
	key, signature, err := decodeKeyAndSignature(raw.KeyAndSignature)
	if err != nil {
		return nil, nil, err
	}
	if !shbls.Verify(signature, key, raw.EthereumAddress.Bytes()) {
		return nil, nil, errors.Errorf("Cannot verify signature")
	}
	return key, signature, nil
}

// Lookup retrieves and verifies the BLS public key the decryptor with the given address has
// registered.
func Lookup(opts *bind.CallOpts, contracts *deployment.Contracts, addr common.Address) (*shbls.PublicKey, error) {
	raw, err := getRawIdentity(opts, contracts, addr)
	if err != nil {
		return nil, err
	}
	key, _, err := raw.GetKeyAndSignature()
	return key, err
}

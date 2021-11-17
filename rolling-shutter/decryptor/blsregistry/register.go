// package blsregistry contains functions to interact with the bls registry contracts.
package blsregistry

import (
	"bytes"
	"context"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/contract"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/medley/txbatch"
)

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

type rawIdentity struct {
	Key             []byte
	Signature       []byte
	EthereumAddress common.Address
}

func getRawIdentity(opts *bind.CallOpts, contracts *deployment.Contracts, addr common.Address) (rawIdentity, error) {
	key, err := contracts.BLSPublicKeyRegistry.Get(opts, addr)
	if err != nil {
		return rawIdentity{}, err
	}

	signature, err := contracts.BLSSignatureRegistry.Get(opts, addr)
	if err != nil {
		return rawIdentity{}, err
	}
	return rawIdentity{
		Key:             key,
		Signature:       signature,
		EthereumAddress: addr,
	}, nil
}

func registerRawIdentity(
	batch *txbatch.TXBatch,
	contracts *deployment.Contracts,
	identity rawIdentity) error {
	from := batch.TransactOpts.From
	ctx := batch.TransactOpts.Context
	callOpts := &bind.CallOpts{Context: ctx}

	currentIdentity, err := getRawIdentity(callOpts, contracts, from)
	if err != nil {
		return err
	}
	registerKey := len(currentIdentity.Key) == 0
	if !registerKey && !bytes.Equal(currentIdentity.Key, identity.Key) {
		return errors.Errorf("Identity already registered: wrong key: have %x, expected %x",
			currentIdentity.Key, identity.Key)
	}

	registerSignature := len(currentIdentity.Signature) == 0
	if !registerSignature && !bytes.Equal(currentIdentity.Signature, identity.Signature) {
		return errors.Errorf("Identity already registered: wrong signature")
	}

	if !registerKey && !registerSignature {
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

	if registerKey {
		tx, err := contracts.BLSPublicKeyRegistry.Register(batch.TransactOpts, n, i, identity.Key)
		if err != nil {
			return err
		}
		batch.Add(tx)
	}
	if registerSignature {
		tx, err := contracts.BLSSignatureRegistry.Register(batch.TransactOpts, n, i, identity.Signature)
		if err != nil {
			return err
		}
		batch.Add(tx)
	}
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
	raw := rawIdentity{
		Key:             shbls.SecretToPublicKey(key).Marshal(),
		Signature:       shbls.Sign(msg, key).Marshal(),
		EthereumAddress: batch.TransactOpts.From,
	}

	return registerRawIdentity(batch, contracts, raw)
}

func (raw *rawIdentity) getKey() (*shbls.PublicKey, error) {
	key := shbls.PublicKey{}
	err := key.Unmarshal(raw.Key)
	if err != nil {
		return nil, err
	}
	signature := shbls.Signature{}
	err = signature.Unmarshal(raw.Signature)
	if err != nil {
		return nil, err
	}
	if !shbls.Verify(&signature, &key, raw.EthereumAddress.Bytes()) {
		return nil, errors.Errorf("Cannot verify signature")
	}
	return &key, nil
}

// Lookup retrieves and verifies the BLS public key the decryptor with the given address has
// registered.
func Lookup(opts *bind.CallOpts, contracts *deployment.Contracts, addr common.Address) (*shbls.PublicKey, error) {
	raw, err := getRawIdentity(opts, contracts, addr)
	if err != nil {
		return nil, err
	}
	return raw.getKey()
}

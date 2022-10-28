package encoding

import (
	"crypto/rand"
	"math/big"

	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/shutter/shlib/shcrypto"
)

func NewEonKeyEnvironment() (*EonKeyEnvironment, error) {
	eke := &EonKeyEnvironment{
		numKeyper: 3,
		threshold: 2,
	}
	err := eke.init()
	return eke, err
}

type EonKeyEnvironment struct {
	gammas             []*shcrypto.Gammas
	eonSecretKeyShares []*shcrypto.EonSecretKeyShare
	numKeyper          int
	threshold          uint64
}

func (e *EonKeyEnvironment) EonPublicKey() *shcrypto.EonPublicKey {
	return shcrypto.ComputeEonPublicKey(e.gammas)
}

func (e *EonKeyEnvironment) init() error {
	ps := []*shcrypto.Polynomial{}
	for i := 0; i < e.numKeyper; i++ {
		p, err := shcrypto.RandomPolynomial(rand.Reader, e.threshold-1)
		if err != nil {
			return err
		}
		ps = append(ps, p)
		e.gammas = append(e.gammas, p.Gammas())
	}

	for i := 0; i < e.numKeyper; i++ {
		vs := []*big.Int{}
		for j := 0; j < e.numKeyper; j++ {
			v := ps[j].EvalForKeyper(i)
			vs = append(vs, v)
		}
		eonSecretKeyShare := shcrypto.ComputeEonSecretKeyShare(vs)
		e.eonSecretKeyShares = append(e.eonSecretKeyShares, eonSecretKeyShare)
	}
	return nil
}

func (e *EonKeyEnvironment) EpochSecretKey(epoch []byte) (*shcrypto.EpochSecretKey, error) {
	epochID := shcrypto.ComputeEpochID(epoch)

	epochSecretKeyShares := []*shcrypto.EpochSecretKeyShare{}
	for i := 0; i < e.numKeyper; i++ {
		epochSecretKeyShare := shcrypto.ComputeEpochSecretKeyShare(e.eonSecretKeyShares[i], epochID)
		epochSecretKeyShares = append(epochSecretKeyShares, epochSecretKeyShare)
	}

	indices := make([]int, 0)
	partialEpochSecretKeyShares := []*shcrypto.EpochSecretKeyShare{}
	for i := 0; i < int(e.threshold); i++ {
		partialEpochSecretKeyShares = append(partialEpochSecretKeyShares, epochSecretKeyShares[i])
		indices = append(indices, i)
	}
	return shcrypto.ComputeEpochSecretKey(
		indices,
		partialEpochSecretKeyShares,
		e.threshold,
	)
}

func (e *EonKeyEnvironment) DecryptPayload(payload []byte, epoch []byte) (*txtypes.ShutterPayload, error) {
	mess := shcrypto.EncryptedMessage{}
	err := mess.Unmarshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, "can't decode encrypted payload")
	}
	secretKey, err := e.EpochSecretKey(epoch)
	if err != nil {
		return nil, errors.Wrap(err, "can't derive epoch secret-key")
	}
	decryptedPayloadBytes, err := mess.Decrypt(secretKey)
	if err != nil {
		return nil, errors.Wrap(err, "can't decrypt payload")
	}
	decryptedPayload, err := txtypes.DecodeShutterPayload(decryptedPayloadBytes)
	if err != nil {
		return nil, errors.Wrap(err, "can't decode decrypted payload")
	}
	return decryptedPayload, nil
}

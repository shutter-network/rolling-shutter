package validatorregistry

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	blst "github.com/supranational/blst/bindings/go"
)

var dst = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

func VerifyAggregateSignature(sig *blst.P2Affine, pks []*blst.P1Affine, msg *AggregateRegistrationMessage) bool {
	if msg.Version < 1 {
		return false
	}
	if len(pks) != int(msg.Count) {
		fmt.Println(len(pks), int(msg.Count))
		return false
	}
	msgHash := crypto.Keccak256(msg.Marshal())
	msgs := make([][]byte, len(pks))
	for i := range pks {
		msgs[i] = msgHash
	}
	return sig.AggregateVerify(true, pks, true, msgs, dst)
}

func VerifySignature(sig *blst.P2Affine, pubkey *blst.P1Affine, msg *LegacyRegistrationMessage) bool {
	msgHash := crypto.Keccak256(msg.Marshal())
	return sig.Verify(true, pubkey, true, msgHash, dst)
}

func CreateAggregateSignature(sks []*blst.SecretKey, msg *AggregateRegistrationMessage) *blst.P2Affine {
	msgHash := crypto.Keccak256(msg.Marshal())
	aggregate := new(blst.P2Aggregate)
	for _, sk := range sks {
		aff := new(blst.P2Affine)
		sig := aff.Sign(sk, msgHash, dst)
		ok := aggregate.Add(sig, true)
		if !ok {
			panic("failure")
		}
	}
	return aggregate.ToAffine()
}

func CreateSignature(sk *blst.SecretKey, msg *LegacyRegistrationMessage) *blst.P2Affine {
	msgHash := crypto.Keccak256(msg.Marshal())
	return new(blst.P2Affine).Sign(sk, msgHash, dst)
}

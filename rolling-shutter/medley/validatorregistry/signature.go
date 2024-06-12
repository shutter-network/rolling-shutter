package validatorregistry

import (
	"github.com/ethereum/go-ethereum/crypto"
	blst "github.com/supranational/blst/bindings/go"
)

var dst = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

func VerifySignature(sig *blst.P2Affine, pubkey *blst.P1Affine, msg *RegistrationMessage) bool {
	msgHash := crypto.Keccak256(msg.Marshal())
	return sig.Verify(true, pubkey, true, msgHash, dst)
}

func CreateSignature(sk *blst.SecretKey, msg *RegistrationMessage) *blst.P2Affine {
	msgHash := crypto.Keccak256(msg.Marshal())
	return new(blst.P2Affine).Sign(sk, msgHash, dst)
}

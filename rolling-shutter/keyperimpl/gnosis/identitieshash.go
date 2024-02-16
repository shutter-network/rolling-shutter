package gnosis

import (
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func computeIdentitiesHash(identityPreimages []identitypreimage.IdentityPreimage) []byte {
	identityPreimagesAsBytes := [][]byte{}
	for _, preimage := range identityPreimages {
		identityPreimagesAsBytes = append(identityPreimagesAsBytes, preimage)
	}
	return crypto.Keccak256(identityPreimagesAsBytes...)
}

func computeIdentitiesHashFromShares(shares []*p2pmsg.KeyShare) []byte {
	identityPreimges := []identitypreimage.IdentityPreimage{}
	for _, share := range shares {
		identityPreimges = append(identityPreimges, identitypreimage.IdentityPreimage(share.EpochID))
	}
	return computeIdentitiesHash(identityPreimges)
}

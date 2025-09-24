package shutterservice

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
		identityPreimges = append(identityPreimges, identitypreimage.IdentityPreimage(share.IdentityPreimage))
	}
	return computeIdentitiesHash(identityPreimges)
}

func computeIdentitiesHashFromKeys(keys []*p2pmsg.Key) []byte {
	identityPreimges := []identitypreimage.IdentityPreimage{}
	for _, key := range keys {
		identityPreimges = append(identityPreimges, identitypreimage.IdentityPreimage(key.IdentityPreimage))
	}
	return computeIdentitiesHash(identityPreimges)
}

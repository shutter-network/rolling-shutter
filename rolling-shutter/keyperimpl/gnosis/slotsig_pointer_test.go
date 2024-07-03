package gnosis

import (
	"bytes"
	"encoding/hex"
	"log"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/gnosisssztypes"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

/*
data: &{InstanceID:42 Eon:2 Slot:10683832 TxPointer:0 IdentityPreimages:[
	{Bytes:[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 163 5 185]}]}
data hash: ff9d6bfe29cce02e04471901e5c8a8c5e9c91fba43e580ded0bab41081b45f2a
msg: instanceID:42 eon:2 keys:{
	identity:"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xa3\x05\xb9"
	key:"\xac\xb9\x14\xdc\x1f\x07v\xfap8;\x069\x1b\xf4-\xa9\x89\xcd/\x1bѮ\xe2\x1c\x91\x1d\x84\x07\x86\x8c_i\xb1i\x16\xce\xf0L\x8d\x9au\xe7\xa0!\x85V$"
}
	gnosis:{slot:10683832 signerIndices:0 signerIndices:1
		signatures:"\xdb\xc8\x18\x8cG\x95\xf1\xec\x8d\x05p\xe5\x1dN\x06Pv\x93헇\x7f\xff\xe2P\xa3\x13\xe0\x883\xbc\x80\"d힕\x19\x9d\x89A\xd84\x82\xe6_\x0c\x9f\x98N\"aT\xb4\xa6\xaf\x95\x1e(%\x1c\xca.\x1f\x01"
		signatures:"\xc5\\/\x96\xdc\x0f\xfdyRHϏ\xf7\x99(\x8d<48\xfbZ\xfaw\xde\xd8sx\x00\x89\xf5L\x08u\xa9~tZ\x91\xae0-\xafWxq\xd0B=4\xda\xe2\x0e@xE\xd3\x02V\x9f\xf6\xa1\xf0\x97\xe5\x01"
	}
keyper set: {KeyperConfigIndex:2 ActivationBlockNumber:10394254 Keypers:[
	0x9A1ba2D523AAB8f7784870B639924103d25Bb714
	0x7b79Ba0f76eE49F6246c0034A2a3445C281a67EB
	0x62F6DC5638250bD9edE84DFBfa54efA263186a4a
	] Threshold:2}
all signatures:
index: 0
signer: 0x9A1ba2D523AAB8f7784870B639924103d25Bb714
signature: dbc8188c4795f1ec8d0570e51d4e06507693ed97877fffe250a313e08833bc802264ed9e95199d8941d83482e65f0c9f984e226154b4a6af951e28251cca2e1f01
valid: false
---
index: 1
signer: 0x7b79Ba0f76eE49F6246c0034A2a3445C281a67EB
signature: c55c2f96dc0ffd795248cf8ff799288d3c3438fb5afa77ded873780089f54c0875a97e745a91ae302daf577871d0423d34dae20e407845d302569ff6a1f097e501
valid: false
---
*/

// this test re-creates the issue #456
func TestSlotDecryptionSignatureTxPointer(t *testing.T) {
	identity := "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xa3\x05\xb9"
	var identities []identitypreimage.IdentityPreimage
	identityPreimage := identitypreimage.IdentityPreimage(identity)
	identities = append(identities, identityPreimage)
	slotDecryptionSignatureData, err := gnosisssztypes.NewSlotDecryptionSignatureData(
		42,
		2,
		10683832,
		0,
		identities,
	)

	dataHash, err := slotDecryptionSignatureData.HashTreeRoot()
	assert.NilError(t, err, "could not hash data")

	expectedDataHash, err := hex.DecodeString("ff9d6bfe29cce02e04471901e5c8a8c5e9c91fba43e580ded0bab41081b45f2a")
	assert.NilError(t, err, "decoding dataHash failed")

	equal := bytes.Equal(dataHash[:], expectedDataHash)
	assert.Equal(t, equal, true, "dataHash does not match expected")

	var keyperAddresses []common.Address
	keyperAddresses = append(keyperAddresses, common.HexToAddress("0x9A1ba2D523AAB8f7784870B639924103d25Bb714"))
	keyperAddresses = append(keyperAddresses, common.HexToAddress("0x7b79Ba0f76eE49F6246c0034A2a3445C281a67EB"))
	keyperAddresses = append(keyperAddresses, common.HexToAddress("0x62F6DC5638250bD9edE84DFBfa54efA263186a4a"))

	var signatures [][]byte

	_sig, err := hex.DecodeString("dbc8188c4795f1ec8d0570e51d4e06507693ed97877fffe250a313e08833bc802264ed9e95199d8941d83482e65f0c9f984e226154b4a6af951e28251cca2e1f01")
	assert.NilError(t, err, "decoding signature failed")
	signatures = append(signatures, _sig)
	_sig, err = hex.DecodeString("c55c2f96dc0ffd795248cf8ff799288d3c3438fb5afa77ded873780089f54c0875a97e745a91ae302daf577871d0423d34dae20e407845d302569ff6a1f097e501")
	assert.NilError(t, err, "decoding signature failed")
	signatures = append(signatures, _sig)
	var validCount int

	var s int64
	var slotnumber uint64
	// to check if slot number was changed between signing time and sending time, we iterate over slot number.
	for s = -200; s < 10; s++ {
		slotnumber = uint64(10683832 + s)
		var i uint64
		// to check if tx-pointer was changed between signing time and sending time, we iterate over tx-pointer.
		for i = 0; i < 500; i++ {
			validCount = 0
			slotDecryptionSignatureData, _ = gnosisssztypes.NewSlotDecryptionSignatureData(
				42,
				2,
				slotnumber,
				i,
				identities,
			)
			// to check if there was some data mixup, we check the signatures against all keypers
			for _, keyperAddress := range keyperAddresses {
				for _, sig := range signatures {
					signatureValid, err := slotDecryptionSignatureData.CheckSignature(sig, keyperAddress)
					assert.NilError(t, err, "signature check failed")
					log.Println("keyper", keyperAddress, "valid", signatureValid)
					if signatureValid {
						log.Printf("tx-pointer was %v", i)
						log.Printf("slot was %v", slotnumber)
						validCount++
						break
					}
				}
			}
			if validCount == 2 {
				log.Printf("tx-pointer was %v", i)
				log.Printf("slot was %v", slotnumber)
				break
			}
		}
	}

	assert.Equal(t, validCount, 2, "not enough valid signatures")
}

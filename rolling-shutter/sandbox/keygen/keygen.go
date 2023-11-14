// simple tool to generate random eon keys and corresponding decryption keys.
package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
)

func main() {
	keygen := testkeygen.NewTestKeyGenerator(&testing.T{}, 3, 2, false)

	var prevEonPublicKey *shcrypto.EonPublicKey
	for i := uint64(0); i < 200; i++ {
		identityPreimage := identitypreimage.Uint64ToIdentityPreimage(i)
		eonPublicKey := keygen.EonPublicKey(identityPreimage)
		decryptionKey := keygen.EpochSecretKey(identityPreimage)

		if prevEonPublicKey == nil || !bytes.Equal(eonPublicKey.Marshal(), prevEonPublicKey.Marshal()) {
			if prevEonPublicKey != nil {
				fmt.Printf("\n")
			}
			fmt.Printf("eon key: %X\n\n", eonPublicKey.Marshal())
			fmt.Printf("epoch id | decryption key\n")
		}
		prevEonPublicKey = eonPublicKey

		fmt.Printf("%X | %X\n", identityPreimage.Bytes(), decryptionKey.Marshal())
	}
}

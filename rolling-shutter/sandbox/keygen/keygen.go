// simple tool to generate random eon keys and corresponding decryption keys.
package main

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
)

func main() {
	keys, err := testkeygen.NewEonKeys(rand.Reader, 3, 2)
	if err != nil {
		panic(err)
	}

	var prevEonPublicKey *shcrypto.EonPublicKey
	for i := uint64(0); i < 200; i++ {
		identityPreimage := identitypreimage.Uint64ToIdentityPreimage(i)
		eonPublicKey := keys.EonPublicKey()
		decryptionKey, err := keys.EpochSecretKey(identityPreimage)
		if err != nil {
			panic(err)
		}

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

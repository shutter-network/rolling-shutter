// simple tool to generate random eon keys and corresponding decryption keys.
package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
)

func main() {
	keygen := testkeygen.NewTestKeyGenerator(&testing.T{}, 3, 2)

	var prevEonPublicKey *shcrypto.EonPublicKey
	for epochID := uint64(0); epochID < 200; epochID++ {
		eonPublicKey := keygen.EonPublicKey(epochID)
		decryptionKey := keygen.EpochSecretKey(epochID)

		if prevEonPublicKey == nil || !bytes.Equal(eonPublicKey.Marshal(), prevEonPublicKey.Marshal()) {
			if prevEonPublicKey != nil {
				fmt.Printf("\n")
			}
			fmt.Printf("eon key: %X\n\n", eonPublicKey.Marshal())
			fmt.Printf("epoch id | decryption key\n")
		}
		prevEonPublicKey = eonPublicKey

		fmt.Printf("%8d | %X\n", epochID, decryptionKey.Marshal())
	}
}

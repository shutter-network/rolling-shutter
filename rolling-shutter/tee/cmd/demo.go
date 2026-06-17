package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/tee"
)

func main() {
	conn, err := tee.DialVerifiedChainDataChannel("127.0.0.1:7772", tee.Config{
		// // Configuration options for SGX verification:
		// SameSigner: false,		// Require that the etheruem-enclave is signed with the same key as we are
		// SignerID:   []byte{},	// Alternative: Specify the SignerID
		// MrEnclave:  []byte{},	// Hash of the code the ethereum-enclave is running
		// ProductID:  new(uint16),	// Number by the signer to identify the ethereum-enclave to distinguish different products/binaries
		// MinISVSVN:  0,			// Minimum (security) version number specified when signing the ethereum enclave

		// Currently we do not use events verified by the ethereum-enclave, so we can just tell it to not extract them.
		// EventExtractionStartBlocknum: ^uint64(0),
		// Contracts:                    nil, // For now: We are not interested in events

		EventExtractionStartBlocknum: 9031200,
		Contracts: []common.Address{
			common.HexToAddress("0x3eDF60dd017aCe33A0220F78741b5581C385A1BA"), // USDZ
		},
	})
	if err != nil {
		panic(err)
	}

	for {
		select {
		case d := <-conn.Headers():
			fmt.Println("Headers: ", d)
		case d := <-conn.Events():
			if len(d.Events) > 0 {
				fmt.Println("Finalized Events: ", d)
			}
		case d := <-conn.Errors():
			fmt.Println("Error: ", d)
			return
		}
	}
}

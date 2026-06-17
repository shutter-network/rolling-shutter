/*
TEE-based hardware encryption functionality for secret data.

This package contains functionality for handling secrets so that they can be
securely stored on — and loaded from — disk. This is a simple serialisation
step, and depending on whether TEE hardware is present (and whether the code is
running in an enclave), these functions either perform hardware sealing and
unsealing, or they simply pass the original data along. This ensures that old
persistence/config files can be loaded as usual, but new files written from
within an enclave will have their secrets sealed.

This package is designed to be minimally invasive for the rest of the codebase.
It does not protect against replay attacks where old persistences are loaded.
The sealed data can only be unsealed by the same CPU that sealed it and only from within a binary that is signed with the same key as the binary that sealed it.

The threat model of this package is to protect a node operator from hackers with admin access that seek to leak his secrets. It does not provide trust in the
node operator's honesty to the outside world. It merely makes the operator trust his own machine to not leak secrets sealed using this package.
*/
package tee

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"

	egoattestation "github.com/edgelesssys/ego/attestation"
	"github.com/edgelesssys/ego/ecrypto"
	egoenclave "github.com/edgelesssys/ego/enclave"
)

// If we are running in SGX, this contains the enclave identity.
var tpmIdentity *egoattestation.Report = nil

func init() {
	ident, hasTPM, err := QueryTPMIdentity()
	if err != nil {
		panic(fmt.Sprintf("Could not query TPM identity: %v", err))
	}
	if hasTPM {
		tpmIdentity = &ident
	}
}

func HasTEE() bool {
	return tpmIdentity != nil
}

func QueryTPMIdentity() (identity egoattestation.Report, exists bool, err error) {
	// Get some information about where we're running (if in a TEE and some
	// details about the TEE we're running in)
	identity, err = egoenclave.GetSelfReport()
	exists = true
	if err != nil {
		if err.Error() == "OE_UNSUPPORTED" {
			exists = false
			err = nil
		} else {
			err = fmt.Errorf("Could not get self-report: %w", err)
		}
	}
	return
}

// All sealed blobs are prefixed with this string, to distinguish them from
// normal blobs. Normal data should never start with this string.
const sealedPrefix = "sealed://"

func blobIsSealed(blob []byte) (sealed bool, payload []byte) {
	if len(blob) >= len(sealedPrefix) && string(blob[:len(sealedPrefix)]) == sealedPrefix {
		return true, blob[len(sealedPrefix):]
	}
	return false, blob
}

// If we are in a TEE, always seal the secret. Otherwise, simply convert to hex.
func SealSecretAsHex(secret []byte) (string, error) {
	return SealSecretAsCustomText(secret, hex.EncodeToString)
}

func SealSecretAsCustomText(secret []byte, encoder func([]byte) string) (string, error) {
	if sealed, _ := blobIsSealed(secret); sealed {
		return "", errors.New("Blob looks like it is already sealed")
	}

	if !HasTEE() {
		return encoder(secret), nil
	}

	sealed, err := ecrypto.SealWithProductKey(secret, nil)
	if err != nil {
		return "", fmt.Errorf("failed to seal secret: %w", err)
	}

	return sealedPrefix + encoder(sealed), nil
}

// If we are in a TEE, always seal the secret. Otherwise, returns the raw bytes.
func SealSecret(secret []byte) ([]byte, error) {
	if sealed, _ := blobIsSealed(secret); sealed {
		return nil, errors.New("Blob looks like it is already sealed")
	}

	if !HasTEE() {
		return secret, nil
	}

	sealed, err := ecrypto.SealWithProductKey(secret, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to seal secret: %w", err)
	}

	ret := []byte(sealedPrefix)
	return append(ret, sealed...), nil
}

// If we are in a TEE, unseal the secret, if it is sealed. Otherwise, simply parse the hex string.
func UnsealSecretFromHex(str string) ([]byte, error) {
	return UnsealSecretFromCustomText(str, hex.DecodeString)
}

func UnsealSecretFromCustomText(str string, decoder func(string) ([]byte, error)) ([]byte, error) {
	if cut, ok := strings.CutPrefix(str, sealedPrefix); ok {
		bytes, err := decoder(cut)
		if err != nil {
			return nil, err
		}

		if !HasTEE() {
			return nil, errors.New("Cannot unseal secret: not in a TEE")
		}

		return ecrypto.Unseal(bytes, nil)
	}

	return decoder(str)
}

func UnsealSecret(secret []byte) ([]byte, error) {
	if sealed, payload := blobIsSealed(secret); sealed {
		if !HasTEE() {
			return nil, errors.New("Cannot unseal secret: not in a TEE")
		}

		return ecrypto.Unseal(payload, nil)
	}

	return secret, nil
}

// This function is only for testing purposes, when the non-TEE behaviour is
// needed somewhere.
//
// Executes the passed function with temporarily disabled TEE functionality
// (during the execution, HasTEE() returns false). If this function is called
// outside of tests, it panics.
func DoWithoutTEE(the_thing func()) {
	if !testing.Testing() {
		panic("This function is for tests only")
	}

	restore := tpmIdentity
	defer func() { tpmIdentity = restore }()
	tpmIdentity = nil

	the_thing()
}

package tee

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

// Start of utlities

func RequiresTEE(t *testing.T) {
	if !HasTEE() {
		t.Skip("Test requires TEE, but we are not running on a TEE")
	}
}

func RunWithoutTEE(t *testing.T, name string, test func(t *testing.T)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		DoWithoutTEE(func() { test(t) })
	})
}

func RunWithTEE(t *testing.T, name string, test func(t *testing.T)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		RequiresTEE(t)
		test(t)
	})
}

// End of utilities

// Start of test cases

func TestBlobIsSealed(t *testing.T) {
	t.Run("Sealed input", func(t *testing.T) {
		input := append([]byte(sealedPrefix), 1, 2, 3, 4)
		isSealed, payload := blobIsSealed(input)
		assert.Equal(t, isSealed, true, "isSealed")
		assert.DeepEqual(t, payload, []byte{1, 2, 3, 4})
	})

	t.Run("Unsealed input", func(t *testing.T) {
		input := []byte{5, 6, 7, 8}
		isSealed, payload := blobIsSealed(input)
		assert.Equal(t, isSealed, false, "isSealed")
		assert.DeepEqual(t, payload, []byte{5, 6, 7, 8})
	})
}

func TestSealSecretAsHex(t *testing.T) {
	RunWithoutTEE(t, "Without TEE", func(t *testing.T) {
		input := []byte{0x12, 0x34, 0x56, 0x78}
		output, err := SealSecretAsHex(input)
		assert.NilError(t, err, "seal")
		assert.Equal(t, output, "12345678", "seal")

		reverse, err := UnsealSecretFromHex(output)
		assert.NilError(t, err, "unseal")
		assert.DeepEqual(t, reverse, input)
	})

	RunWithTEE(t, "HW sealing", func(t *testing.T) {
		input := []byte{0x9a, 0xbc, 0xde, 0xf0}
		output, err := SealSecretAsHex(input)
		assert.NilError(t, err, "seal")
		assert.Assert(t, strings.HasPrefix(output, sealedPrefix), "seal")

		reverse, err := UnsealSecretFromHex(output)
		assert.NilError(t, err, "unseal")
		assert.DeepEqual(t, reverse, input)
	})
}

func TestSealSecret(t *testing.T) {
	RunWithoutTEE(t, "Without TEE", func(t *testing.T) {
		input := []byte{0x12, 0x34, 0x56, 0x78}
		output, err := SealSecret(input)
		assert.NilError(t, err, "seal")
		assert.DeepEqual(t, output, input)

		reverse, err := UnsealSecret(output)
		assert.NilError(t, err, "unseal")
		assert.DeepEqual(t, reverse, input)
	})

	RunWithTEE(t, "HW sealing", func(t *testing.T) {
		input := []byte{0x9a, 0xbc, 0xde, 0xf0}
		output, err := SealSecret(input)
		assert.NilError(t, err, "seal")
		isSealed, _ := blobIsSealed(output)
		assert.Assert(t, isSealed, "seal")

		reverse, err := UnsealSecret(output)
		assert.NilError(t, err, "unseal")
		assert.DeepEqual(t, reverse, input)
	})
}

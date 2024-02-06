package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"syscall/js"

	"github.com/shutter-network/shutter/shlib/shcrypto"
)

var (
	uint8Array        js.Value
	uint8ClampedArray js.Value
	jsRegistry        js.Value
)

const (
	registryJavaScriptName = "__wasm_bridge__"
	initializedPromiseName = "_initialized"
)

func main() {
	uint8Array = js.Global().Get("Uint8Array")
	uint8ClampedArray = js.Global().Get("Uint8ClampedArray")
	jsRegistry = js.Global().Get(registryJavaScriptName)

	jsRegistry.Set("encrypt", encrypt)
	jsRegistry.Set("decrypt", decrypt)
	jsRegistry.Set("verifyDecryptionKey", verifyDecryptionKey)

	// Tell JS we're loaded
	jsRegistry.Get(initializedPromiseName).Invoke()

	select {}
}

var encrypt = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	if len(args) != 4 {
		return encodeResult(nil, fmt.Errorf("expected 4 arguments, got %d", len(args)))
	}
	messageArg := args[0]
	eonPublicKeyArg := args[1]
	epochIDArg := args[2]
	sigmaArg := args[3]

	message, err := decodeMessageArg(messageArg)
	if err != nil {
		return encodeResult(nil, err)
	}
	eonPublicKey, err := decodeEonPublicKeyArg(eonPublicKeyArg)
	if err != nil {
		return encodeResult(nil, err)
	}
	epochID, err := decodeEpochIDArgPoint(epochIDArg)
	if err != nil {
		return encodeResult(nil, err)
	}
	sigma, err := decodeSigmaArg(sigmaArg)
	if err != nil {
		return encodeResult(nil, err)
	}

	encryptedMessage := shcrypto.Encrypt(message, eonPublicKey, epochID, sigma)
	return encodeResult(encryptedMessage.Marshal(), nil)
})

var decrypt = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return encodeResult(nil, fmt.Errorf("expected 4 arguments, got %d", len(args)))
	}
	encryptedMessageArg := args[0]
	decryptionKeyArg := args[1]

	encryptedMessage, err := decodeEncryptedMessageArg(encryptedMessageArg)
	if err != nil {
		return encodeResult(nil, err)
	}
	decryptionKey, err := decodeDecryptionKeyArg(decryptionKeyArg)
	if err != nil {
		return encodeResult(nil, err)
	}

	message, err := encryptedMessage.Decrypt(decryptionKey)
	if err != nil {
		return encodeResult(nil, fmt.Errorf("failed to decrypt message: %s", err))
	}
	return encodeResult(message, nil)
})

var verifyDecryptionKey = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	if len(args) != 3 {
		return encodeResult(nil, fmt.Errorf("expected 4 arguments, got %d", len(args)))
	}
	decryptionKeyArg := args[0]
	eonPublicKeyArg := args[1]
	epochIDArg := args[2]

	decryptionKey, err := decodeDecryptionKeyArg(decryptionKeyArg)
	if err != nil {
		return encodeResult(nil, err)
	}
	eonPublicKey, err := decodeEonPublicKeyArg(eonPublicKeyArg)
	if err != nil {
		return encodeResult(nil, err)
	}
	epochID, err := decodeEpochIDArgBytes(epochIDArg)
	if err != nil {
		return encodeResult(nil, err)
	}

	ok, err := shcrypto.VerifyEpochSecretKey(decryptionKey, eonPublicKey, epochID)
	if err != nil {
		return encodeResult(nil, err)
	}
	return ok
})

func encodeResult(encryptedMessage []byte, err error) string {
	if err != nil {
		return "Error: " + err.Error()
	}
	return "0x" + hex.EncodeToString(encryptedMessage)
}

func decodeMessageArg(arg js.Value) ([]byte, error) {
	return decodeBytesArg(arg, "message")
}

func decodeEonPublicKeyArg(arg js.Value) (*shcrypto.EonPublicKey, error) {
	b, err := decodeBytesArg(arg, "eonPublicKey")
	if err != nil {
		return nil, err
	}

	p := new(shcrypto.EonPublicKey)
	err = p.Unmarshal(b)
	if err != nil {
		return nil, fmt.Errorf("invalid eon public key: %s", err)
	}

	return p, nil
}

func decodeEpochIDArgBytes(arg js.Value) ([]byte, error) {
	b, err := decodeBytesArg(arg, "epochID")
	if err != nil {
		return nil, err
	}
	if len(b) != shcrypto.BlockSize {
		return nil, fmt.Errorf("epochID must be %d bytes, got %d", shcrypto.BlockSize, len(b))
	}
	return b, nil
}

func decodeEpochIDArgPoint(args js.Value) (*shcrypto.EpochID, error) {
	b, err := decodeEpochIDArgBytes(args)
	if err != nil {
		return nil, err
	}
	p := shcrypto.ComputeEpochID(b)
	return p, nil
}

func decodeSigmaArg(arg js.Value) (shcrypto.Block, error) {
	var s shcrypto.Block

	b, err := decodeBytesArg(arg, "sigma")
	if err != nil {
		return s, err
	}
	if len(b) != shcrypto.BlockSize {
		return s, fmt.Errorf("sigma must be %d bytes, got %d", shcrypto.BlockSize, len(b))
	}

	copy(s[:], b)
	return s, nil
}

func decodeEncryptedMessageArg(arg js.Value) (*shcrypto.EncryptedMessage, error) {
	b, err := decodeBytesArg(arg, "encryptedMessage")
	if err != nil {
		return nil, err
	}

	m := new(shcrypto.EncryptedMessage)
	err = m.Unmarshal(b)
	if err != nil {
		return nil, fmt.Errorf("invalid encrypted message: %s", err)
	}

	return m, nil
}

func decodeDecryptionKeyArg(arg js.Value) (*shcrypto.EpochSecretKey, error) {
	b, err := decodeBytesArg(arg, "eonPublicKey")
	if err != nil {
		return nil, err
	}

	k := new(shcrypto.EpochSecretKey)
	err = k.Unmarshal(b)
	if err != nil {
		return nil, fmt.Errorf("invalid decryption key: %s", err)
	}

	return k, nil
}

func decodeBytesArg(arg js.Value, name string) ([]byte, error) {
	if !(arg.InstanceOf(uint8Array) || arg.InstanceOf(uint8ClampedArray)) {
		return nil, fmt.Errorf("argument %s must be of type Uint8Array", name)
	}
	b := make([]byte, arg.Get("length").Int())
	js.CopyBytesToGo(b, arg)
	return b, nil
}

func decodeUint64Arg(arg js.Value, name string) (uint64, error) {
	b, err := decodeBytesArg(arg, name)
	if err != nil {
		return 0, err
	}
	if len(b) != 8 {
		return 0, fmt.Errorf("%s must be 8 bytes, got %d", name, len(b))
	}
	return binary.BigEndian.Uint64(b), nil
}

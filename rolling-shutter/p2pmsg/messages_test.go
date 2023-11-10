package p2pmsg

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/trace"
)

func marshalUnmarshalMessage[M Message](t *testing.T, message M, traceContext *TraceContext) (M, *TraceContext) { //nolint:thelper
	var (
		err        error
		ok         bool
		msgBytes   []byte
		newMessage M
		unmshl     any
	)

	msgBytes, err = Marshal(message, traceContext)
	assert.NilError(t, err)

	unmshl, unmshlTc, err := Unmarshal(msgBytes)
	assert.NilError(t, err)
	if !trace.IsEnabled() {
		assert.Assert(t, unmshlTc == nil)
	}
	newMessage, ok = unmshl.(M)
	assert.Assert(t, ok)
	return newMessage, unmshlTc
}

type testConfig struct {
	identityPreimage identitypreimage.IdentityPreimage
	blockNumber      uint64
	instanceID       uint64
	tkg              *testkeygen.TestKeyGenerator
}

func defaultTestConfig(t *testing.T) testConfig {
	t.Helper()

	identityPreimage, _ := identitypreimage.BigToIdentityPreimage(common.Big2)
	return testConfig{
		identityPreimage: identityPreimage,
		blockNumber:      uint64(0),
		instanceID:       uint64(42),
		tkg:              testkeygen.NewTestKeyGenerator(t, 1, 1, false),
	}
}

func TestDecryptionKey(t *testing.T) {
	cfg := defaultTestConfig(t)
	validSecretKey := cfg.tkg.EpochSecretKey(cfg.identityPreimage).Marshal()

	orig := &DecryptionKey{
		EpochID:    cfg.identityPreimage.Bytes(),
		InstanceID: cfg.instanceID,
		Key:        validSecretKey,
	}
	m, tc := marshalUnmarshalMessage(t, orig, nil)
	assert.Assert(t, tc == nil)
	assert.DeepEqual(t, orig, m, cmpopts.IgnoreUnexported(DecryptionKey{}))
}

func TestDecryptionTrigger(t *testing.T) {
	cfg := defaultTestConfig(t)
	txs := [][]byte{
		[]byte("tx1"),
		[]byte("tx2"),
		[]byte("tx3"),
	}

	privKey, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)

	orig, err := NewSignedDecryptionTrigger(cfg.instanceID, cfg.identityPreimage, cfg.blockNumber, HashByteList(txs), privKey)
	assert.NilError(t, err)
	m, tc := marshalUnmarshalMessage(t, orig, nil)
	assert.Assert(t, tc == nil)
	assert.DeepEqual(t, orig, m, cmpopts.IgnoreUnexported(DecryptionTrigger{}))
}

func TestDecryptionKeyShare(t *testing.T) {
	cfg := defaultTestConfig(t)
	keyperIndex := uint64(0)
	keyshare := cfg.tkg.EpochSecretKeyShare(cfg.identityPreimage, keyperIndex).Marshal()

	orig := &DecryptionKeyShares{
		InstanceID:  cfg.instanceID,
		KeyperIndex: keyperIndex,
		Shares: []*KeyShare{{
			EpochID: cfg.identityPreimage.Bytes(),
			Share:   keyshare,
		}},
	}
	m, tc := marshalUnmarshalMessage(t, orig, nil)
	assert.Assert(t, tc == nil)
	assert.DeepEqual(t, orig, m, cmpopts.IgnoreUnexported(DecryptionKeyShares{}, KeyShare{}))
}

func TestEonPublicKey(t *testing.T) {
	cfg := defaultTestConfig(t)
	eonPublicKey := cfg.tkg.EonPublicKey(cfg.identityPreimage).Marshal()
	activationBlock := uint64(2)

	privKey, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	eon := uint64(5)
	keyperConfigIndex := uint64(6)
	orig, err := NewSignedEonPublicKey(
		cfg.instanceID, eonPublicKey, activationBlock, keyperConfigIndex, eon, privKey,
	)
	assert.NilError(t, err)

	m, tc := marshalUnmarshalMessage(t, orig, nil)
	assert.Assert(t, tc == nil)
	assert.DeepEqual(t, orig, m, cmpopts.IgnoreUnexported(EonPublicKey{}))
}

func TestTraceContext(t *testing.T) {
	trace.SetEnabled()
	defer trace.SetDisabled()

	tc := &TraceContext{
		TraceID:    []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		SpanID:     []byte{1, 1, 1, 1, 1, 1, 1, 1},
		TraceFlags: []byte{1},
		TraceState: "tracestate",
	}
	cfg := defaultTestConfig(t)
	validSecretKey := cfg.tkg.EpochSecretKey(cfg.identityPreimage).Marshal()
	msg := &DecryptionKey{
		EpochID:    cfg.identityPreimage.Bytes(),
		InstanceID: cfg.instanceID,
		Key:        validSecretKey,
	}
	_, newTc := marshalUnmarshalMessage(t, msg, tc)
	assert.DeepEqual(t, tc, newTc, cmpopts.IgnoreUnexported(TraceContext{}))
}

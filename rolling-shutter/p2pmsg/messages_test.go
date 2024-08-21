package p2pmsg

import (
	"crypto/rand"
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
	keys             *testkeygen.EonKeys
}

func defaultTestConfig(t *testing.T) testConfig {
	t.Helper()

	identityPreimage := identitypreimage.BigToIdentityPreimage(common.Big2)
	keys, err := testkeygen.NewEonKeys(rand.Reader, 1, 1)
	assert.NilError(t, err)
	return testConfig{
		identityPreimage: identityPreimage,
		blockNumber:      uint64(0),
		instanceID:       uint64(42),
		keys:             keys,
	}
}

func TestDecryptionKeys(t *testing.T) {
	cfg := defaultTestConfig(t)
	validSecretKey, err := cfg.keys.EpochSecretKey(cfg.identityPreimage)
	assert.NilError(t, err)

	orig := &DecryptionKeys{
		InstanceId: cfg.instanceID,
		Keys: []*Key{
			{
				Identity: cfg.identityPreimage.Bytes(),
				Key:      validSecretKey.Marshal(),
			},
		},
	}
	m, tc := marshalUnmarshalMessage(t, orig, nil)
	assert.Assert(t, tc == nil)
	assert.DeepEqual(t, orig, m, cmpopts.IgnoreUnexported(DecryptionKeys{}, Key{}))
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
	keyshare := cfg.keys.EpochSecretKeyShare(cfg.identityPreimage, int(keyperIndex)).Marshal()

	orig := &DecryptionKeyShares{
		InstanceId:  cfg.instanceID,
		KeyperIndex: keyperIndex,
		Shares: []*KeyShare{{
			EpochId: cfg.identityPreimage.Bytes(),
			Share:   keyshare,
		}},
	}
	m, tc := marshalUnmarshalMessage(t, orig, nil)
	assert.Assert(t, tc == nil)
	assert.DeepEqual(t, orig, m, cmpopts.IgnoreUnexported(DecryptionKeyShares{}, KeyShare{}))
}

func TestEonPublicKey(t *testing.T) {
	cfg := defaultTestConfig(t)
	eonPublicKey := cfg.keys.EonPublicKey().Marshal()
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
		TraceId:    []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		SpanId:     []byte{1, 1, 1, 1, 1, 1, 1, 1},
		TraceFlags: []byte{1},
		TraceState: "tracestate",
	}
	cfg := defaultTestConfig(t)
	validSecretKey, err := cfg.keys.EpochSecretKey(cfg.identityPreimage)
	assert.NilError(t, err)
	msg := &DecryptionKeys{
		InstanceId: cfg.instanceID,
		Keys: []*Key{
			{
				Identity: cfg.identityPreimage.Bytes(),
				Key:      validSecretKey.Marshal(),
			},
		},
	}
	_, newTc := marshalUnmarshalMessage(t, msg, tc)
	assert.DeepEqual(t, tc, newTc, cmpopts.IgnoreUnexported(TraceContext{}))
}

package shmsg

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"

	shcrypto "github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
	"github.com/shutter-network/shutter/shuttermint/medley/testkeygen"
)

func marshalUnmarshalMessage[M P2PMessage](t *testing.T, message M) M {
	t.Helper()

	var (
		err        error
		ok         bool
		msgBytes   []byte
		newMessage M
		unmshl     any
	)
	const sender = "foo/sender"

	msgBytes, err = proto.Marshal(message)
	assert.NilError(t, err)

	unmshl, err = Unmarshal(message.Topic(), msgBytes)
	assert.NilError(t, err)
	newMessage, ok = unmshl.(M)
	assert.Assert(t, ok)
	return newMessage
}

type testConfig struct {
	epochID    uint64
	instanceID uint64
	tkg        *testkeygen.TestKeyGenerator
}

func defaultTestConfig(t *testing.T) testConfig {
	t.Helper()

	return testConfig{
		epochID:    epochid.New(2, 0),
		instanceID: uint64(42),
		tkg:        testkeygen.NewTestKeyGenerator(t, 1, 1),
	}
}

func TestNewPolyCommitmentMsg(t *testing.T) {
	eon := uint64(10)
	threshold := uint64(5)
	poly, err := shcrypto.RandomPolynomial(rand.Reader, threshold)
	assert.NilError(t, err)
	gammas := poly.Gammas()

	msgContainer := NewPolyCommitment(eon, gammas)
	msg := msgContainer.GetPolyCommitment()
	assert.Assert(t, msg != nil)

	assert.Equal(t, eon, msg.Eon)
	assert.Equal(t, int(threshold)+1, len(msg.Gammas))
	for i := 0; i < int(threshold)+1; i++ {
		gammaBytes := msg.Gammas[i]
		assert.DeepEqual(t, gammaBytes, (*gammas)[i].Marshal())
	}
}

func TestNewPolyEvalMsg(t *testing.T) {
	eon := uint64(10)
	receiver := common.BigToAddress(big.NewInt(0xaabbcc))
	encryptedEval := []byte("secret")

	msgContainer := NewPolyEval(eon, []common.Address{receiver}, [][]byte{encryptedEval})
	msg := msgContainer.GetPolyEval()
	assert.Assert(t, msg != nil)

	assert.Equal(t, eon, msg.Eon)
	assert.DeepEqual(t, receiver.Bytes(), msg.Receivers[0])
}

func TestNewP2PMessageFromTopic(t *testing.T) {
	for _, mess := range messageTypes {
		newInstance, err := NewP2PMessageFromTopic(mess.Topic())
		assert.NilError(t, err)
		assert.Equal(t, reflect.TypeOf(newInstance), reflect.TypeOf(mess))
	}
	newInstance, err := NewP2PMessageFromTopic("notopicknown")
	assert.ErrorContains(t, err, "No message type found")
	assert.Assert(t, newInstance == nil)
}

func TestDecryptionKey(t *testing.T) {
	cfg := defaultTestConfig(t)
	validSecretKey := cfg.tkg.EpochSecretKey(cfg.epochID).Marshal()

	orig := &DecryptionKey{
		EpochID:    cfg.epochID,
		InstanceID: cfg.instanceID,
		Key:        validSecretKey,
	}
	m := marshalUnmarshalMessage(t, orig)
	assert.Equal(t, orig.EpochID, m.EpochID)
	assert.Equal(t, orig.InstanceID, m.InstanceID)
	assert.Assert(t, bytes.Compare(orig.Key, m.Key) == 0)
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

	orig, err := NewSignedDecryptionTrigger(cfg.instanceID, cfg.epochID, txs, privKey)
	assert.NilError(t, err)
	m := marshalUnmarshalMessage(t, orig)

	assert.Equal(t, orig.EpochID, m.EpochID)
	assert.Equal(t, orig.InstanceID, m.InstanceID)
	assert.Assert(t, bytes.Compare(orig.TransactionsHash, m.TransactionsHash) == 0)
	assert.Assert(t, bytes.Compare(orig.Signature, m.Signature) == 0)
}

func TestDecryptionKeyShare(t *testing.T) {
	cfg := defaultTestConfig(t)
	keyperIndex := uint64(0)
	keyshare := cfg.tkg.EpochSecretKeyShare(cfg.epochID, keyperIndex).Marshal()

	orig := &DecryptionKeyShare{
		EpochID:     cfg.epochID,
		InstanceID:  cfg.instanceID,
		Share:       keyshare,
		KeyperIndex: keyperIndex,
	}
	m := marshalUnmarshalMessage(t, orig)

	assert.Equal(t, orig.EpochID, m.EpochID)
	assert.Equal(t, orig.InstanceID, m.InstanceID)
	assert.Equal(t, orig.KeyperIndex, m.KeyperIndex)
	assert.Assert(t, bytes.Compare(orig.Share, m.Share) == 0)
}

func TestEonPublicKey(t *testing.T) {
	cfg := defaultTestConfig(t)
	eonPublicKey := cfg.tkg.EonPublicKey(cfg.epochID).Marshal()
	keyperIndex := uint64(0)
	activationBlock := uint64(0)

	privKey, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)

	orig, err := NewSignedEonPublicKey(cfg.instanceID, eonPublicKey, activationBlock, keyperIndex, privKey)
	assert.NilError(t, err)

	m := marshalUnmarshalMessage(t, orig)

	assert.Equal(t, orig.InstanceID, m.InstanceID)
	assert.Assert(t, bytes.Compare(orig.PublicKey, m.PublicKey) == 0)
	assert.Equal(t, orig.ActivationBlock, m.ActivationBlock)
	assert.Equal(t, orig.KeyperIndex, m.KeyperIndex)
	assert.Assert(t, bytes.Compare(orig.Signature, m.Signature) == 0)
}

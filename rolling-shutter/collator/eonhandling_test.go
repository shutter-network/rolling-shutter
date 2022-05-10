package collator

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/commondb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type keyper struct {
	address string
	index   uint64
	msg     *shmsg.EonPublicKey
}

func (k *keyper) handleMsg(ctx context.Context, c *collator) error {
	if k.msg == nil {
		return errors.New("Message not initialized")
	}
	ok, err := c.validateEonPublicKey(ctx, k.msg)
	if err != nil {
		return err
	}
	if ok {
		_, err := c.handleEonPublicKey(ctx, k.msg)
		return err
	}
	return nil
}

type setupEonKeysParams struct {
	instanceID      uint64
	activationBlock uint64
	eventIndex      uint64
	threshold       uint64
	eonPubKey       []byte
	keypers         []*ecdsa.PrivateKey
}

func setupEonKeys(ctx context.Context, t *testing.T, dbtx commondb.DBTX, params setupEonKeysParams) []keyper {
	t.Helper()

	kprs := make([]keyper, 0)

	for i, ethKey := range params.keypers {
		var (
			err error
			ok  bool
			msg *shmsg.EonPublicKey
		)

		msg, err = shmsg.NewSignedEonPublicKey(
			params.instanceID,
			params.eonPubKey,
			params.activationBlock,
			uint64(i),
			5,
			2,
			ethKey,
		)
		assert.NilError(t, err)
		addr := ethcrypto.PubkeyToAddress(ethKey.PublicKey)
		kprs = append(kprs, keyper{address: addr.Hex(), index: uint64(i), msg: msg})
		ok, err = msg.VerifySignature(addr)
		assert.Check(t, ok)
		assert.NilError(t, err)
	}
	keyperSet := make([]string, 0)
	for _, k := range kprs {
		keyperSet = append(keyperSet, k.address)
	}

	db := commondb.New(dbtx)
	err := db.InsertKeyperSet(ctx, commondb.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(params.eventIndex),
		Keypers:               keyperSet,
		ActivationBlockNumber: int64(params.activationBlock),
		Threshold:             int32(params.threshold),
	})
	assert.NilError(t, err)

	return kprs
}

func checkDBResult(t *testing.T, kpr []keyper, msgs []cltrdb.GetEonPublicKeyMessagesRow) {
	t.Helper()

	assert.Check(t, len(msgs) > 0)
	var err error

	for _, m := range msgs {
		k := kpr[m.KeyperIndex]
		assert.Equal(t, k.msg.Candidate.ActivationBlock, uint64(m.ActivationBlockNumber))
		assert.Check(t, bytes.Equal(k.msg.Candidate.PublicKey, m.EonPublicKey))

		var unmshldTemp shmsg.P2PMessage
		unmshldTemp, err = shmsg.Unmarshal(k.msg.Topic(), m.MsgBytes)
		assert.NilError(t, err)

		unmshld, ok := unmshldTemp.(*shmsg.EonPublicKey)
		assert.Check(t, ok)
		assert.Equal(t, k.msg.Candidate.ActivationBlock, unmshld.Candidate.ActivationBlock)
		assert.Check(t, bytes.Equal(k.msg.Candidate.PublicKey, unmshld.Candidate.PublicKey))
	}
}

func TestHandleEonKeyIntegration(t *testing.T) {
	var (
		eonPubKey, eonPubKeyBefore, eonPubKeyNoThreshold []byte
		err                                              error
		dbEon, dbEonBefore                               []cltrdb.GetEonPublicKeyMessagesRow
	)

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, dbpool, closedb := testdb.NewCollatorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)
	tkgBefore := testkeygen.NewTestKeyGenerator(t, 3, 2)
	tkg := testkeygen.NewTestKeyGenerator(t, 3, 2)

	eonPubKeyNoThreshold, _ = tkgBefore.EonPublicKey(1).GobEncode()
	eonPubKeyBefore, _ = tkgBefore.EonPublicKey(1000).GobEncode()
	eonPubKey, _ = tkg.EonPublicKey(2000).GobEncode()

	kpr1, _ := ethcrypto.GenerateKey()
	kpr2, _ := ethcrypto.GenerateKey()
	kpr3, _ := ethcrypto.GenerateKey()

	activationBlockBefore := uint64(42)
	activationBlockNoThreshold := uint64(43)
	activationBlock := uint64(1042)

	// Insert pubkey with not enough signatures
	keypersNoThreshold := setupEonKeys(ctx, t, dbpool, setupEonKeysParams{
		instanceID:      config.InstanceID,
		eventIndex:      uint64(0),
		activationBlock: activationBlockNoThreshold,
		eonPubKey:       eonPubKeyNoThreshold,
		threshold:       tkg.Threshold,
		keypers:         []*ecdsa.PrivateKey{kpr1},
	})
	assert.Check(t, len(keypersNoThreshold) > 0)

	// Insert pubkeys with enough signatures and new key / eventindex/ keyperset
	// but same activation-block
	keypersBefore := setupEonKeys(ctx, t, dbpool, setupEonKeysParams{
		instanceID:      config.InstanceID,
		eventIndex:      uint64(1),
		activationBlock: activationBlockBefore,
		eonPubKey:       eonPubKeyBefore,
		threshold:       tkg.Threshold,
		keypers:         []*ecdsa.PrivateKey{kpr1, kpr2, kpr3},
	})
	assert.Check(t, len(keypersBefore) > 0)

	keypers := setupEonKeys(ctx, t, dbpool, setupEonKeysParams{
		instanceID:      config.InstanceID,
		eventIndex:      uint64(2),
		activationBlock: activationBlock,
		eonPubKey:       eonPubKey,
		threshold:       tkg.Threshold,
		keypers:         []*ecdsa.PrivateKey{kpr3, kpr1, kpr2},
	})
	assert.Check(t, len(keypers) > 0)

	// HACK: Only partially instantiating the collator.
	// This works until the handler/validator functions use something else than
	// the database-pool
	c := collator{dbpool: dbpool}

	for _, k := range keypersBefore {
		err = k.handleMsg(ctx, &c)
		assert.NilError(t, err)
	}
	for _, k := range keypersNoThreshold {
		err = k.handleMsg(ctx, &c)
		assert.NilError(t, err)
	}
	for _, k := range keypers {
		err = k.handleMsg(ctx, &c)
		assert.NilError(t, err)
	}

	// Although the no-threshold reaching pubkey message have a
	// later activation block, they should not get retrieved
	// because they should not get considered valid at this point
	dbEonBefore, err = db.GetEonPublicKeyMessages(ctx, 500)
	assert.NilError(t, err)
	assert.Check(t, len(dbEonBefore) == 3)
	// This should only return the messages with enough signatures
	checkDBResult(t, keypersBefore, dbEonBefore)

	dbEon, err = db.GetEonPublicKeyMessages(ctx, 1050)
	assert.Check(t, len(dbEon) == 3)
	assert.NilError(t, err)
	checkDBResult(t, keypers, dbEon)
}

func TestHandleEonAmbiguityFailsIntegration(t *testing.T) {
	// THE GOAL IS TO MAKE THIS FAIL! (see #238)
	// (testing module does not have expected-to-fail marker)

	var (
		eonPubKey, eonPubKeyNoThreshold []byte
		err                             error
		dbEon                           []cltrdb.GetEonPublicKeyMessagesRow
	)

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, dbpool, closedb := testdb.NewCollatorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)
	tkgBefore := testkeygen.NewTestKeyGenerator(t, 3, 2)
	tkg := testkeygen.NewTestKeyGenerator(t, 3, 2)

	eonPubKeyNoThreshold, _ = tkgBefore.EonPublicKey(1).GobEncode()
	eonPubKey, _ = tkg.EonPublicKey(1000).GobEncode()

	kpr1, _ := ethcrypto.GenerateKey()
	kpr2, _ := ethcrypto.GenerateKey()
	kpr3, _ := ethcrypto.GenerateKey()

	activationBlock := uint64(42)

	// Insert pubkey with not enough signatures
	keypersNoThreshold := setupEonKeys(ctx, t, dbpool, setupEonKeysParams{
		instanceID:      config.InstanceID,
		eventIndex:      uint64(0),
		activationBlock: activationBlock,
		eonPubKey:       eonPubKeyNoThreshold,
		threshold:       tkg.Threshold,
		keypers:         []*ecdsa.PrivateKey{kpr1},
	})
	assert.Check(t, len(keypersNoThreshold) > 0)

	// Insert pubkeys with enough signatures and new key / eventindex/ keyperset
	// but same activation-block
	keypers := setupEonKeys(ctx, t, dbpool, setupEonKeysParams{
		instanceID:      config.InstanceID,
		eventIndex:      uint64(1),
		activationBlock: activationBlock,
		eonPubKey:       eonPubKey,
		threshold:       tkg.Threshold,
		keypers:         []*ecdsa.PrivateKey{kpr1, kpr2, kpr3},
	})
	assert.Check(t, len(keypers) > 0)

	// HACK: Only partially instantiating the collator.
	// This works until the handler/validator functions use something else than
	// the database-pool
	c := collator{dbpool: dbpool}

	for _, k := range keypersNoThreshold {
		err = k.handleMsg(ctx, &c)
		assert.NilError(t, err)
	}

	for _, k := range keypers {
		// the validation should fail due to the
		// ambiguity
		err = k.handleMsg(ctx, &c)
		assert.NilError(t, err)
	}

	// This is actually NOT what we expect to happen!

	dbEon, err = db.GetEonPublicKeyMessages(ctx, 500)
	assert.NilError(t, err)
	// The messages are not returned because of the
	// too low threshold,
	// but the ambiguous keyper set still hinders the
	// correct messages being validated correctly and
	// thus they will not be handled (inserted in the db)
	assert.Check(t, len(dbEon) == 0)
}

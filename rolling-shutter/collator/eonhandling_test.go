package collator

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	enctime "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/time"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testkeygen"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p/p2ptest"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func newTestConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := config.New()
	err := configuration.SetExampleValuesRecursive(cfg)
	assert.NilError(t, err)
	cfg.EpochDuration = &enctime.Duration{Duration: 1 * time.Second}
	return cfg
}

type keyper struct {
	address string
	index   uint64
	msg     *p2pmsg.EonPublicKey
}

type setupEonKeysParams struct {
	instanceID        uint64
	activationBlock   uint64
	keyperConfigIndex uint64
	threshold         uint64
	eonPubKey         []byte
	eon               uint64
	keypers           []*ecdsa.PrivateKey
}

func setupEonKeys(
	ctx context.Context,
	t *testing.T,
	dbtx chainobsdb.DBTX,
	params setupEonKeysParams,
) []keyper {
	t.Helper()

	kprs := make([]keyper, 0)

	for i, ethKey := range params.keypers {
		var (
			err error
			ok  bool
			msg *p2pmsg.EonPublicKey
		)

		msg, err = p2pmsg.NewSignedEonPublicKey(
			params.instanceID,
			params.eonPubKey,
			params.activationBlock,
			params.keyperConfigIndex, // keyperConfigIndex
			params.eon,               // eon
			ethKey,
		)
		assert.NilError(t, err)
		addr := ethcrypto.PubkeyToAddress(ethKey.PublicKey)
		kprs = append(kprs, keyper{address: addr.Hex(), index: uint64(i), msg: msg})
		ok, err = p2pmsg.VerifySignature(msg, addr)
		assert.Check(t, ok)
		assert.NilError(t, err)
	}
	keyperSet := make([]string, 0)
	for _, k := range kprs {
		keyperSet = append(keyperSet, k.address)
	}

	db := chainobsdb.New(dbtx)
	err := db.InsertKeyperSet(ctx, chainobsdb.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(params.keyperConfigIndex),
		Keypers:               keyperSet,
		ActivationBlockNumber: int64(params.activationBlock),
		Threshold:             int32(params.threshold),
	})
	assert.NilError(t, err)

	return kprs
}

func checkDBResult(
	t *testing.T,
	kpr []keyper,
	pubkey cltrdb.EonPublicKeyCandidate,
	votes []cltrdb.EonPublicKeyVote,
) {
	t.Helper()

	assert.Check(t, len(votes) > 0)
	keyperBySender := make(map[string]keyper)
	for _, k := range kpr {
		keyperBySender[k.address] = k
	}

	for _, v := range votes {
		k, ok := keyperBySender[v.Sender]
		assert.Check(t, ok)
		assert.Equal(t, k.msg.ActivationBlock, uint64(pubkey.ActivationBlockNumber))
		assert.Check(t, bytes.Equal(k.msg.PublicKey, pubkey.EonPublicKey))
		pubkey, err := ethcrypto.SigToPub(pubkey.Hash, v.Signature)
		assert.NilError(t, err)
		recoveredAddress := ethcrypto.PubkeyToAddress(*pubkey)
		assert.Equal(t, recoveredAddress.Hex(), v.Sender)
	}
}

func TestHandleEonKeyIntegration(t *testing.T) {
	var (
		eonPubKey, eonPubKeyBefore, eonPubKeyNoThreshold []byte
		err                                              error
	)

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, dbpool, closedb := testdb.NewCollatorTestDB(ctx, t)
	defer closedb()
	testConfig := newTestConfig(t)
	tkgBefore := testkeygen.NewTestKeyGenerator(t, 3, 2)
	tkg := testkeygen.NewTestKeyGenerator(t, 3, 2)

	epoch1 := epochid.Uint64ToEpochID(1)
	epoch1000 := epochid.Uint64ToEpochID(1000)
	epoch2000 := epochid.Uint64ToEpochID(2000)

	eonPubKeyNoThreshold, _ = tkgBefore.EonPublicKey(epoch1).GobEncode()
	eonPubKeyBefore, _ = tkgBefore.EonPublicKey(epoch1000).GobEncode()
	eonPubKey, _ = tkg.EonPublicKey(epoch2000).GobEncode()

	kpr1, _ := ethcrypto.GenerateKey()
	kpr2, _ := ethcrypto.GenerateKey()
	kpr3, _ := ethcrypto.GenerateKey()

	activationBlockBefore := uint64(42)
	activationBlockNoThreshold := uint64(43)
	activationBlock := uint64(1042)

	// Insert pubkey with not enough signatures
	keypersNoThreshold := setupEonKeys(ctx, t, dbpool, setupEonKeysParams{
		instanceID:        testConfig.InstanceID,
		eon:               1,
		keyperConfigIndex: uint64(0),
		activationBlock:   activationBlockNoThreshold,
		eonPubKey:         eonPubKeyNoThreshold,
		threshold:         tkg.Threshold,
		keypers:           []*ecdsa.PrivateKey{kpr1},
	})
	assert.Check(t, len(keypersNoThreshold) > 0)

	// Insert pubkeys with enough signatures and new key / keyperConfigIndex / keyperset
	// but same activation-block
	keypersBefore := setupEonKeys(ctx, t, dbpool, setupEonKeysParams{
		instanceID:        testConfig.InstanceID,
		eon:               2,
		keyperConfigIndex: uint64(1),
		activationBlock:   activationBlockBefore,
		eonPubKey:         eonPubKeyBefore,
		threshold:         tkg.Threshold,
		keypers:           []*ecdsa.PrivateKey{kpr1, kpr2, kpr3},
	})
	assert.Check(t, len(keypersBefore) > 0)

	keypers := setupEonKeys(ctx, t, dbpool, setupEonKeysParams{
		instanceID:        testConfig.InstanceID,
		eon:               3,
		keyperConfigIndex: uint64(2),
		activationBlock:   activationBlock,
		eonPubKey:         eonPubKey,
		threshold:         tkg.Threshold,
		keypers:           []*ecdsa.PrivateKey{kpr3, kpr1, kpr2},
	})
	assert.Check(t, len(keypers) > 0)

	var handler p2p.MessageHandler = &eonPublicKeyHandler{dbpool: dbpool, config: testConfig}

	for _, k := range keypersBefore {
		p2ptest.MustHandleMessage(t, handler, ctx, k.msg)
	}
	for _, k := range keypersNoThreshold {
		p2ptest.MustHandleMessage(t, handler, ctx, k.msg)
	}
	for _, k := range keypers {
		p2ptest.MustHandleMessage(t, handler, ctx, k.msg)
	}

	// Although the no-threshold reaching pubkey message have a
	// later activation block, they should not get retrieved
	// because they should not get considered valid at this point

	pubkey, err := db.FindEonPublicKeyForBlock(ctx, 500)
	assert.NilError(t, err)
	votes, err := db.FindEonPublicKeyVotes(ctx, pubkey.Hash)
	assert.NilError(t, err)
	assert.Equal(t, len(votes), 3)
	checkDBResult(t, keypersBefore, pubkey, votes)

	pubkey, err = db.FindEonPublicKeyForBlock(ctx, 1050)
	assert.NilError(t, err)
	votes, err = db.FindEonPublicKeyVotes(ctx, pubkey.Hash)
	assert.NilError(t, err)
	assert.Equal(t, len(votes), 3)
	checkDBResult(t, keypers, pubkey, votes)
}

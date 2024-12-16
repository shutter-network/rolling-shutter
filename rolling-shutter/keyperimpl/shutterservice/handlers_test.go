package shutterservice

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"gotest.tools/assert"

	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/serviceztypes"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func TestHandleDecryptionKeySharesThresholdNotReached(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	obsKeyperDB := obskeyperdatabase.New(dbpool)
	privateKey, sender, err := generateRandomAccount()
	assert.NilError(t, err)
	keyperIndex := uint64(1)
	keyperConfigIndex := uint64(1)

	err = obsKeyperDB.InsertKeyperSet(ctx, obskeyperdatabase.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(keyperConfigIndex),
		ActivationBlockNumber: 1,
		Keypers:               shdb.EncodeAddresses([]common.Address{sender}),
		Threshold:             2,
	})
	assert.NilError(t, err)

	identityPreimages := []serviceztypes.IdentityPreimage{}
	for i := 0; i < 3; i++ {
		identityPreimage := serviceztypes.IdentityPreimage{
			Bytes: intTo32ByteArray(i),
		}
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	decryptionData := &serviceztypes.DecryptionSignatureData{
		InstanceID:        config.GetInstanceID(),
		Eon:               keyperConfigIndex,
		IdentityPreimages: identityPreimages,
	}

	signature, err := decryptionData.ComputeSignature(privateKey)
	assert.NilError(t, err)

	keys := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)
	var handler p2p.MessageHandler = &DecryptionKeySharesHandler{dbpool: dbpool}

	// threshold is two, so no outgoing message after first input
	shares := []*p2pmsg.KeyShare{}
	for _, identityPreimage := range identityPreimages {
		share := &p2pmsg.KeyShare{
			IdentityPreimage: identityPreimage.Bytes,
			Share:            keys.EpochSecretKeyShare(identityPreimage.Bytes, 0).Marshal(),
		}
		shares = append(shares, share)
	}
	msg := &p2pmsg.DecryptionKeyShares{
		InstanceId:  config.GetInstanceID(),
		Eon:         keyperConfigIndex,
		KeyperIndex: 0,
		Shares:      shares,
		Extra: &p2pmsg.DecryptionKeyShares_Service{
			Service: &p2pmsg.ShutterServiceDecryptionKeySharesExtra{
				Signature: signature,
			},
		},
	}
	validation, err := handler.ValidateMessage(ctx, msg)
	assert.NilError(t, err)
	assert.Equal(t, validation, pubsub.ValidationAccept)
	msgs, err := handler.HandleMessage(ctx, msg)
	assert.NilError(t, err)
	assert.Equal(t, len(msgs), 0)
}

func TestHandleDecryptionKeySharesThresholdReached(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	obsKeyperDB := obskeyperdatabase.New(dbpool)
	keyperCoreDB := corekeyperdatabase.New(dbpool)
	keyper1PrivateKey, keyper1Address, err := generateRandomAccount()
	assert.NilError(t, err)
	keyper2PrivateKey, keyper2Address, err := generateRandomAccount()
	assert.NilError(t, err)
	keyperIndex := uint64(1)
	keyperConfigIndex := uint64(1)

	err = obsKeyperDB.InsertKeyperSet(ctx, obskeyperdatabase.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(keyperConfigIndex),
		ActivationBlockNumber: 1,
		Keypers:               shdb.EncodeAddresses([]common.Address{keyper1Address, keyper2Address}),
		Threshold:             2,
	})
	assert.NilError(t, err)
	identityPreimages := []serviceztypes.IdentityPreimage{}
	for i := 0; i < 3; i++ {
		identityPreimage := serviceztypes.IdentityPreimage{
			Bytes: intTo32ByteArray(i),
		}
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	decryptionData := &serviceztypes.DecryptionSignatureData{
		InstanceID:        config.GetInstanceID(),
		Eon:               keyperConfigIndex,
		IdentityPreimages: identityPreimages,
	}

	keyper1Signature, err := decryptionData.ComputeSignature(keyper1PrivateKey)
	assert.NilError(t, err)

	keys := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)

	var handler p2p.MessageHandler = &DecryptionKeySharesHandler{dbpool: dbpool}

	// threshold is two, so no outgoing message after first input
	shares := []*p2pmsg.KeyShare{}
	for _, identityPreimage := range identityPreimages {
		share := &p2pmsg.KeyShare{
			IdentityPreimage: identityPreimage.Bytes,
			Share:            keys.EpochSecretKeyShare(identityPreimage.Bytes, 0).Marshal(),
		}
		shares = append(shares, share)
	}
	msg := &p2pmsg.DecryptionKeyShares{
		InstanceId:  config.GetInstanceID(),
		Eon:         keyperConfigIndex,
		KeyperIndex: 0,
		Shares:      shares,
		Extra: &p2pmsg.DecryptionKeyShares_Service{
			Service: &p2pmsg.ShutterServiceDecryptionKeySharesExtra{
				Signature: keyper1Signature,
			},
		},
	}
	validation, err := handler.ValidateMessage(ctx, msg)
	assert.NilError(t, err, "validation returned error")
	assert.Equal(t, validation, pubsub.ValidationAccept)

	msgs, err := handler.HandleMessage(ctx, msg)
	assert.NilError(t, err)
	assert.Equal(t, len(msgs), 0)

	keyper2Signature, err := decryptionData.ComputeSignature(keyper2PrivateKey)
	assert.NilError(t, err)

	// now thrsehold will be reached after this message causing to send out
	// the message
	shares = []*p2pmsg.KeyShare{}
	encodedDecryptionKeys := [][]byte{}
	for _, identityPreimage := range identityPreimages {
		share := &p2pmsg.KeyShare{
			IdentityPreimage: identityPreimage.Bytes,
			Share:            keys.EpochSecretKeyShare(identityPreimage.Bytes, 2).Marshal(),
		}
		shares = append(shares, share)

		decKey, _ := generateRandom32Bytes()
		_, err := keyperCoreDB.InsertDecryptionKey(ctx, corekeyperdatabase.InsertDecryptionKeyParams{
			Eon:           int64(keyperConfigIndex),
			EpochID:       identityPreimage.Bytes,
			DecryptionKey: decKey,
		})
		assert.NilError(t, err)

		encodedDecryptionKeys = append(encodedDecryptionKeys, decKey)
	}

	msg = &p2pmsg.DecryptionKeyShares{
		InstanceId:  config.GetInstanceID(),
		Eon:         keyperConfigIndex,
		KeyperIndex: 1,
		Shares:      shares,
		Extra: &p2pmsg.DecryptionKeyShares_Service{
			Service: &p2pmsg.ShutterServiceDecryptionKeySharesExtra{
				Signature: keyper2Signature,
			},
		},
	}

	validation, err = handler.ValidateMessage(ctx, msg)
	assert.NilError(t, err)
	assert.Equal(t, validation, pubsub.ValidationAccept)
	msgs, err = handler.HandleMessage(ctx, msg)
	assert.NilError(t, err)
	assert.Equal(t, len(msgs), 1)
	decKeyMsg, ok := msgs[0].(*p2pmsg.DecryptionKeys)
	assert.Equal(t, ok, true)
	assert.Equal(t, decKeyMsg.InstanceId, config.GetInstanceID())
	assert.Equal(t, len(decKeyMsg.Keys), len(identityPreimages))
	for i, key := range decKeyMsg.Keys {
		assert.Check(t, bytes.Equal(key.IdentityPreimage, identityPreimages[i].Bytes))
		assert.Check(t, bytes.Equal(key.Key, encodedDecryptionKeys[i]))
	}
}

func TestValidateAndHandleDecryptionKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	obsKeyperDB := obskeyperdatabase.New(dbpool)

	keyperConfigIndex := uint64(1)
	keyper1Index := uint64(0)
	keyper2Index := uint64(1)
	keyper1PrivateKey, keyper1Address, err := generateRandomAccount()
	assert.NilError(t, err)
	keyper2PrivateKey, keyper2Address, err := generateRandomAccount()
	assert.NilError(t, err)

	err = obsKeyperDB.InsertKeyperSet(ctx, obskeyperdatabase.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(keyperConfigIndex),
		ActivationBlockNumber: 1,
		Keypers:               shdb.EncodeAddresses([]common.Address{keyper1Address, keyper2Address}),
		Threshold:             2,
	})
	assert.NilError(t, err)

	identityPreimages := []serviceztypes.IdentityPreimage{}
	for i := 0; i < 3; i++ {
		identityPreimage := serviceztypes.IdentityPreimage{
			Bytes: intTo32ByteArray(i),
		}
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	decryptionData := &serviceztypes.DecryptionSignatureData{
		InstanceID:        config.GetInstanceID(),
		Eon:               keyperConfigIndex,
		IdentityPreimages: identityPreimages,
	}

	keyper1Signature, err := decryptionData.ComputeSignature(keyper1PrivateKey)
	assert.NilError(t, err)
	keyper2Signature, err := decryptionData.ComputeSignature(keyper2PrivateKey)
	assert.NilError(t, err)

	keyperIndex := uint64(1)

	keys := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)

	var handler p2p.MessageHandler = &DecryptionKeysHandler{dbpool: dbpool}
	encodedDecryptionKeys := [][]byte{}
	for _, identityPreimage := range identityPreimages {
		decryptionKey, err := keys.EpochSecretKey(identityPreimage.Bytes)
		assert.NilError(t, err)
		encodedDecryptionKey := decryptionKey.Marshal()
		encodedDecryptionKeys = append(encodedDecryptionKeys, encodedDecryptionKey)
	}

	decryptionKeys := []*p2pmsg.Key{}
	for i, identityPreimage := range identityPreimages {
		key := &p2pmsg.Key{
			IdentityPreimage: identityPreimage.Bytes,
			Key:              encodedDecryptionKeys[i],
		}
		decryptionKeys = append(decryptionKeys, key)
	}

	msg := &p2pmsg.DecryptionKeys{
		InstanceId: config.GetInstanceID(),
		Eon:        keyperConfigIndex,
		Keys:       decryptionKeys,
		Extra: &p2pmsg.DecryptionKeys_Service{
			Service: &p2pmsg.ShutterServiceDecryptionKeysExtra{
				SignerIndices: []uint64{keyper1Index, keyper2Index},
				Signature:     [][]byte{keyper1Signature, keyper2Signature},
			},
		},
	}
	validation, err := handler.ValidateMessage(ctx, msg)
	assert.NilError(t, err)
	assert.Equal(t, validation, pubsub.ValidationAccept)
	msgs, err := handler.HandleMessage(ctx, msg)
	assert.NilError(t, err)

	assert.Check(t, len(msgs) == 0)
}

func TestInValidateDecryptionKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	t.Cleanup(dbclose)

	obsKeyperDB := obskeyperdatabase.New(dbpool)

	keyperConfigIndex := uint64(1)
	keyper1Index := uint64(1)
	keyper2Index := uint64(2)
	keyper1PrivateKey, keyper1Address, err := generateRandomAccount()
	assert.NilError(t, err)
	keyper2PrivateKey, keyper2Address, err := generateRandomAccount()
	assert.NilError(t, err)

	err = obsKeyperDB.InsertKeyperSet(ctx, obskeyperdatabase.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(keyperConfigIndex),
		ActivationBlockNumber: 1,
		Keypers:               shdb.EncodeAddresses([]common.Address{keyper1Address, keyper2Address}),
		Threshold:             2,
	})
	assert.NilError(t, err)

	identityPreimages := []serviceztypes.IdentityPreimage{}
	for i := 0; i < 3; i++ {
		identityPreimage := serviceztypes.IdentityPreimage{
			Bytes: intTo32ByteArray(i),
		}
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	decryptionData := &serviceztypes.DecryptionSignatureData{
		InstanceID:        config.GetInstanceID(),
		Eon:               keyperConfigIndex,
		IdentityPreimages: identityPreimages,
	}

	keyper1Signature, err := decryptionData.ComputeSignature(keyper1PrivateKey)
	assert.NilError(t, err)
	keyper2Signature, err := decryptionData.ComputeSignature(keyper2PrivateKey)
	assert.NilError(t, err)

	keyperIndex := uint64(1)

	keys := testsetup.InitializeEon(ctx, t, dbpool, config, keyperIndex)

	var handler p2p.MessageHandler = &DecryptionKeysHandler{dbpool: dbpool}
	encodedDecryptionKeys := [][]byte{}
	for _, identityPreimage := range identityPreimages {
		decryptionKey, err := keys.EpochSecretKey(identityPreimage.Bytes)
		assert.NilError(t, err)
		encodedDecryptionKey := decryptionKey.Marshal()
		encodedDecryptionKeys = append(encodedDecryptionKeys, encodedDecryptionKey)
	}

	decryptionKeys := []*p2pmsg.Key{}
	for i, identityPreimage := range identityPreimages {
		key := &p2pmsg.Key{
			IdentityPreimage: identityPreimage.Bytes,
			Key:              encodedDecryptionKeys[i],
		}
		decryptionKeys = append(decryptionKeys, key)
	}

	msg := &p2pmsg.DecryptionKeys{
		InstanceId: config.GetInstanceID(),
		Eon:        keyperConfigIndex,
		Keys:       decryptionKeys,
		Extra: &p2pmsg.DecryptionKeys_Service{
			Service: &p2pmsg.ShutterServiceDecryptionKeysExtra{
				SignerIndices: []uint64{keyper1Index, keyper2Index},
				Signature:     [][]byte{keyper1Signature, keyper2Signature},
			},
		},
	}
	validation, err := handler.ValidateMessage(ctx, msg)
	assert.Error(t, err, "signer index out of range")
	assert.Equal(t, validation, pubsub.ValidationReject)
}

func intTo32ByteArray(num int) []byte {
	b := make([]byte, 32)
	tempBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(tempBuffer, uint64(num))
	copy(b[24:], tempBuffer)
	return b
}

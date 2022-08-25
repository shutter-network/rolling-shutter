package keyper

import (
	"context"
	"encoding/json"
	"log"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprtopics"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func (kpr *keyper) makeMessagesValidators() map[string]pubsub.Validator {
	validators := make(map[string]pubsub.Validator)
	validators[kprtopics.DecryptionKey] = kpr.validateDecryptionKey
	validators[kprtopics.DecryptionKeyShare] = kpr.validateDecryptionKeyShare
	validators[kprtopics.EonPublicKey] = kpr.validateEonPublicKey
	validators[kprtopics.DecryptionTrigger] = kpr.validateDecryptionTrigger

	return validators
}

func (kpr *keyper) validateDecryptionKey(ctx context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	p2pMessage := new(p2p.Message)
	if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
		return false
	}
	msg, err := unmarshalP2PMessage(p2pMessage)
	if err != nil {
		return false
	}
	if msg.GetInstanceID() != kpr.config.InstanceID {
		return false
	}

	key, ok := msg.(*decryptionKey)
	if !ok {
		return false
	}

	if _, err := epochid.BytesToEpochID(key.EpochID); err != nil {
		log.Printf("invalid epoch id")
		return false
	}
	if key.Eon > math.MaxInt64 {
		log.Printf("eon %d overflows int64", key.Eon)
		return false
	}

	dkgResultDB, err := kpr.db.GetDKGResult(ctx, int64(key.Eon))
	if err == pgx.ErrNoRows {
		log.Printf("no DKG result found for eon %d", key.Eon)
		return false
	}
	if err != nil {
		log.Printf("failed to get dkg result for eon %d from db", key.Eon)
		return false
	}
	if !dkgResultDB.Success {
		log.Printf("no successful DKG result found for eon %d", key.Eon)
		return false
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		log.Printf("error while decoding pure DKG result for eon %d", key.Eon)
		return false
	}
	epochSecretKey, err := key.GetEpochSecretKey()
	if err != nil {
		return false
	}

	ok, err = shcrypto.VerifyEpochSecretKey(epochSecretKey, pureDKGResult.PublicKey, key.EpochID)
	if err != nil {
		log.Printf("error while checking epoch secret key for epoch %v", key.EpochID)
		return false
	}
	return ok
}

func (kpr *keyper) validateDecryptionKeyShare(ctx context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	p2pMessage := new(p2p.Message)
	if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
		return false
	}
	msg, err := unmarshalP2PMessage(p2pMessage)
	if err != nil {
		return false
	}
	if msg.GetInstanceID() != kpr.config.InstanceID {
		return false
	}

	keyShare, ok := msg.(*decryptionKeyShare)
	if !ok {
		return false
	}

	if _, err := epochid.BytesToEpochID(keyShare.EpochID); err != nil {
		log.Printf("invalid epoch id")
		return false
	}
	if keyShare.Eon > math.MaxInt64 {
		log.Printf("eon %d overflows int64", keyShare.Eon)
		return false
	}

	dkgResultDB, err := kpr.db.GetDKGResult(ctx, int64(keyShare.Eon))
	if err == pgx.ErrNoRows {
		log.Printf("no DKG result found for eon %d", keyShare.Eon)
		return false
	}
	if err != nil {
		log.Printf("failed to get dkg result for eon %d from db", keyShare.Eon)
		return false
	}
	if !dkgResultDB.Success {
		log.Printf("no successful DKG result found for eon %d", keyShare.Eon)
		return false
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		log.Printf("error while decoding pure DKG result for eon %d", keyShare.Eon)
		return false
	}

	epochSecretKeyShare, err := keyShare.GetEpochSecretKeyShare()
	if err != nil {
		return false
	}

	ok = shcrypto.VerifyEpochSecretKeyShare(
		epochSecretKeyShare,
		pureDKGResult.PublicKeyShares[keyShare.KeyperIndex],
		shcrypto.ComputeEpochID(keyShare.EpochID),
	)
	return ok
}

func (kpr *keyper) validateEonPublicKey(_ context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	p2pMessage := new(p2p.Message)
	if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
		return false
	}
	msg, err := unmarshalP2PMessage(p2pMessage)
	if err != nil {
		return false
	}
	return msg.GetInstanceID() == kpr.config.InstanceID
}

func (kpr *keyper) validateDecryptionTrigger(ctx context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	p2pMessage := new(p2p.Message)
	if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
		return false
	}
	msg, err := unmarshalP2PMessage(p2pMessage)
	if err != nil {
		return false
	}

	trigger, ok := msg.(*decryptionTrigger)
	if !ok {
		return false
	}

	if trigger.GetInstanceID() != kpr.config.InstanceID {
		log.Printf("instance ID mismatch (want=%d, have=%d)", kpr.config.InstanceID, trigger.GetInstanceID())
		return false
	}
	if _, err := epochid.BytesToEpochID(trigger.EpochID); err != nil {
		log.Printf("invalid epoch id")
		return false
	}

	blk := trigger.BlockNumber
	if blk > math.MaxInt64 {
		log.Printf("block number %d overflows int64", blk)
		return false
	}
	chainCollator, err := kpr.db.GetChainCollator(ctx, int64(blk))
	if err == pgx.ErrNoRows {
		log.Printf("got decryption trigger with no collator for given block number: %d", blk)
		return false
	}
	if err != nil {
		log.Printf("error while getting collator from db for block number: %d", blk)
		return false
	}

	collator, err := shdb.DecodeAddress(chainCollator.Collator)
	if err != nil {
		log.Printf("error while converting collator from string to address: %s", chainCollator.Collator)
		return false
	}

	shTrigger := (*shmsg.DecryptionTrigger)(trigger)
	signatureValid, err := shTrigger.VerifySignature(collator)
	if err != nil {
		log.Printf("error while verifying decryption trigger signature for epoch: %d", trigger.EpochID)
		return false
	}

	return signatureValid
}

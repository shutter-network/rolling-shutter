package keyper

import (
	"context"
	"encoding/json"
	"log"

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

	activationBlockNumber := epochid.BlockNumber(key.epochID)
	dkgResultDB, err := kpr.db.GetDKGResultForBlockNumber(ctx, int64(activationBlockNumber))
	if err == pgx.ErrNoRows {
		return false
	}
	if err != nil {
		log.Printf("failed to get dkg result for epoch %d from db", key.epochID)
		return false
	}
	if !dkgResultDB.Success {
		return false
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		log.Printf("error while decoding pure DKG result for epoch %d", key.epochID)
		return false
	}

	ok, err = shcrypto.VerifyEpochSecretKey(key.key, pureDKGResult.PublicKey, key.epochID)
	if err != nil {
		log.Printf("error while checking epoch secret key for epoch %v", key.epochID)
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

	activationBlockNumber := epochid.BlockNumber(keyShare.epochID)
	dkgResultDB, err := kpr.db.GetDKGResultForBlockNumber(ctx, int64(activationBlockNumber))
	if err == pgx.ErrNoRows {
		return false
	}
	if err != nil {
		log.Printf("failed to get dkg result for epoch %d from db", keyShare.epochID)
		return false
	}
	if !dkgResultDB.Success {
		return false
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		log.Printf("error while decoding pure DKG result for epoch %d", keyShare.epochID)
		return false
	}

	ok = shcrypto.VerifyEpochSecretKeyShare(
		keyShare.share,
		pureDKGResult.PublicKeyShares[keyShare.keyperIndex],
		shcrypto.ComputeEpochID(keyShare.epochID),
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
	if msg.GetInstanceID() != kpr.config.InstanceID {
		return false
	}

	t, ok := msg.(*decryptionTrigger)
	if !ok {
		return false
	}
	blk := epochid.BlockNumber(t.EpochID)
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

	trigger := (*shmsg.DecryptionTrigger)(t)
	signatureValid, err := trigger.VerifySignature(collator)
	if err != nil {
		log.Printf("error while verifying decryption trigger signature for epoch: %d", t.EpochID)
		return false
	}

	return signatureValid
}

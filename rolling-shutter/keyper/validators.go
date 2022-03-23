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
	key, err := unmarshalDecryptionKey(p2pMessage)
	if err != nil {
		return false
	}
	if key.GetInstanceID() != kpr.config.InstanceID {
		return false
	}

	activationBlockNumber := epochid.BlockNumber(key.EpochID)
	dkgResultDB, err := kpr.db.GetDKGResultForBlockNumber(ctx, int64(activationBlockNumber))
	if err == pgx.ErrNoRows {
		return false
	}
	if err != nil {
		log.Printf("failed to get dkg result for epoch %d from db", key.EpochID)
		return false
	}
	if !dkgResultDB.Success {
		return false
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		log.Printf("error while decoding pure DKG result for epoch %d", key.EpochID)
		return false
	}
	epochSecretKey, err := key.GetEpochSecretKey()
	if err != nil {
		return false
	}

	ok, err := shcrypto.VerifyEpochSecretKey(epochSecretKey, pureDKGResult.PublicKey, key.EpochID)
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
	keyShare, err := unmarshalDecryptionKeyShare(p2pMessage)
	if err != nil {
		return false
	}
	if keyShare.GetInstanceID() != kpr.config.InstanceID {
		return false
	}

	activationBlockNumber := epochid.BlockNumber(keyShare.EpochID)
	dkgResultDB, err := kpr.db.GetDKGResultForBlockNumber(ctx, int64(activationBlockNumber))
	if err == pgx.ErrNoRows {
		return false
	}
	if err != nil {
		log.Printf("failed to get dkg result for epoch %d from db", keyShare.EpochID)
		return false
	}
	if !dkgResultDB.Success {
		return false
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		log.Printf("error while decoding pure DKG result for epoch %d", keyShare.EpochID)
		return false
	}
	epochSecretKeyShare, err := keyShare.GetEpochSecretKeyShare()
	if err != nil {
		return false
	}
	return shcrypto.VerifyEpochSecretKeyShare(
		epochSecretKeyShare,
		pureDKGResult.PublicKeyShares[keyShare.KeyperIndex],
		shcrypto.ComputeEpochID(keyShare.EpochID),
	)
}

func (kpr *keyper) validateEonPublicKey(_ context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	p2pMessage := new(p2p.Message)
	if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
		return false
	}
	msg, err := p2pMessage.Unmarshal()
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
	trigger, err := unmarshalDecryptionTrigger(p2pMessage)
	if err != nil {
		return false
	}
	if trigger.GetInstanceID() != kpr.config.InstanceID {
		return false
	}

	blk := epochid.BlockNumber(trigger.EpochID)
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

	signatureValid, err := trigger.VerifySignature(collator)
	if err != nil {
		log.Printf("error while verifying decryption trigger signature for epoch: %d", trigger.EpochID)
		return false
	}

	return signatureValid
}

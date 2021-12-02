package keyper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprtopics"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func (kpr *keyper) makeMessagesValidators() map[string]pubsub.Validator {
	db := kprdb.New(kpr.dbpool)
	validators := make(map[string]pubsub.Validator)
	validators[kprtopics.DecryptionKey] = kpr.makeDecryptionKeyValidator(db)
	validators[kprtopics.DecryptionKeyShare] = kpr.makeKeyShareValidator(db)
	validators[kprtopics.EonPublicKey] = kpr.makeEonPublicKeyValidator()
	validators[kprtopics.DecryptionTrigger] = kpr.makeDecryptionTriggerValidator()

	return validators
}

func (kpr *keyper) makeDecryptionKeyValidator(db *kprdb.Queries) pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
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
			panic("unmarshalled non decryption key message in decryption key validator")
		}

		activationBlockNumber := medley.ActivationBlockNumberFromEpochID(key.epochID)
		dkgResultDB, err := db.GetDKGResultForBlockNumber(ctx, activationBlockNumber)
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
}

func (kpr *keyper) makeKeyShareValidator(db *kprdb.Queries) pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
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
			panic("unmarshalled non decryption key share message in decryption key share validator")
		}

		activationBlockNumber := medley.ActivationBlockNumberFromEpochID(keyShare.epochID)
		dkgResultDB, err := db.GetDKGResultForBlockNumber(ctx, activationBlockNumber)
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
}

func (kpr *keyper) makeEonPublicKeyValidator() pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
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
}

func (kpr *keyper) makeDecryptionTriggerValidator() pubsub.Validator {
	return func(ctx context.Context, peerID peer.ID, libp2pMessage *pubsub.Message) bool {
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
			panic("unmarshalled non decryption trigger message in decryption trigger validator")
		}
		blk := medley.ActivationBlockNumberFromEpochID(t.EpochID)
		collatorString, err := kpr.db.GetChainCollator(ctx, blk)
		if err == pgx.ErrNoRows {
			fmt.Printf("got decryption trigger with no collators for given block number: %d", blk)
			return false
		}
		if err != nil {
			fmt.Printf("error while getting collator from db for block nubmer: %d", blk)
			return false
		}

		collator, err := shdb.DecodeAddress(collatorString)
		if err != nil {
			fmt.Printf("error while converting collator from string to address: %s", collatorString)
			return false
		}

		trigger := (*shmsg.DecryptionTrigger)(t)
		signatureValid, err := trigger.VerifySignature(collator)
		if err != nil {
			fmt.Printf("error while verifying decryption trigger signature for epoch: %d", t.EpochID)
			return false
		}

		return signatureValid
	}
}

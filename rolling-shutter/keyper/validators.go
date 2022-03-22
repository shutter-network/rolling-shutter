package keyper

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func makeValidator[M shmsg.P2PMessage](registry map[string]pubsub.Validator, valFunc func(context.Context, M) bool) {
	var messProto M
	topic := messProto.Topic()

	_, exists := registry[topic]
	if exists {
		// This is likely not intended and happens when different messages return the same P2PMessage.Topic().
		// Currently a topic is mapped 1 to 1 to a message type (instead of using an envelope for unmarshalling)

		// Instead of silently overwriting the old validator, rather panic.
		// TODO maybe allow for chaining of successively registered validator functions per topic
		panic(fmt.Sprintf("Can't register more than one validator per topic (topic: '%s', message-type: '%s')", topic, reflect.TypeOf(messProto)))
	}

	registry[topic] = func(ctx context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
		var (
			key M
			ok  bool
		)

		message := p2p.Message{
			Topic:    *libp2pMessage.Topic,
			Message:  libp2pMessage.Data,
			SenderID: libp2pMessage.GetFrom().Pretty(),
		}
		if strings.Compare(message.Topic, topic) != 0 {
			// This should not happen, if so then we registered the validator function on the wrong topic
			log.Printf("topic mismatch (message-topic: '%s', validator-topic: '%s')", message.Topic, topic)
			return false
		}
		unmshl, err := message.Unmarshal()
		if err != nil {
			log.Printf("error while unmarshalling Message in validator (topic: '%s', error: '%s')", topic, err.Error())
			return false
		}

		key, ok = unmshl.(M)
		if !ok {
			// Either this is a programming error or someone sent the wrong message for that topic.
			log.Printf("type assertion failed while unmarshaling message (topic: '%s'). Received message of type '%s'. Is this a valid message with incorrect type for that topic?", topic, reflect.TypeOf(unmshl))
			return false
		}

		return valFunc(ctx, key)
	}
}

func (kpr *keyper) makeMessagesValidators() map[string]pubsub.Validator {
	validators := make(map[string]pubsub.Validator)

	makeValidator(validators, kpr.validateDecryptionKey)
	makeValidator(validators, kpr.validateDecryptionKeyShare)
	makeValidator(validators, kpr.validateEonPublicKey)
	makeValidator(validators, kpr.validateDecryptionTrigger)

	return validators
}

func (kpr *keyper) validateDecryptionKey(ctx context.Context, key *shmsg.DecryptionKey) bool {
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

func (kpr *keyper) validateDecryptionKeyShare(ctx context.Context, keyShare *shmsg.DecryptionKeyShare) bool {
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

func (kpr *keyper) validateEonPublicKey(_ context.Context, key *shmsg.EonPublicKey) bool {
	return key.GetInstanceID() == kpr.config.InstanceID
}

func (kpr *keyper) validateDecryptionTrigger(ctx context.Context, trigger *shmsg.DecryptionTrigger) bool {
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

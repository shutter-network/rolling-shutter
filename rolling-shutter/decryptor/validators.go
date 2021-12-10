package decryptor

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrtopics"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func (d *Decryptor) makeMessagesValidators() map[string]pubsub.Validator {
	return map[string]pubsub.Validator{
		dcrtopics.DecryptionSignature:           d.validateDecryptionSignature,
		dcrtopics.AggregatedDecryptionSignature: d.validateAggregatedDecryptionSignature,
		dcrtopics.CipherBatch:                   d.validateCipherBatch,
		dcrtopics.DecryptionKey:                 d.validateDecryptionKey,
	}
}

func (d *Decryptor) validateCipherBatch(ctx context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	p2pMessage := new(p2p.Message)
	if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
		return false
	}
	msg, err := unmarshalP2PMessage(p2pMessage)
	if err != nil {
		return false
	}

	if msg.GetInstanceID() != d.Config.InstanceID {
		return false
	}

	cipherBatch, ok := msg.(*cipherBatch)
	if !ok {
		return false
	}

	// check that it's signed by the collator
	activationBlockNumber := epochid.BlockNumber(cipherBatch.DecryptionTrigger.EpochID)
	collatorDBEntry, err := d.db.GetChainCollator(ctx, int64(activationBlockNumber))
	if err == pgx.ErrNoRows {
		log.Printf("error getting collator from db: %s", err)
		return false
	}
	if collatorDBEntry.Collator == "" {
		log.Printf("no collator for activation block number %d", activationBlockNumber)
		return false
	}
	collator, err := shdb.DecodeAddress(collatorDBEntry.Collator)
	if err != nil {
		log.Printf("invalid collator entry: %+v", collatorDBEntry)
		return false
	}
	ok, err = cipherBatch.DecryptionTrigger.VerifySignature(collator)
	if err != nil {
		log.Printf("failed to verify collator signature: %s", err)
		return false
	}
	if !ok {
		return false
	}

	// check the transaction hash matches the given transactions
	if !bytes.Equal(shmsg.HashTransactions(cipherBatch.Transactions), cipherBatch.DecryptionTrigger.TransactionsHash) {
		return false
	}

	return true
}

func (d *Decryptor) validateDecryptionKey(ctx context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	p2pMessage := new(p2p.Message)
	if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
		return false
	}
	msg, err := unmarshalP2PMessage(p2pMessage)
	if err != nil {
		return false
	}
	if msg.GetInstanceID() != d.Config.InstanceID {
		return false
	}

	key, ok := msg.(*decryptionKey)
	if !ok {
		return false
	}

	activationBlockNumber := epochid.BlockNumber(key.epochID)
	eonPublicKeyBytes, err := d.db.GetEonPublicKey(ctx, int64(activationBlockNumber))
	if err == pgx.ErrNoRows {
		log.Printf("received decryption key for epoch %d for which we don't have an eon public key", key.epochID)
		return false
	}
	if err != nil {
		log.Printf("error while getting eon public key from database for epoch ID %v", key.epochID)
		return false
	}
	eonPublicKey := new(shcrypto.EonPublicKey)
	err = eonPublicKey.Unmarshal(eonPublicKeyBytes)
	if err != nil {
		log.Printf("error while unmarshalling eon public key for epoch %v", key.epochID)
		return false
	}
	ok, err = shcrypto.VerifyEpochSecretKey(key.key, eonPublicKey, key.epochID)
	if err != nil {
		log.Printf("error while checking epoch secret key for epoch %v", key.epochID)
		return false
	}
	return ok
}

func (d *Decryptor) validateDecryptionSignature(ctx context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	p2pMessage := new(p2p.Message)
	if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
		return false
	}
	msg, err := unmarshalP2PMessage(p2pMessage)
	if err != nil {
		return false
	}

	if msg.GetInstanceID() != d.Config.InstanceID {
		return false
	}

	signature, ok := msg.(*decryptionSignature)
	if !ok {
		return false
	}

	activationBlockNumber := epochid.BlockNumber(signature.epochID)
	decryptorIndexes := signature.signers.GetIndexes()
	if len(decryptorIndexes) != 1 {
		return false
	}
	decryptorSetMember, err := d.db.GetDecryptorSetMember(ctx, dcrdb.GetDecryptorSetMemberParams{
		ActivationBlockNumber: int64(activationBlockNumber),
		Index:                 decryptorIndexes[0],
	})
	if err == pgx.ErrNoRows {
		return false
	}
	if err != nil {
		log.Printf("error while getting decryptor set member from database: %s", err)
		return false
	}
	if !decryptorSetMember.SignatureValid {
		return false
	}

	key := new(shbls.PublicKey)
	if err := key.Unmarshal(decryptorSetMember.BlsPublicKey); err != nil {
		return false
	}
	return shbls.Verify(signature.signature, key, signature.signedHash.Bytes())
}

func (d *Decryptor) validateAggregatedDecryptionSignature(ctx context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	p2pMessage := new(p2p.Message)
	if err := json.Unmarshal(libp2pMessage.Data, p2pMessage); err != nil {
		return false
	}
	msg, err := unmarshalP2PMessage(p2pMessage)
	if err != nil {
		return false
	}

	if msg.GetInstanceID() != d.Config.InstanceID {
		return false
	}

	signature, ok := msg.(*aggregatedDecryptionSignature)
	if !ok {
		return false
	}

	activationBlockNumber := epochid.BlockNumber(signature.epochID)
	decryptorIndexes := signature.signers.GetIndexes()
	if len(decryptorIndexes) == 0 {
		return false
	}
	decryptorSet, err := d.db.GetDecryptorSet(ctx, int64(activationBlockNumber))
	if err != nil {
		log.Printf("failed to get decryptor set from db for block number %d", activationBlockNumber)
		return false
	}

	keys := make([]*shbls.PublicKey, 0, len(decryptorIndexes))
	for _, decryptorIndex := range decryptorIndexes {
		decryptorSetMember, ok := dcrdb.SearchDecryptorSetRowsForIndex(decryptorSet, decryptorIndex)
		if !ok {
			log.Printf(
				"failed to find decryptor for activation block number %d and index %d in db",
				activationBlockNumber, decryptorIndex,
			)
			return false
		}
		if !decryptorSetMember.SignatureValid {
			return false
		}

		key := new(shbls.PublicKey)
		if err := key.Unmarshal(decryptorSetMember.BlsPublicKey); err != nil {
			log.Printf("failed to unmarshal BLS public key of decryptor %s in db", decryptorSetMember.Address)
			return false
		}
		keys = append(keys, key)
	}

	aggregatedKey := shbls.AggregatePublicKeys(keys)
	return shbls.Verify(signature.signature, aggregatedKey, signature.signedHash.Bytes())
}

package decryptor

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func handleDecryptionKeyInput(
	ctx context.Context,
	config Config,
	db *dcrdb.Queries,
	key *decryptionKey,
) ([]shmsg.P2PMessage, error) {
	eonPublicKeyBytes, err := db.GetEonPublicKey(ctx, medley.Uint64EpochIDToBytes(key.epochID))
	if err == pgx.ErrNoRows {
		return nil, errors.Errorf(
			"received decryption key for epoch %d for which we don't have an eon public key",
			key.epochID,
		)
	}
	if err != nil {
		return nil, err
	}
	eonPublicKey := new(shcrypto.EonPublicKey)
	err = eonPublicKey.Unmarshal(eonPublicKeyBytes)
	if err != nil {
		return nil, err
	}
	ok, err := checkEpochSecretKey(key.key, eonPublicKey, key.epochID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Errorf("received decryption key does not match eon public key for epoch %d", key.epochID)
	}

	keyBytes, _ := key.key.GobEncode()
	tag, err := db.InsertDecryptionKey(ctx, dcrdb.InsertDecryptionKeyParams{
		EpochID: medley.Uint64EpochIDToBytes(key.epochID),
		Key:     keyBytes,
	})
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple keys for same epoch %d", key.epochID)
		return nil, nil
	}
	return handleEpoch(ctx, config, db, key.epochID)
}

func handleCipherBatchInput(
	ctx context.Context,
	config Config,
	db *dcrdb.Queries,
	cipherBatch *cipherBatch,
) ([]shmsg.P2PMessage, error) {
	tag, err := db.InsertCipherBatch(ctx, dcrdb.InsertCipherBatchParams{
		EpochID:      medley.Uint64EpochIDToBytes(cipherBatch.EpochID),
		Transactions: cipherBatch.Transactions,
	})
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple cipherbatches for same epoch %d", cipherBatch.EpochID)
		return nil, nil
	}
	return handleEpoch(ctx, config, db, cipherBatch.EpochID)
}

// handleEpoch produces, store, and output a signature if we have both the cipher batch and key for given epoch.
func handleEpoch(
	ctx context.Context,
	config Config,
	db *dcrdb.Queries,
	epochID uint64,
) ([]shmsg.P2PMessage, error) {
	epochIDBytes := medley.Uint64EpochIDToBytes(epochID)
	cipherBatch, err := db.GetCipherBatch(ctx, epochIDBytes)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	decryptionKeyDB, err := db.GetDecryptionKey(ctx, epochIDBytes)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	log.Printf("decrypting batch for epoch %d", epochID)

	decryptionKey := new(shcrypto.EpochSecretKey)
	err = decryptionKey.GobDecode(decryptionKeyDB.Key)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid decryption key for epoch %d in db", epochID)
	}

	decryptedBatch := decryptCipherBatch(cipherBatch.Transactions, decryptionKey)
	signingData := DecryptionSigningData{
		InstanceID:     config.InstanceID,
		EpochID:        epochID,
		CipherBatch:    cipherBatch.Transactions,
		DecryptedBatch: decryptedBatch,
	}
	signedHash := signingData.Hash().Bytes()
	signatureBytes := signingData.Sign(config.SigningKey).Marshal()
	signerIndex := int64(0) // TODO: find this value

	insertParams := dcrdb.InsertDecryptionSignatureParams{
		EpochID:     epochIDBytes,
		SignedHash:  signedHash,
		SignerIndex: signerIndex,
		Signature:   signatureBytes,
	}
	tag, err := db.InsertDecryptionSignature(ctx, insertParams)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple signatures with same (epoch id, signer index): (%d, %d)",
			epochID, signerIndex)
		return nil, nil
	}

	msgs := []shmsg.P2PMessage{}
	// TODO: handle signer bitfield
	msgs = append(msgs, &shmsg.AggregatedDecryptionSignature{
		InstanceID:          config.InstanceID,
		EpochID:             epochID,
		SignedHash:          signedHash,
		AggregatedSignature: signatureBytes,
		SignerBitfield:      []byte(""), // TODO
	})
	return msgs, nil
}

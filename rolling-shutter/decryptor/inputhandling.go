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
	key *shmsg.DecryptionKey,
) ([]shmsg.P2PMessage, error) {
	tag, err := db.InsertDecryptionKey(ctx, dcrdb.InsertDecryptionKeyParams{
		EpochID: medley.Uint64EpochIDToBytes(key.EpochID),
		Key:     key.Key,
	})
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple keys for same epoch %d", key.EpochID)
		return nil, nil
	}
	return handleEpoch(ctx, config, db, key.EpochID)
}

func handleCipherBatchInput(
	ctx context.Context,
	config Config,
	db *dcrdb.Queries,
	cipherBatch *shmsg.CipherBatch,
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
	signingData := decryptionSigningData{
		instanceID:     0,
		epochID:        epochID,
		cipherBatch:    cipherBatch.Transactions,
		decryptedBatch: decryptedBatch,
	}
	signedHash := signingData.hash().Bytes()
	signatureBytes := signingData.sign(config.SigningKey).Marshal()
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
	// TODO: handle instanceID and signer bitfield
	msgs = append(msgs, &shmsg.AggregatedDecryptionSignature{
		InstanceID:          0,
		EpochID:             epochID,
		SignedHash:          signedHash,
		AggregatedSignature: signatureBytes,
		SignerBitfield:      []byte(""), // TODO
	})
	return msgs, nil
}

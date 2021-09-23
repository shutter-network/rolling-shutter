package decryptor

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
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
		EpochID: medley.Uint64EpochIDToBytes(cipherBatch.EpochID),
		Data:    cipherBatch.Data,
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

	key, err := db.GetDecryptionKey(ctx, epochIDBytes)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	signature, err := signBatch(ctx, cipherBatch, key)
	if err != nil {
		return nil, err
	}

	tag, err := db.InsertDecryptionSignature(ctx, dcrdb.InsertDecryptionSignatureParams(signature))
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple signatures with same (epoch id, signer index): (%d, %d)",
			signature.EpochID, signature.SignerIndex)
		return nil, nil
	}

	msgs := []shmsg.P2PMessage{}
	// TODO: handle instanceID and signer bitfield
	msgs = append(msgs, &shmsg.AggregatedDecryptionSignature{
		InstanceID:          0,
		EpochID:             medley.BytesEpochIDToUint64(signature.EpochID),
		SignedHash:          signature.SignedHash,
		AggregatedSignature: signature.Signature,
		SignerBitfield:      []byte(""),
	})
	return msgs, nil
}

func signBatch(
	_ context.Context, cipherBatch dcrdb.DecryptorCipherBatch, _ dcrdb.DecryptorDecryptionKey) (
	dcrdb.DecryptorDecryptionSignature,
	error) { //nolint //The error is always nil only because it is placeholder
	// TODO: handle signer index
	return dcrdb.DecryptorDecryptionSignature{
		EpochID:     cipherBatch.EpochID,
		SignedHash:  []byte("Placeholder"),
		SignerIndex: 0,
		Signature:   []byte("Placeholder"),
	}, nil
}

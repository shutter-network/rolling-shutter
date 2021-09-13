package decryptor

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

// handleInputs handles the inputs received on the input channel in an endless loop. It only
// returns if the context is done. Errors are logged.
func (d *Decryptor) handleInputs(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v := <-d.inputChannel:
			err := d.handleInput(ctx, v)
			if err != nil {
				log.Printf("error handling input: %s", err)
			}
		}
	}
}

func (d *Decryptor) handleInput(ctx context.Context, value interface{}) error {
	switch v := value.(type) {
	case *shmsg.DecryptionKey:
		return d.handleDecryptionKeyInput(ctx, v)
	case *shmsg.CipherBatch:
		return d.handleCipherBatchInput(ctx, v)
	default:
		return errors.Errorf("received input of invalid type: %T", v)
	}
}

func (d *Decryptor) handleDecryptionKeyInput(ctx context.Context, key *shmsg.DecryptionKey) error {
	tag, err := d.db.InsertDecryptionKey(ctx, dcrdb.InsertDecryptionKeyParams{
		EpochID: int64(key.EpochID),
		Key:     key.Key,
	})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple keys for same epoch %d", key.EpochID)
	}
	return d.handleEpoch(ctx, int64(key.EpochID))
}

func (d *Decryptor) handleCipherBatchInput(ctx context.Context, cipherBatch *shmsg.CipherBatch) error {
	tag, err := d.db.InsertCipherBatch(ctx, dcrdb.InsertCipherBatchParams{
		EpochID: int64(cipherBatch.EpochID),
		Data:    cipherBatch.Data,
	})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple cipherbatches for same epoch %d", cipherBatch.EpochID)
	}
	return d.handleEpoch(ctx, int64(cipherBatch.EpochID))
}

// handleEpoch produces, store, and output a signature if we have both the cipher batch and key for given epoch.
func (d *Decryptor) handleEpoch(ctx context.Context, epochID int64) error {
	cipherBatch, err := d.db.GetCipherBatch(ctx, epochID)
	if err == pgx.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	key, err := d.db.GetDecryptionKey(ctx, epochID)
	if err == pgx.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	signature, err := d.signBatch(ctx, cipherBatch, key)
	if err != nil {
		return err
	}

	tag, err := d.db.InsertDecryptionSignature(ctx, dcrdb.InsertDecryptionSignatureParams(signature))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to store multiple signatures with same (epoch id, signer index): (%d, %d)",
			signature.EpochID, signature.SignerIndex)
	}

	// TODO: handle instanceID and signer bitfield
	aggregatedSignature := &shmsg.AggregatedDecryptionSignature{
		InstanceID:          0,
		EpochID:             uint64(signature.EpochID),
		SignedHash:          signature.SignedHash,
		AggregatedSignature: signature.Signature,
		SignerBitfield:      []byte(""),
	}
	d.outputChannel <- aggregatedSignature

	return nil
}

func (d *Decryptor) signBatch(
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

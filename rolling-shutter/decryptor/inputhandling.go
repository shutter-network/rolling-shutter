package decryptor

import (
	"context"
	"log"

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
	return nil
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
	return nil
}

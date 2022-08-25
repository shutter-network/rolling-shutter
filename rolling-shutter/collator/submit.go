package collator

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shuttermint/collator/cltrdb"
)

func insertTx(ctx context.Context, dbpool *pgxpool.Pool, insertTxParams cltrdb.InsertTxParams) error {
	if len(insertTxParams.EpochID) != 8 {
		return errors.Errorf("EpochID must be exactly 8 bytes")
	}
	txEpoch := insertTxParams.EpochID

	return dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := cltrdb.New(tx)

		// Disallow starting the next epoch
		_, err := tx.Exec(ctx, "LOCK TABLE decryption_trigger IN SHARE MODE")
		if err != nil {
			return err
		}
		epoch, _, err := getNextEpochID(ctx, db)
		if err != nil {
			return err
		}
		if string(txEpoch) != string(epoch.Bytes()) {
			return errors.Errorf("transaction for past epoch")
		}
		return db.InsertTx(ctx, insertTxParams)
	})
}

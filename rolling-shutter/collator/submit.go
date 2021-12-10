package collator

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shuttermint/collator/cltrdb"
	"github.com/shutter-network/shutter/shuttermint/shdb"
)

func insertTx(ctx context.Context, dbpool *pgxpool.Pool, insertTxParams cltrdb.InsertTxParams) error {
	if len(insertTxParams.EpochID) != 8 {
		return errors.Errorf("EpochID must be exactly 8 bytes")
	}
	txEpoch := shdb.DecodeUint64(insertTxParams.EpochID)

	tx, err := dbpool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	db := cltrdb.New(tx)

	// Disallow starting the next epoch
	_, err = tx.Exec(ctx, "LOCK TABLE collator.decryption_trigger IN SHARE MODE")
	if err != nil {
		return err
	}
	epoch, err := getNextEpochID(ctx, db)
	if err != nil {
		return err
	}
	if txEpoch < epoch {
		return errors.Errorf("transaction for past epoch")
	}
	err = db.InsertTx(ctx, insertTxParams)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

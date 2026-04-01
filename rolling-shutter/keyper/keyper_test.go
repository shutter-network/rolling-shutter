package keyper

import (
	"context"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"gotest.tools/assert"

	obskeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	keyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func TestHandleOnChainKeyperSetChangesSkipsPastInvalidConfig(t *testing.T) {
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, keyperdb.Definition)
	t.Cleanup(dbclose)

	currentKeypers := []common.Address{
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		common.HexToAddress("0x0000000000000000000000000000000000000002"),
	}
	duplicateAddr := common.HexToAddress("0x0000000000000000000000000000000000000003")
	nextKeypers := []common.Address{
		common.HexToAddress("0x0000000000000000000000000000000000000004"),
		common.HexToAddress("0x0000000000000000000000000000000000000005"),
	}

	err := dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		q := keyperdb.New(tx)
		cq := obskeyperdb.New(tx)

		if err := q.InsertBatchConfig(ctx, keyperdb.InsertBatchConfigParams{
			KeyperConfigIndex:     1,
			Height:                1,
			Keypers:               shdb.EncodeAddresses(currentKeypers),
			Threshold:             2,
			Started:               true,
			ActivationBlockNumber: 10,
		}); err != nil {
			return err
		}
		if err := cq.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
			KeyperConfigIndex:     2,
			ActivationBlockNumber: 11,
			Keypers:               shdb.EncodeAddresses([]common.Address{duplicateAddr, duplicateAddr}),
			Threshold:             2,
		}); err != nil {
			return err
		}
		return cq.InsertKeyperSet(ctx, obskeyperdb.InsertKeyperSetParams{
			KeyperConfigIndex:     3,
			ActivationBlockNumber: 12,
			Keypers:               shdb.EncodeAddresses(nextKeypers),
			Threshold:             2,
		})
	})
	assert.NilError(t, err)

	kpr := &KeyperCore{
		config: &kprconfig.Config{
			Shuttermint: &kprconfig.ShuttermintConfig{
				DKGStartBlockDelta: 0,
			},
		},
	}

	err = dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return kpr.handleOnChainKeyperSetChanges(ctx, tx, 100)
	})
	assert.NilError(t, err)

	queries := keyperdb.New(dbpool)
	lastSent, err := queries.GetLastBatchConfigProcessed(ctx)
	assert.NilError(t, err)
	assert.Equal(t, lastSent, int64(2))

	err = dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return kpr.handleOnChainKeyperSetChanges(ctx, tx, 100)
	})
	assert.NilError(t, err)

	lastSent, err = queries.GetLastBatchConfigProcessed(ctx)
	assert.NilError(t, err)
	assert.Equal(t, lastSent, int64(3))

	msg, err := queries.GetNextShutterMessage(ctx)
	assert.NilError(t, err)
	assert.Check(t, strings.Contains(msg.Description, "config-index=3"))
}

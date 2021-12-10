package keyper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/keyper/shutterevents"
)

type ShuttermintDriver struct {
	shmcl            client.Client
	dbpool           *pgxpool.Pool
	shuttermintState *ShuttermintState
}

var errEmptyBlockchain = errors.New("empty shuttermint blockchain")

func getLastCommittedHeight(ctx context.Context, shmcl client.Client) (int64, error) {
	latestBlock, err := shmcl.Block(ctx, nil)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get last committed height of shuttermint chain")
	}
	if latestBlock.Block == nil || latestBlock.Block.LastCommit == nil {
		return 0, errEmptyBlockchain
	}
	return latestBlock.Block.LastCommit.Height, nil
}

// SyncAppWithDB queries the shutter tendermint app for events and inserts them into our database.
func SyncAppWithDB(
	ctx context.Context,
	shmcl client.Client,
	dbpool *pgxpool.Pool,
	shuttermintState *ShuttermintState) error {
	a := ShuttermintDriver{
		shmcl:            shmcl,
		dbpool:           dbpool,
		shuttermintState: shuttermintState,
	}
	return a.sync(ctx)
}

func (smdrv *ShuttermintDriver) sync(ctx context.Context) error {
	q := kprdb.New(smdrv.dbpool)
	oldMeta, err := q.TMGetSyncMeta(ctx)
	if err != nil {
		return err
	}
	currentBlock := oldMeta.CurrentBlock

	lastCommittedHeight, err := getLastCommittedHeight(ctx, smdrv.shmcl)
	if errors.Is(err, errEmptyBlockchain) {
		log.Printf("Shuttermint blockchain still empty")
		return nil
	}
	if err != nil {
		return err
	}

	if lastCommittedHeight == currentBlock {
		return nil
	}

	err = smdrv.fetchEvents(ctx, currentBlock, lastCommittedHeight)
	return err
}

func makeEvents(height int64, events []abcitypes.Event) []shutterevents.IEvent {
	var res []shutterevents.IEvent
	for _, ev := range events {
		x, err := shutterevents.MakeEvent(ev, height)
		if err != nil {
			log.Printf("Error: malformed event: %+v ev=%+v", err, ev)
		} else {
			res = append(res, x)
		}
	}
	return res
}

func (smdrv *ShuttermintDriver) fetchEvents(ctx context.Context, heightFrom, lastCommittedHeight int64) error {
	const perQuery = 500
	currentBlock := heightFrom
	logProgress := currentBlock+perQuery < lastCommittedHeight

	for currentBlock := heightFrom; currentBlock < lastCommittedHeight; currentBlock += perQuery {
		height := currentBlock + perQuery
		if height > lastCommittedHeight {
			height = lastCommittedHeight
		}
		query := fmt.Sprintf("tx.height >= %d and tx.height <= %d", currentBlock+1, height)
		if logProgress {
			log.Printf("Fetch events: query=%s targetHeight=%d", query, lastCommittedHeight)
		}
		// tendermint silently caps the perPage value at 100, make sure to stay below, otherwise
		// our exit condition is wrong and the log.Fatalf will trigger a panic below; see
		// https://github.com/shutter-network/shutter/issues/50
		perPage := 100
		page := 1
		total := 0
		txs := []*coretypes.ResultTx{}
		for {
			res, err := smdrv.shmcl.TxSearch(ctx, query, false, &page, &perPage, "")
			if err != nil {
				return errors.Wrap(err, "failed to fetch shuttermint txs")
			}

			total += len(res.Txs)
			txs = append(txs, res.Txs...)
			if page*perPage >= res.TotalCount {
				if total != res.TotalCount {
					log.Fatalf("internal error. got %d transactions, expected %d transactions from shuttermint for height %d..%d",
						total,
						res.TotalCount,
						currentBlock+1,
						lastCommittedHeight)
				}
				break
			}
			page++
		}
		err := smdrv.handleTransactions(ctx, txs, currentBlock, height, lastCommittedHeight)
		if err != nil {
			return err
		}
	}
	return nil
}

func (smdrv *ShuttermintDriver) innerHandleTransactions(
	ctx context.Context, queries *kprdb.Queries,
	txs []*coretypes.ResultTx,
	oldCurrentBlock, newCurrentBlock, lastCommittedHeight int64) error {
	oldMeta, err := queries.TMGetSyncMeta(ctx)
	if err != nil {
		return err
	}
	if oldMeta.CurrentBlock != oldCurrentBlock {
		return errors.Errorf(
			"wrong current block stored in database: stored=%d expected=%d",
			oldMeta.CurrentBlock, oldCurrentBlock)
	}
	err = smdrv.shuttermintState.Load(ctx, queries)
	if err != nil {
		return err
	}
	err = queries.TMSetSyncMeta(ctx, kprdb.TMSetSyncMetaParams{
		CurrentBlock:        newCurrentBlock,
		LastCommittedHeight: lastCommittedHeight,
		SyncTimestamp:       time.Now(),
	})
	if err != nil {
		return err
	}
	for _, tx := range txs {
		err = smdrv.shuttermintState.shiftPhases(ctx, queries, tx.Height)
		if err != nil {
			return err
		}

		events := makeEvents(tx.Height, tx.TxResult.GetEvents())
		for _, ev := range events {
			err = smdrv.shuttermintState.HandleEvent(ctx, queries, ev)
			if err != nil {
				return err
			}
		}
	}

	// XXX We should move the following call into the BeforeSaveHook
	err = smdrv.shuttermintState.shiftPhases(ctx, queries, newCurrentBlock)
	if err != nil {
		return err
	}
	err = smdrv.shuttermintState.BeforeSaveHook(ctx, queries)
	if err != nil {
		return err
	}
	return smdrv.shuttermintState.Save(ctx, queries)
}

func (smdrv *ShuttermintDriver) handleTransactions(
	ctx context.Context,
	txs []*coretypes.ResultTx,
	oldCurrentBlock, newCurrentBlock, lastCommittedHeight int64) error {
	tx, err := smdrv.dbpool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to start tx")
	}
	defer func() {
		err = tx.Rollback(ctx)
		if !errors.Is(err, pgx.ErrTxClosed) {
			smdrv.shuttermintState.Invalidate()
		}
	}()
	queries := kprdb.New(tx)
	err = smdrv.innerHandleTransactions(ctx, queries, txs, oldCurrentBlock, newCurrentBlock, lastCommittedHeight)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

package smobserver

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/shutterevents"
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
	shuttermintState *ShuttermintState,
) error {
	a := ShuttermintDriver{
		shmcl:            shmcl,
		dbpool:           dbpool,
		shuttermintState: shuttermintState,
	}
	return a.sync(ctx)
}

func (smdrv *ShuttermintDriver) sync(ctx context.Context) error {
	q := database.New(smdrv.dbpool)
	oldMeta, err := q.TMGetSyncMeta(ctx)
	if err != nil {
		return err
	}
	currentBlock := oldMeta.CurrentBlock

	lastCommittedHeight, err := getLastCommittedHeight(ctx, smdrv.shmcl)
	if errors.Is(err, errEmptyBlockchain) {
		log.Info().Msg("Shuttermint blockchain still empty")
		return nil
	}
	if err != nil {
		return err
	}

	if lastCommittedHeight == currentBlock {
		return nil
	}

	err = smdrv.fetchEvents2(ctx, currentBlock, lastCommittedHeight)
	return err
}

func makeEvents(height int64, events []abcitypes.Event) []shutterevents.IEvent {
	var res []shutterevents.IEvent
	for _, ev := range events {
		x, err := shutterevents.MakeEvent(ev, height)
		if err != nil {
			log.Error().Err(err).Str("event", ev.String()).Msg("malformed event")
		} else {
			res = append(res, x)
		}
	}
	return res
}

func (smdrv *ShuttermintDriver) fetchEvents2(ctx context.Context, heightFrom, lastCommittedHeight int64) error {
	for currentBlock := heightFrom + 1; currentBlock < lastCommittedHeight; currentBlock++ {
		results, err := smdrv.shmcl.BlockResults(ctx, &currentBlock)
		if err != nil {
			return err
		}
		err = smdrv.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
			return smdrv.handleBlock(ctx, database.New(tx), results, lastCommittedHeight)
		})
		if err != nil {
			smdrv.shuttermintState.Invalidate()
			return err
		}
	}
	return nil
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
			log.Info().Str("query", query).Int64("target-height", lastCommittedHeight).Msg("fetch events")
		}
		// tendermint silently caps the perPage value at 100, make sure to stay below, otherwise
		// our exit condition is wrong and the log.Fatal will trigger a panic below; see
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
					log.Fatal().
						Int("num-fetched-transactions", total).
						Int("num-expected-transactions", res.TotalCount).
						Int64("from-height", currentBlock+1).
						Int64("to-height", lastCommittedHeight).
						Msg("internal error. fetched fewer shuttermint transactions than expected")
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

func (smdrv *ShuttermintDriver) handleBlock(
	ctx context.Context,
	queries *database.Queries,
	block *coretypes.ResultBlockResults,
	lastCommittedHeight int64,
) error {
	oldMeta, err := queries.TMGetSyncMeta(ctx)
	if err != nil {
		return err
	}
	if block.Height != oldMeta.CurrentBlock+1 {
		return errors.Errorf(
			"blocks should be handled in order, expected %d, got %d",
			oldMeta.CurrentBlock+1,
			block.Height,
		)
	}

	err = queries.TMSetSyncMeta(ctx, database.TMSetSyncMetaParams{
		CurrentBlock:        block.Height,
		LastCommittedHeight: lastCommittedHeight,
		SyncTimestamp:       time.Now(),
	})
	if err != nil {
		return err
	}

	err = smdrv.shuttermintState.shiftPhases(ctx, queries, block.Height)
	if err != nil {
		return err
	}

	handleEvents := func(abciEvents []abcitypes.Event) error {
		events := makeEvents(block.Height, abciEvents)
		for _, ev := range events {
			err = smdrv.shuttermintState.HandleEvent(ctx, queries, ev)
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = handleEvents(block.BeginBlockEvents)
	if err != nil {
		return err
	}
	for _, txres := range block.TxsResults {
		err = handleEvents(txres.Events)
		if err != nil {
			return err
		}
	}
	err = handleEvents(block.EndBlockEvents)
	if err != nil {
		return err
	}

	return nil
}

func (smdrv *ShuttermintDriver) innerHandleTransactions(
	ctx context.Context, queries *database.Queries,
	txs []*coretypes.ResultTx,
	oldCurrentBlock, newCurrentBlock, lastCommittedHeight int64,
) error {
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
	err = queries.TMSetSyncMeta(ctx, database.TMSetSyncMetaParams{
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
	oldCurrentBlock, newCurrentBlock, lastCommittedHeight int64,
) error {
	err := smdrv.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return smdrv.innerHandleTransactions(
			ctx, database.New(tx), txs, oldCurrentBlock, newCurrentBlock, lastCommittedHeight,
		)
	})
	if err != nil {
		smdrv.shuttermintState.Invalidate()
	}
	return err
}

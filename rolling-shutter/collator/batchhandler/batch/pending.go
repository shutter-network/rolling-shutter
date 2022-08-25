package batch

import (
	"context"
	"time"

	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type Pending struct {
	previous  State
	processed bool
}

func (tr *Pending) StateEnum() StateEnum {
	return PendingState
}

func (tr *Pending) Process(batch *Batch) *StateChangeResult {
	if tr.processed {
		return nil
	}
	return &StateChangeResult{
		EpochID:               batch.EpochID(),
		FromState:             tr.previous.StateEnum(),
		ToState:               tr.StateEnum(),
		P2PMessages:           []shmsg.P2PMessage{},
		SequencerTransactions: []txtypes.TxData{},
		Errors:                []StateChangeError{},
	}
}

func (tr *Pending) Post(batch *Batch) State {
	// FIXME the block query and ApplyTx
	// calls could be problematic,
	// because they can block the transition
	// function for quite some time

	ctx := context.Background()

	// first apply all the pooled transactions
	for _, tx := range batch.txPool.Pop() {
		err := batch.ApplyTx(ctx, tx)
		if err != nil {
			batch.Log().Error().Err(err).Msg("error while applying cached transaction")
			err = errors.Wrap(err, "transaction could not be applied")
			tx.Result <- transaction.Result{Err: err, Success: false}
			close(tx.Result)
		}
	}
	tr.processed = true
	return tr
}

func (tr *Pending) OnStateChangePrevious(_ *Batch, _ StateChangeResult) State {
	return tr
}

func (tr *Pending) OnEpochTick(batch *Batch, _ time.Time) State {
	batch.Log().Debug().Msg("received epoch tick in pending state")
	return &Committed{previous: tr}
}

func (tr *Pending) OnDecryptionKey(_ *Batch, _ []byte) State {
	return tr
}

func (tr *Pending) OnTransaction(batch *Batch, tx *transaction.Pending) State {
	ctx := context.Background()

	// apply incoming txs directly while we wait for
	// the next epoch tick

	// FIXME the  ApplyTx
	// calls could be problematic,
	// because they can block the transition
	// function for quite some time
	// ->> we could introduce a

	err := batch.ApplyTx(ctx, tx)
	if err != nil {
		batch.Log().Debug().Err(err).Msg("error while applying transaction")
		err = errors.Wrap(err, "transaction could not be applied")
		tx.Result <- transaction.Result{Err: err, Success: false}
		close(tx.Result)
	}
	return tr
}

func (tr *Pending) OnBatchConfirmation(_ *Batch, _ epochid.EpochID) State {
	return tr
}

func (tr *Pending) OnStop(_ *Batch) State {
	return &Stopping{previous: tr}
}

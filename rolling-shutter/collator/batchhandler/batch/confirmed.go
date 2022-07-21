package batch

import (
	"time"

	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type Confirmed struct {
	previous  State
	processed bool
}

func (tr Confirmed) StateEnum() StateEnum {
	return ConfirmedState
}

func (tr Confirmed) Process(batch *Batch) *StateChangeResult {
	if tr.processed {
		return nil
	}
	for _, tx := range batch.committedTxs.Transactions() {
		tx.Result <- transaction.Result{Err: nil, Success: true}
		close(tx.Result)
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

func (tr Confirmed) Post(_ *Batch) State {
	tr.processed = true
	return tr
}

func (tr Confirmed) OnStateChangePrevious(batch *Batch, stateChange StateChangeResult) State {
	return tr
}

func (tr Confirmed) OnEpochTick(batch *Batch, tickTime time.Time) State {
	return tr
}

func (tr Confirmed) OnDecryptionKey(batch *Batch, decryptionKey []byte) State {
	return tr
}

func (tr Confirmed) OnTransaction(batch *Batch, tx *transaction.Pending) State {
	return tr
}

func (tr Confirmed) OnBatchConfirmation(batch *Batch, epochID epochid.EpochID) State {
	return tr
}

func (tr Confirmed) OnStop(batch *Batch) State {
	return Stopping{previous: tr}
}

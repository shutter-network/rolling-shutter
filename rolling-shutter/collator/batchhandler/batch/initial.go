package batch

import (
	"time"

	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type Initial struct {
	processed bool
}

func (tr Initial) StateEnum() StateEnum {
	return InitialState
}

func (tr Initial) Process(batch *Batch) *StateChangeResult {
	if tr.processed {
		return nil
	}
	return &StateChangeResult{
		EpochID:               batch.EpochID(),
		FromState:             NoState,
		ToState:               tr.StateEnum(),
		P2PMessages:           []shmsg.P2PMessage{},
		SequencerTransactions: []txtypes.TxData{},
		Errors:                []StateChangeError{},
	}
}

func (tr Initial) Post(batch *Batch) State {
	if batch.previous == nil {
		return Pending{previous: tr}
	}
	batch.subscription = batch.previous.Broker.Subscribe()
	tr.processed = true
	return tr
}

func (tr Initial) OnStateChangePrevious(batch *Batch, stateChange StateChangeResult) State {
	// previous state change
	if stateChange.ToState == ConfirmedState {
		// if the previous batch was committed,
		// we can now transition to state pending
		// for this batch
		batch.Log().Debug().Msg("previous batch cofirmed")
		return Pending{previous: tr}
	}
	return tr
}

func (tr Initial) OnEpochTick(batch *Batch, tickTime time.Time) State {
	return tr
}

func (tr Initial) OnDecryptionKey(batch *Batch, decryptionKey []byte) State {
	return tr
}

func (tr Initial) OnTransaction(batch *Batch, tx *transaction.Pending) State {
	// enqueue / sort the tx into the batches tx pool
	// without verification
	batch.txPool.Enqueue(tx)
	return tr
}

func (tr Initial) OnBatchConfirmation(batch *Batch, epochID epochid.EpochID) State {
	return tr
}

func (tr Initial) OnStop(batch *Batch) State {
	return Stopping{previous: tr}
}

package batch

import (
	"time"

	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type Stopping struct {
	previous  State
	processed bool
}

func (tr *Stopping) StateEnum() StateEnum { return StoppingState }
func (tr *Stopping) String() string       { return "stopping" }

func (tr *Stopping) Process(batch *Batch) *StateChangeResult {
	if tr.processed {
		return nil
	}
	if batch.subscription != nil {
		batch.previous.Broker.Unsubscribe(batch.subscription)
		batch.subscription = nil
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

func (tr *Stopping) Post(batch *Batch) State {
	tr.processed = true
	return &Stopped{previous: tr}
}

func (tr *Stopping) OnStateChangePrevious(batch *Batch, stateChange StateChangeResult) State {
	return tr
}

func (tr *Stopping) OnEpochTick(batch *Batch, tickTime time.Time) State { return tr }

func (tr *Stopping) OnDecryptionKey(batch *Batch, decryptionKey []byte) State {
	return tr
}

func (tr *Stopping) OnTransaction(batch *Batch, tx *transaction.Pending) State {
	err := errors.New("the batch this transaction is signed for has already been committed")
	tx.Result <- transaction.Result{Err: err, Success: false}
	close(tx.Result)
	return tr
}

func (tr *Stopping) OnBatchConfirmation(batch *Batch, epochID epochid.EpochID) State {
	return tr
}

func (tr *Stopping) OnStop(batch *Batch) State { return &Stopped{previous: tr} }

type Stopped struct {
	processed bool
	previous  State
}

func (tr *Stopped) StateEnum() StateEnum { return NoState }
func (tr *Stopped) String() string       { return "nostate" }
func (tr *Stopped) Process(batch *Batch) *StateChangeResult {
	return &StateChangeResult{
		EpochID:               batch.EpochID(),
		FromState:             tr.previous.StateEnum(),
		ToState:               tr.StateEnum(),
		P2PMessages:           []shmsg.P2PMessage{},
		SequencerTransactions: []txtypes.TxData{},
		Errors:                []StateChangeError{},
	}
}

func (tr *Stopped) Post(batch *Batch) State {
	tr.processed = true
	return tr
}

func (tr *Stopped) OnStateChangePrevious(batch *Batch, stateChange StateChangeResult) State {
	return tr
}
func (tr *Stopped) OnEpochTick(batch *Batch, tickTime time.Time) State       { return tr }
func (tr *Stopped) OnDecryptionKey(batch *Batch, decryptionKey []byte) State { return tr }
func (tr *Stopped) OnTransaction(batch *Batch, tx *transaction.Pending) State {
	err := errors.New("the batch this transaction is signed for has already been committed")
	tx.Result <- transaction.Result{Err: err, Success: false}
	close(tx.Result)
	return tr
}

func (tr *Stopped) OnBatchConfirmation(batch *Batch, epochID epochid.EpochID) State {
	return tr
}
func (tr *Stopped) OnStop(batch *Batch) State { return tr }

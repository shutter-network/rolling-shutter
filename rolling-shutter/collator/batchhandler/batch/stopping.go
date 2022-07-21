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

func (tr Stopping) StateEnum() StateEnum { return StoppingState }
func (tr Stopping) String() string       { return "stopping" }

func (tr Stopping) Process(batch *Batch) *StateChangeResult {
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

func (tr Stopping) Post(batch *Batch) State {
	tr.processed = true
	return StoppedTransition{}
}

func (tr Stopping) OnStateChangePrevious(batch *Batch, stateChange StateChangeResult) State {
	return tr
}

func (tr Stopping) OnEpochTick(batch *Batch, tickTime time.Time) State { return tr }

func (tr Stopping) OnDecryptionKey(batch *Batch, decryptionKey []byte) State {
	return tr
}

func (tr Stopping) OnTransaction(batch *Batch, tx *transaction.Pending) State {
	err := errors.New("the batch this transaction is signed for has already been committed")
	tx.Result <- transaction.Result{Err: err, Success: false}
	close(tx.Result)
	return tr
}

func (tr Stopping) OnBatchConfirmation(batch *Batch, epochID epochid.EpochID) State {
	return tr
}

func (tr Stopping) OnStop(batch *Batch) State { return StoppedTransition{} }

type StoppedTransition struct{}

func (tr StoppedTransition) StateEnum() StateEnum { return NoState }
func (tr StoppedTransition) String() string       { return "nostate" }
func (tr StoppedTransition) Process(batch *Batch) *StateChangeResult {
	return nil
}
func (tr StoppedTransition) Post(batch *Batch) State { return tr }
func (tr StoppedTransition) OnStateChangePrevious(batch *Batch, stateChange StateChangeResult) State {
	return tr
}
func (tr StoppedTransition) OnEpochTick(batch *Batch, tickTime time.Time) State       { return tr }
func (tr StoppedTransition) OnDecryptionKey(batch *Batch, decryptionKey []byte) State { return tr }
func (tr StoppedTransition) OnTransaction(batch *Batch, tx *transaction.Pending) State {
	err := errors.New("the batch this transaction is signed for has already been committed")
	tx.Result <- transaction.Result{Err: err, Success: false}
	close(tx.Result)
	return tr
}

func (tr StoppedTransition) OnBatchConfirmation(batch *Batch, epochID epochid.EpochID) State {
	return tr
}
func (tr StoppedTransition) OnStop(batch *Batch) State { return tr }

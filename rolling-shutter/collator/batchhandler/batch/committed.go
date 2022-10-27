package batch

import (
	"errors"
	"time"

	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type Committed struct {
	previous  State
	processed bool
}

func (tr *Committed) StateEnum() StateEnum {
	return CommittedState
}

func (tr *Committed) Process(batch *Batch) *StateChangeResult {
	if tr.processed {
		return nil
	}
	msgs := []shmsg.P2PMessage{&shmsg.DecryptionTrigger{
		InstanceID:       batch.instanceID,
		EpochID:          batch.EpochID().Bytes(),
		BlockNumber:      batch.L1BlockNumber,
		TransactionsHash: batch.Hash(),
	}}
	return &StateChangeResult{
		EpochID:               batch.EpochID(),
		FromState:             tr.previous.StateEnum(),
		ToState:               tr.StateEnum(),
		P2PMessages:           msgs,
		SequencerTransactions: []txtypes.TxInner{},
		Errors:                []StateChangeError{},
	}
}

func (tr *Committed) Post(_ *Batch) State {
	tr.processed = true
	return tr
}

func (tr *Committed) OnStateChangePrevious(_ *Batch, _ StateChangeResult) State {
	return tr
}

func (tr *Committed) OnEpochTick(_ *Batch, _ time.Time) State {
	return tr
}

func (tr *Committed) OnDecryptionKey(batch *Batch, decryptionKey []byte) State {
	batch.decryptionKey = decryptionKey
	return &Decrypted{previous: tr}
}

func (tr *Committed) OnTransaction(_ *Batch, tx *transaction.Pending) State {
	err := errors.New("the batch this transaction is signed for has already been committed")
	tx.Result <- transaction.Result{Err: err, Success: false}
	close(tx.Result)
	return tr
}

func (tr *Committed) OnBatchConfirmation(_ *Batch, _ epochid.EpochID) State {
	return tr
}

func (tr *Committed) OnStop(_ *Batch) State {
	return &Stopping{previous: tr}
}

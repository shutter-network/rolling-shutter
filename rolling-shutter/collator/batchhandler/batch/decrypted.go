package batch

import (
	"math/big"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type Decrypted struct {
	previous  State
	processed bool
}

func (tr *Decrypted) StateEnum() StateEnum {
	return DecryptedState
}

func (tr *Decrypted) Process(batch *Batch) *StateChangeResult {
	if tr.processed {
		return nil
	}
	log.Debug().Int("num-shutter-txs", batch.committedTxs.Len()).Msg("construction BatchTx")
	ts := time.Now().Unix()
	btxData := &txtypes.BatchTx{
		ChainID:       batch.ChainID,
		DecryptionKey: batch.decryptionKey,
		BatchIndex:    batch.Index(),
		L1BlockNumber: batch.L1BlockNumber,
		Timestamp:     big.NewInt(ts),
		Transactions:  batch.committedTxs.Bytes(),
	}
	txs := make([]txtypes.TxData, 1)
	txs[0] = btxData
	return &StateChangeResult{
		EpochID:               batch.EpochID(),
		FromState:             tr.previous.StateEnum(),
		ToState:               tr.StateEnum(),
		P2PMessages:           []shmsg.P2PMessage{},
		SequencerTransactions: txs,
		Errors:                []StateChangeError{},
	}
}

func (tr *Decrypted) Post(_ *Batch) State {
	tr.processed = true
	return tr
}

func (tr *Decrypted) OnStateChangePrevious(_ *Batch, _ StateChangeResult) State {
	return tr
}

func (tr *Decrypted) OnEpochTick(_ *Batch, _ time.Time) State {
	return tr
}

func (tr *Decrypted) OnDecryptionKey(_ *Batch, _ []byte) State {
	return tr
}

func (tr *Decrypted) OnTransaction(_ *Batch, tx *transaction.Pending) State {
	err := errors.New("the batch this transaction is signed for has already been committed")
	tx.Result <- transaction.Result{Err: err, Success: false}
	close(tx.Result)
	return tr
}

func (tr *Decrypted) OnBatchConfirmation(batch *Batch, epochID epochid.EpochID) State {
	if epochid.Equal(batch.EpochID(), epochID) {
		return &Confirmed{previous: tr}
	}
	return tr
}

func (tr *Decrypted) OnStop(_ *Batch) State {
	return &Stopping{previous: tr}
}

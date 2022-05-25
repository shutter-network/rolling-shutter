package batch

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/shutter-network/txtypes/types"
)

func NewPendingTx(signer txtypes.Signer, txBytes []byte) (*PendingTransaction, error) {
	var tx txtypes.Transaction
	err := tx.UnmarshalBinary(txBytes)
	if err != nil {
		return nil, err
	}

	sender, err := signer.Sender(&tx)
	if err != nil {
		return nil, err
	}

	pendingTx := &PendingTransaction{
		txBytes: txBytes,
		tx:      &tx,
		sender:  sender,
	}
	return pendingTx, nil
}

// PendingTransaction is a wrapper struct that associates additional
// data to an incoming shutter transaction from a user.
// It is used to keep track of sender, gas-fees and receive-time of
// a shutter transaction.
type PendingTransaction struct {
	tx       *txtypes.Transaction
	txBytes  []byte
	sender   common.Address
	minerFee *big.Int
	gasCost  *big.Int
	time     time.Time
}

// SetReceived sets the time an incoming transaction was
// received. It should be called with a `t=nil` value immediatelY
// after the transaction was received.
func (pt *PendingTransaction) SetReceived(t *time.Time) {
	if t == nil {
		pt.time = time.Now()
		return
	}
	pt.time = *t
}

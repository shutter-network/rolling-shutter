package transaction

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/shutter-network/txtypes/types"
)

type Result struct {
	Success bool
	Err     error
}

// Pending is a wrapper struct that associates additional
// data to an incoming shutter transaction from a user.
// It is used to keep track of sender, gas-fees and receive-time of
// a shutter transaction.
type Pending struct {
	Tx          *txtypes.Transaction
	TxBytes     []byte
	Sender      common.Address
	MinerFee    *big.Int
	GasCost     *big.Int
	ReceiveTime time.Time
	Result      chan Result
}

func NewPending(signer txtypes.Signer, txBytes []byte, receiveTime time.Time) (*Pending, error) {
	var tx txtypes.Transaction
	err := tx.UnmarshalBinary(txBytes)
	if err != nil {
		return nil, err
	}

	sender, err := signer.Sender(&tx)
	if err != nil {
		return nil, err
	}

	pendingTx := &Pending{
		Tx:          &tx,
		TxBytes:     txBytes,
		Sender:      sender,
		MinerFee:    &big.Int{},
		GasCost:     &big.Int{},
		ReceiveTime: receiveTime,
		// buffered channel since this resembles a "Promise"
		Result: make(chan Result, 1),
	}
	return pendingTx, nil
}

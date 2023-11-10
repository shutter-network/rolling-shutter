package database

import (
	txtypes "github.com/shutter-network/txtypes/types"
)

// UnmarshalTransactions unmarshals the given slice of transactions. It returns an error if any of
// the transactions cannot be unmarshaled.
func UnmarshalTransactions(txs []Transaction) ([]txtypes.Transaction, error) {
	var unmarshalledTxs []txtypes.Transaction

	for _, t := range txs {
		tx := txtypes.Transaction{}
		err := tx.UnmarshalBinary(t.TxBytes)
		if err != nil {
			return nil, err
		}
		unmarshalledTxs = append(unmarshalledTxs, tx)
	}
	return unmarshalledTxs, nil
}

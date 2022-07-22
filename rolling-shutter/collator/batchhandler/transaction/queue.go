package transaction

import (
	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func NewTransactionQueue() *TransactionQueue {
	return &TransactionQueue{
		txqueue: make([]*Pending, 0),
		Senders: make(map[common.Address]bool, 0),
	}
}

// TransactionQueue is a container struct allowing
// for append-only operation on a list of `Pending`.
// TransactionQueue implements convenience methods to
// join two queues, generate the hash of all queued transactions
// and show the set of all transactions sender addresses.
type TransactionQueue struct {
	txqueue []*Pending
	Senders map[common.Address]bool
}

func (q *TransactionQueue) JoinRight(other *TransactionQueue) *TransactionQueue {
	txqueue := make([]*Pending, q.Len()+other.Len())
	n := copy(txqueue, q.txqueue)
	copy(txqueue[n:], other.txqueue)

	senders := make(map[common.Address]bool, len(q.Senders))
	for addr, v := range q.Senders {
		senders[addr] = v
	}
	for addr, v := range other.Senders {
		senders[addr] = v
	}
	return &TransactionQueue{txqueue: txqueue, Senders: senders}
}

func (q *TransactionQueue) Hash() []byte {
	txHashes := make([][]byte, q.Len())
	for i, t := range q.Transactions() {
		txHashes[i] = t.Hash().Bytes()
	}
	return shmsg.HashTransactions(txHashes)
}

func (q *TransactionQueue) Transactions() []*txtypes.Transaction {
	txs := make([]*txtypes.Transaction, len(q.txqueue))
	for i, p := range q.txqueue {
		txs[i] = p.Tx
	}
	return txs
}

func (q *TransactionQueue) Bytes() [][]byte {
	l := make([][]byte, q.Len())
	for i, tx := range q.txqueue {
		l[i] = tx.TxBytes
	}
	return l
}

func (q *TransactionQueue) Len() int { return len(q.txqueue) }

func (q *TransactionQueue) Enqueue(tx *Pending) {
	q.Senders[tx.Sender] = true
	q.txqueue = append(q.txqueue, tx)
}

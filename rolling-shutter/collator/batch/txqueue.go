package batch

import (
	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func NewTransactionQueue() *TransactionQueue {
	return &TransactionQueue{
		txqueue: make([]*PendingTransaction, 0),
		Senders: make(map[common.Address]bool, 0),
	}
}

// TransactionQueue is a container struct allowing
// for append-only operation on a list of `PendingTransaction`.
// TransactionQueue implements convenience methods to
// join two queues, generate the hash of all queued transactions
// and show the set of all transactions sender addresses.
type TransactionQueue struct {
	txqueue []*PendingTransaction
	Senders map[common.Address]bool
}

func (q *TransactionQueue) JoinRight(other *TransactionQueue) *TransactionQueue {
	txqueue := make([]*PendingTransaction, q.Len()+other.Len())
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
		txs[i] = p.tx
	}
	return txs
}

func (q *TransactionQueue) Bytes() [][]byte {
	l := make([][]byte, q.Len())
	for i, tx := range q.txqueue {
		l[i] = tx.txBytes
	}
	return l
}

func (q *TransactionQueue) Len() int { return len(q.txqueue) }

func (q *TransactionQueue) Enqueue(tx *PendingTransaction) {
	q.Senders[tx.sender] = true
	q.txqueue = append(q.txqueue, tx)
}

// TotalByteSize returns the sum of the size of all transactions measured in bytes.
func (q *TransactionQueue) TotalByteSize() int {
	bs := q.Bytes()
	s := 0
	for _, b := range bs {
		s += len(b)
	}
	return s
}

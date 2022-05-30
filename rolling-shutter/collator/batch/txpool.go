package batch

import (
	"container/heap"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/shutter-network/txtypes/types"
)

// SortedUint64s represents a set of batch indices
// that remain sorted upon inserting and removing
// values to the set.
type SortedUint64s []uint64

func (bt SortedUint64s) Len() int { return len(bt) }
func (bt *SortedUint64s) Swap(i, j int) {
	(*bt)[i], (*bt)[j] = (*bt)[j], (*bt)[i]
}
func (bt SortedUint64s) Less(i, j int) bool { return bt[i] < bt[j] }
func (bt SortedUint64s) Search(x uint64) int {
	return sort.Search(bt.Len(), func(i int) bool { return bt[i] >= x })
}

// searchDoesExist returns the index where x should be
// inserted in order for the internal array to remain in sorted order.
// The secondary return value indicates wether x is already
// present in Batches.
func (bt SortedUint64s) searchDoesExist(x uint64) (int, bool) {
	i := bt.Search(x)
	if i < bt.Len() && bt[i] == x {
		// x is present at index i
		return i, true
	}
	// x is not present in array
	// but i is the index where it would be inserted.
	return i, false
}

func (bt SortedUint64s) ToUint64s() []uint64 {
	return bt
}

// Has indicates wether x is already present in Batches.
func (bt SortedUint64s) Has(x uint64) bool {
	if bt.Len() == 0 {
		return false
	}
	_, exists := bt.searchDoesExist(x)
	return exists
}

// Removes removes x from SortedUint64s.
// Since a value can only be inserted once, the remove operation
// will leave SortedUint64s without any entry with value x.
func (bt *SortedUint64s) Remove(x uint64) {
	if bt.Len() == 0 {
		return
	}
	i, exists := bt.searchDoesExist(x)
	if exists {
		*bt = append((*bt)[:i], (*bt)[i+1:]...)
	}
}

// Insert inserts x into SortedUint64s, if x doesn't yet exist in SortedUint64s.
// If x already exists, Insert does nothing.
// The insert operation will leave SortedUint64s in sorted order.
func (bt *SortedUint64s) Insert(x uint64) {
	var i int
	if bt.Len() == 0 {
		i = 0
	} else {
		var exists bool
		i, exists = bt.searchDoesExist(x)
		if exists {
			return
		}
	}
	*bt = append(*bt, 0)
	copy((*bt)[i+1:], (*bt)[i:])
	(*bt)[i] = x
}

// SortableTransactions implements the container/heap interface.
// It allows to sort `PendingTransaction` transactions by Nonce,
// and if the nonces are the same the transaction with
// earlier receival time will come first in the sorted slice.
type SortableTransactions []*PendingTransaction

func (s SortableTransactions) Len() int { return len(s) }
func (s SortableTransactions) Less(i, j int) bool {
	// If the nonces are equal, use the time the transaction was last seen for
	// deterministic sorting
	if s[i].tx.Nonce() == s[j].tx.Nonce() {
		return s[j].receiveTime.Before(s[i].receiveTime)
	}
	return s[i].tx.Nonce() < s[j].tx.Nonce()
}
func (s SortableTransactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *SortableTransactions) Push(x interface{}) {
	*s = append(*s, x.(*PendingTransaction))
}

func (s *SortableTransactions) Pop() interface{} {
	old := *s
	n := len(old)
	if n == 0 {
		return nil
	}
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

func (s *SortableTransactions) Peek() interface{} {
	old := *s
	n := len(old)
	if n == 0 {
		return nil
	}
	return old[n-1]
}

func NewTransactionPool(signer txtypes.Signer) *TransactionPool {
	return &TransactionPool{
		txs:          make(map[common.Address]*SortableTransactions, 0),
		batchSenders: make(map[uint64]map[common.Address]bool, 0),
		signer:       signer,
		batches:      make(SortedUint64s, 0),
	}
}

// TransactionPool represents a set of transactions that can pop
// (remove and return) transactions in a profit-maximizing sorted order per batch.
type TransactionPool struct {
	signer txtypes.Signer
	mux    sync.Mutex
	// Per account nonce-sorted list of transactions
	txs map[common.Address]*SortableTransactions
	// The set of tx sender addresses per batch
	batchSenders map[uint64]map[common.Address]bool
	// The sorted list of batch-indices currently in the Pool
	batches SortedUint64s
}

func (t *TransactionPool) Batches() SortedUint64s {
	t.mux.Lock()
	defer t.mux.Unlock()

	return t.batches
}

// Senders returns the set of all sender addresses that have transactions
// pooled for that `batchIndex`.
func (t *TransactionPool) Senders(batchIndex uint64) []common.Address {
	t.mux.Lock()
	defer t.mux.Unlock()

	senders := make([]common.Address, 0)
	sendersMap, exists := t.batchSenders[batchIndex]
	if len(sendersMap) > 0 && exists {
		for s := range sendersMap {
			senders = append(senders, s)
		}
	}
	return senders
}

// Pop returns a list of all transactions for that `batchIndex`
// and remove it from the pool. If there are no transactions for
// that `batchIndex` it returns an empty list.
// If there were transactions for batches prior to the popped batch,
// those transactions are removed from the pool - they are NOT
// returned by Pop.
func (t *TransactionPool) Pop(batchIndex uint64) []*PendingTransaction {
	t.mux.Lock()
	defer t.mux.Unlock()
	txs := make([]*PendingTransaction, 0)
	senders, ok := t.batchSenders[batchIndex]
	if !ok {
		return txs
	}
	for addr := range senders {
		tq := t.txs[addr]
		for {
			if len(*tq) == 0 {
				// We popped the queue empty!
				break
			}
			p, _ := heap.Pop(tq).(*PendingTransaction)
			if p.tx.BatchIndex() < batchIndex {
				// forget this tx
				continue
			} else if p.tx.BatchIndex() > batchIndex {
				// process this transaction at a later Pop()
				// push on the heap again
				heap.Push(tq, p)
				break
			} else {
				txs = append(txs, p)
			}
		}
	}
	delete(t.batchSenders, batchIndex)
	t.batches.Remove(batchIndex)
	return txs
}

// Push appends a transaction for that `batchIndex` to the
// tail of the list of transactions.
// It also adds the transaction's sender address to the
// set of senders for that `batchIndex`, retrievable by
// the Sender() method.
func (t *TransactionPool) Push(pending *PendingTransaction) {
	t.mux.Lock()
	defer t.mux.Unlock()

	batch := pending.tx.BatchIndex()
	senders, exists := t.batchSenders[batch]
	if !exists {
		senders = make(map[common.Address]bool, 1)
		t.batchSenders[batch] = senders
	}

	senders[pending.sender] = true
	txs, exists := t.txs[pending.sender]
	if !exists {
		tx := make(SortableTransactions, 0)
		txs = &tx
		heap.Init(txs)
		t.txs[pending.sender] = txs
	}
	heap.Push(txs, pending)
	t.batches.Insert(batch)
}

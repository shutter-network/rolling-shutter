package transaction

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func NewQueue() *Queue {
	return &Queue{
		txqueue: make([]*Pending, 0),
		senders: make(map[common.Address]bool, 0),
	}
}

// `Queue` is a container struct allowing
// for append-only operation on a list of `Pending`.
// `Queue` implements convenience methods to
// join two queues, generate the hash of all queued transactions
// and show the set of all transactions sender addresses.
type Queue struct {
	mu      sync.RWMutex
	txqueue []*Pending
	senders map[common.Address]bool
}

func (q *Queue) JoinRight(other *Queue) *Queue {
	q.mu.RLock()
	other.mu.RLock()
	defer other.mu.RUnlock()
	defer q.mu.RUnlock()

	txqueue := make([]*Pending, len(q.txqueue)+len(other.txqueue))
	n := copy(txqueue, q.txqueue)
	copy(txqueue[n:], other.txqueue)

	senders := make(map[common.Address]bool, len(q.senders))
	for addr, v := range q.senders {
		senders[addr] = v
	}
	for addr, v := range other.senders {
		senders[addr] = v
	}
	return &Queue{txqueue: txqueue, senders: senders}
}

func (q *Queue) Hash() []byte {
	q.mu.RLock()
	defer q.mu.RUnlock()
	txHashes := make([][]byte, q.Len())
	for i, t := range q.txqueue {
		txHashes[i] = t.Tx.Hash().Bytes()
	}
	return shmsg.HashByteList(txHashes)
}

func (q *Queue) Transactions() []*Pending {
	q.mu.RLock()
	defer q.mu.RUnlock()
	txqueue := make([]*Pending, len(q.txqueue))
	copy(txqueue, q.txqueue)
	return txqueue
}

func (q *Queue) Pop() []*Pending {
	q.mu.Lock()
	defer q.mu.Unlock()
	txqueue := q.txqueue
	q.txqueue = make([]*Pending, 0)
	q.senders = make(map[common.Address]bool)
	return txqueue
}

func (q *Queue) Bytes() [][]byte {
	q.mu.RLock()
	defer q.mu.RUnlock()
	l := make([][]byte, q.Len())
	for i, tx := range q.txqueue {
		l[i] = tx.TxBytes
	}
	return l
}

func (q *Queue) Len() int {
	return len(q.txqueue)
}

func (q *Queue) Enqueue(tx *Pending) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.senders[tx.Sender] = true
	q.txqueue = append(q.txqueue, tx)
}

// TotalByteSize returns the sum of the size of all transactions measured in bytes.
func (q *Queue) TotalByteSize() int {
	bs := q.Bytes()
	s := 0
	for _, b := range bs {
		s += len(b)
	}
	return s
}

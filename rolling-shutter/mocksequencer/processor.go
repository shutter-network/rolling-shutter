package mocksequencer

import (
	"context"
	"math/big"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"
)

const (
	BaseFee  = 22000
	GasLimit = 2200000000
)

type blockData struct {
	baseFee  *big.Int
	gasLimit uint64
}

type activationBlockMap[T any] struct {
	mux sync.RWMutex
	mp  map[uint64]T
}

func newActivationBlockMap[T any]() *activationBlockMap[T] {
	return &activationBlockMap[T]{
		mp:  map[uint64]T{},
		mux: sync.RWMutex{},
	}
}

func (a *activationBlockMap[T]) Set(val T, block uint64) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.mp[block] = val
}

func (a *activationBlockMap[T]) Find(block uint64) (T, error) {
	var (
		foundVal T
		i        int
	)
	a.mux.RLock()
	defer a.mux.RUnlock()

	blocks := make([]uint64, len(a.mp))
	for k := range a.mp {
		blocks[i] = k
		i++
	}
	if len(blocks) == 0 {
		return foundVal, errors.New("no value was set")
	}

	// sort in descending order
	sort.Slice(blocks, func(i, j int) bool { return blocks[i] > blocks[j] })

	// search for first index that is within the range
	idx := sort.Search(len(blocks), func(i int) bool {
		return blocks[i] <= block
	})

	if idx == len(blocks) {
		// nothing found, this means the queried block is lower
		// than the lowest activation block
		return foundVal, errors.New("no value was found")
	}

	activationBlockIndex := blocks[idx]
	foundVal = a.mp[activationBlockIndex]
	return foundVal, nil
}

type SequencerProcessor struct {
	port       int16
	nonces     map[string]map[string]uint64
	collators  *activationBlockMap[common.Address]
	eonKeys    *activationBlockMap[[]byte]
	chainID    *big.Int
	blocks     map[string]blockData
	txs        map[string]*txtypes.Transaction
	batchIndex uint64
	signer     txtypes.Signer
}

func New(chainID *big.Int, port int16) *SequencerProcessor {
	sequencer := &SequencerProcessor{
		port:       port,
		nonces:     map[string]map[string]uint64{},
		collators:  newActivationBlockMap[common.Address](),
		eonKeys:    newActivationBlockMap[[]byte](),
		chainID:    chainID,
		blocks:     map[string]blockData{},
		txs:        map[string]*txtypes.Transaction{},
		signer:     txtypes.NewLondonSigner(chainID),
		batchIndex: 0,
	}
	baseFee := big.NewInt(BaseFee)
	sequencer.setBlock(baseFee, GasLimit, "latest")
	return sequencer
}

func (proc *SequencerProcessor) setBlock(baseFee *big.Int, gasLimit uint64, block string) {
	b, exists := proc.blocks[block]
	if !exists {
		b = blockData{baseFee: baseFee, gasLimit: gasLimit}
		proc.blocks[block] = b
		return
	}
	b.baseFee = baseFee
	b.gasLimit = gasLimit
}

func (proc *SequencerProcessor) setNonce(a common.Address, nonce uint64, block string) {
	nc, exists := proc.nonces[block]
	if !exists {
		nc = make(map[string]uint64, 0)
		proc.nonces[block] = nc
	}
	nc[a.Hex()] = nonce
}

func (proc *SequencerProcessor) getNonce(a common.Address, block string) uint64 {
	nonce := uint64(0)
	nc, exists := proc.nonces[block]
	if !exists {
		nc = make(map[string]uint64, 0)
		proc.nonces[block] = nc
	}
	nonce, exists = nc[a.Hex()]
	if !exists {
		proc.setNonce(a, nonce, block)
	}
	return nonce
}

func (proc *SequencerProcessor) processEncryptedTx(
	ctx context.Context,
	gasPool *core.GasPool,
	batchIndex, batchL1BlockNumber uint64,
	txBytes, decryptionKey, eonKey []byte,
) error {
	var tx txtypes.Transaction
	err := tx.UnmarshalBinary(txBytes)
	if err != nil {
		return errors.Wrap(err, "can't unmarshal incoming bytes to transaction")
	}
	if tx.Type() != txtypes.ShutterTxType {
		return errors.New("no shutter tx type")
	}

	sender, err := proc.signer.Sender(&tx)
	if err != nil {
		return errors.New("sender not recoverable")
	}

	shutterTxStr, _ := tx.MarshalJSON()
	log.Ctx(ctx).Info().Str("signer", sender.Hex()).RawJSON("transaction", shutterTxStr).Msg("received shutter transaction")

	if tx.L1BlockNumber() != batchL1BlockNumber {
		return errors.New("l1-block-number mismatch")
	}

	if tx.BatchIndex() != batchIndex {
		return errors.New("batch-index mismatch")
	}
	_ = decryptionKey
	_ = eonKey
	nonce := proc.getNonce(sender, "latest")
	if tx.Nonce() != nonce+1 {
		return errors.New("nonce mismatch")
	}
	err = gasPool.SubGas(tx.Gas())
	if err != nil {
		// can return core.ErrGasLimitReached
		return err
	}

	proc.setNonce(sender, nonce+1, "latest")
	return nil
}

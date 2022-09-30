package mocksequencer

import (
	"context"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"
)

const (
	BaseFee            = 22000
	GasLimit           = 2200000000
	allowedL1Deviation = 5
	l1PollInterVal     = 7 * time.Second
)

type BlockData struct {
	BaseFee  *big.Int
	GasLimit uint64
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

type Processor struct {
	URL        string
	Nonces     map[string]map[string]uint64
	Collators  *activationBlockMap[common.Address]
	EonKeys    *activationBlockMap[[]byte]
	ChainID    *big.Int
	Blocks     map[string]BlockData
	Txs        map[string]*txtypes.Transaction
	BatchIndex uint64
	Signer     txtypes.Signer

	mux           sync.RWMutex
	l1BlockNumber uint64
	l1RPCURL      string
}

func New(chainID *big.Int, sequencerURL, l1RPCURL string) *Processor {
	sequencer := &Processor{
		URL:           sequencerURL,
		Nonces:        map[string]map[string]uint64{},
		Collators:     newActivationBlockMap[common.Address](),
		EonKeys:       newActivationBlockMap[[]byte](),
		ChainID:       chainID,
		Blocks:        map[string]BlockData{},
		Txs:           map[string]*txtypes.Transaction{},
		BatchIndex:    0,
		Signer:        txtypes.NewLondonSigner(chainID),
		mux:           sync.RWMutex{},
		l1BlockNumber: 0,
		l1RPCURL:      l1RPCURL,
	}
	baseFee := big.NewInt(BaseFee)
	sequencer.setBlock(baseFee, GasLimit, "latest")
	return sequencer
}

func (proc *Processor) RunBackgroundTasks(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	setError := func(err error) <-chan error {
		errChan <- err
		close(errChan)
		return errChan
	}

	ticker := time.NewTicker(l1PollInterVal)
	l1Client, err := ethclient.DialContext(ctx, proc.l1RPCURL)
	if err != nil {
		return setError(errors.Wrap(err, "error connecting to layer 1 JSON RPC endpoint"))
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				setError(ctx.Err())
				return
			case <-ticker.C:
				newBlockNumber, err := l1Client.BlockNumber(ctx)
				if err != nil {
					setError(errors.Wrap(err, "error retrieving block-number from layer 1 RPC"))
					return
				}
				proc.mux.Lock()
				proc.l1BlockNumber = newBlockNumber
				proc.mux.Unlock()
			}
		}
	}()
	return errChan
}

func (proc *Processor) setBlock(baseFee *big.Int, gasLimit uint64, block string) {
	b, exists := proc.Blocks[block]
	if !exists {
		b = BlockData{BaseFee: baseFee, GasLimit: gasLimit}
		proc.Blocks[block] = b
		return
	}
	b.BaseFee = baseFee
	b.GasLimit = gasLimit
}

func (proc *Processor) setNonce(a common.Address, nonce uint64, block string) {
	nc, exists := proc.Nonces[block]
	if !exists {
		nc = make(map[string]uint64, 0)
		proc.Nonces[block] = nc
	}
	nc[a.Hex()] = nonce
}

func (proc *Processor) GetNonce(a common.Address, block string) uint64 {
	nonce := uint64(0)
	nc, exists := proc.Nonces[block]
	if !exists {
		nc = make(map[string]uint64, 0)
		proc.Nonces[block] = nc
	}
	nonce, exists = nc[a.Hex()]
	if !exists {
		proc.setNonce(a, nonce, block)
	}
	return nonce
}

func (proc *Processor) validateBatch(tx *txtypes.Transaction) error {
	if tx.Type() != txtypes.BatchTxType {
		return errors.New("unexpected transaction type")
	}

	if tx.ChainId().Cmp(proc.ChainID) != 0 {
		return errors.New("chain-id mismatch")
	}

	if proc.BatchIndex != tx.BatchIndex()-1 {
		return errors.New("incorrect batch-index for next batch")
	}

	collator, err := proc.Collators.Find(tx.L1BlockNumber())
	if err != nil {
		return errors.Wrap(err, "collator validation failed")
	}
	sender, err := proc.Signer.Sender(tx)
	if err != nil {
		return errors.Wrap(err, "error recovering batch tx sender")
	}
	if collator != sender {
		return errors.Wrap(err, "not signed by correct collator")
	}

	// all checks passed, the batch-tx is valid (disregarding validity of included encrypted transactions)
	return nil
}

func (proc *Processor) ProcessEncryptedTx(
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

	sender, err := proc.Signer.Sender(&tx)
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
	nonce := proc.GetNonce(sender, "latest")
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

func (proc *Processor) SubmitBatch(ctx context.Context, batchTx *txtypes.Transaction) (string, error) {
	var gasPool core.GasPool

	err := proc.validateBatch(batchTx)

	// marshal to JSON just for logging output
	txStr, _ := batchTx.MarshalJSON()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).RawJSON("transaction", txStr).Msg("received invalid batch transaction")
		return "", errors.Wrap(err, "batch-tx invalid")
	}

	proc.mux.RLock()
	currentL1Block := proc.l1BlockNumber
	proc.mux.RUnlock()

	l1BlockDeviation := batchTx.L1BlockNumber() - currentL1Block
	if l1BlockDeviation < 0 {
		// the collator is lacking behind
		l1BlockDeviation = -l1BlockDeviation
	}
	if l1BlockDeviation > allowedL1Deviation {
		return "", errors.Errorf("the 'L1BlockNumber' deviates more than the allowed %d blocks.", allowedL1Deviation)
	}

	eonKey, err := proc.EonKeys.Find(batchTx.L1BlockNumber())
	if err != nil {
		err = errors.Wrap(err, "no eon key found for batch transaction")
		log.Ctx(ctx).Error().Err(err).Msg("error while retrieving eon key")
		return "", err
	}

	gasPool.AddGas(GasLimit)
	for _, shutterTx := range batchTx.Transactions() {
		err := proc.ProcessEncryptedTx(ctx, &gasPool, batchTx.BatchIndex(), batchTx.L1BlockNumber(), shutterTx, batchTx.DecryptionKey(), eonKey)
		if err != nil {
			// those are conditions that the collator can check,
			// so an error here means the whole batch is invalid
			err := errors.Wrap(err, "transaction invalid")
			return "", err
		}
		log.Info().Msg("successfully applied shutter-tx")
	}

	sender, _ := proc.Signer.Sender(batchTx)
	log.Ctx(ctx).Info().Str("signer", sender.Hex()).RawJSON("transaction", txStr).Msg("received batch transaction")
	proc.BatchIndex = batchTx.BatchIndex()

	log.Ctx(ctx).Info().Str("batch", hexutil.EncodeUint64(proc.BatchIndex)).Msg("started new batch")

	return batchTx.Hash().Hex(), nil
}

package mocksequencer

import (
	"context"
	"encoding/binary"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/batch"
)

const (
	BaseFee            = 22000
	GasLimit           = 2200000000
	AllowedL1Deviation = 5
	L1PollInterVal     = 1 * time.Second
)

var (
	coinbase              = common.HexToAddress("0x0000000000000000000000000000000000000000")
	allowedL1DeviationBig = big.NewInt(AllowedL1Deviation)
)

type BlockData struct {
	mux            sync.Mutex
	Hash           common.Hash
	BaseFee        *big.Int
	GasLimit       uint64
	Number         uint64
	Transactions   txtypes.Transactions
	Nonces         map[common.Address]uint64
	Balances       map[common.Address]*big.Int
	gasPool        *core.GasPool
	feeBeneficiary common.Address
}

func (b *BlockData) ApplyTx(tx *txtypes.Transaction, sender common.Address) error {
	b.mux.Lock()
	defer b.mux.Unlock()

	nonce := b.GetNonce(sender)
	if tx.Nonce() != nonce {
		return errors.New("nonce mismatch")
	}
	balance := b.GetBalance(sender)

	// deductable from the sender's account
	gasCost := batch.CalculateGasCost(tx, b.BaseFee)
	// add to the gasBeneficiary's account (collator)
	priorityFee := batch.CalculatePriorityFee(tx, b.BaseFee)

	if balance.Cmp(gasCost) == -1 {
		return errors.New("insufficient funds for gas fee")
	}

	err := b.gasPool.SubGas(tx.Gas())
	if err != nil {
		// can return core.ErrGasLimitReached
		return err
	}
	// if this didn't error, the transaction has
	// to be committed to the block now

	b.SetBalance(sender, big.NewInt(0).Sub(balance, gasCost))

	gbBalance := b.GetBalance(b.feeBeneficiary)
	b.SetBalance(b.feeBeneficiary, big.NewInt(0).Add(gbBalance, priorityFee))

	b.SetNonce(sender, nonce+1)
	b.Transactions = append(b.Transactions, tx)
	return nil
}

func (b *BlockData) SetBalance(a common.Address, balance *big.Int) {
	b.Balances[a] = balance
}

func (b *BlockData) GetBalance(a common.Address) *big.Int {
	balance, ok := b.Balances[a]
	if !ok {
		return big.NewInt(0)
	}
	return balance
}

func (b *BlockData) SetNonce(a common.Address, nonce uint64) {
	b.Nonces[a] = nonce
}

func (b *BlockData) GetNonce(a common.Address) uint64 {
	nonce := b.Nonces[a]
	return nonce
}

func CreateNextBlockData(baseFee *big.Int, gasLimit uint64, feeBeneficiary common.Address, previous *BlockData) *BlockData {
	gasPool := core.GasPool(gasLimit)
	bd := &BlockData{
		mux:            sync.Mutex{},
		Hash:           common.Hash{},
		BaseFee:        baseFee,
		GasLimit:       gasLimit,
		Number:         0,
		Transactions:   []*txtypes.Transaction{},
		Nonces:         map[common.Address]uint64{},
		Balances:       map[common.Address]*big.Int{},
		gasPool:        &gasPool,
		feeBeneficiary: feeBeneficiary,
	}

	d := crypto.NewKeccakState()
	if previous != nil {
		// copy the maps
		for addr, nonce := range previous.Nonces {
			bd.Nonces[addr] = nonce
		}
		for addr, balance := range previous.Balances {
			bd.Balances[addr] = balance
		}
		bd.Number = previous.Number + 1
		d.Write(previous.Hash[:])
	}
	// block-hash is simply the keccak256 of
	// the byte representation of the block-number
	buf := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutUvarint(buf, bd.Number)
	d.Write(buf)

	d.Read(bd.Hash[:])

	return bd
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
	URL       string
	Collators *activationBlockMap[common.Address]
	EonKeys   *activationBlockMap[[]byte]
	ChainID   *big.Int

	LatestBlock common.Hash
	Blocks      map[common.Hash]*BlockData
	Txs         map[common.Hash]TransactionIdentifier
	BatchIndex  uint64
	Signer      txtypes.Signer

	Mux           sync.RWMutex
	l1BlockNumber uint64
	l1RPCURL      string
}

type TransactionIdentifier struct {
	BlockHash common.Hash
	Index     int
}

func New(chainID *big.Int, sequencerURL, l1RPCURL string) *Processor {
	sequencer := &Processor{
		URL:           sequencerURL,
		Collators:     newActivationBlockMap[common.Address](),
		EonKeys:       newActivationBlockMap[[]byte](),
		ChainID:       chainID,
		LatestBlock:   common.Hash{},
		Blocks:        make(map[common.Hash]*BlockData),
		Txs:           map[common.Hash]TransactionIdentifier{},
		BatchIndex:    0,
		Signer:        txtypes.NewLondonSigner(chainID),
		Mux:           sync.RWMutex{},
		l1BlockNumber: 0,
		l1RPCURL:      l1RPCURL,
	}
	blockData := CreateNextBlockData(big.NewInt(BaseFee), GasLimit, coinbase, nil)
	sequencer.setLatestBlock(blockData)
	return sequencer
}

func (proc *Processor) RunBackgroundTasks(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	setError := func(err error) <-chan error {
		errChan <- err
		close(errChan)
		return errChan
	}

	ticker := time.NewTicker(L1PollInterVal)
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
				proc.Mux.Lock()
				if newBlockNumber > proc.l1BlockNumber {
					proc.l1BlockNumber = newBlockNumber
					log.Debug().Int("l1-block-number", int(newBlockNumber)).Msg("updated state from layer 1 node")
				} else if newBlockNumber < proc.l1BlockNumber {
					log.Warn().Int("new-l1-block-number", int(newBlockNumber)).
						Int("cached-l1-block-number", int(proc.l1BlockNumber)).
						Msg("l1 cache inconsistency")
				}
				proc.Mux.Unlock()
			}
		}
	}()
	return errChan
}

func (proc *Processor) GetBlock(blockNrOrHash ethrpc.BlockNumberOrHash) (block *BlockData, err error) {
	if blockNrOrHash.BlockHash != nil {
		var ok bool
		block, ok = proc.Blocks[*blockNrOrHash.BlockHash]
		if !ok {
			return nil, errors.New("block not found")
		}
	} else if blockNrOrHash.BlockNumber != nil {
		var ok bool
		// only possible for "latest" (-1) currently
		if *blockNrOrHash.BlockNumber != ethrpc.LatestBlockNumber {
			return nil, errors.New("provided block number argument currently not supported")
		}
		block, ok = proc.Blocks[proc.LatestBlock]
		if !ok {
			return nil, errors.New("block not found")
		}
	} else {
		return nil, errors.New("block argument invalid")
	}
	return block, nil
}

// setNextBlock sets the provided block-data as the next
// latest-block and persist the data in the internal
// datastructures for later retrieval.
// The Sequencer write-lock has to held during this operation,
// when there is potential concurrent access.
func (proc *Processor) setLatestBlock(b *BlockData) {
	proc.Blocks[b.Hash] = b
	// make the transactions locateable by hash
	for i, tx := range b.Transactions {
		proc.Txs[tx.Hash()] = TransactionIdentifier{BlockHash: b.Hash, Index: i}
	}
	proc.LatestBlock = b.Hash
}

func (proc *Processor) validateBatch(tx *txtypes.Transaction) (common.Address, error) {
	var sender common.Address
	if tx.Type() != txtypes.BatchTxType {
		return sender, errors.New("unexpected transaction type")
	}

	if tx.ChainId().Cmp(proc.ChainID) != 0 {
		return sender, errors.New("chain-id mismatch")
	}

	if proc.BatchIndex != tx.BatchIndex()-1 {
		return sender, errors.New("incorrect batch-index for next batch")
	}

	collator, err := proc.Collators.Find(tx.L1BlockNumber())
	if err != nil {
		return sender, errors.Wrap(err, "collator validation failed")
	}
	sender, err = proc.Signer.Sender(tx)
	if err != nil {
		return sender, errors.Wrap(err, "error recovering batch tx sender")
	}
	if collator != sender {
		return sender, errors.Wrap(err, "not signed by correct collator")
	}

	// all checks passed, the batch-tx is valid (disregarding validity of included encrypted transactions)
	return sender, nil
}

func (proc *Processor) ProcessEncryptedTx(
	ctx context.Context,
	batchIndex, batchL1BlockNumber uint64,
	txBytes, decryptionKey, eonKey []byte,
	pendingBlock *BlockData,
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
	// TODO decrypt encrypted payload and decode rlp to transaction

	err = pendingBlock.ApplyTx(&tx, sender)
	if err != nil {
		return errors.Wrap(err, "couldn't apply transaction")
	}
	return nil
}

func (proc *Processor) SubmitBatch(ctx context.Context, batchTx *txtypes.Transaction) (string, error) {
	var gasPool core.GasPool

	proc.Mux.Lock()
	defer proc.Mux.Unlock()

	collator, err := proc.validateBatch(batchTx)

	// marshal to JSON just for logging output
	txStr, _ := batchTx.MarshalJSON()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).RawJSON("transaction", txStr).Msg("received invalid batch transaction")
		return "", errors.Wrap(err, "batch-tx invalid")
	}

	currentL1Block := proc.l1BlockNumber
	latestBlock, ok := proc.Blocks[proc.LatestBlock]
	if !ok {
		// TODO panic
		// internal server error
		return "", errors.New("latest block is not initialized")
	}

	// calculate l1-block-number deviation between collator and sequencer:
	// delta = batch-l1-blocknumber - sequencer-l1-block-number
	l1BlockDeviation := new(big.Int).Sub(
		new(big.Int).SetUint64(batchTx.L1BlockNumber()),
		new(big.Int).SetUint64(currentL1Block),
	)
	if l1BlockDeviation.CmpAbs(allowedL1DeviationBig) == 1 {
		// the deviation between batch-tx's l1-block-number and sequencer's
		// known l1-block-number is greater than allowed:
		// |delta| > |maximum-delta|
		return "", errors.Errorf("the 'L1BlockNumber' deviates more than the allowed %d blocks", allowedL1Deviation)
	}

	eonKey, err := proc.EonKeys.Find(batchTx.L1BlockNumber())
	if err != nil {
		err = errors.Wrap(err, "no eon key found for batch transaction")
		log.Ctx(ctx).Error().Err(err).Msg("error while retrieving eon key")
		return "", err
	}

	pendingBlock := CreateNextBlockData(big.NewInt(BaseFee), GasLimit, collator, latestBlock)
	gasPool.AddGas(GasLimit)
	for _, shutterTx := range batchTx.Transactions() {
		err := proc.ProcessEncryptedTx(
			ctx,
			batchTx.BatchIndex(),
			batchTx.L1BlockNumber(),
			shutterTx,
			batchTx.DecryptionKey(),
			eonKey,
			pendingBlock,
		)
		if err != nil {
			// those are conditions that the collator can check,
			// so an error here means the whole batch is invalid
			err := errors.Wrap(err, "transaction invalid")
			return "", err
		}
		log.Info().Msg("successfully applied shutter-tx")
	}

	// commit the "state change"
	proc.BatchIndex = batchTx.BatchIndex()
	proc.setLatestBlock(pendingBlock)
	proc.Mux.Unlock()

	log.Ctx(ctx).Info().Str("signer", collator.Hex()).RawJSON("transaction", txStr).Msg("included batch transaction")
	log.Ctx(ctx).Info().Str("batch", hexutil.EncodeUint64(proc.BatchIndex)).Msg("started new batch")

	return batchTx.Hash().Hex(), nil
}

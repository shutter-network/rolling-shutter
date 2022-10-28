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
	rpcerrors "github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/errors"
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
		err := errors.New("nonce mismatch for payload transaction")
		return rpcerrors.TransactionRejected(err)
	}
	balance := b.GetBalance(sender)

	// deductable from the sender's account
	gasCost := batch.CalculateGasCost(tx, b.BaseFee)
	// add to the gasBeneficiary's account (collator)
	priorityFee := batch.CalculatePriorityFee(tx, b.BaseFee)

	if balance.Cmp(gasCost) == -1 {
		err := errors.New("insufficient funds for gas fee")
		return rpcerrors.TransactionRejected(err)
	}

	err := b.gasPool.SubGas(tx.Gas())
	if err != nil {
		return rpcerrors.TransactionRejected(core.ErrGasLimitReached)
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
	_, _ = d.Read(bd.Hash[:])
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

type Sequencer struct {
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

func New(chainID *big.Int, sequencerURL, l1RPCURL string) *Sequencer {
	sequencer := &Sequencer{
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

func (proc *Sequencer) RunBackgroundTasks(ctx context.Context) <-chan error {
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

func (proc *Sequencer) GetBlock(blockNrOrHash ethrpc.BlockNumberOrHash) (block *BlockData, err error) {
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
		err := errors.New("block argument invalid")
		return nil, rpcerrors.ParseError(err)
	}
	return block, nil
}

// setNextBlock sets the provided block-data as the next
// latest-block and persist the data in the internal
// datastructures for later retrieval.
// The Sequencer write-lock has to held during this operation,
// when there is potential concurrent access.
func (proc *Sequencer) setLatestBlock(b *BlockData) {
	proc.Blocks[b.Hash] = b
	// make the transactions locateable by hash
	for i, tx := range b.Transactions {
		log.Info().Msg("added transaction")
		proc.Txs[tx.Hash()] = TransactionIdentifier{BlockHash: b.Hash, Index: i}
	}
	proc.LatestBlock = b.Hash
}

func (proc *Sequencer) validateBatch(tx *txtypes.Transaction) (common.Address, error) {
	var sender common.Address
	if tx.Type() != txtypes.BatchTxType {
		err := errors.New("unexpected transaction type")
		return sender, rpcerrors.ParseError(err)
	}

	if tx.ChainId().Cmp(proc.ChainID()) != 0 {
		err := errors.New("chain-id mismatch")
		return sender, rpcerrors.TransactionRejected(err)
	}

	if proc.BatchIndex != tx.BatchIndex()-1 {
		err := errors.New("incorrect batch-index for next batch")
		return sender, rpcerrors.TransactionRejected(err)
	}

	collator, err := proc.Collators.Find(tx.L1BlockNumber())
	if err != nil {
		err := errors.Wrap(err, "collator validation failed")
		return sender, rpcerrors.TransactionRejected(err)
	}
	sender, err = proc.Signer.Sender(tx)
	if err != nil {
		err := errors.Wrap(err, "error recovering batch tx sender")
		return sender, rpcerrors.TransactionRejected(err)
	}
	if collator != sender {
		err := errors.Wrap(err, "not signed by correct collator")
		return sender, rpcerrors.TransactionRejected(err)
	}

	// all checks passed, the batch-tx is valid (disregarding validity of included encrypted transactions)
	return sender, nil
}

func (proc *Sequencer) ProcessEncryptedTx(
	ctx context.Context,
	batchIndex, batchL1BlockNumber uint64,
	txBytes []byte,
	epochSecretKey *shcrypto.EpochSecretKey,
	pendingBlock *BlockData,
) error {
	var tx txtypes.Transaction
	err := tx.UnmarshalBinary(txBytes)
	if err != nil {
		return errors.Wrap(err, "can't unmarshal incoming bytes")
	}
	if tx.Type() != txtypes.ShutterTxType {
		return errors.New("wrong transaction type")
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

	payload, err := decryptPayload(tx.EncryptedPayload(), epochSecretKey)
	if err != nil {
		return errors.Wrap(err, "couldn't decrypt payload")
	}

	txInner := tx.TxInner()
	shutterTx, ok := txInner.(*txtypes.ShutterTx)
	if !ok {
		return errors.New("could not extract ShutterTx")
	}
	shutterTx.Payload = payload
	shutterTxWithPayload := txtypes.NewTx(shutterTx)

	err = pendingBlock.ApplyTx(shutterTxWithPayload, sender)
	if err != nil {
		return errors.Wrap(err, "couldn't apply transaction")
	}
	return nil
}

func (proc *Sequencer) SubmitBatch(ctx context.Context, batchTx *txtypes.Transaction) (string, error) {
	var gasPool core.GasPool

	proc.Mux.Lock()
	defer proc.Mux.Unlock()

	if batchTx == nil {
		return "", errors.New("missing argument: batch-transaction")
	}

	collator, err := proc.validateBatch(batchTx)
	if err != nil {
		return "", errors.Wrap(err, "failed to validate batch")
	}

	// marshal to JSON just for logging output
	txStr, err2 := batchTx.MarshalJSON()
	if err2 != nil {
		panic(err2)
	}

	if err != nil {
		log.Ctx(ctx).Error().Err(err).RawJSON("transaction", txStr).Msg("received invalid batch transaction")
		return "", err
	}

	currentL1Block := proc.l1BlockNumber
	latestBlock, ok := proc.Blocks[proc.LatestBlock]
	if !ok {
		err := errors.New("latest block is not initialized")
		return "", rpcerrors.InternalServerError(err)
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
		err := errors.Errorf("the 'L1BlockNumber' deviates more than the allowed %d blocks", AllowedL1Deviation)
		return "", rpcerrors.TransactionRejected(err)
	}

	pendingBlock := CreateNextBlockData(big.NewInt(BaseFee), GasLimit, collator, latestBlock)
	gasPool.AddGas(GasLimit)

	epochSecretKey := &shcrypto.EpochSecretKey{}
	err = epochSecretKey.GobDecode(batchTx.DecryptionKey())
	if err != nil {
		err = errors.Wrap(err, "couldn't decode decryption key")
		return "", rpcerrors.TransactionRejected(err)
	}

	for _, shutterTx := range batchTx.Transactions() {
		err := proc.ProcessEncryptedTx(
			ctx,
			batchTx.BatchIndex(),
			batchTx.L1BlockNumber(),
			shutterTx,
			epochSecretKey,
			pendingBlock,
		)
		if err != nil {
			// those are conditions that the collator can check,
			// so an error here means the whole batch is invalid
			err := errors.Wrap(err, "invalid shutter transaction in batch")
			return "", rpcerrors.TransactionRejected(err)
		}
		log.Info().Msg("successfully applied shutter-tx")
	}

	// commit the "state change"
	proc.BatchIndex = batchTx.BatchIndex()
	proc.setLatestBlock(pendingBlock)

	log.Ctx(ctx).Info().Str("signer", collator.Hex()).RawJSON("transaction", txStr).Msg("included batch transaction")
	log.Ctx(ctx).Info().Str("batch", hexutil.EncodeUint64(proc.BatchIndex)).Msg("started new batch")

	return batchTx.Hash().Hex(), nil
}

func decryptPayload(messageBytes []byte, epochSecretKey *shcrypto.EpochSecretKey) (*txtypes.ShutterPayload, error) {
	message := &shcrypto.EncryptedMessage{}
	err := message.Unmarshal(messageBytes)
	if err != nil {
		return nil, errors.Wrap(err, "can't unmarshal message")
	}
	decryptedBytes, err := message.Decrypt(epochSecretKey)
	if err != nil {
		return nil, errors.Wrap(err, "can't decrypt message")
	}
	return txtypes.DecodeShutterPayload(decryptedBytes)
}

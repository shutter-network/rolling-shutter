package mocksequencer

import (
	"context"
	"encoding/json"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/shutter/shlib/shcrypto"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/encoding"
	rpcerrors "github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/errors"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

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
	if len(a.mp) == 0 {
		return foundVal, errors.New("no value was set")
	}

	blocks := make([]uint64, len(a.mp))
	for k := range a.mp {
		blocks[i] = k
		i++
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

func New(
	config *Config,
) service.Service {
	// TODO query from the node
	sequencer := &Sequencer{
		URL:                  config.HTTPListenAddress,
		Collators:            newActivationBlockMap[common.Address](),
		EonKeys:              newActivationBlockMap[[]byte](),
		Mux:                  sync.RWMutex{},
		l1BlockNumber:        0,
		Config:               config,
		sentTransactions:     make(map[common.Hash]bool),
		sentTransactionsLock: sync.RWMutex{},
	}
	return sequencer
}

type Sequencer struct {
	Mux       sync.RWMutex
	dbpool    *pgxpool.Pool
	contracts *deployment.Contracts
	p2p       *p2p.P2PHandler

	Config               *Config
	URL                  string
	Collators            *activationBlockMap[common.Address]
	EonKeys              *activationBlockMap[[]byte]
	Signer               txtypes.Signer
	sentTransactions     map[common.Hash]bool
	sentTransactionsLock sync.RWMutex

	L2BackendRPC    *ethrpc.Client
	L2BackendClient *ethclient.Client

	lastSequencerBatchSubmitted time.Time
	active                      bool
	l1BlockNumber               uint64
}

func (proc *Sequencer) Start(ctx context.Context, runner service.Runner) error {
	var err error

	// TODO we need to define / init the mocksequencer db?
	// At least the play scripts should consider this
	dbpool, err := pgxpool.Connect(ctx, proc.Config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	runner.Defer(dbpool.Close)
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("connected to database")
	proc.dbpool = dbpool

	// proc.p2p, err = p2p.New(proc.Config.P2P)
	// if err != nil {
	// 	return err
	// }

	proc.L2BackendRPC, err = ethrpc.Dial(proc.Config.L2BackendURL.String())
	if err != nil {
		return errors.Wrap(err, "couldn't connect to L2 backend")
	}
	proc.L2BackendClient = ethclient.NewClient(proc.L2BackendRPC)

	contracts, err := deployment.NewContracts(proc.L2BackendClient, proc.Config.DeploymentDir)
	if err != nil {
		return err
	}
	proc.contracts = contracts

	chainID, err := proc.L2BackendClient.ChainID(ctx)
	if err != nil {
		return err
	}
	proc.Signer = txtypes.NewLondonSigner(chainID)

	return runner.StartService(proc.getServices()...)
}

type TransactionIdentifier struct {
	BlockHash common.Hash
	Index     int
}

func (proc *Sequencer) RUnlock() {
	proc.Mux.RUnlock()
}

func (proc *Sequencer) RLock() {
	proc.Mux.RLock()
}

func (proc *Sequencer) ChainID(ctx context.Context) (*big.Int, error) {
	return proc.L2BackendClient.ChainID(ctx)
}

func (proc *Sequencer) Active() bool {
	proc.RLock()
	defer proc.RUnlock()
	return proc.active
}

func (proc *Sequencer) BatchIndex(ctx context.Context) (uint64, error) {
	return proc.L2BackendClient.BlockNumber(ctx)
}

func (proc *Sequencer) validateBatch(ctx context.Context, tx *txtypes.Transaction) (common.Address, error) {
	var sender common.Address
	if tx.Type() != txtypes.BatchTxType {
		err := errors.New("unexpected transaction type")
		return sender, rpcerrors.ParseError(err)
	}
	chainId, err := proc.ChainID(ctx)
	if err != nil {
		err := errors.New("internal error, can't query backend chain-id")
		return sender, rpcerrors.ParseError(err)
	}

	if tx.Type() != txtypes.BatchTxType {
		err := errors.New("unexpected transaction type")
		return sender, rpcerrors.ParseError(err)
	}

	if tx.ChainId().Cmp(chainId) != 0 {
		err := errors.New("chain-id mismatch")
		return sender, rpcerrors.TransactionRejected(err)
	}

	batchIndex, err := proc.BatchIndex(ctx)
	if err != nil {
		err := errors.New("internal error, can't query backend block-number")
		return sender, rpcerrors.ParseError(err)
	}
	if batchIndex != tx.BatchIndex()-1 {
		err := errors.New("incorrect batch-index for next batch")
		return sender, rpcerrors.TransactionRejected(err)
	}

	collator, err := proc.Collators.Find(tx.L1BlockNumber())
	if err != nil {
		err := errors.Wrapf(err, "no collator registered for block-number %d", tx.L1BlockNumber())
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

	logEncryptedShutterTx(&tx)

	// NOTE: for the  proof-of-concept, remove this check.
	// this is because we don't implemented the l1-lookahead
	// yet, that would allow to infer the l1 block number
	// for future batches as well
	// if tx.L1BlockNumber() != batchL1BlockNumber {
	// 	return errors.New("l1-block-number mismatch")
	// }
	_ = batchL1BlockNumber

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
	logDecryptedShutterTx(shutterTxWithPayload, shutterTx.Payload)
	err = proc.ForwardTxToL2Backend(ctx, shutterTxWithPayload, &sender)
	if err != nil {
		return errors.Wrap(err, "couldn't apply transaction")
	}
	return nil
}

func (proc *Sequencer) SendTransaction(
	ctx context.Context,
	tx *coretypes.Transaction,
	from *common.Address,
) (*common.Hash, error) {
	var txHash *common.Hash
	txJSON := encoding.ToTransactionJSON(tx, from)
	mshTx, _ := json.Marshal(txJSON)
	log.Info().RawJSON("transaction", mshTx).Msg("sending transaction to backend")
	err := proc.L2BackendRPC.CallContext(ctx, &txHash, "eth_sendTransaction", txJSON)
	return txHash, err
}

func (proc *Sequencer) ForwardTxToL2Backend(
	ctx context.Context,
	tx *txtypes.Transaction,
	sender *common.Address,
) error {
	decrTx := &coretypes.DynamicFeeTx{
		ChainID:   tx.ChainId(),
		Nonce:     tx.Nonce(),
		GasTipCap: tx.GasTipCap(),
		GasFeeCap: tx.GasFeeCap(),
		Gas:       tx.Gas(),
		To:        tx.To(),
		Value:     tx.Value(),
		Data:      tx.Data(),
	}

	newTx := coretypes.NewTx(decrTx)
	err := proc.ImpersonateAccount(ctx, sender)
	if err != nil {
		log.Error().Err(err).Msg("error during 'impersonate'")
	}
	txHash, sendErr := proc.SendTransaction(ctx, newTx, sender)
	// XXX somewhere here the handler once crashed,
	// during receiving the first transaction
	if sendErr == nil {
		proc.sentTransactionsLock.Lock()
		proc.sentTransactions[*txHash] = true
		proc.sentTransactionsLock.Unlock()
	}
	err = proc.StopImpersonateAccount(ctx, sender)
	if err != nil {
		log.Error().Err(err).Msg("error during 'stop impersonate'")
	}
	return sendErr
}

func (proc *Sequencer) ImpersonateAccount(ctx context.Context, account *common.Address) error {
	return proc.L2BackendRPC.CallContext(ctx, nil, "hardhat_impersonateAccount", account.Hex())
}

func (proc *Sequencer) StopImpersonateAccount(ctx context.Context, account *common.Address) error {
	return proc.L2BackendRPC.CallContext(ctx, nil, "hardhat_stopImpersonatingAccount", account.Hex())
}

func (proc *Sequencer) DisableAutomine(ctx context.Context) error {
	// TODO enable the reverse order, so that we can also enable it again

	// TODO --> we can't pass a bool as result since it is no pointer.
	// figure out how to set the bool result, maybe with raw-Json?
	// proc.L2BackendRPC.CallContext(ctx, status, "hardhat_getAutomine")
	err := proc.L2BackendRPC.CallContext(ctx, nil, "evm_setAutomine", false)
	if err != nil {
		return err
	}
	err = proc.L2BackendRPC.CallContext(ctx, nil, "evm_setAutomine", false)
	if err != nil {
		return err
	}

	err = proc.L2BackendRPC.CallContext(ctx, nil, "evm_setIntervalMining", 0)
	if err != nil {
		return err
	}
	return nil
}

func (proc *Sequencer) MineBlock(ctx context.Context) error {
	err := proc.L2BackendRPC.CallContext(ctx, nil, "evm_mine")
	if err != nil {
		return errors.Wrap(err, "error during 'evm_mine' call")
	}
	log.Info().Msg("mined block")
	return nil
}

func (proc *Sequencer) SetCollator(address common.Address, activationBlock uint64) {
	proc.Collators.Set(address, activationBlock)
}

func logBatchTx(batchTx *txtypes.Transaction) {
	log.Info().
		Uint64("l1-block-number", batchTx.L1BlockNumber()).
		Bytes("decryption key", batchTx.DecryptionKey()).
		Uint64("batch-index", batchTx.BatchIndex()).
		Uint64("nonce", batchTx.Nonce()).
		Interface("timestamp", batchTx.Timestamp()).
		Interface("num-transactions", len(batchTx.Transactions())).
		Msg("received submitted batch")
	log.Debug().
		Interface("num-transactions", batchTx.Transactions()).
		Msg("received encrypted shutter transactions in batch")
}

func logEncryptedShutterTx(tx *txtypes.Transaction) {
	log.Info().
		Uint64("chain-id", tx.ChainId().Uint64()).
		Uint64("nonce", tx.Nonce()).
		Uint64("gas-tip-cap", tx.GasTipCap().Uint64()).
		Uint64("gas-fee-cap", tx.GasFeeCap().Uint64()).
		Uint64("gas", tx.Gas()).
		Uint64("l1-block-number", tx.L1BlockNumber()).
		Uint64("batch-index", tx.BatchIndex()).
		Str("encrypted-payload", common.Bytes2Hex(tx.EncryptedPayload())).
		Msg("received encrypted shutter transaction")
}

func logDecryptedShutterTx(tx *txtypes.Transaction, payload *txtypes.ShutterPayload) {
	payloadEvent := zerolog.Dict().
		Str("to", payload.To.Hex()).
		Str("data", common.Bytes2Hex(payload.Data)).
		Str("value", payload.Value.String())
	log.Info().
		Uint64("chain-id", tx.ChainId().Uint64()).
		Uint64("nonce", tx.Nonce()).
		Uint64("gas-tip-cap", tx.GasTipCap().Uint64()).
		Uint64("gas-fee-cap", tx.GasFeeCap().Uint64()).
		Uint64("gas", tx.Gas()).
		Uint64("l1-block-number", tx.L1BlockNumber()).
		Uint64("batch-index", tx.BatchIndex()).
		Dict("decrypted-payload", payloadEvent).
		Msg("successfully decrypted shutter transaction")
}

func (proc *Sequencer) SubmitBatch(
	ctx context.Context,
	batchTx *txtypes.Transaction,
) (string, error) {
	logBatchTx(batchTx)

	proc.Mux.Lock()
	defer proc.Mux.Unlock()

	if batchTx == nil {
		return "", errors.New("missing argument: batch-transaction")
	}

	collator, err := proc.validateBatch(ctx, batchTx)
	if err != nil {
		return "", errors.Wrap(err, "failed to validate batch")
	}

	// marshal to JSON just for logging output
	txStr, err2 := batchTx.MarshalJSON()
	if err2 == nil {
		if err != nil {
			log.Error().Err(err).RawJSON("transaction", txStr).Msg("received invalid batch transaction")
		} else {
			log.Info().RawJSON("transaction", txStr).Msg("received batch tx")
		}
	}
	if err != nil {
		return "", err
	}

	// TODO query
	currentL1Block := proc.l1BlockNumber

	// calculate l1-block-number deviation between collator and sequencer:
	// delta = batch-l1-blocknumber - sequencer-l1-block-number
	l1BlockDeviation := new(big.Int).Sub(
		new(big.Int).SetUint64(batchTx.L1BlockNumber()),
		new(big.Int).SetUint64(currentL1Block),
	)

	allowedL1DeviationBig := new(big.Int).SetUint64(proc.Config.MaxBlockDeviation)
	if l1BlockDeviation.CmpAbs(allowedL1DeviationBig) == 1 {
		// the deviation between batch-tx's l1-block-number and sequencer's
		// known l1-block-number is greater than allowed:
		// |delta| > |maximum-delta|
		log.Error().
			Uint64("batchtx-l1-blocknum", batchTx.L1BlockNumber()).
			Uint64("current-l1-blocknum", currentL1Block).
			Uint64("max-block-deviation", proc.Config.MaxBlockDeviation).
			Msg("rejecting batchtx (block number deviation)")
		err := errors.Errorf(
			"the 'L1BlockNumber' deviates more than the allowed %d blocks",
			proc.Config.MaxBlockDeviation,
		)
		return "", rpcerrors.TransactionRejected(err)
	}

	epochSecretKey := &shcrypto.EpochSecretKey{}
	err = epochSecretKey.GobDecode(batchTx.DecryptionKey())
	if err != nil {
		err = errors.Wrap(err, "couldn't decode decryption key")
		return "", rpcerrors.TransactionRejected(err)
	}

	proc.lastSequencerBatchSubmitted = time.Now()

	for _, shutterTx := range batchTx.Transactions() {
		err := proc.ProcessEncryptedTx(
			ctx,
			batchTx.BatchIndex(),
			batchTx.L1BlockNumber(),
			shutterTx,
			epochSecretKey,
		)
		if err != nil {
			// those are conditions that the collator can check,
			// so an error here means the whole batch is invalid
			err := errors.Wrap(err, "invalid shutter transaction in batch")
			return "", rpcerrors.TransactionRejected(err)
		}
		log.Info().Msg("successfully applied shutter-tx")
	}

	log.Info().Str("signer", collator.Hex()).RawJSON("transaction", txStr).Msg("included batch transaction")

	err = proc.MineBlock(ctx)
	if err != nil {
		return "", errors.Wrap(err, "couldn't propose batch")
	}
	log.Info().Uint64("batch-index", batchTx.BatchIndex()).Msg("started new batch")

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

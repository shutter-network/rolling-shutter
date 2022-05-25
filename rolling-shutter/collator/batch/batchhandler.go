package batch

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

const (
	//  The maximum number of future batches allowed for transaction submittal
	TransactionAcceptanceInterval = 5
)

// computeNextEpochID takes an epoch id as parameter and returns the id of the epoch following it.
// The function also depends on the current mainchain block number and the configured execution
// block delay. The result will encode a block number and a sequence number. The sequence number
// will be the sequence number of the previous epoch id plus one. The block number will be
// max(current block number - execution block delay, block number encoded in previous epoch id, 0).
func computeNextEpochID(epochID uint64, currentBlockNumber uint32, executionBlockDelay uint32) uint64 {
	executionBlockNumber := uint32(0)
	if currentBlockNumber >= executionBlockDelay {
		executionBlockNumber = currentBlockNumber - executionBlockDelay
	}

	previousExecutionBlockNumber := epochid.BlockNumber(epochID)
	if executionBlockNumber < previousExecutionBlockNumber {
		executionBlockNumber = previousExecutionBlockNumber
	}

	sequenceNumber := epochid.SequenceNumber(epochID)
	return epochid.New(sequenceNumber+1, executionBlockNumber)
}

// GetNextEpochID gets the epochID that will be used for the next decryption trigger or cipher batch.
func GetNextEpochID(ctx context.Context, db *cltrdb.Queries) (uint64, error) {
	epochID, err := db.GetNextEpochID(ctx)
	if err != nil {
		// There should already be an epochID in the database so not finding a row is an error
		return 0, err
	}
	return shdb.DecodeUint64(epochID), nil
}

func NewBatchHandler(cfg config.Config, dbpool *pgxpool.Pool) (*BatchHandler, error) {
	ctx := context.Background()

	l2Client, err := rpc.Dial(cfg.SequencerURL)
	if err != nil {
		return nil, err
	}
	l2EthClient := ethclient.NewClient(l2Client)
	chainID, err := l2EthClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	signer := txtypes.LatestSignerForChainID(chainID)

	return &BatchHandler{
		l2Client:    l2Client,
		l2EthClient: l2EthClient,
		config:      cfg,
		signer:      signer,
		txpool:      NewTransactionPool(signer),
		dbpool:      dbpool,
	}, nil
}

// BatchHandler is a threadsafe handler to process the following actions
// - validate and enqueue incoming tx's either to the currently pending batch
//    or for later validation in the transaction pool
// - handle incoming decryption keys and send the currently pending batch-transaction
//    to the sequencer
// - start next epoch and initiate the broadcasting of the decryption-trigger to
//    the keypers
type BatchHandler struct {
	mux      sync.Mutex
	l2Client *rpc.Client
	// l2EthClient is the RPC l2Client wrapped by ethclient.Client
	l2EthClient *ethclient.Client
	config      config.Config
	txpool      *TransactionPool
	signer      txtypes.Signer
	dbpool      *pgxpool.Pool
	latestBatch *Batch
}

func (bh *BatchHandler) Signer() txtypes.Signer {
	return bh.signer
}

// LatestEpochID returns the epoch-id associated to
// the current batch that is yet to be submitted to the sequencer.
func (bh *BatchHandler) LatestEpochID() uint64 {
	return bh.latestBatch.EpochID()
}

func (bh *BatchHandler) privateKey() *ecdsa.PrivateKey {
	return bh.config.EthereumKey
}

// EnqueueTx handles the potential addition of a user's transaction
// to the latest local batch or in case of future transactions to the transaction pool.
// Future transactions can't be queued if they have a batch-index that is too
// far in the future - the allowed difference is set by the `TransactionAcceptanceInterval` (const).
// Transactions that are queued in the transaction pool can't be validated yet,
// because the chain-state (balance, nonce) at the future batch is yet to be known.
// Thus, even successfully queued transactions can't be guaranteed to be considered
// later on for inclusion in the corresponding batch.
// Due to the limitation that a user HAS to encrypt a transaction to a very specific
// batch, a global priority-queue like transaction-pool like on Layer1 ethereum
// is not possible. Users have to re-submit a transaction that did not make it
// to the specific batch explicitly.
func (bh *BatchHandler) EnqueueTx(ctx context.Context, txBytes []byte) error {
	var tx txtypes.Transaction
	err := tx.UnmarshalBinary(txBytes)
	if err != nil {
		return errors.New("can't decode transaction bytes")
	}

	if tx.BatchIndex() > math.MaxUint32 {
		return errors.New("batch index overflow")
	}

	err = bh.dbpool.BeginFunc(ctx, func(dbtx pgx.Tx) error {
		db := cltrdb.New(dbtx)
		currentEpochID, err := GetNextEpochID(ctx, db)
		if err != nil {
			return err
		}

		// we implicitly assume the user always means the current
		// block number
		blockNumber := epochid.BlockNumber(currentEpochID)
		currentBatchIndex := uint64(epochid.SequenceNumber(currentEpochID))

		if tx.BatchIndex() < currentBatchIndex {
			// This will also be the case when we already started the
			// next epoch but did not successfully receive the
			// decryption key and got a confirmation from the
			// sequencer for state inclusion
			return errors.New("historic batch index")
		} else if tx.BatchIndex() > currentBatchIndex+TransactionAcceptanceInterval {
			// only allow future batch indices some batches ahead
			return errors.New("batch too far in the future")
		}

		pending, err := NewPendingTx(bh.Signer(), txBytes)
		if err != nil {
			return err
		}
		// Set the transactions received timestamp to the current time.
		pending.SetReceived(nil)
		txEpoch := epochid.New(uint32(tx.BatchIndex()), blockNumber)

		if err := db.InsertTx(ctx, cltrdb.InsertTxParams{
			TxHash:  tx.Hash().Bytes(),
			EpochID: shdb.EncodeUint64(txEpoch),
			TxBytes: txBytes,
		}); err != nil {
			return errors.Wrap(err, "can't insert tx into db")
		}

		// We could be in the process of updating the latest batch from a different thread,
		// so this needs a sync barrier.
		// This makes sure we don't still forward a transaction to the txpool, although
		// the new batch hast just been initialized with the queued transactions from
		// the txpool.
		bh.mux.Lock()
		defer bh.mux.Unlock()
		var threshold uint64
		if bh.latestBatch == nil {
			// this will push all transactions
			// to the txpool, because we don't have
			// a Batch initialized yet
			threshold = currentBatchIndex - 1
		} else {
			// this will push all transactions
			// other than the latest batch index to the txpool.

			threshold = bh.latestBatch.BatchIndex()
		}
		if tx.BatchIndex() > threshold {
			// push to the tx pool for future processing
			bh.txpool.Push(pending)
			return nil
		}
		// If we are currently in between batches
		// (next epoch started, but latestBatch not updated yet due to incomplete
		// HandleDecryptionKey), no transaction will have passed until this point!
		// This means that all transactions will go to the pool and no transaction
		// goes to the latestBatch during that phase.

		// Don't allow starting a new epoch while that tx is still processed
		_, err = dbtx.Exec(ctx, "LOCK TABLE decryption_trigger IN SHARE ROW EXCLUSIVE MODE")
		if err != nil {
			return err
		}
		// Only transactions with tx.BatchIndex() == bh.latestBatch.BatchIndex()
		// will be pushed to the batch
		return bh.ProcessTx(ctx, pending)
	})

	return err
}

// ProcessTx validates a transaction and includes it in the latest batch,
// if the transaction is valid and applicable.
// Otherwise, an error is thrown specifying why the transaction could not
// be included.
// ProcessTx should only ever be called for transactions that are meant to be
// included specifically in the latest batch.
func (bh *BatchHandler) ProcessTx(ctx context.Context, tx *PendingTransaction) error {
	batch := bh.latestBatch
	if batch == nil {
		return errors.New("batch is not submittable")
	}
	ok, err := batch.ValidateTx(ctx, tx)
	if !ok {
		return errors.Wrap(err, "tx not valid")
	}

	if err != nil {
		return err
	}
	ok, err = batch.ApplyTx(ctx, tx)
	if !ok {
		return errors.Wrap(err, "can't apply tx")
	}

	return err
}

// HandleDecryptionKey handles incoming decryption keys as received by the keypers after the decryption trigger
// was sent out from the collator.
// HandleDecryptionKey finalizes the latestBatch by constructing the corresponding Batch-transaction and
// sending that to the sequencer.
// The function blocks until the sequencer has successfully updated the chain-state with the submitted Batch-transaction,
// and only then creates a new latestBatch. This is due to the latest-batch state validation
// relies on the polled state from the sequencer and only the sequencer knows how the state (balances and nonces)
// progressed during the waited upon state update.
func (bh *BatchHandler) HandleDecryptionKey(ctx context.Context, epochID uint64, decryptionKey []byte) ([]shmsg.P2PMessage, error) {
	var (
		outMessages []shmsg.P2PMessage
		nextEpochID uint64
	)

	if bh.LatestEpochID() != epochID {
		return nil, errors.New("received decryption key for wrong batch")
	}

	// this is the blockNumber the collator originally advertised for the epochID
	// the users will sign that at some point (see #270)
	blockNumber := epochid.BlockNumber(epochID)
	btx, err := bh.latestBatch.SignedBatchTx(bh.privateKey(), decryptionKey, uint64(blockNumber))
	if err != nil {
		return nil, err
	}

	err = sendTransaction(ctx, bh.l2Client, btx)
	if err != nil {
		return nil, err
	}

	// Wait until the batch-tx is processed and the rollup's latest state is updated
	// to include the transactions in that batch
	// We need this because the `Batch` struct relies on the rollup's latest state for tx validation etc.
	err = waitConfirmation(ctx, bh.l2Client, btx)
	if err != nil {
		return nil, err
	}

	// Sync barrier so that incoming transactions will only be funneled to the txpool or batch state
	// after the new batch state has been created
	bh.mux.Lock()
	defer bh.mux.Unlock()

	err = bh.dbpool.BeginFunc(ctx, func(dbtx pgx.Tx) error {
		db := cltrdb.New(dbtx)
		nextEpochID, err = GetNextEpochID(ctx, db)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Only now that the state is updated in the sequencer, we can create the new Batch.
	// The Batch relies on polling the new state from the sequencer.
	// We could track nonces ourselves from the local previous state,
	// but we have to let the sequencer update the balances with the decrypted transactions.
	newBatch, err := NewCachedPendingBatch(ctx, nextEpochID, bh.l2EthClient)
	if err != nil {
		return nil, err
	}
	bh.latestBatch = newBatch
	for _, tx := range bh.txpool.Pop(newBatch.BatchIndex()) {
		err := bh.ProcessTx(ctx, tx)
		if err != nil {
			err = errors.Wrapf(err, "pooled tx invalid (hash=%s), dropped", tx.tx.Hash())
			fmt.Println(err)
			continue
		}
	}
	return outMessages, nil
}

// InitialiseEpoch sets the state of the BatchHandler class upon startup
// and looks for pending, but not updated transactions in the database.
// Note that the collatordb's GetNextEpochID() is used to determine the
// current pending epoch-id.
func (bh *BatchHandler) InitialiseEpoch(ctx context.Context) error {
	return bh.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		// Disallow processing transactions at the same time
		_, err := tx.Exec(ctx, "LOCK TABLE decryption_trigger IN SHARE ROW EXCLUSIVE MODE")
		if err != nil {
			return err
		}
		db := cltrdb.New(tx)
		epochID, err := GetNextEpochID(ctx, db)
		if err != nil {
			return err
		}
		// either read in the latest unsubmitted txs from the DB,
		// or create a new, empty batch
		batch, err := bh.reconstructBatchFromDB(ctx, db, epochID)
		if err != nil {
			return errors.Wrap(err, "batch is non-submittable")
		}
		bh.latestBatch = batch
		return nil
	})
}

// StartNextEpoch puts a DecryptionTrigger in the collatordb to be sent to the keypers.
// StartNextEpoch progresses the pending EpochID to the next value which will stop
// transactions encrypted for the old epoch-id to be accepted by the collator.
func (bh *BatchHandler) StartNextEpoch(ctx context.Context, currentBlockNumber uint32) ([]shmsg.P2PMessage, error) {
	var outMessages []shmsg.P2PMessage

	err := bh.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		// Disallow processing transactions at the same time
		_, err := tx.Exec(ctx, "LOCK TABLE decryption_trigger IN SHARE ROW EXCLUSIVE MODE")
		if err != nil {
			return err
		}

		db := cltrdb.New(tx)
		epochID, err := GetNextEpochID(ctx, db)
		if err != nil {
			return err
		}

		trigger, err := shmsg.NewSignedDecryptionTrigger(
			bh.config.InstanceID, epochID, bh.latestBatch.Transactions().Hash(), bh.config.EthereumKey,
		)
		if err != nil {
			return err
		}

		// Write back the generated trigger to the database
		if err := db.InsertTrigger(ctx, cltrdb.InsertTriggerParams{
			EpochID:   shdb.EncodeUint64(trigger.EpochID),
			BatchHash: trigger.TransactionsHash,
		}); err != nil {
			return err
		}

		nextEpochID := computeNextEpochID(epochID, currentBlockNumber, bh.config.ExecutionBlockDelay)
		if err := db.SetNextEpochID(ctx, shdb.EncodeUint64(nextEpochID)); err != nil {
			return err
		}

		// This will be read from the DB and broadcasted over the P2P Messaging layer.
		// We then wait for the decryption key to be handled by BatchHandler.HandleDecryptionKey()
		outMessages = []shmsg.P2PMessage{trigger}
		return err
	})
	return outMessages, err
}

// reconstructBatchFromDB will re-apply already persisted and thus validated transactions for that epoch
// to the in-memory batch-cache.
// This method should only be used for immediate recovery of lost, unsubmitted batches,
// e.g. due to a crash. It is only applicable for a pending, unsubmitted batch, since it relies on
// the current chain-state of the sequencer node.
func (bh *BatchHandler) reconstructBatchFromDB(ctx context.Context, db *cltrdb.Queries, epochID uint64) (*Batch, error) {
	batch, err := NewCachedPendingBatch(ctx, epochID, bh.l2EthClient)
	if err != nil {
		return nil, err
	}

	// Tx's are returned sorted in the db-insert order, and thus in the submitting
	// insert order.
	// This way we can replay all transactions on the batch
	transactions, err := db.GetTransactionsByEpoch(ctx, shdb.EncodeUint64(epochID))
	if err != nil {
		return nil, err
	}

	for _, txBytes := range transactions {
		var tx txtypes.Transaction
		err := tx.UnmarshalBinary(txBytes)
		if err != nil {
			// This shouldn't happen, since the tx was once instantiated and validated
			// before being written to DB
			err = errors.Wrap(err, "error during unmarshaling transactions from DB")
			fmt.Println(err)
			continue
		}
		p, err := batch.NewPendingTx(&tx, txBytes)
		if err != nil {
			// This shouldn't happen, since the tx was once instantiated and validated
			// before being written to DB
			err = errors.Wrap(err, "error during recovering transactions from DB")
			fmt.Println(err)
			continue
		}
		ok, err := batch.ValidateTx(ctx, p)
		if err != nil || !ok {
			// This shouldn't happen, since the tx was once instantiated and validated
			// before being written to DB
			err = errors.Wrap(err, "validation error during recovering transactions from DB")
			fmt.Println(err)
			continue
		}
		ok, err = batch.ApplyTx(ctx, p)
		if err != nil || !ok {
			// This shouldn't happen, since the tx was once instantiated and validated
			// before being written to DB
			err = errors.Wrap(err, "error during applying recovered transactions from DB")
			fmt.Println(err)
			continue
		}
	}
	return batch, nil
}

// sendTransaction uses the raw rpc.Client instead of the usual ethclient.Client wrapper
// because we want to use the modified txtypes marshaling here instead of the one from the
// go-ethereum repository.
func sendTransaction(ctx context.Context, client *rpc.Client, tx *txtypes.Transaction) error {
	data, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	f := func() (string, error) {
		var result string
		err := client.CallContext(ctx, &result, "eth_sendRawTransaction", hexutil.Encode(data))
		if err != nil {
			return result, err
		}
		return result, nil
	}
	_, err = medley.Retry(ctx, f)
	if err != nil {
		return errors.Wrap(err, "can't send transaction")
	}
	return nil
}

// waitConfirmation polls for the transaction-receipt of the sent out batch-transaction.
// Currently it only retries several times until it fails.
// waitConfirmation returns nil when the transaction has been confirmed successfully.
func waitConfirmation(ctx context.Context, client *rpc.Client, tx *txtypes.Transaction) error {
	f := func() (*txtypes.Receipt, error) {
		var result txtypes.Receipt
		err := client.CallContext(ctx, &result, "eth_getTransactionReceipt", tx.Hash().Hex())
		if err != nil {
			return nil, err
		}
		return &result, nil
	}

	receipt, err := medley.Retry(ctx, f)
	if err != nil {
		return errors.Wrapf(err, "can't retrieve receipt for tx-hash: %s", tx.Hash().Hex())
	}
	if receipt.TxHash.Hex() != tx.Hash().Hex() {
		return errors.New("couldn't poll result")
	}
	if receipt.Status == txtypes.ReceiptStatusFailed {
		return errors.New("receipt status failed")
	}
	return nil
}

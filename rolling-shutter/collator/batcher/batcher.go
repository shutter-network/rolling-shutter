package batcher

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

var (
	ErrWrongChainID             = errors.New("transaction has wrong chainid")
	ErrWongTxType               = errors.New("only encrypted shutter transactions allowed")
	ErrBatchIndexInPast         = errors.New("batch index in the past")
	ErrBatchIndexTooFarInFuture = errors.New("batch index is too far in the future")
	ErrWaitForSequencer         = errors.New("waiting for sequencer to generate a new block")
	ErrBatchAlreadyExists       = errors.New("batch already exists")
	ErrNoEonPublicKey           = errors.New("no eon public key found")
)

type Batcher struct {
	l2Client            L2ClientReader
	l1EthClient         *ethclient.Client
	config              *config.Config
	signer              txtypes.Signer
	dbpool              *pgxpool.Pool
	nextBatchChainState *ChainState
	mux                 sync.Mutex
}

func newBatcherFromClients(
	ctx context.Context,
	cfg *config.Config,
	dbpool *pgxpool.Pool,
	l1EthClient *ethclient.Client,
	l2Client L2ClientReader,
) (*Batcher, error) {
	// This will already do a query to the l2-client
	chainID, err := l2Client.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	signer := txtypes.LatestSignerForChainID(chainID)

	btchr := &Batcher{
		l2Client:            l2Client,
		l1EthClient:         l1EthClient,
		config:              cfg,
		signer:              signer,
		dbpool:              dbpool,
		nextBatchChainState: nil,
	}
	err = btchr.initChainState(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to init the chain state")
		// It's fine to log and ignore this error. We'll be able to handle
		// nextBatchChainState being nil.
	}
	return btchr, nil
}

func NewBatcher(ctx context.Context, cfg *config.Config, dbpool *pgxpool.Pool) (*Batcher, error) {
	// l1 client initialisation
	l1EthClient, err := ethclient.DialContext(ctx, cfg.Ethereum.EthereumURL)
	if err != nil {
		return nil, err
	}

	l2Client, err := NewRPCClient(ctx, cfg.SequencerURL)
	if err != nil {
		return nil, err
	}
	return newBatcherFromClients(ctx, cfg, dbpool, l1EthClient, l2Client)
}

// initializeNextBatch populates the next_batch table with a valid value.
func (btchr *Batcher) initializeNextBatch(
	ctx context.Context,
	db *database.Queries,
) (identitypreimage.IdentityPreimage, uint64, error) {
	var (
		nextIdentityPreimage identitypreimage.IdentityPreimage
		l1BlockNumber        uint64
		err                  error
	)
	l1BlockNumber, err = btchr.l1EthClient.BlockNumber(ctx)
	if err != nil {
		return nextIdentityPreimage, l1BlockNumber, err
	}
	if l1BlockNumber > math.MaxInt64 {
		return nextIdentityPreimage, l1BlockNumber, errors.Errorf("block number too big: %d", l1BlockNumber)
	}

	l2batchIndex, err := btchr.l2Client.GetBatchIndex(ctx)
	if err != nil {
		return nextIdentityPreimage, l1BlockNumber, err
	}
	nextIdentityPreimage = identitypreimage.Uint64ToIdentityPreimage(l2batchIndex + 1)
	err = db.SetNextBatch(ctx, database.SetNextBatchParams{
		EpochID:       nextIdentityPreimage.Bytes(),
		L1BlockNumber: int64(l1BlockNumber),
	})
	return nextIdentityPreimage, l1BlockNumber, err
}

// earlyValidateTx validates a transaction for some basic properties.
func (btchr *Batcher) earlyValidateTx(tx *txtypes.Transaction) error {
	if tx.ChainId().Cmp(btchr.signer.ChainID()) != 0 {
		return ErrWrongChainID
	}
	if tx.Type() != txtypes.ShutterTxType {
		return ErrWongTxType
	}
	return nil
}

func (btchr *Batcher) EnsureChainState(ctx context.Context) error {
	btchr.mux.Lock()
	defer btchr.mux.Unlock()
	if btchr.nextBatchChainState != nil {
		return nil
	}
	return btchr.initChainState(ctx)
}

// initChainState initializes the nextBatchChainState field. It makes sure that the l2Client is up
// to date, i.e. has applied the latest transactions. If the client is not up to date or another
// error happens that prevents us from initializing the field, this method will set the field to
// nil and return an error.
func (btchr *Batcher) initChainState(ctx context.Context) error {
	btchr.nextBatchChainState = nil
	db := database.New(btchr.dbpool)
	nextBatchEpochID, _, err := batchhandler.GetNextBatch(ctx, db)
	if err == pgx.ErrNoRows {
		// the DB does not have a NextBatch set yet.
		// this happens for a first time initialisation
		nextBatchEpochID, _, err = btchr.initializeNextBatch(ctx, db)
		if err != nil {
			return errors.Wrap(err, "failed to initialize 'NextBatch' in DB")
		}
	} else if err != nil {
		return err
	}
	nextBatchIndex := nextBatchEpochID.Uint64()

	l2batchIndex, err := btchr.l2Client.GetBatchIndex(ctx)
	if err != nil {
		return err
	}
	if l2batchIndex >= nextBatchIndex {
		// something is seriously wrong here, as the sequencer has already produced a block
		return ErrBatchAlreadyExists
	} else if l2batchIndex < nextBatchIndex-1 {
		// need to wait for the sequencer to produce the block
		log.Info().Uint64("batch-index", l2batchIndex).Uint64("next-batch-index", nextBatchIndex).
			Msg("wait for sequencer")
		return ErrWaitForSequencer
	}
	block, err := btchr.l2Client.GetBlockInfo(ctx)
	if err != nil {
		return err
	}
	btchr.nextBatchChainState = NewChainState(
		btchr.signer,
		block.BaseFee(),
		block.GasLimit(),
		nextBatchEpochID,
	)
	err = btchr.loadAndApplyTransactions(ctx, db)
	if err != nil {
		// If we fail to load and apply the transaction, the state in nextBatchChainState
		// is invalid. So, let's set it to nil.
		btchr.nextBatchChainState = nil
		return err
	}
	log.Info().Uint64("batch-index", btchr.nextBatchChainState.identityPreimage.Uint64()).
		Msg("loaded chain state")
	return nil
}

// loadAndApplyTransactions loads transactions from the database for the current batch.
func (btchr *Batcher) loadAndApplyTransactions(ctx context.Context, db *database.Queries) error {
	txs, err := db.GetNonRejectedTransactionsByEpoch(ctx, btchr.nextBatchChainState.identityPreimage.Bytes())
	if err != nil {
		return err
	}

	unmarshalledTxs, err := database.UnmarshalTransactions(txs)
	if err != nil {
		return err
	}
	err = btchr.ensureAccountsInitialized(ctx, unmarshalledTxs)
	if err != nil {
		return err
	}
	err = btchr.applyTransactions(ctx, unmarshalledTxs, txs)
	return err
}

// applyTransactions tries to apply each transaction from the given list of transactions. The
// caller must make sure that each sender account is already initialized and only transactions with
// status 'new' or 'committed' are passed to this function. This function updates the state of
// 'new' transactions to either 'committed' or 'rejected', i.e. it commits to transactions being
// included in the current batch.
func (btchr *Batcher) applyTransactions(
	ctx context.Context,
	unmarshalledTxs []txtypes.Transaction,
	txs []database.Transaction,
) error {
	db := database.New(btchr.dbpool)
	for i := range unmarshalledTxs {
		err := btchr.nextBatchChainState.CanApplyTx(
			&unmarshalledTxs[i],
			uint64(len(txs[i].TxBytes)),
		)
		if txs[i].Status == database.TxstatusNew {
			var newStatus database.Txstatus
			if err == nil {
				newStatus = database.TxstatusCommitted
			} else {
				newStatus = database.TxstatusRejected
			}
			err = db.SetTransactionStatus(ctx, database.SetTransactionStatusParams{
				TxHash: txs[i].TxHash,
				Status: newStatus,
			})
			if err != nil {
				return err
			}
			if newStatus == database.TxstatusCommitted {
				btchr.nextBatchChainState.ApplyTx(&unmarshalledTxs[i], uint64(len(txs[i].TxBytes)))
			}
		} else if err == nil {
			btchr.nextBatchChainState.ApplyTx(&unmarshalledTxs[i], uint64(len(txs[i].TxBytes)))
		} else if err != nil {
			panic("Cannot apply committed tx")
		}
	}
	return nil
}

func (btchr *Batcher) closeBatchImpl(
	ctx context.Context,
	db *database.Queries,
	l1blockNumber int64,
) error {
	nextBatchEpochID, _, err := batchhandler.GetNextBatch(ctx, database.New(btchr.dbpool))
	if err != nil {
		return err
	}

	// Lookup the current eon public key for the given block.
	_, err = db.FindEonPublicKeyForBlock(ctx, l1blockNumber)
	if err == pgx.ErrNoRows {
		log.Info().Int64("l1BlockNumber", l1blockNumber).Msg("no eon public key found")
		return ErrNoEonPublicKey
	} else if err != nil {
		return err
	}

	// Mark all new TXs as rejected
	err = db.RejectNewTransactions(ctx, nextBatchEpochID.Bytes())
	if err != nil {
		return err
	}
	txs, err := db.GetCommittedTransactionsByEpoch(ctx, nextBatchEpochID.Bytes())
	if err != nil {
		return err
	}
	txsHash := hashTransactions(txs)

	// Write back the generated trigger to the database
	err = db.InsertTrigger(ctx, database.InsertTriggerParams{
		EpochID:       nextBatchEpochID.Bytes(),
		BatchHash:     txsHash,
		L1BlockNumber: l1blockNumber,
	})
	if err != nil {
		return err
	}

	newEpoch := batchhandler.ComputeNextEpochID(nextBatchEpochID)

	return db.SetNextBatch(ctx, database.SetNextBatchParams{
		EpochID:       newEpoch.Bytes(),
		L1BlockNumber: l1blockNumber,
	})
}

// CloseBatch closes the current batch.
func (btchr *Batcher) CloseBatch(ctx context.Context) error {
	btchr.mux.Lock()
	defer btchr.mux.Unlock()

	if btchr.nextBatchChainState == nil {
		err := btchr.initChainState(ctx)
		if err != nil {
			return err
		}
	}

	l1blockNumber, err := btchr.l1EthClient.BlockNumber(ctx)
	if err != nil {
		return err
	}

	err = btchr.dbpool.BeginFunc(ctx, func(dbtx pgx.Tx) error {
		return btchr.closeBatchImpl(ctx, database.New(dbtx), int64(l1blockNumber))
	})
	if err != nil {
		return err
	}
	btchr.nextBatchChainState = nil
	return nil
}

// EnqueueTx handles the potential addition of a user's transaction to the latest local batch or in
// case of future transactions to the transaction pool.
// Future transactions can't be queued if they have a batch-index that is too far in the future.
// The allowed difference is set by the `BatchIndexAcceptenceInterval` config entry.
// Transactions that are queued in the transaction pool can't be validated yet, because the
// chain-state (balance, nonce) at the future batch is yet to be known.
// Thus, even successfully queued transactions can't be guaranteed to be considered later on for
// inclusion in the corresponding batch.
// Due to the limitation that a user has to encrypt a transaction to a very specific
// batch, a global priority-queue-like transaction-pool (as on layer-1 Ethereum)
// is not possible. Users have to re-submit a transaction that did not make it
// to the specific batch explicitly.
func (btchr *Batcher) EnqueueTx(ctx context.Context, txBytes []byte) error {
	var err error
	receiveTime := time.Now()
	_ = receiveTime
	tx := &txtypes.Transaction{}
	err = tx.UnmarshalBinary(txBytes)
	if err != nil {
		return err
	}

	err = btchr.earlyValidateTx(tx)
	if err != nil {
		return err
	}
	// Please consider the following as part of the tx validation. We should be able to extract
	// the sender from the transaction. Do not move this to the 'if txInNextBatch' branch.
	account, err := btchr.signer.Sender(tx)
	if err != nil {
		return err
	}

	btchr.mux.Lock()
	defer btchr.mux.Unlock()

	if btchr.nextBatchChainState == nil {
		err = btchr.initChainState(ctx)
		if err != nil {
			log.Info().Err(err).Msg("cannot load chain state")
		}
	}

	db := database.New(btchr.dbpool)
	nextBatchEpochID, _, err := batchhandler.GetNextBatch(ctx, db)
	if err != nil {
		return err
	}
	nextBatchIndex := nextBatchEpochID.Uint64()

	if tx.BatchIndex() < nextBatchIndex {
		return ErrBatchIndexInPast
	} else if tx.BatchIndex() >= nextBatchIndex+uint64(btchr.config.BatchIndexAcceptenceInterval) {
		return ErrBatchIndexTooFarInFuture
	}

	txInNextBatch := btchr.nextBatchChainState != nil && tx.BatchIndex() == nextBatchIndex

	txstatus := database.TxstatusNew
	if txInNextBatch {
		// If the tx goes into the next batch, we ensure it can be applied by calling
		// CanApplyTx after making sure we have the current nonce and balance for the
		// sender's account.
		err = btchr.ensureAccountInitialized(ctx, account)
		if err != nil {
			return err
		}
		err = btchr.nextBatchChainState.CanApplyTx(tx, uint64(len(txBytes)))
		if err != nil {
			return err
		}
		txstatus = database.TxstatusCommitted
	}

	err = btchr.dbpool.BeginFunc(ctx, func(dbtx pgx.Tx) error {
		identityPreimage := identitypreimage.Uint64ToIdentityPreimage(tx.BatchIndex()).Bytes()
		return database.New(dbtx).InsertTx(ctx, database.InsertTxParams{
			TxHash:  tx.Hash().Bytes(),
			EpochID: identityPreimage,
			TxBytes: txBytes,
			Status:  txstatus,
		})
	})
	if err != nil {
		return err
	}

	if txInNextBatch {
		btchr.nextBatchChainState.ApplyTx(tx, uint64(len(txBytes)))
	}

	return nil
}

// ensureAccountInitialized ensures that we do have the nonce and balance stored in
// nextBatchChainState for the given address. It uses the l2EthClient to get that information via
// RPC if necessary.
func (btchr *Batcher) ensureAccountInitialized(ctx context.Context, account common.Address) error {
	if !btchr.nextBatchChainState.IsAccountInitialized(account) {
		info, err := btchr.l2Client.GetAccountInfo(ctx, account)
		if err != nil {
			return err
		}
		btchr.nextBatchChainState.InitializeAccount(account, info.Balance, info.Nonce)
	}
	return nil
}

// ensureAccountsInitialized ensures that we do have the nonce and balance stored in
// nextBatchChainState for all senders of the given transactions.
func (btchr *Batcher) ensureAccountsInitialized(
	ctx context.Context,
	txs []txtypes.Transaction,
) error {
	for i := range txs {
		account, err := btchr.signer.Sender(&txs[i])
		if err != nil {
			return err
		}

		err = btchr.ensureAccountInitialized(ctx, account)
		if err != nil {
			return err
		}
	}
	return nil
}

func hashTransactions(txs []database.Transaction) []byte {
	txHashes := make([][]byte, len(txs))
	for i, t := range txs {
		txHashes[i] = t.TxHash
	}
	// Hash the list of transaction hashes
	return p2pmsg.HashByteList(txHashes)
}

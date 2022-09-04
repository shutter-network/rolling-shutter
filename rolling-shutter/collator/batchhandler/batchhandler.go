package batchhandler

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/batch"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

const (
	//  The maximum number of batches in the pool of Batches
	SizeBatchPool = 5
)

// computeNextEpochID takes an epoch id as parameter and returns the id of the epoch following it.
func computeNextEpochID(epochID epochid.EpochID) (epochid.EpochID, error) {
	n := epochID.Big()
	nextN := new(big.Int).Add(n, common.Big1)
	return epochid.BigToEpochID(nextN)
}

// GetNextBatch gets the epochID and block number that will be used in the next batch.
func GetNextBatch(ctx context.Context, db *cltrdb.Queries) (epochid.EpochID, uint64, error) {
	b, err := db.GetNextBatch(ctx)
	if err != nil {
		// There should already be an epochID in the database so not finding a row is an error
		return epochid.EpochID{}, 0, err
	}
	epochID, err := epochid.BytesToEpochID(b.EpochID)
	if err != nil {
		return epochid.EpochID{}, 0, err
	}
	if b.L1BlockNumber < 0 {
		return epochid.EpochID{}, 0, errors.Errorf("negative l1 block number in db")
	}
	l1BlockNumber := uint64(b.L1BlockNumber)
	return epochID, l1BlockNumber, nil
}

// NewBatchHandler initializes a new instance of BatchHandler.
// NewBatchHandler connects to the sequencer and queries
// some node information (chain-id, latest block) and if we recover
// pending transactions from the database it will also query state
// information for the corresponding accounts (nonce, balance) in
// order to validate and apply the transactions to the current pending batch.
func NewBatchHandler(cfg config.Config, dbpool *pgxpool.Pool) (*BatchHandler, error) {
	ctx := context.Background()

	l2Client, err := rpc.Dial(cfg.SequencerURL)
	if err != nil {
		return nil, err
	}
	l1EthClient, err := ethclient.Dial(cfg.EthereumURL)
	if err != nil {
		return nil, err
	}
	l2EthClient := ethclient.NewClient(l2Client)
	// This will already do a query to the l2-client
	chainID, err := l2EthClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	signer := txtypes.LatestSignerForChainID(chainID)

	bh := &BatchHandler{
		l2Client:               l2Client,
		l2EthClient:            l2EthClient,
		l1EthClient:            l1EthClient,
		config:                 cfg,
		signer:                 signer,
		dbpool:                 dbpool,
		mux:                    sync.RWMutex{},
		batches:                []*batch.Batch{},
		outMessage:             make(chan shmsg.P2PMessage),
		ticker:                 nil,
		confirmedBatchesBroker: nil,
		confirmedBatch:         make(chan epochid.EpochID),
		stopSignal:             make(chan struct{}),
	}
	return bh, nil
}

type BatchHandler struct {
	l2Client *rpc.Client
	// l2EthClient is the RPC l2Client wrapped by ethclient.Client
	l2EthClient *ethclient.Client
	l1EthClient *ethclient.Client
	config      config.Config
	signer      txtypes.Signer
	dbpool      *pgxpool.Pool
	mux         sync.RWMutex
	batches     []*batch.Batch
	outMessage  chan shmsg.P2PMessage
	ticker      *EpochTicker
	stopSignal  chan struct{}

	confirmedBatchesBroker *medley.Broker[epochid.EpochID]
	confirmedBatch         chan epochid.EpochID
}

func (bh *BatchHandler) Signer() txtypes.Signer {
	return bh.signer
}

func (bh *BatchHandler) Messages() <-chan shmsg.P2PMessage {
	return bh.outMessage
}

func (bh *BatchHandler) ConfirmedBatch() chan<- epochid.EpochID {
	return bh.confirmedBatch
}

// EnqueueTx handles the potential addition of a user's transaction
// to the latest local batch or in case of future transactions to the transaction pool.
// Future transactions can't be queued if they have a batch-index that is too
// far in the future - the allowed difference is set by the `TransactionAcceptanceInterval` (const).
// Transactions that are queued in the transaction pool can't be validated yet,
// because the chain-state (balance, nonce) at the future batch is yet to be known.
// Thus, even successfully queued transactions can't be guaranteed to be considered
// later on for inclusion in the corresponding batch.
// Due to the limitation that a user has to encrypt a transaction to a very specific
// batch, a global priority-queue-like transaction-pool (as on layer-1 Ethereum)
// is not possible. Users have to re-submit a transaction that did not make it
// to the specific batch explicitly.
// The return value is the promise of the transaction result.
// For a successful transaction, this will only receive a value once the transaction's batch
// has been confirmed by the sequencer.
// An unsuccessful transaction can fail anytime in-between submittal and delegation to the sequencer,
// but mainly fails on validation as the Batch becomes pending.
func (bh *BatchHandler) EnqueueTx(ctx context.Context, txBytes []byte) <-chan transaction.Result {
	var tx txtypes.Transaction
	// create the result promise.
	result := make(chan transaction.Result, 1)
	err := tx.UnmarshalBinary(txBytes)
	if err != nil {
		result <- transaction.Result{
			Err:     errors.New("can't decode transaction bytes"),
			Success: false,
		}
		return result
	}
	if tx.Type() != txtypes.ShutterTxType {
		result <- transaction.Result{
			Err:     errors.New("only encrypted shutter transactions allowed"),
			Success: false,
		}
		return result
	}
	txEpoch := epochid.Uint64ToEpochID(tx.BatchIndex())

	receiveTime := time.Now()
	pending, err := transaction.NewPending(bh.Signer(), txBytes, receiveTime)
	if err != nil {
		result <- transaction.Result{Err: err, Success: false}
		return result
	}

	bh.mux.RLock()
	// dispatch to the correct `Batch` object
	b := bh.getBatch(txEpoch)
	bh.mux.RUnlock()

	if b == nil {
		result <- transaction.Result{
			Err:     errors.Errorf("no batch found in batchhandler for epoch-id %s", txEpoch.String()),
			Success: false,
		}
		return result
	}
	select {
	case b.Transaction <- pending:
	case <-ctx.Done():
		result <- transaction.Result{
			Err:     errors.New("server stopped"),
			Success: false,
		}
		return result
	}
	return pending.Result
}

// Run is the main entrypoint for the BatchHandler's
// go-routines.
// Run initializes the pool of currently active batches
// and starts the go-routines that handle the individual
// StateChangeResult's of the Batches state-transitions.
func (bh *BatchHandler) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start the EpochTicker in order to distribute
	// new epoch events to the batches.
	bh.ticker = StartNewEpochTicker(bh.config.EpochDuration)
	defer bh.ticker.Stop()

	// Start the broker that reads the newly confirmed
	// batches from the sequencer and distributes that
	// to the batches.
	bh.confirmedBatchesBroker = medley.StartNewBroker[epochid.EpochID](false)
	defer close(bh.confirmedBatchesBroker.Publish)

	// Fill the pool of batches based on the `SizeBatchPool` constant
	// and start the batches Run go-routines.
	for i := 0; i < SizeBatchPool; i++ {
		_, err := bh.appendHeadBatch(ctx)
		if err != nil {
			return errors.Wrap(err, "couldn't create new batch")
		}
	}

	// if any of the spawned HandleStateTransitions
	// and batch.Run() methods return an error, this context is
	// canceled and thus all other routines under this
	// errorgroup's umbrella are canceled
	eg, errctx := errgroup.WithContext(ctx)

	bh.mux.RLock()
	for _, b := range bh.batches {
		// new variable to avoid the loop variable b being
		// captured by func literal
		bb := b
		eg.Go(func() error {
			// Start the StateChangeResult handler for all the batches
			// in the pool of batches.
			return bh.HandleStateTransitions(errctx, eg, bb)
		})
	}
	bh.mux.RUnlock()
	for {
		select {
		case val, ok := <-bh.confirmedBatch:
			if !ok {
				bh.confirmedBatch = nil
				continue
			}
			bh.confirmedBatchesBroker.Publish <- val
		case <-bh.stopSignal:
			// deferred cancel cancels the HandleStateTransitions
			return nil
		case <-errctx.Done():
			return errctx.Err()
		}
	}
}

func (bh *BatchHandler) getInitialEpoch(ctx context.Context) (epochid.EpochID, uint64, error) {
	var (
		epochID       epochid.EpochID
		l1BlockNumber uint64
	)
	err := bh.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		var err error
		db := cltrdb.New(tx)
		epochID, l1BlockNumber, err = GetNextBatch(ctx, db)
		if err == pgx.ErrNoRows {
			// TODO query the current blocknumber
			// TODO initial epoch maybe a config parameter?
			return err
		} else if err != nil {
			return err
		} else {
			return nil
		}
	})
	return epochID, l1BlockNumber, err
}

func (bh *BatchHandler) Stop() {
	close(bh.stopSignal)
}

// appendHeadBatch initializes a new batch for the next future epoch-id that is not yet
// in the pool of batches and appends that to the "HEAD" position of the pool of batches.
func (bh *BatchHandler) appendHeadBatch(ctx context.Context) (*batch.Batch, error) {
	bh.mux.Lock()
	var (
		headBatch     *batch.Batch
		epochID       epochid.EpochID
		l1BlockNumber uint64
		err           error
	)
	if len(bh.batches) == 0 {
		headBatch = nil
		epochID, l1BlockNumber, err = bh.getInitialEpoch(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		headBatch = bh.batches[len(bh.batches)-1]
		epochID, _ = computeNextEpochID(headBatch.EpochID())
		// TODO guess the l1blocknumber based on recent timings
		l1BlockNumber = uint64(42)
	}

	newBatch, err := batch.New(ctx, bh.config.InstanceID, epochID, l1BlockNumber, bh.l2EthClient, headBatch)
	if err != nil {
		return nil, err
	}
	bh.batches = append(bh.batches, newBatch)
	newBatch.Log().Debug().
		Int("num-batches", len(bh.batches)).
		Str("block-number", fmt.Sprint(newBatch.L1BlockNumber)).
		Msg("added HEAD batch to batch pool")
	bh.mux.Unlock()
	return newBatch, nil
}

func (bh *BatchHandler) getBatch(epoch epochid.EpochID) *batch.Batch {
	for _, candidate := range bh.batches {
		if bytes.Equal(candidate.EpochID().Bytes(), epoch.Bytes()) {
			return candidate
		}
	}
	return nil
}

// `removeBatch` removes a batch from the pool of batches.
// Should be called after the batch has stopped it's state-transition
// and reached it's end of lifetime (confirmed and applied in the rollup).
func (bh *BatchHandler) removeBatch(b *batch.Batch) {
	bh.mux.Lock()
	defer bh.mux.Unlock()

	index := -1
	for i, a := range bh.batches {
		if a == b {
			index = i
			break
		}
	}
	if index >= 0 {
		// remove the batch at the found index from the slice via re-slicing
		bh.batches = append(bh.batches[:index], bh.batches[index+1:]...)
		log.Debug().Int("num-batches", len(bh.batches)).Msg("removed batch from pool")
	}
}

// HandleStateTransitions tracks the StateChangeResults emitted by a certain batch
// and takes action on some state-changes:
// - when the batch becomes "committed",
// a new HEAD-batch will be appended to the pool of batches
// - when the batch becomes "confirmed" or the context is canceled,
// the cleanup function is called, the batch is stopped
// and the HandleStateTransitions returns with a optional error as return value.
// HandleStateTransitions takes care of registering and deregistering
// observers for the confirmed batches and epoch ticks.
// It also delegates the input to and the output from the batches state-transition
// loop:
// - received confirmed batches and epoch ticks from the respective broker subscriptions are
// delegated to the batches Input channels
// - outbound P2P-messages and sequencer transactions are processed from the StateChangeResult
// and delegated to the messaging / rpc layer.
// - errors occurring during the state changes are logged.
func (bh *BatchHandler) HandleStateTransitions(ctx context.Context, eg *errgroup.Group, b *batch.Batch) error {
	var stopChan chan error

	batchState := b.Broker.Subscribe(1)
	confirmedBatch := bh.confirmedBatchesBroker.Subscribe(0)
	epochTick := bh.ticker.Subscribe()

	eg.Go(func() error {
		return b.Run(ctx, epochTick)
	})

	stop := func() {
		b.Broker.Unsubscribe(batchState)
		bh.removeBatch(b)
		bh.ticker.Unsubscribe(epochTick)
		bh.confirmedBatchesBroker.Unsubscribe(confirmedBatch)
		stopChan = b.Stop()
		b.Log().Debug().Msg("stop func executed")
	}
	ctxDone := ctx.Done()

	for {
		select {
		case stopErr := <-stopChan:
			return stopErr

		case confirmed, ok := <-confirmedBatch:
			if !ok {
				confirmedBatch = nil
				continue
			}
			b.Log().Debug().Msg("received confirmed batch in batchhandler")
			b.ConfirmedBatch <- confirmed

		case stateTransition, ok := <-batchState:
			if !ok {
				batchState = nil
				continue
			}
			b.Log().Debug().Int("to-state", int(stateTransition.ToState)).Msg("received state change in batchhandler")
			switch stateTransition.ToState {
			case batch.NoState:
			case batch.InitialState:
			case batch.PendingState:
			case batch.CommittedState:
				// add a new batch to the list of batches
				headBatch, err := bh.appendHeadBatch(ctx)
				if err != nil {
					b.Log().Error().Err(err).Msg("couldn't create new batch, stopping.")
					// stopping this batch and by means of the error group
					// all other batches.
					// we don't have any graceful handling of this currently,
					// so it's best to shut down the batchhandler entirely
					stop()
					// this ensures stop is not called by other means twice
					stop = func() {}
					continue
				}

				// run the state-transitions of a new batch at the head
				// of the list of batches.
				// this initially allows for transactions to be queued up in the batch
				eg.Go(func() error {
					return bh.HandleStateTransitions(ctx, eg, headBatch)
				})
			case batch.DecryptedState:
			case batch.ConfirmedState:
				stop()
				// this ensures the cleanup is not called by ctx.Done() as well
				stop = func() {}
			case batch.StoppingState:
			}
			for _, transitionError := range stateTransition.Errors {
				if transitionError.Err != nil {
					b.Log().Error().Err(transitionError.Err).Msg("error occurred during state-transition")
				}
			}
			for _, mess := range stateTransition.P2PMessages {
				err := bh.OnOutgoingMessage(ctx, mess)
				if err != nil {
					b.Log().Error().Err(err).Msg("error while handling state-transition messages in batch-handler")
				}
			}
			for _, tx := range stateTransition.SequencerTransactions {
				err := bh.OnSequencerTransaction(ctx, tx)
				if err != nil {
					b.Log().Error().Err(err).Msg("error while handling state-transition transactions in batch-handler")
				}
			}
		case <-ctxDone:
			b.Log().Debug().Msg("context canceled, stopping in HandleStateTransitions")
			ctxDone = nil
			stop()
			// this ensures the cleanup is not called by "Confirmed" state-transition
			// as well
			stop = func() {}
		}
	}
}

// OnOutgoingMessage signs and dispatches outbound P2P-messages
// to the correct handler functions based on protoreflect type introspection.
// Finally the message will be handed to the messaging layer via it's
// send channel.
// Note: This operation blocks if the messaging layer does not consume
// the sent message!
func (bh *BatchHandler) OnOutgoingMessage(ctx context.Context, msg shmsg.P2PMessage) error {
	typ := p2p.GetMessageType(msg)
	if typ == "shmsg.DecryptionTrigger" {
		trigger, _ := msg.(*shmsg.DecryptionTrigger)
		err := shmsg.Sign(trigger, bh.config.EthereumKey)
		if err != nil {
			return err
		}
		msg = trigger
		err = bh.OnOutboundDecryptionTrigger(ctx, trigger)
		if err != nil {
			log.Error().Err(err).Msg("error during processing of DecryptionTrigger")
		}
	}
	bh.outMessage <- msg
	return nil
}

// OnSequencerTransaction signs and sends outbound transactions (mainly BatchTx)
// to the sequencers JSON-RPC endpoint.
// Note: The send operation eventually retries sends if the sequencer is unresponsive
// or an error is returned - the "OnSequencerTransaction" method-call is thus blocking during
// the whole send operation.
func (bh *BatchHandler) OnSequencerTransaction(ctx context.Context, tx txtypes.TxData) error {
	signedTx, err := txtypes.SignNewTx(bh.config.EthereumKey, bh.signer, tx)
	if err != nil {
		// that's a bug!
		return err
	}
	return sendTransaction(ctx, bh.l2Client, signedTx)
}

func (bh *BatchHandler) HandleDecryptionKey(_ context.Context, epochID epochid.EpochID, key []byte) error {
	log.Debug().Str("epoch", epochID.String()).Msg("received decryption key")
	bh.mux.RLock()
	b := bh.getBatch(epochID)
	bh.mux.RUnlock()

	if b == nil {
		err := errors.Errorf("no batch found in batchhandler for epoch-id %s", epochID.String())
		log.Debug().Str("epoch", epochID.String()).Err(err).Msg("error during handle decryption key")
		return nil
	}
	b.DecryptionKey <- key
	return nil
}

// OnOutboundDecryptionTrigger puts a DecryptionTrigger in the collatordb to be sent to the keypers.
// OnOutboundDecryptionTrigger progresses the pending EpochID to the next value which will stop
// transactions encrypted for the old epoch-id to be accepted by the collator.
// TODO maybe pass the batch as argument instead of getting it from the handler
func (bh *BatchHandler) OnOutboundDecryptionTrigger(ctx context.Context, trigger *shmsg.DecryptionTrigger) error {
	epochID, err := epochid.BytesToEpochID(trigger.GetEpochID())
	if err != nil {
		return errors.Wrap(err, "can't decode epoch-id bytes from DecryptionTrigger msg")
	}

	bh.mux.RLock()
	b := bh.getBatch(epochID)
	bh.mux.RUnlock()

	if b == nil {
		return errors.Errorf("no batch found in batchhandler for epoch-id %s", epochID.String())
	}

	err = bh.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := cltrdb.New(tx)

		var err error
		trigger, err = shmsg.NewSignedDecryptionTrigger(
			bh.config.InstanceID,
			epochID,
			b.L1BlockNumber,
			b.Hash(),
			bh.config.EthereumKey,
		)
		if err != nil {
			return err
		}

		// Write back the generated trigger to the database
		if err := db.InsertTrigger(ctx, cltrdb.InsertTriggerParams{
			EpochID:   trigger.EpochID,
			BatchHash: trigger.TransactionsHash,
		}); err != nil {
			return err
		}
		return err
	})
	return err
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
		//
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
	return err
}

func GetBlockNumber(ctx context.Context, client *ethclient.Client) (uint64, error) {
	blk, err := medley.Retry(ctx, func() (uint64, error) {
		return client.BlockNumber(ctx)
	})
	if err != nil {
		return 0, err
	}
	return blk, nil
}

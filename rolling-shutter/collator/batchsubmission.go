package collator

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batcher"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/client"
)

type Submitter struct {
	l1Client  *ethclient.Client
	l2Client  batcher.L2ClientReader
	dbpool    *pgxpool.Pool
	privKey   *ecdsa.PrivateKey
	signer    txtypes.Signer
	sequencer *client.Client
	collator  *collator
}

func NewSubmitter(
	ctx context.Context,
	cfg *config.Config,
	dbpool *pgxpool.Pool,
) (*Submitter, error) {
	l1Client, err := ethclient.Dial(cfg.Ethereum.EthereumURL)
	if err != nil {
		return nil, err
	}

	l2Client, err := batcher.NewRPCClient(ctx, cfg.SequencerURL)
	if err != nil {
		return nil, err
	}
	chainID, err := l2Client.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	signer := txtypes.LatestSignerForChainID(chainID)
	sequencer, err := client.DialContext(ctx, cfg.SequencerURL)
	if err != nil {
		return nil, err
	}
	return &Submitter{
		l1Client:  l1Client,
		l2Client:  l2Client,
		dbpool:    dbpool,
		signer:    signer,
		privKey:   cfg.Ethereum.PrivateKey.Key,
		sequencer: sequencer,
	}, nil
}

// createBatchTx creates the batchtx for the given epoch.
func (submitter *Submitter) createBatchTx(
	ctx context.Context,
	db *cltrdb.Queries,
	epoch epochid.EpochID,
) error {
	decryptionKey, err := db.GetDecryptionKey(ctx, epoch.Bytes())
	if err == pgx.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	logger := log.With().Uint64("epoch", epoch.Uint64()).Logger()
	defer func() {
		if err != nil {
			logger.Error().Err(err).Msg("could not create batchtx")
		}
	}()

	txs, err := db.GetCommittedTransactionsByEpoch(ctx, epoch.Bytes())
	if err != nil {
		return err
	}

	transactions := [][]byte{}
	for _, t := range txs {
		transactions = append(transactions, t.TxBytes)
	}
	// XXX the collator will advertise the L1 block number in the future. Users will include
	// this L1 block number in their signed transactions and the collator will also include it
	// as part of the batchtx. Since we did not implement that functionality yet, we just use
	// the current L1 block number as a temporary workaround.
	l1blocknum, err := submitter.l1Client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	batchTx := txtypes.BatchTx{
		ChainID:       submitter.signer.ChainID(),
		DecryptionKey: decryptionKey.DecryptionKey,
		BatchIndex:    epoch.Uint64(),
		L1BlockNumber: l1blocknum,
		Timestamp:     big.NewInt(time.Now().Unix()),
		Transactions:  transactions,
	}
	tx, err := txtypes.SignNewTx(submitter.privKey, submitter.signer, &batchTx)
	if err != nil {
		return err
	}
	txbytes, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	err = db.InsertBatchTx(ctx, cltrdb.InsertBatchTxParams{
		EpochID:   epoch.Bytes(),
		Marshaled: txbytes,
	})
	if err != nil {
		return err
	}
	logger.Info().Interface("batch-tx", batchTx).Int("num-shutter-tx", len(txs)).Msg("created batchtx and inserted in db")
	return nil
}

// submitBatchTxToSequencer reads the unsubmitted batchtx from the database and tries to submit it
// to the sequencer.
func (submitter *Submitter) submitBatchTxToSequencer(ctx context.Context) error {
	db := cltrdb.New(submitter.dbpool)
	unsubmitted, err := db.GetUnsubmittedBatchTx(ctx)
	if err == pgx.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}
	epoch, err := epochid.BytesToEpochID(unsubmitted.EpochID)
	if err != nil {
		return err
	}
	// XXX is this expected to be +1 the processed batch???
	l2BatchIndex, err := submitter.l2Client.GetBatchIndex(ctx)
	if err != nil {
		return err
	}

	// XXX why is this called here? to keep the "state-machine"
	// running?
	// -> this calls the submitBatch
	defer submitter.collator.signals.newDecryptionKey()

	// This will set the batch submitted,
	// since the batch index did progress to this batch
	if epoch.Uint64() <= l2BatchIndex {
		return db.SetBatchSubmitted(ctx)
	}

	// otherwise we submit the batch
	// FIXME method handler crashed
	_, err = submitter.sequencer.SubmitBatchData(ctx, unsubmitted.Marshaled)
	log.Info().
		Uint64("epoch-id", epoch.Uint64()).
		Uint64("batch-index", l2BatchIndex).
		Err(err).
		Msg("submitted batch data")
	if err != nil {
		time.Sleep(time.Second * 5)
	}
	return err
}

func (submitter *Submitter) submitBatch(ctx context.Context) error {
	db := cltrdb.New(submitter.dbpool)
	unsubmitted, err := db.GetUnsubmittedBatchTx(ctx)
	if err == nil {
		// FIXME this happens for every epoch, although it was already
		// submitted
		log.Info().
			Hex("unsubmitted-epoch", unsubmitted.EpochID).
			Msg("still have an unsubmitted batch")
		// -> this calls the submitBatchTxToSequencer,
		// and if the blocknumber now progressed, will mark the batch
		// submitted, return
		// early and trigger this submitBatch again
		submitter.collator.signals.newBatchTx()
		return nil
	} else if err != pgx.ErrNoRows {
		return err
	}

	// XXX can this be lagging?
	// then we should check wether the batch was already submitted,
	// and trigger the function again
	l2BatchIndex, err := submitter.l2Client.GetBatchIndex(ctx)
	if err != nil {
		return err
	}
	epochIDToSubmit := epochid.Uint64ToEpochID(l2BatchIndex + 1)
	return submitter.createBatchTx(ctx, db, epochIDToSubmit)
}

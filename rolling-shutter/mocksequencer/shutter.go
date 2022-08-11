package mocksequencer

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"
)

type ShutterService struct {
	processor *SequencerProcessor
}

var _ RPCService = (*ShutterService)(nil)

func (s *ShutterService) injectProcessor(p *SequencerProcessor) {
	s.processor = p
}

func (s *ShutterService) name() string {
	return "shutter"
}

func (s *ShutterService) GetBatchIndex() (string, error) {
	return hexutil.EncodeUint64(s.processor.batchIndex), nil
}

func (s *ShutterService) validateBatch(tx *txtypes.Transaction) error {
	if tx.Type() != txtypes.BatchTxType {
		return errors.New("unexpected transaction type")
	}

	if tx.ChainId().Cmp(s.processor.chainID) != 0 {
		return errors.New("chain-id mismatch")
	}

	if s.processor.batchIndex != tx.BatchIndex()-1 {
		return errors.New("incorrect batch-index for next batch")
	}

	collator, err := s.processor.collators.Find(tx.L1BlockNumber())
	if err != nil {
		return errors.Wrap(err, "collator validation failed")
	}
	sender, err := s.processor.signer.Sender(tx)
	if err != nil {
		return errors.Wrap(err, "error recovering batch tx sender")
	}
	if collator != sender {
		return errors.Wrap(err, "not signed by correct collator")
	}

	// all checks passed, the batch-tx is valid (disregarding validity of included encrypted transactions)
	return nil
}

func (s *ShutterService) SubmitBatch(ctx context.Context, batchTransaction string) (string, error) {
	var (
		tx      txtypes.Transaction
		gasPool core.GasPool
	)

	txBytes, err := hexutil.Decode(batchTransaction)
	if err != nil {
		return "", errors.Wrap(err, "can't decode incoming tx bytes")
	}

	err = tx.UnmarshalBinary(txBytes)
	if err != nil {
		return "", errors.Wrap(err, "can't unmarshal incoming bytes to transaction")
	}

	err = s.validateBatch(&tx)
	txStr, _ := tx.MarshalJSON()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).RawJSON("transaction", txStr).Msg("received invalid batch transaction")
		return "", errors.Wrap(err, "batch-tx invalid")
	}

	// for now, just set the current view on the L1 chain to the
	// claimed collator's view
	currentL1BlockNumber := tx.L1BlockNumber()
	eonKey, err := s.processor.eonKeys.Find(currentL1BlockNumber)
	if err != nil {
		err = errors.Wrap(err, "no eon key found for batch transaction")
		log.Ctx(ctx).Error().Err(err).Msg("error while retrieving eon key")
		return "", err
	}

	gasPool.AddGas(GasLimit)
	for _, shutterTx := range tx.Transactions() {
		err := s.processor.processEncryptedTx(ctx, &gasPool, tx.BatchIndex(), tx.L1BlockNumber(), shutterTx, tx.DecryptionKey(), eonKey)
		if err != nil {
			// those are conditions that the collator can check,
			// so an error here means the whole batch is invalid
			err := errors.Wrap(err, "transaction invalid")
			return "", err
		}
		log.Info().Msg("successfully applied shutter-tx")
	}

	sender, _ := s.processor.signer.Sender(&tx)
	log.Ctx(ctx).Info().Str("signer", sender.Hex()).RawJSON("transaction", txStr).Msg("received batch transaction")
	s.processor.batchIndex = tx.BatchIndex()

	log.Ctx(ctx).Info().Str("batch", hexutil.EncodeUint64(s.processor.batchIndex)).Msg("started new batch")

	return tx.Hash().Hex(), nil
}

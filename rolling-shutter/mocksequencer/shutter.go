package mocksequencer

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func (s *ShutterService) SubmitBatch(transaction string) (string, error) {
	txBytes, err := hexutil.Decode(transaction)
	if err != nil {
		return "", errors.Wrap(err, "can't decode incoming tx bytes")
	}

	var tx txtypes.Transaction
	err = tx.UnmarshalBinary(txBytes)
	if err != nil {
		return "", errors.Wrap(err, "can't unmarshal incoming bytes to transaction")
	}
	// TODO process the BatchTx
	// - validate the batchtx
	// - pretty log the batch tx
	// - decrypt and pretty log all transactions

	log.Info().Msg("received batch-tx")
	if s.processor.batchIndex != tx.BatchIndex()-1 {
		return "", errors.New("incorrect batch-index for next batch")
	}
	for _, shutterTx := range tx.Transactions() {
		err := s.processor.processEncryptedTx(shutterTx)
		if err != nil {
			log.Error().Err(err).Msg("tx not processable, dropping transaction")
		} else {
			log.Info().Msg("successfully applied shutter-tx")
		}
	}
	s.processor.batchIndex = tx.BatchIndex()
	log.Info().Str("batch", hexutil.EncodeUint64(s.processor.batchIndex)).Msg("started new batch")
	return tx.Hash().Hex(), nil
}

package rpc

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"

	rpcerrors "github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/errors"
)

type ShutterService struct {
	processor             Sequencer
	batchProductionActive bool
}

// var _ mocksequencer.RPCService = (*ShutterService)(nil)

func (s *ShutterService) InjectProcessor(p Sequencer) {
	s.processor = p
	s.batchProductionActive = s.processor.Active()
}

func (s *ShutterService) Name() string {
	return "shutter"
}

func (s *ShutterService) BatchIndex(ctx context.Context) (hexutil.Uint64, error) {
	// Caching is fine here, since currently it won't become inactice again.
	if !s.batchProductionActive {
		s.batchProductionActive = s.processor.Active()
		// TODO how can we define a new rpc error number?
		return 0, errors.New("batch production has not started yet")
	}
	idx, err := s.processor.BatchIndex(ctx)
	if err != nil {
		return 0, err
	}
	return hexutil.Uint64(idx), nil
}

func (s *ShutterService) GetTransactionByHash(hash common.Hash) (*txtypes.TransactionData, error) {
	// s.processor.RLock()
	// defer s.processor.RUnlock()
	//
	// txID, ok := s.processor.GetTransaction(hash)
	// if !ok {
	// 	// ETH JSON RPC returns "null" when not found
	// 	return nil, nil
	// }
	// blockHash := ethrpc.BlockNumberOrHash{BlockHash: &txID.BlockHash}
	//
	// // rpcTx := tx.TransactionData()
	//
	// return rpcTx, nil
	// TODO
	return nil, nil
}

// FIXME this seemed to have crashed.
// How to find out when?
func (s *ShutterService) SubmitBatch(
	ctx context.Context,
	batchTransaction string,
) (string, error) {
	tx := &txtypes.Transaction{}

	txBytes, err := hexutil.Decode(batchTransaction)
	if err != nil {
		err := errors.Wrap(err, "can't decode incoming tx bytes")
		return "", rpcerrors.ParseError(err)
	}

	err = tx.UnmarshalBinary(txBytes)
	if err != nil {
		err := errors.Wrap(err, "can't unmarshal incoming bytes to transaction")
		return "", rpcerrors.ParseError(err)
	}
	return s.processor.SubmitBatch(ctx, tx)
}

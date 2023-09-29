package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
	rpcerrors "github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/errors"
)

type ShutterService struct {
	processor *mocksequencer.Sequencer
}

var _ mocksequencer.RPCService = (*ShutterService)(nil)

func (s *ShutterService) InjectProcessor(p *mocksequencer.Sequencer) {
	s.processor = p
}

func (s *ShutterService) Name() string {
	return "shutter"
}

func (s *ShutterService) BatchIndex() hexutil.Uint64 {
	return hexutil.Uint64(s.processor.BatchIndex)
}

func (s *ShutterService) GetTransactionByHash(hash common.Hash) (*txtypes.TransactionData, error) {
	s.processor.Mux.RLock()
	defer s.processor.Mux.RUnlock()

	txID, ok := s.processor.Txs[hash]
	if !ok {
		// ETH JSON RPC returns "null" when not found
		return nil, nil
	}
	blockHash := ethrpc.BlockNumberOrHash{BlockHash: &txID.BlockHash}
	b, err := s.processor.GetBlock(blockHash)
	if err != nil {
		// this shouldn't happen
		return nil, rpcerrors.ExtractRPCError(err)
	}
	tx := b.Transactions[txID.Index]

	rpcTx := tx.TransactionData()

	if b.Hash != (common.Hash{}) {
		rpcTx.BlockHash = &b.Hash
		rpcTx.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(b.Number))
		index := uint64(txID.Index)
		rpcTx.TransactionIndex = (*hexutil.Uint64)(&index)

		//nolint:godox //this is not worth an issue at the moment
		// TODO(ezdac) passing this as well would be nice,
		// but it is not required due to the already passed decrypted tx.Payload
		//
		// This would require us to save the decryption key
		// result.DecryptionKey = (*hexutil.Bytes)(&decryptionKey)
	}
	return rpcTx, nil
}

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

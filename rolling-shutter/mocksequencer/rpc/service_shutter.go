package rpc

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
)

type ShutterService struct {
	processor *mocksequencer.Processor
}

var _ mocksequencer.RPCService = (*ShutterService)(nil)

func (s *ShutterService) InjectProcessor(p *mocksequencer.Processor) {
	s.processor = p
}

func (s *ShutterService) Name() string {
	return "shutter"
}

func (s *ShutterService) GetBatchIndex() (string, error) {
	return hexutil.EncodeUint64(s.processor.BatchIndex), nil
}

func (s *ShutterService) SubmitBatch(ctx context.Context, batchTransaction string) (string, error) {
	var tx txtypes.Transaction

	txBytes, err := hexutil.Decode(batchTransaction)
	if err != nil {
		return "", errors.Wrap(err, "can't decode incoming tx bytes")
	}

	err = tx.UnmarshalBinary(txBytes)
	if err != nil {
		return "", errors.Wrap(err, "can't unmarshal incoming bytes to transaction")
	}
	return s.processor.SubmitBatch(ctx, &tx)
}

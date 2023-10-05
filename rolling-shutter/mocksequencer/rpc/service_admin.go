package rpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/encoding"
	rpcerrors "github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/errors"
)

type AdminService struct {
	processor *mocksequencer.Sequencer
}

var _ mocksequencer.RPCService = (*AdminService)(nil)

func (s *AdminService) InjectProcessor(p *mocksequencer.Sequencer) {
	s.processor = p
}

func (s *AdminService) Name() string {
	return "admin"
}

func (s *AdminService) AddCollator(address string, l1BlockNumber uint64) (int, error) {
	var err error
	defer func() {
		log.Info().
			Err(err).
			Str("address", address).
			Uint64("l1-blocknumber", l1BlockNumber).
			Msg("admin method AddCollator called")
	}()

	collator, err := encoding.StringToAddress(address)
	if err != nil {
		return 0, err
	}
	s.processor.Collators.Set(collator, l1BlockNumber)
	return 1, nil
}

func (s *AdminService) SetNonce(address common.Address, nonce *hexutil.Uint64) (int, error) {
	s.processor.Mux.Lock()
	defer s.processor.Mux.Unlock()

	blockHash := ethrpc.BlockNumberOrHash{BlockHash: &s.processor.LatestBlock}
	b, err := s.processor.GetBlock(blockHash)
	if err != nil {
		// this shouldn't happen
		return 0, rpcerrors.ExtractRPCError(err)
	}

	b.SetNonce(address, uint64(*nonce))
	return 1, nil
}

func (s *AdminService) SetBalance(address common.Address, balance *hexutil.Big) (int, error) {
	s.processor.Mux.Lock()
	defer s.processor.Mux.Unlock()

	blockHash := ethrpc.BlockNumberOrHash{BlockHash: &s.processor.LatestBlock}
	b, err := s.processor.GetBlock(blockHash)
	if err != nil {
		// this shouldn't happen
		return 0, rpcerrors.ExtractRPCError(err)
	}
	b.SetBalance(address, balance.ToInt())
	return 1, nil
}

func (s *AdminService) AddEonKey(eonKey string, l1BlockNumber uint64) (int, error) {
	eonKeyBytes, err := hexutil.Decode(eonKey)
	if err != nil {
		err = errors.Wrap(err, "eon key could not be decoded")
		return 0, err
	}
	s.processor.EonKeys.Set(eonKeyBytes, l1BlockNumber)
	return 1, nil
}

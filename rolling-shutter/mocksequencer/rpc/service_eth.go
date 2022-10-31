package rpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
)

type EthService struct {
	processor *mocksequencer.Processor
}

var _ mocksequencer.RPCService = (*EthService)(nil)

func (s *EthService) InjectProcessor(p *mocksequencer.Processor) {
	s.processor = p
}

func (s *EthService) Name() string {
	return "eth"
}

func (s *EthService) GetTransactionCount(address common.Address, blockNrOrHash ethrpc.BlockNumberOrHash) (*hexutil.Uint64, error) {
	s.processor.Mux.RLock()
	defer s.processor.Mux.RUnlock()
	block, err := s.processor.GetBlock(blockNrOrHash)
	if err != nil {
		return nil, err
	}
	nonce := hexutil.Uint64(block.GetNonce(address))
	return &nonce, nil
}

func (s *EthService) GetBalance(address common.Address, blockNrOrHash ethrpc.BlockNumberOrHash) (*hexutil.Big, error) {
	s.processor.Mux.RLock()
	defer s.processor.Mux.RUnlock()
	block, err := s.processor.GetBlock(blockNrOrHash)
	if err != nil {
		return nil, err
	}
	balance := (*hexutil.Big)(block.GetBalance(address))
	return balance, nil
}

func (s *EthService) ChainID() *hexutil.Big {
	return (*hexutil.Big)(s.processor.ChainID())
}

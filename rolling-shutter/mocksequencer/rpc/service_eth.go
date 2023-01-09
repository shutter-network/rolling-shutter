package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
	rpcerrors "github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/errors"
)

// rpcMarshalHeader converts the given header to the RPC output .
func rpcMarshalHeader(head *ethtypes.Header) map[string]interface{} {
	result := map[string]interface{}{
		"number":           (*hexutil.Big)(head.Number),
		"hash":             head.Hash(),
		"parentHash":       head.ParentHash,
		"nonce":            head.Nonce,
		"mixHash":          head.MixDigest,
		"sha3Uncles":       head.UncleHash,
		"logsBloom":        head.Bloom,
		"stateRoot":        head.Root,
		"miner":            head.Coinbase,
		"difficulty":       (*hexutil.Big)(head.Difficulty),
		"extraData":        hexutil.Bytes(head.Extra),
		"size":             hexutil.Uint64(head.Size()),
		"gasLimit":         hexutil.Uint64(head.GasLimit),
		"gasUsed":          hexutil.Uint64(head.GasUsed),
		"timestamp":        hexutil.Uint64(head.Time),
		"transactionsRoot": head.TxHash,
		"receiptsRoot":     head.ReceiptHash,
	}

	if head.BaseFee != nil {
		result["baseFeePerGas"] = (*hexutil.Big)(head.BaseFee)
	}

	return result
}

type EthService struct {
	processor *mocksequencer.Sequencer
}

var _ mocksequencer.RPCService = (*EthService)(nil)

func (s *EthService) InjectProcessor(p *mocksequencer.Sequencer) {
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
		err := errors.New("header for hash not found")
		return nil, rpcerrors.Default(err)
	}
	nonce := hexutil.Uint64(block.GetNonce(address))
	return &nonce, nil
}

func (s *EthService) GetBalance(address common.Address, blockNrOrHash ethrpc.BlockNumberOrHash) (*hexutil.Big, error) {
	s.processor.Mux.RLock()
	defer s.processor.Mux.RUnlock()
	block, err := s.processor.GetBlock(blockNrOrHash)
	if err != nil {
		err := errors.New("header for hash not found")
		return nil, rpcerrors.Default(err)
	}
	balance := (*hexutil.Big)(block.GetBalance(address))
	return balance, nil
}

//nolint:var-naming,revive,stylecheck
func (s *EthService) ChainId() *hexutil.Big {
	return (*hexutil.Big)(s.processor.ChainID())
}

func (s *EthService) GetBlockByNumber(_ context.Context, blockNumber ethrpc.BlockNumber, _ bool) (map[string]interface{}, error) {
	var result map[string]interface{}
	s.processor.Mux.RLock()
	defer s.processor.Mux.RUnlock()

	block, err := s.processor.GetBlock(ethrpc.BlockNumberOrHashWithNumber(blockNumber))
	if err != nil {
		err := errors.New("header for blockNumber not found")
		return nil, rpcerrors.Default(err)
	}

	header := &ethtypes.Header{
		ParentHash: [32]byte{},
		UncleHash:  ethtypes.EmptyUncleHash,
		// Coinbase:    [20]byte{}, // optional
		Root:        ethtypes.EmptyRootHash,
		TxHash:      ethtypes.EmptyRootHash,
		ReceiptHash: [32]byte{},
		Bloom:       [256]byte{},
		Difficulty:  &big.Int{},
		Number:      &big.Int{},
		GasLimit:    block.GasLimit,
		GasUsed:     0,
		Time:        0,
		Extra:       []byte{},
		// MixDigest:   [32]byte{}, // optional
		// Nonce:       [8]byte{},  // optional
		BaseFee: new(big.Int).Set(block.BaseFee),
	}

	result = rpcMarshalHeader(header)
	return result, nil
}

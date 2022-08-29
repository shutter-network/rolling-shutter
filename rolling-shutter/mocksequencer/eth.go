package mocksequencer

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type EthService struct {
	processor *SequencerProcessor
}

var _ RPCService = (*EthService)(nil)

func (s *EthService) injectProcessor(p *SequencerProcessor) {
	s.processor = p
}

func (s *EthService) name() string {
	return "eth"
}

func (s *EthService) GetTransactionCount(address, block string) (string, error) {
	addr, err := stringToAddress(address)
	if err != nil {
		return "", err
	}
	nonce := s.processor.getNonce(addr, block)
	return hexutil.EncodeUint64(nonce), nil
}

func (*EthService) GetBalance(_, _ string) (string, error) {
	// always return constant max amount
	return hexutil.EncodeBig(abi.MaxUint256), nil
}

func (s *EthService) ChainID() (string, error) {
	return hexutil.EncodeBig(s.processor.chainID), nil
}

func (s *EthService) GetBlockByNumber(blockNumber string, _ bool) (json.RawMessage, error) {
	b, exists := s.processor.blocks[blockNumber]
	if !exists {
		return json.RawMessage("\"null\""), nil
	}
	return jsonBlock(b.baseFee, b.gasLimit), nil
}

func jsonBlock(baseFee *big.Int, gasLimit uint64) json.RawMessage {
	var bloom ethtypes.Bloom
	bloomHex, _ := bloom.MarshalText()
	blockString := fmt.Sprintf(`{"baseFeePerGas": "%s",
"difficulty": "0x1",
"extraData": "0x00",
"gasLimit": "%s",
"gasUsed": "0x7defcf",
"hash": "0xc7608dbb166f66c00ca8a7b0674c982b1cc12d390d7b3a3572e9185b583621f7",
"logsBloom": "%s",
"miner": "0x0000000000000000000000000000000000000000",
"mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
"nonce": "0x0000000000000000",
"number": "0x68d6c6",
"parentHash": "0x9c07b52b71bda063c864b57cc28e49397d4eedadf7c91ed83ab776db78cfec8b",
"receiptsRoot": "0x7617c2f379f393dbc5dae56f9095aab437ed3bed63947d708b6a4e54c551964c",
"sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
"size": "0x1509e",
"stateRoot": "0x239f82f65838272e6dfe4ebbd755f6d4f2d12a09aa4df65f8346ab9afd0b2e43",
"timestamp": "0x627ccb76",
"totalDifficulty": "0x99c2ec",
"transactions": [],
"transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
"uncles": []}`, hexutil.EncodeBig(baseFee), hexutil.EncodeUint64(gasLimit), bloomHex)
	return json.RawMessage(strings.ReplaceAll(blockString, "\n", ""))
}

package sequencer

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"
)

type jsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      uint64          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  []interface{}   `json:"params,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type blockData struct {
	baseFee  *big.Int
	gasLimit uint64
}

func decodeRequest(req *http.Request, mess *jsonrpcMessage) error {
	err := json.NewDecoder(req.Body).Decode(mess)
	if err != nil {
		return err
	}
	return nil
}

// RunMockEthServer initializes the MockEthServer instance,
// spawns a new httptest server and sets the running servers
// URL as the MockEthServer.URL field.
// This is the mainly used initializer for running the mock server
// in tests.
// The MockEthServer.Teardown() should be called deferred immediately
// after using this function.
func RunMockEthServer(t *testing.T) *MockEthServer {
	t.Helper()
	mock := &MockEthServer{
		balances:    make(map[string]map[string]*big.Int),
		nonces:      make(map[string]map[string]uint64),
		blocks:      make(map[string]blockData),
		receivedTxs: make(map[string]bool),
		txs:         make(map[string]*txtypes.Transaction),
		hooks:       make([]txHookFunc, 0),
		t:           t,
	}
	mock.HTTPServer = httptest.NewServer(http.HandlerFunc(mock.handle))
	mock.URL = mock.HTTPServer.URL
	return mock
}

type txHookFunc func(me *MockEthServer, tx *txtypes.Transaction) bool

// Single client Mock eth server that implements some Ethereum
// JSON RPC methods (eth_getBalance, eth_getTransactionCount,
//  eth_chainId, eth_sendRawTransaction, eth_getBlockByNumber,
//  eth_getTransactionReceipt).
// Not all methods will strictly follow a Ethereum nodes logic
// in terms of error response and response data integrety.
// The functionality of the MockEthServer is based on the
// needs in the tests and might be extented in the future.
type MockEthServer struct {
	Mux         sync.RWMutex
	URL         string
	t           *testing.T
	balances    map[string]map[string]*big.Int
	nonces      map[string]map[string]uint64
	chainID     *big.Int
	blocks      map[string]blockData
	blockNumber uint64
	receivedTxs map[string]bool
	HTTPServer  *httptest.Server
	txs         map[string]*txtypes.Transaction
	hooks       []txHookFunc
}

// Teardown has to be called if the MockEthServer is run
// by instantiating it with the RunMockEthServer function.
func (me *MockEthServer) Teardown() {
	if me.HTTPServer != nil {
		me.HTTPServer.Close()
	}
}

func (me *MockEthServer) handle(w http.ResponseWriter, r *http.Request) {
	var (
		status int
		mess   jsonrpcMessage
	)

	if r.URL.Path != "/" {
		me.t.Errorf("Expected to request path \"/\", got: %s", r.URL.Path)
	}
	if r.Header.Get("Accept") != "application/json" {
		me.t.Errorf("Expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
	}
	if r.Method != "POST" && r.Method != "GET" {
		me.t.Errorf("Expected Method: POST / GET, got: %s", r.Method)
	}

	err := decodeRequest(r, &mess)
	if err != nil {
		err = errors.Wrapf(err, "error while decoding request body")
		me.t.Error(err)
	}
	log.Debug().Msg(fmt.Sprintf("JSONRPC method call: %s, params %v", mess.Method, mess.Params))

	switch mess.Method {
	case "eth_getBalance":
		me.Mux.RLock()
		status, err = me.getBalance(&mess)
		me.Mux.RUnlock()
	case "eth_getTransactionCount":
		me.Mux.RLock()
		status, err = me.getNonce(&mess)
		me.Mux.RUnlock()
	case "eth_chainId":
		me.Mux.RLock()
		status, err = me.getChainID(&mess)
		me.Mux.RUnlock()
	case "eth_sendRawTransaction":
		// set the write lock here,
		// since we might modify the internal
		// state in the hook-function
		// that is called in `respondRawTx`
		me.Mux.Lock()
		status, err = me.respondRawTx(&mess)
		me.Mux.Unlock()
	case "eth_getBlockByNumber":
		me.Mux.RLock()
		status, err = me.getBlockByNumber(&mess)
		me.Mux.RUnlock()
	case "eth_blockNumber":
		me.Mux.RLock()
		status, err = me.getBlockNumber(&mess)
		me.Mux.RUnlock()
	case "eth_getTransactionReceipt":
		me.Mux.RLock()
		status, err = me.getReceipt(&mess)
		me.Mux.RUnlock()
	default:
		me.t.Errorf("Eth JSON RPC method not known or not supported by MockServer: %s", mess.Method)
		return
	}
	if err != nil {
		err = errors.Wrapf(err, "error while handling method %s", mess.Method)
		me.t.Fatal(err)
	}

	// encode result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	err = encoder.Encode(mess)
	if err != nil {
		err = errors.Wrapf(err, "error while encoding response in mock server")
		me.t.Error(err)
	}
}

func (me *MockEthServer) SetBlockNumber(blockNumber uint64) {
	me.blockNumber = blockNumber
}

// SetBalance is used to set the state of the server for the "eth_getBalance" method.
// SetBalance sets the balance `b` for address `a` at the block `block`.
// `block` has a value like in the JSON RPC spec, so e.g. the current block is
// referred to with "latest".
func (me *MockEthServer) SetBalance(a common.Address, b *big.Int, block string) {
	bl, exists := me.balances[block]
	if !exists {
		bl = make(map[string]*big.Int, 0)
		me.balances[block] = bl
	}
	bl[a.Hex()] = b
}

func (me *MockEthServer) getBalance(mess *jsonrpcMessage) (int, error) {
	if len(mess.Params) > 2 {
		return -1, errors.Errorf("got more parameters then expected: %v", mess.Params)
	}
	if len(mess.Params) != 2 {
		return -1, errors.Errorf("expected exactly 2 parameters, got: %v", mess.Params)
	}
	address := normalizeAddress(mess.Params[0].(string))
	block := mess.Params[1].(string)
	balance := me.balances[block][address]
	res := fmt.Sprintf("%q", hexutil.EncodeBig(balance))
	mess.Result = json.RawMessage(res)
	return 200, nil
}

// SetNonce is used to set the state of the server for the "eth_getTransactionCount" method.
// SetNonce sets the nonce `nonce` for address `a` at the block `block`.
// `block` has a value like in the JSON RPC spec, so e.g. the current block is
// referred to with "latest".
func (me *MockEthServer) SetNonce(a common.Address, nonce uint64, block string) {
	nc, exists := me.nonces[block]
	if !exists {
		nc = make(map[string]uint64, 0)
		me.nonces[block] = nc
	}
	nc[a.Hex()] = nonce
}

func (me *MockEthServer) getNonce(mess *jsonrpcMessage) (int, error) {
	if len(mess.Params) > 2 {
		return -1, errors.Errorf("got more parameters then expected: %v", mess.Params)
	}
	address := normalizeAddress(mess.Params[0].(string))
	block := mess.Params[1].(string)
	nonce := me.nonces[block][address]
	log.Debug().Msg(fmt.Sprintf("nonces:%v, addr:%s, nonce:%d, block:%s", me.nonces, address, nonce, block))

	res := fmt.Sprintf("%q", hexutil.EncodeUint64(nonce))
	mess.Result = json.RawMessage(res)
	return 200, nil
}

// SetChainID is used to set the state of the server for the "eth_chainId" method.
func (me *MockEthServer) SetChainID(c *big.Int) {
	me.chainID = c
}

func (me *MockEthServer) getChainID(mess *jsonrpcMessage) (int, error) {
	if len(mess.Params) > 0 {
		return -1, errors.Errorf("got more parameters then expected: %v", mess.Params)
	}
	res := fmt.Sprintf("%q", hexutil.EncodeBig(me.chainID))
	mess.Result = json.RawMessage(res)
	return 200, nil
}

// SetBlock is used to set the state of the server for the "eth_getBlockByNumber" method.
// SetNonce sets the base-fee `baseFee` and `gasLimit` in the block struct returned for
// block `block`.
// `block` has a value like in the JSON RPC spec, so e.g. the current block is
// referred to with "latest".
// The returned block struct follows the JSON RPC spec, but it relies on returning
// a static dummy raw JSON string, where only base-fee and gas-limit values
// are replaced by the values set in this method.
// Thus, the returned values like parent-hash etc. do not necessarily comply
// with any state-transition and geth-validation logic.
func (me *MockEthServer) SetBlock(baseFee *big.Int, gasLimit uint64, block string) {
	b, exists := me.blocks[block]
	if !exists {
		b = blockData{baseFee: baseFee, gasLimit: gasLimit}
		me.blocks[block] = b
		return
	}
	b.baseFee = baseFee
	b.gasLimit = gasLimit
}

func (me *MockEthServer) getBlockNumber(mess *jsonrpcMessage) (int, error) {
	if len(mess.Params) > 0 {
		return -1, errors.Errorf("got more parameters then expected: %v", mess.Params)
	}
	res := fmt.Sprintf("%q", hexutil.EncodeUint64(me.blockNumber))
	mess.Result = json.RawMessage(res)
	return 200, nil
}

func (me *MockEthServer) getBlockByNumber(mess *jsonrpcMessage) (int, error) {
	if len(mess.Params) > 2 {
		return -1, errors.Errorf("got more parameters then expected: %v", mess.Params)
	}
	// only use the first param,
	// ignore the bool param switching full transaction list / transaction hash list
	blockNum := mess.Params[0].(string)
	b, exists := me.blocks[blockNum]
	if !exists {
		mess.Result = json.RawMessage("\"null\"")
		return 200, nil
	}
	mess.Result = jsonBlock(b.baseFee, b.gasLimit)
	return 200, nil
}

func (me *MockEthServer) getReceipt(mess *jsonrpcMessage) (int, error) {
	if len(mess.Params) > 1 {
		return -1, errors.Errorf("got more parameters then expected: %v", mess.Params)
	}
	txHash := mess.Params[0].(string)
	_, exists := me.txs[txHash]
	if !exists {
		mess.Result = json.RawMessage("null")
		return 200, nil
	}
	mess.Result = jsonReceipt(txHash, big.NewInt(42), true)
	return 200, nil
}

func (me *MockEthServer) respondRawTx(mess *jsonrpcMessage) (int, error) {
	if len(mess.Params) > 1 {
		return -1, errors.Errorf("got more parameters then expected: %v", mess.Params)
	}
	rawTx := mess.Params[0].(string)
	txBytes, err := hexutil.Decode(rawTx)
	if err != nil {
		return 0, errors.Wrap(err, "can't decode incoming tx bytes")
	}

	var tx txtypes.Transaction
	err = tx.UnmarshalBinary(txBytes)
	if err != nil {
		return 0, errors.Wrap(err, "can't unmarshael incoming bytes to transaction")
	}
	me.txs[tx.Hash().Hex()] = &tx

	newHooks := make([]txHookFunc, 0)
	for _, f := range me.hooks {
		// true return value means
		// the function should be removed
		// from the actions
		if !f(me, &tx) {
			newHooks = append(newHooks, f)
		}
	}
	me.hooks = newHooks
	res := fmt.Sprintf("%q", tx.Hash().Hex())
	mess.Result = json.RawMessage(res)
	return 200, nil
}

// ReceivedTransactions returns all unmarshalled transaction structs
// that were received via the "eth_sendRawTransaction" method by the client.
func (me *MockEthServer) ReceivedTransactions() []*txtypes.Transaction {
	txs := make([]*txtypes.Transaction, 0)
	for _, tx := range me.txs {
		txs = append(txs, tx)
	}
	return txs
}

// RegisterTxHook registers a hook function to be executed
// on any incoming transaction that is received from clients
// via the "eth_sendRawTransaction" method.
//
// The passed in function `hook` has to comply to the following interface:
// `type txHookFunc func(me *MockEthServer, tx *txtypes.Transaction) bool`
// The argument `me` represents the current MockEthServer instance and tx
// represents the received unmarshaled client transaction.
// The function can process the transaction in any way u
// (e.g. assert certain values).
// If the function returns true while being executed on an incoming
// transaction, it will get deregistered from the hooks and
// not executed for any further incoming transactions.
func (me *MockEthServer) RegisterTxHook(hook txHookFunc) {
	me.hooks = append(me.hooks, hook)
}

// GetReceivedTx returns the unmarshalled transaction struct
// that was received via the "eth_sendRawTransaction" method by the client
// and matches the transaction hash `txHash`.
func (me *MockEthServer) GetReceivedTx(txHash string) *txtypes.Transaction {
	tx, exists := me.txs[txHash]
	if !exists {
		return nil
	}
	return tx
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

func jsonReceipt(txHash string, blockNumber *big.Int, success bool) json.RawMessage {
	var status string
	if success {
		status = "0x1"
	} else {
		status = "0x0"
	}
	var bloom ethtypes.Bloom
	bloomHex, _ := bloom.MarshalText()
	receiptString := fmt.Sprintf(`{
    "blockHash": "0xc4485dae215696aeb152e6a4f469a00dac2640bd0cbb540275e1e4b416475c52",
    "blockNumber": "%s",
    "contractAddress": null,
    "cumulativeGasUsed": "0xfff8",
    "effectiveGasPrice": "0xee6b2800",
    "from": "0x7b907992d6c5820ff7c80fb9e481780f2bbf30fd",
    "gasUsed": "0xfff8",
    "logs": [],
    "logsBloom": "%s",
    "status": "%s",
    "to": "0xff50ed3d0ec03ac01d4c79aad74928bff48a7b2b",
    "transactionHash": "%s",
    "transactionIndex": "0x0",
    "type": "0x0"
  }`, hexutil.EncodeBig(blockNumber), bloomHex, status, txHash)
	return json.RawMessage(strings.ReplaceAll(receiptString, "\n", ""))
}

func normalizeAddress(addr string) string {
	return common.HexToAddress(addr).String()
}

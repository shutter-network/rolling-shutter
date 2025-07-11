package shutterservice

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/help"
)

// # Trigger definition
// ## Comparison operators
// "eq", "lt", "lte", "gt", "gte" for number types, "match" for string/bytes
//
// ## ABI-Event:
//
//	{
//	 "contract": "0xdead..beef",
//		"signature": "Transfer(from address indexed, to address indexed, amount uint256)",
//		"conditions": [
//			{"to": {"match": "0xdead...beef"}},
//			{"amount": {"gte": 1}}
//		],
//	}
//
// Note: fields that are not referenced in "conditions" are not restricted.
//
// ## RAW-Event:
// A user may not have the event-ABI available, or may not want to share it.
//
//	{
//	 "contract": "0xdead..beef",
//	 "rawsig": "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
//	 "rawconditions": [
//	   	{"topic1": "any"},
//	 	{"topic2": {"match": "0xdead..beef"}},
//	 	{"data": {
//	 		"start": 0,
//			"end": 32, // probably unnecessary, can be derived from "cast" type
//			"cast": "uint256",
//			"gte": 1
//			}
//		},
//	 ],
//	}
//
// Note: in order to allow for successful matching/parsing, _all_ "topics" must be referenced -- "any" allows for no restrictions.
//
// ## Condensed encoding (WIP)
//
// We need this condensed encoding for registering trigger conditions on the blockchain (most likely as an event......)
// [-1:0] version byte  // Note: all byte numbers need to shift 1 to the right, to have the version in there...
// [0:32] address
// [33:64] topic0/raw signature
// [65] OPCODE-MATCH (see event_triggers.py)
// [66:matching_topics_number*32] matching hashes for topics
// [*:end] DATA matches
// Encoding for DATA matches:
// [*:2] offset
// [3] cast-matchtype-size {0: bytes32-match, 1: uint256-lt, 2: uint256-lte, 3: uint256-eq, 4: uint256-gte, 5:uint256-gt}
// [4-36] matchdata
// [$repeat for all data field conditions]

func TestLogFilter(t *testing.T) {
	// The first part of a log record consists of an array of topics. These topics are used to describe
	// the event. The first topic usually consists of the signature (a keccak256 hash) of the name of
	// the event that occurred, including the types (uint256, string, etc.) of its parameters. One exception
	// where this signature is not included as the first topic is when emitting anonymous events. Since
	// topics can only hold a maximum of 32 bytes of data, things like arrays or strings cannot be used as
	// topics reliably. Instead, it should be included as data in the log record, not as a topic. If you
	// were to try including a topic that’s larger than 32 bytes, the topic will be hashed instead. As a
	// result, this hash can only be reversed if you know the original input. In conclusion, topics should
	// only reliably be used for data that strongly narrows down search queries (like addresses).
	//
	// the event signature is produced by hashing the event's name and parameter types with Keccak-256.
	sig := crypto.Keccak256([]byte("Transfer(address,address,uint256,string)"))
	//  ==> "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	assert.Check(t, sig != nil, "sig is nil")

	// var Addresses []common.Address // restricts matches to events created by specific contracts

	// The Topic list restricts matches to particular event topics. Each event has a list
	// of topics. Topics matches a prefix of that list. An empty element slice matches any
	// topic. Non-empty elements represent an alternative that matches any of the
	// contained topics.
	//
	// Examples:
	// {} or nil          matches any topic list
	// {{A}}              matches topic A in first position
	// {{}, {B}}          matches any topic in first position AND B in second position
	// {{A}, {B}}         matches topic A in first position AND B in second position
	// {{A, B}, {C, D}}   matches topic (A OR B) in first position AND (C OR D) in second position
	// var Topics [][]common.Hash
	////////////

	// assert.Check(t, Addresses.len == 0, "false")
	// assert.Check(t, Topics.len == 0, "false")
	// abigen.Bind()
	// backend.simulated
}

type TestSetup struct {
	backend  *simulated.Backend
	address  common.Address
	contract *help.ERC20Basic
	key      *ecdsa.PrivateKey
}

func SetupAndDeploy() (TestSetup, error) {
	setup := TestSetup{}

	// Setup simulated block chain
	key, _ := crypto.GenerateKey()
	setup.key = key
	auth, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1337))
	if err != nil {
		return setup, err
	}
	alloc := make(types.GenesisAlloc)
	alloc[auth.From] = types.Account{Balance: big.NewInt(10000000000000000)}
	blockchain := simulated.NewBackend(alloc)
	setup.backend = blockchain

	// Deploy contract
	address, _, contract, err := help.DeployERC20Basic(
		auth,
		blockchain.Client(),
	)
	if err != nil {
		return setup, fmt.Errorf("failed to deploy %w", err)
	}
	setup.address = address
	setup.contract = contract
	// commit all pending transactions
	blockchain.Commit()
	return setup, nil
}

// Test ERC20 contract gets deployed correctly
func TestDeployContract(t *testing.T) {
	setup, err := SetupAndDeploy()
	assert.NilError(t, err, "setup and deploy failed")

	if len(setup.address.Bytes()) == 0 {
		t.Error("Expected a valid deployment address. Received empty address byte array instead")
	}
}

func ERC20Transfer(t *testing.T, setup TestSetup, from common.Address, to common.Address, amount int64) {
	client := setup.backend.Client()

	head, _ := client.HeaderByNumber(context.Background(), nil)
	gasPrice := new(big.Int).Add(head.BaseFee, big.NewInt(params.GWei))
	chainid, _ := client.ChainID(context.Background())
	nonce, err := client.PendingNonceAt(context.Background(), from)
	assert.NilError(t, err, "failed to get nonce")
	signer := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		return types.SignTx(tx, types.LatestSignerForChainID(chainid), setup.key)
	}

	assert.NilError(t, err, "failed to sign")
	tx, err := setup.contract.Transfer(&bind.TransactOpts{
		From:      from,
		Nonce:     big.NewInt(int64(nonce)),
		GasTipCap: big.NewInt(params.GWei),
		GasFeeCap: gasPrice,
		GasLimit:  2100000,
		Signer:    signer,
	}, to, big.NewInt(amount))
	assert.NilError(t, err, "error with token transfer")
	txHash := tx.Hash()
	assert.NilError(t, err, "error getting block")
	setup.backend.Commit()
	found, pending, err := client.TransactionByHash(context.Background(), txHash)
	assert.NilError(t, err, "error getting tx")
	assert.Check(t, pending == false, "still pending")
	assert.Check(t, found.Hash() == txHash)
	receipt, err := client.TransactionReceipt(context.Background(), txHash)

	assert.NilError(t, err, "error getting receipt")
	assert.Check(t, receipt.Status == 1, "transfer failed")
}

func TestFilterLogsQuery(t *testing.T) {
	setup, err := SetupAndDeploy()
	assert.NilError(t, err, "failure deploying")
	client := setup.backend.Client()
	latest, err := client.BlockNumber(context.Background())
	assert.NilError(t, err, "error getting blocknumber")

	addr := crypto.PubkeyToAddress(setup.key.PublicKey)

	ERC20Transfer(t, setup, addr, addr, 1)

	// topic0 signature
	topic0 := crypto.Keccak256([]byte("Transfer(address,address,uint256,string)"))

	// topic1 address from indexed
	topic1 := topicPad(addr.Bytes())

	// topic2 address to indexed
	topic2 := topicPad(addr.Bytes())

	topics := [][]common.Hash{
		{common.Hash(topic0)},
		{common.Hash(topic1)},
		{common.Hash(topic2)},
	}

	query := ethereum.FilterQuery{
		BlockHash: nil,
		FromBlock: big.NewInt(int64(latest)),
		ToBlock:   nil,
		Addresses: nil,
		Topics:    topics,
	}

	logs, err := client.FilterLogs(context.Background(), query)
	assert.NilError(t, err, "error getting logs")
	assert.Check(t, len(logs) > 0, "found no logs")
	for _, log := range logs {
		// uint256 amount
		amount := big.NewInt(0).SetBytes(log.Data[:32])
		amount256, overflow := uint256.FromBig(amount)
		assert.Check(t, !overflow, "err parsing uint256")
		t.Log(amount, amount256)
		stringdata := log.Data[33:]

		t.Log(string(stringdata))

		val := map[string]any{}
		rlp.Decode(bytes.NewReader(log.Data), val)
		t.Log(val)

	}
}

func TestBloomFilterMatch(t *testing.T) {
	setup, err := SetupAndDeploy()
	assert.NilError(t, err, "failure deploying")
	addr := crypto.PubkeyToAddress(setup.key.PublicKey)

	ERC20Transfer(t, setup, addr, addr, 1)
	latest, err := setup.backend.Client().BlockByNumber(context.Background(), nil)
	assert.NilError(t, err, "error getting block")

	// topic0 signature
	topic0 := crypto.Keccak256([]byte("Transfer(address,address,uint256,string)"))
	assert.Check(t, latest.Bloom().Test(topic0), "could not find topic0")

	// topic1 address from indexed
	topic1 := topicPad(addr.Bytes())
	assert.Check(t, latest.Bloom().Test(topic1), "could not find topic1")

	// topic2 address to indexed
	topic2 := topicPad(addr.Bytes())
	assert.Check(t, latest.Bloom().Test(topic2), "could not find topic2")

	// we could probably calculate `bloom9.go:types.bloomValues` for all topics, and manually match the merged/combined topic
}

func topicPad(data []byte) []byte {
	out := make([]byte, 32)
	copy(out[32-len(data):], data)
	return out
}

func LogFilterTest(t *testing.T) {
}

// LogFilterer provides access to contract log events using a one-off query or continuous
// event subscription.
//
// Logs received through a streaming query subscription may have Removed set to true,
// indicating that the log was reverted due to a chain reorganisation.
// type LogFilterer interface {
// 	FilterLogs(ctx context.Context, q FilterQuery) ([]types.Log, error)
// 	SubscribeFilterLogs(ctx context.Context, q FilterQuery, ch chan<- types.Log) (Subscription, error)
// }

var _ bind.ContractBackend = (simulated.Client)(nil)

var (
	testKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testAddr    = crypto.PubkeyToAddress(testKey.PublicKey)
	testKey2, _ = crypto.HexToECDSA("7ee346e3f7efc685250053bfbafbfc880d58dc6145247053d4fb3cb0f66dfcb2")
	testAddr2   = crypto.PubkeyToAddress(testKey2.PublicKey)
)

func simTestBackend(testAddr common.Address) *simulated.Backend {
	return simulated.NewBackend(
		types.GenesisAlloc{
			testAddr: {Balance: big.NewInt(10000000000000000)},
		},
	)
}

func newTx(sim *simulated.Backend, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	client := sim.Client()

	// create a signed transaction to send
	head, _ := client.HeaderByNumber(context.Background(), nil) // Should be child's, good enough
	gasPrice := new(big.Int).Add(head.BaseFee, big.NewInt(params.GWei))
	addr := crypto.PubkeyToAddress(key.PublicKey)
	chainid, _ := client.ChainID(context.Background())
	nonce, err := client.PendingNonceAt(context.Background(), addr)
	if err != nil {
		return nil, err
	}
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainid,
		Nonce:     nonce,
		GasTipCap: big.NewInt(params.GWei),
		GasFeeCap: gasPrice,
		Gas:       21000,
		To:        &addr,
	})

	return types.SignTx(tx, types.LatestSignerForChainID(chainid), key)
}

// func DeployTokenContract(t *testing.T, sim *simulatedBackend) {
// 	abigen
// }

func TestNewBackend(t *testing.T) {
	sim := simulated.NewBackend(types.GenesisAlloc{})
	defer sim.Close()

	client := sim.Client()
	num, err := client.BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if num != 0 {
		t.Fatalf("expected 0 got %v", num)
	}
	// Create a block
	sim.Commit()
	num, err = client.BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if num != 1 {
		t.Fatalf("expected 1 got %v", num)
	}
}

func TestSendTransaction(t *testing.T) {
	sim := simTestBackend(testAddr)
	defer sim.Close()

	client := sim.Client()
	ctx := context.Background()

	signedTx, err := newTx(sim, testKey)
	if err != nil {
		t.Errorf("could not create transaction: %v", err)
	}
	// send tx to simulated backend
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		t.Errorf("could not add tx to pending block: %v", err)
	}
	sim.Commit()
	block, err := client.BlockByNumber(ctx, big.NewInt(1))
	if err != nil {
		t.Errorf("could not get block at height 1: %v", err)
	}

	if signedTx.Hash() != block.Transactions()[0].Hash() {
		t.Errorf("did not commit sent transaction. expected hash %v got hash %v", block.Transactions()[0].Hash(), signedTx.Hash())
	}
}

func createAndCloseSimBackend() {
	genesisData := types.GenesisAlloc{}
	simulatedBackend := simulated.NewBackend(genesisData)
	defer simulatedBackend.Close()
}

package shutterservice

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/params"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/help"
)

// This should match what is in help/erc20bindings.go
const (
	TestEvtSig     = "Transfer(address,address,address,uint256,string)"
	TestEvtSigFull = "Transfer(address from indexed, address to indexed, address notify indexed, uint256 amount, string foobar)"
)

type TestSetup struct {
	backend                *simulated.Backend
	erc20Address           common.Address
	erc20Contract          *help.ERC20Basic
	triggerRegistryAddress common.Address
	triggerContract        *help.ShutterRegistry
	key                    *ecdsa.PrivateKey
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

	// Deploy erc20Contract
	erc20Address, _, erc20Contract, err := help.DeployERC20Basic(
		auth,
		blockchain.Client(),
	)
	if err != nil {
		return setup, fmt.Errorf("failed to deploy ERC20 %w", err)
	}
	setup.erc20Address = erc20Address
	setup.erc20Contract = erc20Contract

	// Deploy ShutterEventTrigger contract
	triggerRegistryAddress, _, triggerContract, err := help.DeployShutterRegistry(
		auth,
		blockchain.Client(),
	)
	if err != nil {
		return setup, fmt.Errorf("failed to deploy TriggerRegistry %w", err)
	}
	setup.triggerRegistryAddress = triggerRegistryAddress
	setup.triggerContract = triggerContract
	// commit all pending transactions
	blockchain.Commit()
	return setup, nil
}

// Test ERC20 contract gets deployed correctly
func TestEvtDeployContract(t *testing.T) {
	setup, err := SetupAndDeploy()
	assert.NilError(t, err, "setup and deploy failed")

	if len(setup.erc20Address.Bytes()) == 0 {
		t.Error("Expected a valid deployment address. Received empty address byte array instead")
	}
}

func RegisterTrigger(t *testing.T, setup TestSetup, trigger EventTriggerDefinition) {
	client := setup.backend.Client()
	from := crypto.PubkeyToAddress(setup.key.PublicKey)

	head, _ := client.HeaderByNumber(context.Background(), nil)
	gasPrice := new(big.Int).Add(head.BaseFee, big.NewInt(params.GWei))
	chainid, _ := client.ChainID(context.Background())
	nonce, err := client.PendingNonceAt(context.Background(), from)
	assert.NilError(t, err, "failed to get nonce")
	signer := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		return types.SignTx(tx, types.LatestSignerForChainID(chainid), setup.key)
	}

	ttl := uint64(head.Number.Int64() + 10)
	eon := uint64(1)
	identityPrefix := crypto.Keccak256Hash([]byte("test"))

	assert.NilError(t, err, "failed to sign")
	marshaledTrigger := trigger.MarshalBytes()
	tx, err := setup.triggerContract.Register(&bind.TransactOpts{
		From:      from,
		Nonce:     big.NewInt(int64(nonce)),
		GasTipCap: big.NewInt(params.GWei),
		GasFeeCap: gasPrice,
		GasLimit:  2100000,
		Signer:    signer,
	},
		eon, identityPrefix, marshaledTrigger, ttl)
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
	tx, err := setup.erc20Contract.Transfer(&bind.TransactOpts{
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

func TestEvtBloomFilterMatch(t *testing.T) {
	setup, err := SetupAndDeploy()
	assert.NilError(t, err, "failure deploying")
	addr := crypto.PubkeyToAddress(setup.key.PublicKey)

	ERC20Transfer(t, setup, addr, addr, 1)
	latest, err := setup.backend.Client().BlockByNumber(context.Background(), nil)
	assert.NilError(t, err, "error getting block")

	// topic0 signature
	topic0 := crypto.Keccak256([]byte(TestEvtSig))
	assert.Check(t, latest.Bloom().Test(topic0), "could not find topic0")

	// topic1 address from indexed
	topic1 := WordPad(addr.Bytes())
	assert.Check(t, latest.Bloom().Test(topic1), "could not find topic1")

	// topic2 address to indexed
	topic2 := WordPad(addr.Bytes())
	assert.Check(t, latest.Bloom().Test(topic2), "could not find topic2")

	// we could probably calculate `bloom9.go:types.bloomValues` for all topics, and manually match the merged/combined topic
}

func CreateDefinition(contract common.Address, topic1 []byte, topic2 []byte, topic3 []byte, amount int64, target []byte) EventTriggerDefinition {
	definition := EventTriggerDefinition{
		Contract:       contract,
		EventSignature: common.HexToHash("0x1234567890123456789012345678901234567890123456789012345678901234"),
		Conditions: []Condition{
			{
				Location: TopicData{
					number: 1,
				},
				Constraint: MatchConstraint{
					target: topic1,
				},
			},
			{
				Location: TopicData{
					number: 3,
				},
				Constraint: MatchConstraint{
					target: topic3,
				},
			},
			{
				Location: OffsetData{
					argnumber: 0,
					complex:   false,
				},
				Constraint: NumConstraint{
					op:     GTE,
					target: big.NewInt(amount),
				},
			},
			{
				Location: OffsetData{
					argnumber: 1,
					complex:   true,
				},
				Constraint: MatchConstraint{
					target: target,
				},
			},
		},
	}
	return definition
}

func TestEvtEventTriggerDefinition(t *testing.T) {
	setup, err := SetupAndDeploy()
	assert.NilError(t, err, "error during setup")

	senderAddr := crypto.PubkeyToAddress(setup.key.PublicKey)
	zeroAddr := common.HexToAddress("0x00000000000000000000000000000000")
	fiveAddr := common.HexToAddress("0x00000000000000000000000000000005")
	amount := int64(1)
	// stringMatch := []byte("Lets see how long this string can get and what it will look like in the data, I feel like I need to keep going for a bit........")
	stringMatch := []byte("short")

	definition := CreateDefinition(setup.erc20Address, WordPad(senderAddr.Bytes()), nil, WordPad(fiveAddr.Bytes()), int64(1), stringMatch)
	assert.Check(t, len(definition.Conditions) == 4, "something went wrong")
	f := definition.ToFilterQuery()
	assert.Check(t, len(f.Topics) > 0, "no filterquery")

	ERC20Transfer(t, setup, senderAddr, setup.erc20Address, amount)

	checkTopics := true

	logs, err := setup.backend.Client().FilterLogs(context.Background(), f)
	assert.NilError(t, err, "error using filter query")
	assert.Check(t, len(logs) == 1, "filter did not match")
	for _, elog := range logs {
		assert.Check(t, definition.Match(elog, checkTopics) == true, "did not match %v", elog.Data)
	}
	// mismatch on topic2
	ERC20Transfer(t, setup, senderAddr, zeroAddr, 0)

	latest, err := setup.backend.Client().BlockNumber(context.Background())
	assert.NilError(t, err, "no latest block number")
	f.FromBlock = big.NewInt(int64(latest))
	f.ToBlock = big.NewInt(int64(latest))

	logs, err = setup.backend.Client().FilterLogs(context.Background(), f)
	assert.NilError(t, err, "error using filter query")
	for _, elog := range logs {
		assert.Check(t, definition.Match(elog, checkTopics) == false, "did match %v", elog.Data)
	}
	// mismatch on topic2 -- we should not find the event!)
	definition.Conditions = append(definition.Conditions, Condition{
		Location: TopicData{
			number: 2,
		},
		Constraint: MatchConstraint{
			target: WordPad(senderAddr.Bytes()),
		},
	})
	overspecific := definition.ToFilterQuery()
	ERC20Transfer(t, setup, senderAddr, zeroAddr, amount)

	latest, err = setup.backend.Client().BlockNumber(context.Background())
	assert.NilError(t, err, "no latest block number")
	f.FromBlock = big.NewInt(int64(latest))
	f.ToBlock = big.NewInt(int64(latest))

	logs, err = setup.backend.Client().FilterLogs(context.Background(), overspecific)
	assert.NilError(t, err, "error using filter query")
	assert.Check(t, len(logs) == 0, "event filter query incorrect")

	// mismatch on amount
	ERC20Transfer(t, setup, senderAddr, setup.erc20Address, 0)

	latest, err = setup.backend.Client().BlockNumber(context.Background())
	assert.NilError(t, err, "no latest block number")
	f.FromBlock = big.NewInt(int64(latest))
	f.ToBlock = big.NewInt(int64(latest))

	logs, err = setup.backend.Client().FilterLogs(context.Background(), f)
	assert.NilError(t, err, "error using filter query")
	for _, elog := range logs {
		assert.Check(t, definition.Match(elog, checkTopics) == false, "did match %v", elog.Data)
	}

	RegisterTrigger(t, setup, definition)
}

func TestEvtNumConstraintTest(t *testing.T) {
	lt1 := NumConstraint{
		LT,
		big.NewInt(1),
	}
	assert.Check(t, lt1.Test(big.NewInt(0)) == true, "wrong result")
	assert.Check(t, lt1.Test(big.NewInt(1)) == false, "wrong result")
	assert.Check(t, lt1.Test(big.NewInt(2)) == false, "wrong result")

	lte1 := NumConstraint{
		LTE,
		big.NewInt(1),
	}
	assert.Check(t, lte1.Test(big.NewInt(0)) == true, "wrong result")
	assert.Check(t, lte1.Test(big.NewInt(1)) == true, "wrong result")
	assert.Check(t, lte1.Test(big.NewInt(2)) == false, "wrong result")

	eq1 := NumConstraint{
		EQ,
		big.NewInt(1),
	}
	assert.Check(t, eq1.Test(big.NewInt(0)) == false, "wrong result")
	assert.Check(t, eq1.Test(big.NewInt(1)) == true, "wrong result")
	assert.Check(t, eq1.Test(big.NewInt(2)) == false, "wrong result")

	gte1 := NumConstraint{
		GTE,
		big.NewInt(1),
	}
	assert.Check(t, gte1.Test(big.NewInt(0)) == false, "wrong result")
	assert.Check(t, gte1.Test(big.NewInt(1)) == true, "wrong result")
	assert.Check(t, gte1.Test(big.NewInt(2)) == true, "wrong result")

	gt1 := NumConstraint{
		GT,
		big.NewInt(1),
	}
	assert.Check(t, gt1.Test(big.NewInt(0)) == false, "wrong result")
	assert.Check(t, gt1.Test(big.NewInt(1)) == false, "wrong result")
	assert.Check(t, gt1.Test(big.NewInt(2)) == true, "wrong result")
}

func TestEvtRegistry(t *testing.T) {
	setup, err := SetupAndDeploy()
	assert.NilError(t, err, "failed setup and deploy")

	assert.Check(t, setup.backend != nil, "setup is nil")
	def := CreateDefinition(
		setup.erc20Address,
		WordPad(crypto.PubkeyToAddress(setup.key.PublicKey).Bytes()),
		WordPad(crypto.PubkeyToAddress(setup.key.PublicKey).Bytes()),
		WordPad(crypto.PubkeyToAddress(setup.key.PublicKey).Bytes()),
		1, []byte("short"))
	RegisterTrigger(t, setup, def)

	ser := def.MarshalBytes()

	de := EventTriggerDefinition{}

	de.UnmarshalBytes(ser)
	round := de.MarshalBytes()
	assert.Check(t, deepEqual(ser, round), "roundtrip failed\n%v\n%v", ser, de.MarshalBytes())

	logs, err := setup.triggerContract.FilterEventTriggerRegistered(&bind.FilterOpts{
		Start:   uint64(0),
		End:     nil,
		Context: context.Background(),
	},
		[]uint64{1},
	)
	for logs.Next() {
		x := logs.Event.TriggerDefinition
		assert.Check(t, deepEqual(ser, x), "serialization mismatch")
	}
	var newConditions []Condition
	for _, cond := range def.Conditions {
		_, ok := cond.Constraint.(NumConstraint)
		if !ok {
			newConditions = append(newConditions, cond)
		}
	}
	def.Conditions = newConditions
	shorter := def.MarshalBytes()
	assert.Check(t, !deepEqual(shorter, ser), "we lost a constraint")
}

func deepEqual(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, ar := range a {
		if len(b[i]) != len(ar) {
			return false
		}
		for j, v := range ar {
			if b[i][j] != v {
				return false
			}
		}
	}
	return true
}

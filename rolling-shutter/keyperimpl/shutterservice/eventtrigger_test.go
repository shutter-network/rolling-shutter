package shutterservice

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/params"
	testHelperBindings "github.com/shutter-network/contracts/v2/bindings/eventtriggertesthelper"
	triggerRegistryBindings "github.com/shutter-network/contracts/v2/bindings/shuttereventtriggerregistry"
	"gotest.tools/assert"
)

// This should match the Trigger event from the TestHelper contract
const (
	TestEvtSig     = "Trigger(uint64,bytes32,bytes32,bytes32)"
	TestEvtSigFull = "Trigger(uint64 topic1 indexed, bytes32 topic2 indexed, bytes32 data1, bytes32 data2)"
)

type TestSetup struct {
	backend                *simulated.Backend
	testHelperAddress      common.Address
	testHelperContract     *testHelperBindings.Eventtriggertesthelper
	triggerRegistryAddress common.Address
	triggerContract        *triggerRegistryBindings.Shuttereventtriggerregistry
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

	// Deploy testHelperContract
	testHelperAddress, _, testHelperContract, err := testHelperBindings.DeployEventtriggertesthelper(
		auth,
		blockchain.Client(),
	)
	if err != nil {
		return setup, fmt.Errorf("failed to deploy TestHelper %w", err)
	}
	setup.testHelperAddress = testHelperAddress
	setup.testHelperContract = testHelperContract

	// Deploy ShutterEventTrigger contract
	triggerRegistryAddress, _, triggerContract, err := triggerRegistryBindings.DeployShuttereventtriggerregistry(
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

// Test TestHelper contract gets deployed correctly
func TestEvtDeployContract(t *testing.T) {
	setup, err := SetupAndDeploy()
	assert.NilError(t, err, "setup and deploy failed")

	if len(setup.testHelperAddress.Bytes()) == 0 {
		t.Error("Expected a valid deployment address. Received empty address byte array instead")
	}
}

func RegisterTrigger(t *testing.T, setup TestSetup, trigger EventTriggerDefinition) {
	t.Helper()
	client := setup.backend.Client()
	from := crypto.PubkeyToAddress(setup.key.PublicKey)

	head, _ := client.HeaderByNumber(context.Background(), nil)
	gasPrice := new(big.Int).Add(head.BaseFee, big.NewInt(params.GWei))
	chainid, _ := client.ChainID(context.Background())
	nonce, err := client.PendingNonceAt(context.Background(), from)
	assert.NilError(t, err, "failed to get nonce")
	signer := func(_ common.Address, tx *types.Transaction) (*types.Transaction, error) {
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

func EmitTriggerEvent(t *testing.T, setup TestSetup, topic1 uint64, topic2 [32]byte, data1 [32]byte, data2 [32]byte) {
	t.Helper()
	client := setup.backend.Client()

	from := crypto.PubkeyToAddress(setup.key.PublicKey)
	head, _ := client.HeaderByNumber(context.Background(), nil)
	gasPrice := new(big.Int).Add(head.BaseFee, big.NewInt(params.GWei))
	chainid, _ := client.ChainID(context.Background())
	nonce, err := client.PendingNonceAt(context.Background(), from)
	assert.NilError(t, err, "failed to get nonce")
	signer := func(_ common.Address, tx *types.Transaction) (*types.Transaction, error) {
		return types.SignTx(tx, types.LatestSignerForChainID(chainid), setup.key)
	}

	tx, err := setup.testHelperContract.Trigger(&bind.TransactOpts{
		From:      from,
		Nonce:     big.NewInt(int64(nonce)),
		GasTipCap: big.NewInt(params.GWei),
		GasFeeCap: gasPrice,
		GasLimit:  2100000,
		Signer:    signer,
	}, topic1, topic2, data1, data2)
	assert.NilError(t, err, "error with trigger call")
	txHash := tx.Hash()
	setup.backend.Commit()
	found, pending, err := client.TransactionByHash(context.Background(), txHash)
	assert.NilError(t, err, "error getting tx")
	assert.Check(t, pending == false, "still pending")
	assert.Check(t, found.Hash() == txHash)
	receipt, err := client.TransactionReceipt(context.Background(), txHash)

	assert.NilError(t, err, "error getting receipt")
	assert.Check(t, receipt.Status == 1, "trigger call failed")
}

func TestEvtBloomFilterMatch(t *testing.T) {
	setup, err := SetupAndDeploy()
	assert.NilError(t, err, "failure deploying")

	topic2 := [32]byte{0x02}
	EmitTriggerEvent(t, setup, 1, topic2, [32]byte{0x03}, [32]byte{0x04})
	latest, err := setup.backend.Client().BlockByNumber(context.Background(), nil)
	assert.NilError(t, err, "error getting block")

	topic0 := crypto.Keccak256([]byte(TestEvtSig))
	assert.Check(t, latest.Bloom().Test(topic0), "could not find topic0")

	assert.Check(t, latest.Bloom().Test(WordPad([]byte{0x01})), "could not find topic1")
	assert.Check(t, latest.Bloom().Test(topic2[:]), "could not find topic2")

	// we could probably calculate `bloom9.go:types.bloomValues` for all topics, and manually match the merged/combined topic
}

func CreateDefinition(contract common.Address, topic1 uint64, topic2 [32]byte, data1 [32]byte, data2 [32]byte) EventTriggerDefinition {
	transferEventID := common.BytesToHash(crypto.Keccak256([]byte(TestEvtSig)))
	definition := EventTriggerDefinition{
		Contract:       contract,
		EventSignature: transferEventID,
		Conditions: []Condition{
			{
				Location: TopicData{
					number: 1, // topic1 (uint64)
				},
				Constraint: MatchConstraint{
					target: WordPad([]byte{byte(topic1)}),
				},
			},
			{
				Location: TopicData{
					number: 2, // topic2 (bytes32)
				},
				Constraint: MatchConstraint{
					target: topic2[:],
				},
			},
			{
				Location: OffsetData{
					argnumber: 0, // data1 (bytes32)
					complex:   false,
				},
				Constraint: MatchConstraint{
					target: data1[:],
				},
			},
			{
				Location: OffsetData{
					argnumber: 1, // data2 (bytes32)
					complex:   false,
				},
				Constraint: MatchConstraint{
					target: data2[:],
				},
			},
		},
	}
	return definition
}

func TestEvtEventTriggerDefinition(t *testing.T) {
	setup, err := SetupAndDeploy()
	assert.NilError(t, err, "error during setup")

	definition := CreateDefinition(
		setup.testHelperAddress,
		1,
		[32]byte{0x02},
		[32]byte{0x03},
		[32]byte{0x04})
	assert.Check(t, len(definition.Conditions) == 4, "something went wrong")
	f := definition.ToFilterQuery()
	assert.Check(t, len(f.Topics) > 0, "no filterquery")

	EmitTriggerEvent(t, setup, 1, [32]byte{0x02}, [32]byte{0x03}, [32]byte{0x04})

	checkTopics := true

	logs, err := setup.backend.Client().FilterLogs(context.Background(), f)
	assert.NilError(t, err, "error using filter query")
	assert.Check(t, len(logs) == 1, "filter did not match")
	for _, elog := range logs {
		assert.Check(t, definition.Match(elog, checkTopics) == true, "did not match %v", elog.Data)
	}
	// mismatch on topic2
	EmitTriggerEvent(t, setup, 0, [32]byte{0x04}, [32]byte{0x05}, [32]byte{0x06})

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
			target: WordPad([]byte{0x04}),
		},
	})
	overspecific := definition.ToFilterQuery()
	EmitTriggerEvent(t, setup, 1, [32]byte{0x07}, [32]byte{0x08}, [32]byte{0x09})

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
		setup.testHelperAddress,
		1,
		[32]byte{0x01},
		[32]byte{0x02},
		[32]byte{0x03})
	RegisterTrigger(t, setup, def)

	ser := def.MarshalBytes()

	de := EventTriggerDefinition{}

	err = de.UnmarshalBytes(ser)
	assert.NilError(t, err, "unmarshal failed")
	round := de.MarshalBytes()
	assert.Check(t, bytes.Equal(ser, round), "roundtrip failed\n%v\n%v", ser, de.MarshalBytes())

	logs, err := setup.triggerContract.FilterEventTriggerRegistered(&bind.FilterOpts{
		Start:   uint64(0),
		End:     nil,
		Context: context.Background(),
	},
		[]uint64{1},
	)
	for logs.Next() {
		x := logs.Event.TriggerDefinition
		assert.Check(t, bytes.Equal(ser, x), "serialization mismatch")
	}
}

func TestEventTriggerMarshalUnmarshal(t *testing.T) {
	contractAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	tests := []struct {
		name       string
		definition EventTriggerDefinition
	}{
		{
			name: "empty conditions",
			definition: EventTriggerDefinition{
				Contract:       contractAddr,
				EventSignature: common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
				Conditions:     []Condition{},
			},
		},
		{
			name: "simple topic condition",
			definition: EventTriggerDefinition{
				Contract:       contractAddr,
				EventSignature: common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
				Conditions: []Condition{
					{
						Location: TopicData{number: 1},
						Constraint: MatchConstraint{
							target: WordPad(common.HexToAddress("0xdeadbeef").Bytes()),
						},
					},
				},
			},
		},
		{
			name: "multiple topic conditions",
			definition: EventTriggerDefinition{
				Contract:       contractAddr,
				EventSignature: common.HexToHash("0x3333333333333333333333333333333333333333333333333333333333333333"),
				Conditions: []Condition{
					{
						Location: TopicData{number: 1},
						Constraint: MatchConstraint{
							target: WordPad(common.HexToAddress("0xdeadbeef").Bytes()),
						},
					},
					{
						Location: TopicData{number: 2},
						Constraint: MatchConstraint{
							target: WordPad(common.HexToAddress("0xcafebabe").Bytes()),
						},
					},
					{
						Location: TopicData{number: 3},
						Constraint: MatchConstraint{
							target: WordPad(common.HexToAddress("0x12345678").Bytes()),
						},
					},
				},
			},
		},
		{
			name: "data conditions with numeric constraints",
			definition: EventTriggerDefinition{
				Contract:       contractAddr,
				EventSignature: common.HexToHash("0x4444444444444444444444444444444444444444444444444444444444444444"),
				Conditions: []Condition{
					{
						Location: OffsetData{argnumber: 0, complex: false},
						Constraint: NumConstraint{
							op:     GTE,
							target: big.NewInt(1000),
						},
					},
					{
						Location: OffsetData{argnumber: 1, complex: false},
						Constraint: NumConstraint{
							op:     LT,
							target: big.NewInt(5000),
						},
					},
				},
			},
		},
		{
			name: "mixed topic and data conditions",
			definition: EventTriggerDefinition{
				Contract:       contractAddr,
				EventSignature: common.HexToHash("0x5555555555555555555555555555555555555555555555555555555555555555"),
				Conditions: []Condition{
					{
						Location: TopicData{number: 1},
						Constraint: MatchConstraint{
							target: WordPad(common.HexToAddress("0xdeadbeef").Bytes()),
						},
					},
					{
						Location: OffsetData{argnumber: 0, complex: false},
						Constraint: NumConstraint{
							op:     EQ,
							target: big.NewInt(42),
						},
					},
					{
						Location: OffsetData{argnumber: 1, complex: true},
						Constraint: MatchConstraint{
							target: WordPad([]byte("complex data condition")),
						},
					},
				},
			},
		},
		{
			name: "complex data condition",
			definition: EventTriggerDefinition{
				Contract:       contractAddr,
				EventSignature: common.HexToHash("0x6666666666666666666666666666666666666666666666666666666666666666"),
				Conditions: []Condition{
					{
						Location: OffsetData{argnumber: 0, complex: true},
						Constraint: MatchConstraint{
							target: append(WordPad([]byte("complex word 1")), WordPad([]byte("complex word 2"))...),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled := tt.definition.MarshalBytes()
			assert.Check(t, len(marshaled) > 0, "marshaled data should not be empty")

			var unmarshaled EventTriggerDefinition
			err := unmarshaled.UnmarshalBytes(marshaled)
			assert.NilError(t, err, "unmarshalling should succeed")

			fmt.Printf("Unmarshaled definition: %+v\n", unmarshaled)
			fmt.Printf("Original definition: %+v\n", tt.definition)
			fmt.Printf("Marshaled bytes: %x\n", marshaled)

			remarshal := unmarshaled.MarshalBytes()
			assert.Check(t, bytes.Equal(marshaled, remarshal), "remarshaled data should match original marshaled data")

			assert.Check(t, unmarshaled.Contract == tt.definition.Contract, "contract address should be preserved")
			assert.Check(t, unmarshaled.EventSignature == tt.definition.EventSignature, "event signature should be preserved")
			assert.Check(t, len(unmarshaled.Conditions) == len(tt.definition.Conditions), "number of conditions should be preserved")
			for i, cond := range unmarshaled.Conditions {
				assert.Check(t, reflect.DeepEqual(cond.Location, tt.definition.Conditions[i].Location), "condition location should be preserved")
				assert.Check(t, reflect.DeepEqual(cond.Constraint, tt.definition.Conditions[i].Constraint), "condition constraint should be preserved")
			}
		})
	}
}

func TestEventTriggerMarshalUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		expectError   bool
		errorContains string
	}{
		{
			name:          "empty data",
			data:          []byte{},
			expectError:   true,
			errorContains: "failed to read version",
		},
		{
			name: "wrong version",
			data: []byte{
				0x99, // wrong version
			},
			expectError:   true,
			errorContains: "version mismatch",
		},
		{
			name: "incomplete data - missing contract",
			data: []byte{
				VERSION, // correct version but missing contract data
			},
			expectError: true,
		},
		{
			name:        "incomplete data - truncated contract",
			data:        append([]byte{VERSION}, make([]byte, 10)...), // version + partial contract
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var definition EventTriggerDefinition
			err := definition.UnmarshalBytes(tt.data)

			if tt.expectError {
				assert.Check(t, err != nil, "expected an error but got none")
				if tt.errorContains != "" {
					assert.Check(t, strings.Contains(err.Error(), tt.errorContains),
						"error message should contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				assert.NilError(t, err, "unexpected error")
			}
		})
	}
}

// TestTopicPatternGeneration tests the TopicPattern method
func TestTopicPatternGeneration(t *testing.T) {
	tests := []struct {
		name            string
		conditions      []Condition
		expectedPattern byte
	}{
		{
			name:            "no topic conditions",
			conditions:      []Condition{},
			expectedPattern: 0b000,
		},
		{
			name: "topic1 only",
			conditions: []Condition{
				{Location: TopicData{number: 1}, Constraint: MatchConstraint{target: []byte("test")}},
			},
			expectedPattern: 0b100,
		},
		{
			name: "topic2 only",
			conditions: []Condition{
				{Location: TopicData{number: 2}, Constraint: MatchConstraint{target: []byte("test")}},
			},
			expectedPattern: 0b010,
		},
		{
			name: "topic3 only",
			conditions: []Condition{
				{Location: TopicData{number: 3}, Constraint: MatchConstraint{target: []byte("test")}},
			},
			expectedPattern: 0b001,
		},
		{
			name: "topic1 and topic3",
			conditions: []Condition{
				{Location: TopicData{number: 1}, Constraint: MatchConstraint{target: []byte("test")}},
				{Location: TopicData{number: 3}, Constraint: MatchConstraint{target: []byte("test")}},
			},
			expectedPattern: 0b101,
		},
		{
			name: "all topics",
			conditions: []Condition{
				{Location: TopicData{number: 1}, Constraint: MatchConstraint{target: []byte("test")}},
				{Location: TopicData{number: 2}, Constraint: MatchConstraint{target: []byte("test")}},
				{Location: TopicData{number: 3}, Constraint: MatchConstraint{target: []byte("test")}},
			},
			expectedPattern: 0b111,
		},
		{
			name: "mixed topic and data conditions",
			conditions: []Condition{
				{Location: TopicData{number: 1}, Constraint: MatchConstraint{target: []byte("test")}},
				{Location: OffsetData{argnumber: 0}, Constraint: NumConstraint{op: EQ, target: big.NewInt(42)}},
				{Location: TopicData{number: 3}, Constraint: MatchConstraint{target: []byte("test")}},
			},
			expectedPattern: 0b101,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definition := EventTriggerDefinition{
				Contract:       common.HexToAddress("0x1234"),
				EventSignature: common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				Conditions:     tt.conditions,
			}

			pattern := definition.TopicPattern()
			assert.Check(t, pattern == tt.expectedPattern,
				"expected pattern %b, got %b", tt.expectedPattern, pattern)
		})
	}
}

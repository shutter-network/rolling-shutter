package app

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/shutterevents"
)

// GenesisAppState is used to hold the initial list of keypers, who will bootstrap the system by
// providing the first real BatchConfig to be used. We use common.MixedcaseAddress to hold the list
// of keypers as that one serializes as checksum address.
type GenesisAppState struct {
	Keypers    []common.MixedcaseAddress `json:"keypers"`
	Threshold  uint64                    `json:"threshold"`
	InitialEon uint64                    `json:"initialEon"`
}

func NewGenesisAppState(keypers []common.Address, threshold int, initialEon uint64) GenesisAppState {
	appState := GenesisAppState{
		Threshold:  uint64(threshold),
		InitialEon: initialEon,
	}
	for _, k := range keypers {
		appState.Keypers = append(appState.Keypers, common.NewMixedcaseAddress(k))
	}
	return appState
}

// GetKeypers returns the keypers defined in the GenesisAppState.
func (appState *GenesisAppState) GetKeypers() []common.Address {
	var res []common.Address
	for _, k := range appState.Keypers {
		res = append(res, k.Address())
	}
	return res
}

// ReflectDeepEquals is used to compare any two objects by the Voting generics type using
// reflect.DeepEqual.
type ReflectDeepEquals[T any] struct{}

func (ReflectDeepEquals[T]) Equals(a, b T) bool {
	return reflect.DeepEqual(a, b)
}

// BatchConfigEquals is used to compare two BatchConfig structs by the Voting generics type.
type BatchConfigEquals = ReflectDeepEquals[BatchConfig]

// ComparableEquals is used to compare two Comparables by the Voting generics type.
type ComparableEquals[T comparable] struct{}

func (ComparableEquals[T]) Equals(a, b T) bool {
	return a == b
}

type DKGSuccessVoting = Voting[bool, ComparableEquals[bool]]

// ConfigVoting is used to let the keypers vote on new BatchConfigs to be added.
type ConfigVoting = Voting[BatchConfig, BatchConfigEquals]

// NewConfigVoting creates a ConfigVoting struct.
func NewConfigVoting() ConfigVoting {
	return NewVoting[BatchConfig, BatchConfigEquals]()
}

// ValidatorPubkey holds the raw 32 byte ed25519 public key to be used as tendermint validator key
// We use this is a map key, so don't use a byte slice.
type ValidatorPubkey struct {
	Ed25519pubkey string
}

func (vp ValidatorPubkey) String() string {
	return fmt.Sprintf("ed25519:%s", hex.EncodeToString([]byte(vp.Ed25519pubkey)))
}

// Powermap maps a ValidatorPubkey to the validators voting power.
type Powermap map[ValidatorPubkey]int64

// NewValidatorPubkey creates a new ValidatorPubkey from a 32 byte ed25519 raw pubkey. See
// https://docs.tendermint.com/master/spec/abci/apps.html#validator-updates for more information
func NewValidatorPubkey(pubkey []byte) (ValidatorPubkey, error) {
	if len(pubkey) != ed25519.PublicKeySize {
		return ValidatorPubkey{}, errors.Errorf("pubkey must be 32 bytes")
	}
	return ValidatorPubkey{Ed25519pubkey: string(pubkey)}, nil
}

// ShutterApp holds our data structures used for the tendermint app.
type ShutterApp struct {
	Configs      []*BatchConfig
	DKGMap       map[uint64]*DKGInstance // map eon to DKGInstance
	ConfigVoting ConfigVoting
	// EonStartVotings map[uint64]*EonStartVoting
	Gobpath         string
	LastSaved       time.Time
	LastBlockHeight int64
	Identities      map[common.Address]ValidatorPubkey
	BlocksSeen      map[common.Address]uint64
	Validators      Powermap
	EONCounter      uint64
	DevMode         bool
	CheckTxState    *CheckTxState
	NonceTracker    *NonceTracker
	ChainID         string
}

// CheckTxState is a part of the state used by CheckTx calls that is reset at every commit.
type CheckTxState struct {
	Members      map[common.Address]bool
	TxCounts     map[common.Address]int
	NonceTracker *NonceTracker
}

// NonceTracker tracks which nonces have been used and which have not.
type NonceTracker struct {
	RandomNonces map[common.Address]map[uint64]bool
}

type SenderReceiverPair struct {
	Sender, Receiver common.Address
}

// DKGInstance manages the state of one eon key generation instance.
type DKGInstance struct {
	Config              BatchConfig
	Eon                 uint64
	SuccessVoting       DKGSuccessVoting
	PolyEvalsSeen       map[SenderReceiverPair]struct{}
	PolyCommitmentsSeen map[common.Address]struct{}
	AccusationsSeen     map[common.Address]struct{}
	ApologiesSeen       map[common.Address]struct{}
}

type (
	Accusation     = shutterevents.Accusation
	Apology        = shutterevents.Apology
	BatchConfig    = shutterevents.BatchConfig
	PolyCommitment = shutterevents.PolyCommitment
	PolyEval       = shutterevents.PolyEval
)

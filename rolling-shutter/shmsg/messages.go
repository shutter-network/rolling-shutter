package shmsg

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"google.golang.org/protobuf/reflect/protoreflect"

	shcrypto "github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrtopics"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprtopics"
)

// P2PMessage can be send via the p2p protocol.
type P2PMessage interface {
	protoreflect.ProtoMessage
	ImplementsP2PMessage()
	GetInstanceID() uint64
	Topic() string
}

func (*DecryptionTrigger) ImplementsP2PMessage() {
}

func (*DecryptionTrigger) Topic() string {
	return kprtopics.DecryptionTrigger
}

func (*DecryptionKeyShare) ImplementsP2PMessage() {
}

func (*DecryptionKeyShare) Topic() string {
	return kprtopics.DecryptionKeyShare
}

func (*DecryptionKey) ImplementsP2PMessage() {
}

func (*DecryptionKey) Topic() string {
	return dcrtopics.DecryptionKey
}

func (*CipherBatch) ImplementsP2PMessage() {
}

func (*CipherBatch) Topic() string {
	return dcrtopics.CipherBatch
}

func (*DecryptionSignature) ImplementsP2PMessage() {
}

func (*DecryptionSignature) Topic() string {
	return dcrtopics.DecryptionSignature
}

func (*AggregatedDecryptionSignature) ImplementsP2PMessage() {
}

func (*AggregatedDecryptionSignature) Topic() string {
	return dcrtopics.AggregatedDecryptionSignature
}

// NewBatchConfig creates a new BatchConfig message.
func NewBatchConfig(
	startBatchIndex uint64,
	keypers []common.Address,
	threshold uint64,
	configIndex uint64,
	started bool,
	validatorsUpdated bool,
) *Message {
	var keypersBytes [][]byte
	for _, k := range keypers {
		keypersBytes = append(keypersBytes, k.Bytes())
	}

	return &Message{
		Payload: &Message_BatchConfig{
			BatchConfig: &BatchConfig{
				StartBatchIndex:   startBatchIndex,
				Keypers:           keypersBytes,
				Threshold:         threshold,
				ConfigIndex:       configIndex,
				Started:           started,
				ValidatorsUpdated: validatorsUpdated,
			},
		},
	}
}

// NewApology creates a new apology message used in the DKG process. This message reveals the
// polyEvals, that where sent encrypted via the PolyEval messages to each accuser.
func NewApology(eon uint64, accusers []common.Address, polyEvals []*big.Int) *Message {
	if len(accusers) != len(polyEvals) {
		panic("bad call to NewApology")
	}

	var accusersBytes [][]byte
	for _, a := range accusers {
		accusersBytes = append(accusersBytes, a.Bytes())
	}

	var polyEvalsBytes [][]byte
	for _, e := range polyEvals {
		polyEvalsBytes = append(polyEvalsBytes, e.Bytes())
	}

	return &Message{
		Payload: &Message_Apology{
			Apology: &Apology{
				Eon:       eon,
				Accusers:  accusersBytes,
				PolyEvals: polyEvalsBytes,
			},
		},
	}
}

func NewAccusation(eon uint64, accused []common.Address) *Message {
	accusedBytes := [][]byte{}
	for _, a := range accused {
		accusedBytes = append(accusedBytes, a.Bytes())
	}
	return &Message{
		Payload: &Message_Accusation{
			Accusation: &Accusation{
				Eon:     eon,
				Accused: accusedBytes,
			},
		},
	}
}

// NewPolyCommitment creates a new poly commitment message containing gamma values.
func NewPolyCommitment(eon uint64, gammas *shcrypto.Gammas) *Message {
	gammaBytes := [][]byte{}
	for _, gamma := range *gammas {
		gammaBytes = append(gammaBytes, gamma.Marshal())
	}

	return &Message{
		Payload: &Message_PolyCommitment{
			PolyCommitment: &PolyCommitment{
				Eon:    eon,
				Gammas: gammaBytes,
			},
		},
	}
}

// NewPolyEval creates a new poly eval message.
func NewPolyEval(eon uint64, receivers []common.Address, encryptedEvals [][]byte) *Message {
	rs := [][]byte{}
	for _, receiver := range receivers {
		rs = append(rs, receiver.Bytes())
	}

	return &Message{
		Payload: &Message_PolyEval{
			PolyEval: &PolyEval{
				Eon:            eon,
				Receivers:      rs,
				EncryptedEvals: encryptedEvals,
			},
		},
	}
}

// NewEonStartVote creates a new eon start vote message.
func NewEonStartVote(startBatchIndex uint64) *Message {
	return &Message{
		Payload: &Message_EonStartVote{
			EonStartVote: &EonStartVote{
				StartBatchIndex: startBatchIndex,
			},
		},
	}
}

// NewBatchConfigStarted creates a new BatchConfigStarted message.
func NewBatchConfigStarted(configIndex uint64) *Message {
	return &Message{
		Payload: &Message_BatchConfigStarted{
			BatchConfigStarted: &BatchConfigStarted{
				BatchConfigIndex: configIndex,
			},
		},
	}
}

// NewCheckIn creates a new CheckIn message.
func NewCheckIn(validatorPublicKey []byte, encryptionKey *ecies.PublicKey) *Message {
	encryptionKeyECDSA := encryptionKey.ExportECDSA()
	return &Message{
		Payload: &Message_CheckIn{
			CheckIn: &CheckIn{
				ValidatorPublicKey:  validatorPublicKey,
				EncryptionPublicKey: crypto.CompressPubkey(encryptionKeyECDSA),
			},
		},
	}
}

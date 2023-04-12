package shmsg

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protoreflect"

	shcrypto "github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrtopics"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprtopics"
	"github.com/shutter-network/shutter/shuttermint/snapshot/snptopics"
)

// P2PMessage can be send via the p2p protocol.
type P2PMessage interface {
	protoreflect.ProtoMessage
	ImplementsP2PMessage()
	GetInstanceID() uint64
	Topic() string
	LogInfo() string
}

func (*DecryptionTrigger) ImplementsP2PMessage() {
}

func (trigger *DecryptionTrigger) LogInfo() string {
	return fmt.Sprintf("DecryptionTrigger{epochid=%X}", trigger.EpochID)
}

func (*DecryptionTrigger) Topic() string {
	return kprtopics.DecryptionTrigger
}

func (*DecryptionKeyShare) ImplementsP2PMessage() {
}

func (share *DecryptionKeyShare) GetEpochSecretKeyShare() (*shcrypto.EpochSecretKeyShare, error) {
	epochSecretKeyShare := new(shcrypto.EpochSecretKeyShare)
	if err := epochSecretKeyShare.Unmarshal(share.GetShare()); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key share P2P message")
	}
	return epochSecretKeyShare, nil
}

func (share *DecryptionKeyShare) LogInfo() string {
	return fmt.Sprintf(
		"DecryptionKeyShare{epochid=%X, keyperIndex=%d}",
		share.EpochID,
		share.KeyperIndex,
	)
}

func (*DecryptionKeyShare) Topic() string {
	return kprtopics.DecryptionKeyShare
}

func (*DecryptionKey) ImplementsP2PMessage() {
}

func (key *DecryptionKey) LogInfo() string {
	return fmt.Sprintf("DecryptionKey{epochid=%X}", key.EpochID)
}

func (key *DecryptionKey) GetEpochSecretKey() (*shcrypto.EpochSecretKey, error) {
	epochSecretKey := new(shcrypto.EpochSecretKey)
	if err := epochSecretKey.Unmarshal(key.GetKey()); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key P2P message")
	}
	return epochSecretKey, nil
}

func (*DecryptionKey) Topic() string {
	return dcrtopics.DecryptionKey
}

func (*CipherBatch) ImplementsP2PMessage() {
}

func (batch *CipherBatch) LogInfo() string {
	return fmt.Sprintf(
		"CipherBatch{epochid=%X, num tx=%d}",
		batch.DecryptionTrigger.EpochID,
		len(batch.Transactions),
	)
}

func (*CipherBatch) Topic() string {
	return dcrtopics.CipherBatch
}

func (cipherBatch *CipherBatch) GetInstanceID() uint64 {
	return cipherBatch.DecryptionTrigger.GetInstanceID()
}

func (*DecryptionSignature) ImplementsP2PMessage() {
}

func (sig *DecryptionSignature) LogInfo() string {
	return fmt.Sprintf(
		"DecryptionSignature{epochid=%X}",
		sig.EpochID,
	)
}

func (*DecryptionSignature) Topic() string {
	return dcrtopics.DecryptionSignature
}

func (*AggregatedDecryptionSignature) ImplementsP2PMessage() {
}

func (ads *AggregatedDecryptionSignature) LogInfo() string {
	return fmt.Sprintf(
		"AggregatedDecryptionSignature{epochid=%X}",
		ads.EpochID,
	)
}

func (*AggregatedDecryptionSignature) Topic() string {
	return dcrtopics.AggregatedDecryptionSignature
}

func (*EonPublicKey) ImplementsP2PMessage() {
}

func (e *EonPublicKey) LogInfo() string {
	return fmt.Sprintf(
		"EonPublicKey{eon=%d}",
		e.Eon,
	)
}

func (*EonPublicKey) Topic() string {
	return kprtopics.EonPublicKey
}

func (*TimedEpoch) ImplementsP2PMessage() {}

func (te *TimedEpoch) LogInfo() string {
	return fmt.Sprintf("TimedEpoch{EpochID=%X, NotBefore=%d}", te.EpochID, te.NotBefore)
}

func (*TimedEpoch) Topic() string {
	return snptopics.TimedEpoch
}

// NewBatchConfig creates a new BatchConfig message.
func NewBatchConfig(
	activationBlockNumber uint64,
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
				ActivationBlockNumber: activationBlockNumber,
				Keypers:               keypersBytes,
				Threshold:             threshold,
				ConfigIndex:           configIndex,
				Started:               started,
				ValidatorsUpdated:     validatorsUpdated,
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
func NewEonStartVote(activationBlockNumber uint64) *Message {
	return &Message{
		Payload: &Message_EonStartVote{
			EonStartVote: &EonStartVote{
				ActivationBlockNumber: activationBlockNumber,
			},
		},
	}
}

// NewBlockSeen creates a new BlockSeen message. The keypers send this when they see a new block on
// the main chain that possibly leads to starting a batch config.
func NewBlockSeen(blockNumber uint64) *Message {
	return &Message{
		Payload: &Message_BlockSeen{
			BlockSeen: &BlockSeen{
				BlockNumber: blockNumber,
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

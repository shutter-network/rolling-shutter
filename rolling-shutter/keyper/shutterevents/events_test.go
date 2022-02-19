package shutterevents_test

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shtest"
	"github.com/shutter-network/shutter/shuttermint/keyper/shutterevents"
)

var (
	polynomial *shcrypto.Polynomial
	gammas     shcrypto.Gammas
	eon        = uint64(64738)
	sender     = common.BytesToAddress([]byte("foo"))
	addresses  = []common.Address{
		common.BigToAddress(big.NewInt(1)),
		common.BigToAddress(big.NewInt(2)),
		common.BigToAddress(big.NewInt(3)),
	}
)

func init() {
	var err error
	polynomial, err = shcrypto.RandomPolynomial(rand.Reader, 3)
	if err != nil {
		panic(err)
	}

	gammas = *polynomial.Gammas()
}

var eciesPublicKeyComparer = gocmp.Comparer(func(x, y *ecies.PublicKey) bool {
	return reflect.DeepEqual(x, y)
})

// roundtrip checks that the given IEvent round-trips, i.e. it can be serialized as an ABCI Event
// and deserialized back again to an equal value.
func roundtrip(t *testing.T, ev shutterevents.IEvent) {
	t.Helper()
	ev2, err := shutterevents.MakeEvent(ev.MakeABCIEvent(), 0)
	assert.NilError(t, err)
	assert.DeepEqual(t, ev, ev2, shtest.BigIntComparer, eciesPublicKeyComparer)
}

func TestAccusation(t *testing.T) {
	ev := &shutterevents.Accusation{
		Eon:     eon,
		Sender:  sender,
		Accused: addresses,
	}
	roundtrip(t, ev)
}

func TestEmptyAccusation(t *testing.T) {
	ev := &shutterevents.Accusation{
		Eon:    eon,
		Sender: sender,
	}
	roundtrip(t, ev)
}

func TestApology(t *testing.T) {
	accusers := addresses
	var polyEval []*big.Int
	for i := 0; i < len(accusers); i++ {
		eval := big.NewInt(int64(100 + i))
		polyEval = append(polyEval, eval)
	}
	ev := &shutterevents.Apology{
		Eon:      eon,
		Sender:   sender,
		Accusers: accusers,
		PolyEval: polyEval,
	}
	roundtrip(t, ev)
}

func TestEmptyApology(t *testing.T) {
	ev := &shutterevents.Apology{
		Eon:    eon,
		Sender: sender,
	}
	roundtrip(t, ev)
}

func TestBatchConfig(t *testing.T) {
	ev := &shutterevents.BatchConfig{
		ActivationBlockNumber: 111,
		Threshold:             2,
		Keypers:               addresses,
		ConfigIndex:           uint64(0xffffffffffffffff),
	}
	roundtrip(t, ev)
	// XXX should we implement this or drop the field?
	// ev.Started = true
	// roundtrip(t, ev)
}

func TestBatchConfigStarted(t *testing.T) {
	ev := &shutterevents.BatchConfigStarted{
		ConfigIndex: uint64(0xffffffffffffffff),
	}
	roundtrip(t, ev)
}

func TestCheckIn(t *testing.T) {
	privateKeyECDSA, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	publicKey := ecies.ImportECDSAPublic(&privateKeyECDSA.PublicKey)
	ev := &shutterevents.CheckIn{Sender: sender, EncryptionPublicKey: publicKey}
	roundtrip(t, ev)
}

func TestEonStarted(t *testing.T) {
	ev := &shutterevents.EonStarted{Eon: eon, ActivationBlockNumber: 9999, ConfigIndex: 567}
	roundtrip(t, ev)
}

func TestPolyCommitment(t *testing.T) {
	ev := &shutterevents.PolyCommitment{
		Eon:    eon,
		Sender: sender,
		Gammas: &gammas,
	}
	roundtrip(t, ev)
}

func TestPolyEval(t *testing.T) {
	var receivers []common.Address
	var encryptedEvals [][]byte
	for i := 1; i < 10; i++ {
		receivers = append(receivers, common.BigToAddress(new(big.Int).SetUint64(uint64(i))))
		encryptedEvals = append(encryptedEvals, []byte(fmt.Sprintf("encrypted: %d", i)))
	}
	ev := &shutterevents.PolyEval{
		Eon:            eon,
		Sender:         sender,
		Receivers:      receivers,
		EncryptedEvals: encryptedEvals,
	}
	roundtrip(t, ev)
}

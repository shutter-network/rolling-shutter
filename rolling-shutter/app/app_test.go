package app

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/shutter-network/shutter/shlib/shtest"
)

func TestNewShutterApp(t *testing.T) {
	app := NewShutterApp()
	assert.Equal(t, len(app.Configs), 1, "Configs should contain exactly one guard element")
	assert.Assert(t, is.DeepEqual(app.Configs[0], &BatchConfig{}), "Bad guard element")
}

func TestGetBatch(t *testing.T) {
	app := NewShutterApp()

	err := app.addConfig(BatchConfig{
		ConfigIndex:           1,
		ActivationBlockNumber: 100,
		Threshold:             1,
		Keypers:               addr,
	})
	assert.NilError(t, err)

	err = app.addConfig(BatchConfig{
		ConfigIndex:           2,
		ActivationBlockNumber: 200,
		Threshold:             2,
		Keypers:               addr,
	})
	assert.NilError(t, err)

	err = app.addConfig(BatchConfig{
		ConfigIndex:           3,
		ActivationBlockNumber: 300,
		Threshold:             3,
		Keypers:               addr,
	})
	assert.NilError(t, err)

	assert.Equal(t, uint64(0), app.getBatchState(0).Config.Threshold)
	assert.Equal(t, uint64(0), app.getBatchState(99).Config.Threshold)
	assert.Equal(t, uint64(1), app.getBatchState(100).Config.Threshold)
	assert.Equal(t, uint64(1), app.getBatchState(101).Config.Threshold)
	assert.Equal(t, uint64(1), app.getBatchState(199).Config.Threshold)
	assert.Equal(t, uint64(2), app.getBatchState(200).Config.Threshold)
	assert.Equal(t, uint64(3), app.getBatchState(1000).Config.Threshold)
}

func TestAddConfig(t *testing.T) {
	app := NewShutterApp()

	err := app.addConfig(BatchConfig{
		ConfigIndex:           1,
		ActivationBlockNumber: 100,
		Threshold:             1,
		Keypers:               addr,
	})
	assert.NilError(t, err)

	err = app.addConfig(BatchConfig{
		ConfigIndex:           2,
		ActivationBlockNumber: 99,
		Threshold:             1,
		Keypers:               addr,
	})
	assert.Assert(t, err != nil, "Expected error, ActivationBlockNumber must not decrease")

	err = app.addConfig(BatchConfig{
		ConfigIndex:           1,
		ActivationBlockNumber: 100,
		Threshold:             1,
		Keypers:               addr,
	})
	assert.Assert(t, err != nil, "Expected error, ConfigIndex must increase")

	err = app.addConfig(BatchConfig{
		ConfigIndex:           2,
		ActivationBlockNumber: 100,
		Threshold:             2,
		Keypers:               addr,
	})
	assert.NilError(t, err)
}

func TestGobDKG(t *testing.T) {
	var eon uint64 = 201
	var err error
	keypers := addr
	dkg := NewDKGInstance(BatchConfig{
		ConfigIndex:           1,
		ActivationBlockNumber: 100,
		Threshold:             1,
		Keypers:               keypers,
	}, eon)

	err = dkg.RegisterAccusationMsg(Accusation{
		Sender:  keypers[0],
		Eon:     eon,
		Accused: []common.Address{keypers[1]},
	})
	assert.NilError(t, err)

	err = dkg.RegisterApologyMsg(Apology{
		Sender:   keypers[0],
		Eon:      eon,
		Accusers: []common.Address{keypers[1]},
	})
	assert.NilError(t, err)

	err = dkg.RegisterPolyCommitmentMsg(PolyCommitment{
		Sender: keypers[0],
		Eon:    eon,
	})
	assert.NilError(t, err)

	err = dkg.RegisterPolyEvalMsg(PolyEval{
		Sender:         keypers[0],
		Eon:            eon,
		Receivers:      []common.Address{keypers[1]},
		EncryptedEvals: [][]byte{{}},
	})
	assert.NilError(t, err)

	shtest.EnsureGobable(t, &dkg, new(DKGInstance))
}

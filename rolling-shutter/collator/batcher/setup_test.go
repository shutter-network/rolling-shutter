package batcher

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/sequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

const numAccounts = 5

// From the gocmp Example on Reporter interface:

// DiffReporter is a simple custom reporter that only records differences
// detected during comparison.
type DiffReporter struct {
	path  gocmp.Path
	diffs []string
}

func (r *DiffReporter) PushStep(ps gocmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *DiffReporter) Report(rs gocmp.Result) {
	if !rs.Equal() {
		vx, vy := r.path.Last().Values()
		r.diffs = append(r.diffs, fmt.Sprintf("%#v:\n\t-: %+v\n\t+: %+v\n", r.path, vx, vy))
	}
}

func (r *DiffReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *DiffReporter) String() string {
	return strings.Join(r.diffs, "\n")
}

func assertEqual(t *testing.T, cancel func(error), x, y any, opts ...gocmp.Option) { //nolint:unparam
	t.Helper()
	rep := &DiffReporter{}
	opts = append(opts, gocmp.Reporter(rep))
	res := gocmp.Equal(x, y, opts...)
	if !res {
		// NOTE: this does not have the AST based reporting
		// of the expression being evaluated.
		// This would be nice to have, but is not a must.
		msg := "assertEqual failed " + rep.String()
		t.Error(msg)
		cancel(errors.New(msg))
	}
}

func compareBigInt(a, b *big.Int) bool {
	return a.Cmp(b) == 0
}

func compareByte(a, b []byte) bool {
	return bytes.Equal(a, b)
}

func newTestConfig(t *testing.T) config.Config {
	t.Helper()

	ethereumKey, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	return config.Config{
		EthereumURL:                  "http://127.0.0.1:8454",
		SequencerURL:                 "http://127.0.0.1:8455",
		EthereumKey:                  ethereumKey,
		ExecutionBlockDelay:          uint32(5),
		InstanceID:                   123,
		EpochDuration:                2 * time.Second,
		BatchIndexAcceptenceInterval: 5,
	}
}

type TestParams struct {
	GasLimit       uint64
	InitialBalance *big.Int
	BaseFee        *big.Int
	TxGasTipCap    *big.Int
	TxGasFeeCap    *big.Int
	InitialEpochID epochid.EpochID
	EpochDuration  time.Duration
}

type Fixture struct {
	Config      config.Config
	EthL1Server *sequencer.MockEthServer
	EthL2Server *sequencer.MockEthServer
	Address     common.Address
	Coinbase    common.Address
	Batcher     *Batcher
	Params      TestParams
	DB          *cltrdb.Queries
	ChainID     *big.Int
	Keys        [numAccounts]*ecdsa.PrivateKey
}

func Setup(ctx context.Context, t *testing.T, params TestParams) *Fixture {
	t.Helper()

	var keys [numAccounts]*ecdsa.PrivateKey
	for i := 0; i < numAccounts; i++ {
		k, err := ethcrypto.GenerateKey()
		assert.NilError(t, err)
		keys[i] = k
	}

	cfg := newTestConfig(t)
	if params.EpochDuration != 0 {
		cfg.EpochDuration = params.EpochDuration
	}

	ethL1 := sequencer.RunMockEthServer(t)
	t.Cleanup(ethL1.Teardown)
	cfg.EthereumURL = ethL1.URL

	ethL2 := sequencer.RunMockEthServer(t)
	t.Cleanup(ethL2.Teardown)
	cfg.SequencerURL = ethL2.URL
	ethL2.SetBatchIndex(params.InitialEpochID.Uint64() - 1)

	db, dbpool, dbteardown := testdb.NewCollatorTestDB(ctx, t)
	t.Cleanup(dbteardown)

	address := ethcrypto.PubkeyToAddress(keys[0].PublicKey)
	chainID := big.NewInt(199)
	gasLimit := params.GasLimit
	coinbase := common.HexToAddress("0x0000000000000000000000000000000000000000")

	// Set the values on the dummy rpc server
	ethL1.SetBlockNumber(1)
	ethL2.SetBalance(address, params.InitialBalance, "latest")
	ethL2.SetBalance(coinbase, big.NewInt(0), "latest")
	ethL2.SetNonce(address, uint64(0), "latest")
	ethL2.SetChainID(chainID)
	ethL2.SetBlock(params.BaseFee, gasLimit, "latest")

	// set initial ("next") epoch id and block number manually,
	// this is usually done in the collator and not in the handler

	err := db.SetNextBatch(ctx, cltrdb.SetNextBatchParams{
		EpochID:       params.InitialEpochID.Bytes(),
		L1BlockNumber: 42,
	})
	assert.NilError(t, err)
	batcher, err := NewBatcher(ctx, cfg, dbpool)
	assert.NilError(t, err)
	return &Fixture{
		Config:      cfg,
		EthL1Server: ethL1,
		EthL2Server: ethL2,
		Address:     address,
		Coinbase:    coinbase,
		Batcher:     batcher,
		Params:      params,
		DB:          db,
		ChainID:     chainID,
		Keys:        keys,
	}
}

func (fix *Fixture) AddEonPublicKey(ctx context.Context, t *testing.T) {
	t.Helper()
	hash := []byte{1, 2, 3}
	pubkey := []byte{4, 5, 6}
	err := fix.DB.InsertEonPublicKeyCandidate(ctx, cltrdb.InsertEonPublicKeyCandidateParams{
		Hash:         hash,
		EonPublicKey: pubkey,
	})
	assert.NilError(t, err)
	err = fix.DB.ConfirmEonPublicKey(ctx, hash)
	assert.NilError(t, err)
}

func (fix *Fixture) MakeTx(t *testing.T, accountIndex, batchIndex, nonce, gas int) ([]byte, []byte) {
	t.Helper()
	assert.Check(t, accountIndex >= 0 && accountIndex < numAccounts)
	// construct a valid transaction
	txData := &txtypes.ShutterTx{
		ChainID:          fix.ChainID,
		Nonce:            uint64(nonce),
		GasTipCap:        fix.Params.TxGasTipCap,
		GasFeeCap:        fix.Params.TxGasFeeCap,
		Gas:              uint64(gas),
		EncryptedPayload: []byte("foo"),
		BatchIndex:       uint64(batchIndex),
	}
	tx, err := txtypes.SignNewTx(fix.Keys[accountIndex], txtypes.LatestSignerForChainID(fix.ChainID), txData)
	assert.NilError(t, err)

	// marshal tx to bytes
	txBytes, err := tx.MarshalBinary()
	assert.NilError(t, err)
	return txBytes, tx.Hash().Bytes()
}

// `p2pMessagingMock` shortcuts the entire P2P messaging layer that would
// normally handle the communication with the keyper-set.
// If the `p2pMessagingMock` receives a DecryptionTrigger,
// it will wait a fixed amount of time and then
// call the BatchHandler's HandleDecryptionKey with a fixed decryption-key
// for that epoch.
// The data of the DecryptionTrigger is not fully validated.
func p2pMessagingMock(
	ctx context.Context,
	t *testing.T,
	failFunc func(error),
	cfg config.Config,
	batchHandler *batchhandler.BatchHandler,
) error {
	t.Helper()

	messages := batchHandler.Messages()
	for {
		select {
		case out, ok := <-messages:
			if !ok {
				log.Debug().Msg("P2P message mock: messages closed from outside")
				return nil
			}
			// Emulate communication overhead etc.
			time.Sleep(500 * time.Millisecond)
			typ := out.ProtoReflect().Type().Descriptor().FullName()
			if typ == "shmsg.DecryptionTrigger" {
				trigger, _ := out.(*shmsg.DecryptionTrigger)
				epoch, _ := epochid.BytesToEpochID(trigger.EpochID)

				assertEqual(t, failFunc, trigger.InstanceID, cfg.InstanceID)
				address := ethcrypto.PubkeyToAddress(cfg.EthereumKey.PublicKey)
				signatureCorrect, err := shmsg.VerifySignature(trigger, address)
				assertEqual(t, failFunc, err, nil)
				assertEqual(t, failFunc, signatureCorrect, true)

				// collator successfully received the decryption-key via P2P messaging,
				// pass to the batch-handler
				err = batchHandler.HandleDecryptionKey(ctx, epoch, []byte("decrkey"))
				assertEqual(t, failFunc, err, nil)
			}

		case <-ctx.Done():
			err := ctx.Err()
			if err == context.Canceled {
				return nil
			}
			return err
		}
	}
}

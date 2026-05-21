package txsender

import (
	"context"
	"database/sql"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"gotest.tools/assert"

	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func TestBigIntNumericRoundtrip(t *testing.T) {
	cases := []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(1_000_000_000_000_000_000), // 1 ETH in wei
		new(big.Int).Mul(big.NewInt(1e18), big.NewInt(1e9)),
	}
	for _, want := range cases {
		num, err := bigIntToNumeric(want)
		assert.NilError(t, err)
		got, err := numericToBigInt(num)
		assert.NilError(t, err)
		assert.Equal(t, want.Cmp(got), 0, "roundtrip mismatch: want %s got %s", want, got)
	}
}

func TestNilBigIntEncodesAsZero(t *testing.T) {
	num, err := bigIntToNumeric(nil)
	assert.NilError(t, err)
	got, err := numericToBigInt(num)
	assert.NilError(t, err)
	assert.Equal(t, got.Sign(), 0)
}

func TestNullNumericDecodesAsZero(t *testing.T) {
	got, err := numericToBigInt(pgtype.Numeric{Status: pgtype.Null})
	assert.NilError(t, err)
	assert.Equal(t, got.Sign(), 0)
}

// fakeClient records the calls made into it so tests can assert on which
// chain-side operations the sender performed.
type fakeClient struct {
	mu                       sync.Mutex
	chainID                  *big.Int
	nonce                    uint64
	gasLimit                 uint64
	tipCap                   *big.Int
	baseFee                  *big.Int
	receiptErr               error
	sendErr                  error
	sentTxs                  []*types.Transaction
	receiptCallsByHash       map[common.Hash]int
	sendTransactionCallCount int32
}

func newFakeClient() *fakeClient {
	return &fakeClient{
		chainID:            big.NewInt(1337),
		gasLimit:           21000,
		tipCap:             big.NewInt(2_000_000_000),
		baseFee:            big.NewInt(5_000_000_000),
		receiptErr:         ethereum.NotFound,
		receiptCallsByHash: map[common.Hash]int{},
	}
}

func (c *fakeClient) ChainID(_ context.Context) (*big.Int, error) {
	return c.chainID, nil
}

func (c *fakeClient) PendingNonceAt(_ context.Context, _ common.Address) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := c.nonce
	c.nonce++
	return n, nil
}

func (c *fakeClient) SuggestGasTipCap(_ context.Context) (*big.Int, error) {
	return new(big.Int).Set(c.tipCap), nil
}

func (c *fakeClient) HeaderByNumber(_ context.Context, _ *big.Int) (*types.Header, error) {
	return &types.Header{BaseFee: new(big.Int).Set(c.baseFee)}, nil
}

func (c *fakeClient) EstimateGas(_ context.Context, _ ethereum.CallMsg) (uint64, error) {
	return c.gasLimit, nil
}

func (c *fakeClient) SendTransaction(_ context.Context, tx *types.Transaction) error {
	atomic.AddInt32(&c.sendTransactionCallCount, 1)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sendErr != nil {
		return c.sendErr
	}
	c.sentTxs = append(c.sentTxs, tx)
	return nil
}

func (c *fakeClient) TransactionReceipt(_ context.Context, h common.Hash) (*types.Receipt, error) {
	c.mu.Lock()
	c.receiptCallsByHash[h]++
	c.mu.Unlock()
	return nil, c.receiptErr
}

func newTestSender(t *testing.T, dbpool *pgxpool.Pool) (*TxSender, *fakeClient) {
	t.Helper()
	key, err := crypto.GenerateKey()
	assert.NilError(t, err)
	fc := newFakeClient()
	s := New(Config{
		DBPool:     dbpool,
		Client:     fc,
		PrivateKey: key,
	})
	s.chainID = fc.chainID
	return s, fc
}

// TestPollProcessesBothPhasesInOneIteration verifies the merged poll loop:
// injecting one pending row and one submitted row results in both being
// processed within a single poll iteration. The pending row should be sent
// via the fake client, and the submitted row should have its receipt queried.
func TestPollProcessesBothPhasesInOneIteration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	s, fc := newTestSender(t, dbpool)

	queries := corekeyperdb.New(dbpool)
	zero, err := bigIntToNumeric(big.NewInt(0))
	assert.NilError(t, err)

	// One pending row, to be picked up by the submit phase.
	pendingID, err := queries.InsertPendingTx(ctx, corekeyperdb.InsertPendingTxParams{
		ToAddress: common.HexToAddress("0x000000000000000000000000000000000000dead").Hex(),
		Data:      []byte{0x01, 0x02, 0x03},
		Value:     zero,
	})
	assert.NilError(t, err)

	// One row pre-marked submitted, to be picked up by the confirm phase.
	submittedID, err := queries.InsertPendingTx(ctx, corekeyperdb.InsertPendingTxParams{
		ToAddress: common.HexToAddress("0x00000000000000000000000000000000000000ff").Hex(),
		Data:      []byte{0xaa, 0xbb},
		Value:     zero,
	})
	assert.NilError(t, err)
	preexistingHash := common.HexToHash("0xabc0000000000000000000000000000000000000000000000000000000000001")
	err = queries.MarkTxSubmitted(ctx, corekeyperdb.MarkTxSubmittedParams{
		ID:     submittedID,
		TxHash: sql.NullString{String: preexistingHash.Hex(), Valid: true},
		Nonce:  sql.NullInt64{Int64: 42, Valid: true},
	})
	assert.NilError(t, err)

	s.poll(ctx)

	assert.Equal(t, int32(1), atomic.LoadInt32(&fc.sendTransactionCallCount),
		"submit phase: expected exactly one SendTransaction call (the pending row)")
	pendingAfter, err := queries.GetTxOutboxByID(ctx, pendingID)
	assert.NilError(t, err)
	assert.Equal(t, "submitted", pendingAfter.Status,
		"submit phase: pending row should be marked submitted")

	assert.Equal(t, 1, fc.receiptCallsByHash[preexistingHash],
		"confirm phase: expected one TransactionReceipt call for the pre-existing submitted row")
}

// TestSubmitRowMarksFailedWhenSendFails verifies the mark-before-send ordering:
// when SendTransaction returns an error after the row has already been marked
// submitted, the row must end up in `failed` status (the submitted -> failed
// transition is intentional and recorded via MarkTxFailed).
func TestSubmitRowMarksFailedWhenSendFails(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	s, fc := newTestSender(t, dbpool)
	fc.sendErr = stubError("rpc connection refused")

	queries := corekeyperdb.New(dbpool)
	zero, err := bigIntToNumeric(big.NewInt(0))
	assert.NilError(t, err)
	id, err := queries.InsertPendingTx(ctx, corekeyperdb.InsertPendingTxParams{
		ToAddress: common.HexToAddress("0x000000000000000000000000000000000000dead").Hex(),
		Data:      []byte{0x01, 0x02, 0x03},
		Value:     zero,
	})
	assert.NilError(t, err)

	s.poll(ctx)

	assert.Equal(t, int32(1), atomic.LoadInt32(&fc.sendTransactionCallCount),
		"SendTransaction should be invoked exactly once before failure is recorded")
	row, err := queries.GetTxOutboxByID(ctx, id)
	assert.NilError(t, err)
	assert.Equal(t, "failed", row.Status,
		"row should be marked failed after SendTransaction error")
	assert.Assert(t, row.Error.Valid, "failure error column should be populated")
	assert.Assert(t, row.TxHash.Valid,
		"tx_hash should be persisted (mark-before-send wrote it prior to send)")
}

// TestSubmitRowSkipsSendWhenMarkSubmittedFails verifies that if the DB update
// to mark the row submitted fails, SendTransaction is never called. We force
// the failure by dropping the tx_outbox table mid-flight, then driving
// submitRow directly with a hand-constructed row so the absent table is only
// surfaced at the MarkTxSubmitted step (not earlier in GetPendingTxs).
func TestSubmitRowSkipsSendWhenMarkSubmittedFails(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	s, fc := newTestSender(t, dbpool)

	zero, err := bigIntToNumeric(big.NewInt(0))
	assert.NilError(t, err)
	row := corekeyperdb.TxOutbox{
		ID:        1,
		ToAddress: common.HexToAddress("0x000000000000000000000000000000000000dead").Hex(),
		Data:      []byte{0x01, 0x02, 0x03},
		Value:     zero,
		Status:    "pending",
	}

	_, err = dbpool.Exec(ctx, "DROP TABLE tx_outbox")
	assert.NilError(t, err)

	s.submitRow(ctx, row)

	assert.Equal(t, int32(0), atomic.LoadInt32(&fc.sendTransactionCallCount),
		"SendTransaction must not be called when MarkTxSubmitted fails")
}

// TestEnqueueTxPersistsLabel verifies that a label passed to EnqueueTx is
// stored on the row and readable back via the standard query path.
func TestEnqueueTxPersistsLabel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	const wantLabel = "submitDealing ksi=7 retry=2"

	var id int64
	err := dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		var inner error
		id, inner = EnqueueTx(
			ctx,
			tx,
			common.HexToAddress("0x000000000000000000000000000000000000dead"),
			[]byte{0x01, 0x02, 0x03},
			big.NewInt(0),
			wantLabel,
		)
		return inner
	})
	assert.NilError(t, err)

	row, err := corekeyperdb.New(dbpool).GetTxOutboxByID(ctx, id)
	assert.NilError(t, err)
	assert.Equal(t, wantLabel, row.Label)
}

// TestSubmitRowBuildsDynamicFeeTx verifies that submitRow constructs an
// EIP-1559 DynamicFeeTx with GasTipCap from SuggestGasTipCap and
// GasFeeCap = 2 * baseFee + tipCap from the latest header's BaseFee.
func TestSubmitRowBuildsDynamicFeeTx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, corekeyperdb.Definition)
	t.Cleanup(dbclose)

	s, fc := newTestSender(t, dbpool)
	fc.tipCap = big.NewInt(1_500_000_000)
	fc.baseFee = big.NewInt(7_000_000_000)

	queries := corekeyperdb.New(dbpool)
	zero, err := bigIntToNumeric(big.NewInt(0))
	assert.NilError(t, err)
	_, err = queries.InsertPendingTx(ctx, corekeyperdb.InsertPendingTxParams{
		ToAddress: common.HexToAddress("0x000000000000000000000000000000000000dead").Hex(),
		Data:      []byte{0x01, 0x02, 0x03},
		Value:     zero,
	})
	assert.NilError(t, err)

	s.poll(ctx)

	assert.Equal(t, 1, len(fc.sentTxs), "expected exactly one submitted tx")
	tx := fc.sentTxs[0]
	assert.Equal(t, uint8(types.DynamicFeeTxType), tx.Type(),
		"submitted tx should be EIP-1559 DynamicFeeTx")
	assert.Equal(t, fc.tipCap.String(), tx.GasTipCap().String(),
		"GasTipCap should match SuggestGasTipCap")
	// 2*baseFee + tipCap = 2*7e9 + 1.5e9 = 15.5e9
	wantFeeCap := new(big.Int).Mul(fc.baseFee, big.NewInt(2))
	wantFeeCap.Add(wantFeeCap, fc.tipCap)
	assert.Equal(t, wantFeeCap.String(), tx.GasFeeCap().String(),
		"GasFeeCap should equal 2*baseFee + tipCap")
}

type stubError string

func (e stubError) Error() string { return string(e) }

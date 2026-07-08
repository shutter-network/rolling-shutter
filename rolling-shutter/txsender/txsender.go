// Package txsender implements a database-backed transaction outbox sender.
//
// The DKG participation loop (and any other producer) writes a pending row to
// the `tx_outbox` table with ABI-packed calldata, in the same database
// transaction as its business state. TxSender polls those rows, signs and
// submits them to the chain, then watches for receipts and updates the row
// status accordingly.
//
// Status lifecycle: pending -> submitted -> confirmed | failed.
package txsender

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

// Client is the subset of go-ethereum's ethclient.Client functionality that
// TxSender depends on. Kept narrow so tests can swap in a fake.
type Client interface {
	ChainID(ctx context.Context) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

// Config carries the dependencies needed to run TxSender.
type Config struct {
	DBPool       *pgxpool.Pool
	Client       Client
	PrivateKey   *ecdsa.PrivateKey
	PollInterval time.Duration
}

// TxSender is a service that drains the `tx_outbox` table and reports
// receipts back to it. It runs a single poll loop that, on each tick, first
// submits all pending rows in id order, then checks all submitted rows for
// receipts.
type TxSender struct {
	cfg     Config
	address common.Address
	chainID *big.Int
}

// New constructs a TxSender. It does not perform any network calls; chain
// state (chain ID, nonce) is fetched lazily on first send.
func New(cfg Config) *TxSender {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 1 * time.Second
	}
	addr := crypto.PubkeyToAddress(cfg.PrivateKey.PublicKey)
	return &TxSender{cfg: cfg, address: addr}
}

// Start implements service.Service. A single poll loop runs both phases
// sequentially per tick: submit pending rows, then check submitted rows.
func (s *TxSender) Start(ctx context.Context, runner service.Runner) error {
	chainID, err := s.cfg.Client.ChainID(ctx)
	if err != nil {
		return errors.Wrap(err, "tx_outbox sender: read chain id")
	}
	s.chainID = chainID
	log.Info().
		Str("from", s.address.Hex()).
		Str("chain-id", chainID.String()).
		Dur("poll-interval", s.cfg.PollInterval).
		Msg("starting tx outbox sender")

	runner.Go(func() error { return s.pollLoop(ctx) })
	return nil
}

func (s *TxSender) pollLoop(ctx context.Context) error {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.poll(ctx)
		}
	}
}

// poll runs one iteration of the sender: it first submits all pending rows,
// then checks all submitted rows for receipts. The phases are sequential so
// that a row enqueued and submitted within the same tick can also be checked
// for a receipt on the next tick without ordering surprises.
func (s *TxSender) poll(ctx context.Context) {
	if err := s.processPending(ctx); err != nil {
		log.Error().Err(err).Msg("tx outbox: submit phase failed")
	}
	if err := s.processSubmitted(ctx); err != nil {
		log.Error().Err(err).Msg("tx outbox: confirm phase failed")
	}
}

// processPending picks up every pending row in id order and submits it. Rows
// are processed sequentially because nonce ordering matters; if one row's
// submission fails the caller is left to inspect the row's `error` column.
func (s *TxSender) processPending(ctx context.Context) error {
	queries := corekeyperdb.New(s.cfg.DBPool)
	rows, err := queries.GetPendingTxs(ctx)
	if err != nil {
		return errors.Wrap(err, "list pending tx_outbox rows")
	}
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.submitRow(ctx, row)
	}
	return nil
}

// submitRow signs and submits the transaction for a single outbox row. All
// errors are recorded against the row itself; the loop never aborts because
// of one bad row.
func (s *TxSender) submitRow(ctx context.Context, row corekeyperdb.TxOutbox) {
	queries := corekeyperdb.New(s.cfg.DBPool)

	value, err := numericToBigInt(row.Value)
	if err != nil {
		s.markFailed(ctx, row.ID, row.Label, errors.Wrap(err, "decode tx_outbox value"))
		return
	}
	to := common.HexToAddress(row.ToAddress)

	nonce, err := s.cfg.Client.PendingNonceAt(ctx, s.address)
	if err != nil {
		// Transient network error — leave the row pending so the next tick
		// retries.
		log.Warn().Err(err).Int64("id", row.ID).Str("label", row.Label).Msg("tx outbox: read nonce")
		return
	}
	gasLimit, err := s.cfg.Client.EstimateGas(ctx, ethereum.CallMsg{
		From:  s.address,
		To:    &to,
		Value: value,
		Data:  row.Data,
	})
	if err != nil {
		// Gas estimation failures usually mean the transaction would revert;
		// record it as failed so an operator can inspect the calldata.
		s.markFailed(ctx, row.ID, row.Label, errors.Wrap(err, "estimate gas"))
		return
	}
	tipCap, err := s.cfg.Client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Warn().Err(err).Int64("id", row.ID).Str("label", row.Label).Msg("tx outbox: suggest gas tip cap")
		return
	}
	latestHeader, err := s.cfg.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Warn().Err(err).Int64("id", row.ID).Str("label", row.Label).Msg("tx outbox: read latest header")
		return
	}
	if latestHeader.BaseFee == nil {
		s.markFailed(ctx, row.ID, row.Label, errors.New("latest header has no base fee (non-EIP-1559 chain)"))
		return
	}
	// GasFeeCap = 2 * baseFee + tipCap. The 2x multiplier provides headroom
	// for base-fee growth across a few blocks of inclusion delay.
	feeCap := new(big.Int).Mul(latestHeader.BaseFee, big.NewInt(2))
	feeCap.Add(feeCap, tipCap)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   s.chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     value,
		Data:      row.Data,
	})
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(s.chainID), s.cfg.PrivateKey)
	if err != nil {
		s.markFailed(ctx, row.ID, row.Label, errors.Wrap(err, "sign transaction"))
		return
	}

	// Mark the row submitted BEFORE broadcasting. If the process crashes
	// between this update and the SendTransaction call, the row stays
	// `submitted` with its tx hash stored; the confirm loop then polls for the
	// receipt. This is preferable to the inverse ordering (send first, mark
	// after), which would leave a crashed row as `pending` and cause a
	// duplicate broadcast on restart with a fresh nonce.
	if err := queries.MarkTxSubmitted(ctx, corekeyperdb.MarkTxSubmittedParams{
		ID:     row.ID,
		TxHash: sql.NullString{String: signed.Hash().Hex(), Valid: true},
		Nonce:  sql.NullInt64{Int64: int64(nonce), Valid: true}, //nolint:gosec // G115: Ethereum nonces fit well within int64
	}); err != nil {
		log.Error().Err(err).
			Int64("id", row.ID).
			Str("label", row.Label).
			Str("tx-hash", signed.Hash().Hex()).
			Msg("tx outbox: mark submitted failed; skipping send")
		return
	}

	if err := s.cfg.Client.SendTransaction(ctx, signed); err != nil {
		// SendTransaction can fail for transient (RPC) or terminal (already
		// known, replacement underpriced) reasons. Mark the row failed in all
		// cases — MarkTxFailed has no precondition on the current status so
		// the submitted -> failed transition is valid. Re-broadcast logic for
		// stuck submitted rows (without on-chain inclusion) is out of scope.
		s.markFailed(ctx, row.ID, row.Label, errors.Wrap(err, "send transaction"))
		return
	}

	log.Info().
		Int64("id", row.ID).
		Str("label", row.Label).
		Str("to", row.ToAddress).
		Str("tx-hash", signed.Hash().Hex()).
		Uint64("nonce", nonce).
		Msg("tx outbox: submitted")
}

// processSubmitted polls every submitted row for a receipt. Rows without a
// receipt yet are left alone for the next tick.
func (s *TxSender) processSubmitted(ctx context.Context) error {
	queries := corekeyperdb.New(s.cfg.DBPool)
	rows, err := queries.GetSubmittedTxs(ctx)
	if err != nil {
		return errors.Wrap(err, "list submitted tx_outbox rows")
	}
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.checkReceipt(ctx, row)
	}
	return nil
}

func (s *TxSender) checkReceipt(ctx context.Context, row corekeyperdb.TxOutbox) {
	if !row.TxHash.Valid {
		// Submitted without a tx hash is a bug — flag it and stop polling.
		s.markFailed(ctx, row.ID, row.Label, errors.New("submitted row has no tx_hash"))
		return
	}
	queries := corekeyperdb.New(s.cfg.DBPool)
	hash := common.HexToHash(row.TxHash.String)
	receipt, err := s.cfg.Client.TransactionReceipt(ctx, hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return
		}
		log.Warn().Err(err).Int64("id", row.ID).Str("label", row.Label).Str("tx-hash", row.TxHash.String).
			Msg("tx outbox: fetch receipt")
		return
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		if err := queries.MarkTxConfirmed(ctx, row.ID); err != nil {
			log.Error().Err(err).Int64("id", row.ID).Str("label", row.Label).Msg("tx outbox: mark confirmed")
			return
		}
		log.Info().Int64("id", row.ID).Str("label", row.Label).Str("tx-hash", row.TxHash.String).
			Msg("tx outbox: confirmed")
		return
	}

	s.markFailed(ctx, row.ID, row.Label, errors.Errorf("tx receipt status %d", receipt.Status))
}

func (s *TxSender) markFailed(ctx context.Context, id int64, label string, cause error) {
	log.Error().Err(cause).Int64("id", id).Str("label", label).Msg("tx outbox: marking failed")
	queries := corekeyperdb.New(s.cfg.DBPool)
	if err := queries.MarkTxFailed(ctx, corekeyperdb.MarkTxFailedParams{
		ID:    id,
		Error: sql.NullString{String: cause.Error(), Valid: true},
	}); err != nil {
		log.Error().Err(err).Int64("id", id).Str("label", label).Msg("tx outbox: mark failed update")
	}
}

// EnqueueTx inserts a pending row into `tx_outbox`. Producers use this from
// within their own database transactions so the intent to send a transaction
// is committed atomically with the state that motivated it.
//
// The label is an opaque, caller-supplied string. TxSender stores it as-is
// and includes it in every log line for the row so operators can identify
// what a transaction is for without decoding calldata.
func EnqueueTx(
	ctx context.Context,
	tx pgx.Tx,
	to common.Address,
	data []byte,
	value *big.Int,
	label string,
) (int64, error) {
	num, err := bigIntToNumeric(value)
	if err != nil {
		return 0, errors.Wrap(err, "encode value")
	}
	return corekeyperdb.New(tx).InsertPendingTx(ctx, corekeyperdb.InsertPendingTxParams{
		ToAddress: to.Hex(),
		Data:      data,
		Value:     num,
		Label:     label,
	})
}

func bigIntToNumeric(v *big.Int) (pgtype.Numeric, error) {
	if v == nil {
		v = new(big.Int)
	}
	var num pgtype.Numeric
	if err := num.Set(v.String()); err != nil {
		return pgtype.Numeric{}, err
	}
	return num, nil
}

func numericToBigInt(num pgtype.Numeric) (*big.Int, error) {
	if num.Status != pgtype.Present {
		return new(big.Int), nil
	}
	var s string
	if err := num.AssignTo(&s); err != nil {
		return nil, err
	}
	bi, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, errors.Errorf("not a base-10 integer: %q", s)
	}
	return bi, nil
}

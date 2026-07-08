package dkg

import (
	"context"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/contracts/v2/bindings/ecieskeyregistry"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/txsender"
)

// MaybeRegisterECIESKey enqueues a `registerKey` outbox row for the ECIES
// key registry if we are a member of the given keyper set and our public
// key is not yet present in `ecies_keys`. The DB cache is populated by the
// host keyper's ECIES syncer (both initial poll and live events) and is the
// source of truth for prior registrations.
//
// Per ADR 0004 this writes to `tx_outbox` instead of calling the registry
// directly; `TxSender` handles signing and submission. The check is
// idempotent — once the registry event is indexed back into `ecies_keys`,
// subsequent calls become no-ops.
//
// The destination address is the manager's configured `ECIESRegistryAddr`
// (a single registry serves all keyper sets). This is intended to be called
// by the host keyper's Keyper Set Syncer handler once per discovered keyper
// set, at the same architectural level as `HandleBlock`.
func (m *Manager) MaybeRegisterECIESKey(ctx context.Context, keyperSetIndex int64) error {
	return m.cfg.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		obsQueries := obskeyper.New(tx)
		coreQueries := corekeyperdb.New(tx)

		keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, keyperSetIndex)
		if err != nil {
			return errors.Wrapf(err, "fetch keyper set %d", keyperSetIndex)
		}
		ownIndex, err := keyperSet.GetIndex(m.cfg.OwnAddress)
		if err != nil {
			// Not a member.
			return nil
		}

		exists, err := coreQueries.ExistsECIESKey(ctx, shdb.EncodeAddress(m.cfg.OwnAddress))
		if err != nil {
			return errors.Wrap(err, "query ecies_keys for own address")
		}
		if exists {
			log.Debug().
				Int64("keyper-set-index", keyperSetIndex).
				Uint64("keyper-index", ownIndex).
				Msg("ECIES key already registered for keyper set, skipping")
			return nil
		}

		// `crypto/ecies.PrivateKey` embeds an ECDSA public key on its
		// `PublicKey.ExportECDSA()` accessor. We need raw secp256k1 public
		// key bytes (the on-chain format) for the registry.
		ecdsaPub := m.cfg.ECIESPrivateKey.ExportECDSA().PublicKey
		pubKey := ethcrypto.FromECDSAPub(&ecdsaPub)

		abi, err := ecieskeyregistry.EcieskeyregistryMetaData.GetAbi()
		if err != nil {
			return errors.Wrap(err, "load ECIESKeyRegistry ABI")
		}
		data, err := abi.Pack(
			"registerKey",
			uint64(keyperSetIndex), //nolint:gosec // G115: keyper set index is bounded by the on-chain contract
			ownIndex,
			pubKey,
		)
		if err != nil {
			return errors.Wrap(err, "pack registerKey calldata")
		}
		label := fmt.Sprintf("registerKey ksi=%d", keyperSetIndex)
		outboxID, err := txsender.EnqueueTx(ctx, tx, m.cfg.ECIESRegistryAddr, data, nil, label)
		if err != nil {
			return errors.Wrap(err, "enqueue registerKey tx")
		}
		log.Info().
			Int64("keyper-set-index", keyperSetIndex).
			Uint64("keyper-index", ownIndex).
			Int64("tx-outbox-id", outboxID).
			Msg("enqueued ECIES key registration")
		return nil
	})
}

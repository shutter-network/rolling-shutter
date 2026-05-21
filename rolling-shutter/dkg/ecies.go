package dkg

import (
	"context"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/txsender"
)

// maybeRegisterECIESKey enqueues a `registerKey` outbox row for the ECIES
// key registry if we are a member of the given eon's keyper set and our
// public key is not yet present in `ecies_keys`. The DB cache is populated
// by the host keyper's ECIES syncer (both initial poll and live events) and
// is the source of truth for prior registrations.
//
// Per ADR 0004 this writes to `tx_outbox` instead of calling the registry
// directly; `TxSender` handles signing and submission. The check is
// idempotent — once the registry event is indexed back into `ecies_keys`,
// subsequent calls become no-ops.
//
// The destination address is the manager's configured `ECIESRegistryAddr`
// (a single registry serves all keyper sets).
func (m *Manager) maybeRegisterECIESKey(ctx context.Context, eon corekeyperdb.Eon) error {
	// Resolve membership: pull the keyper set, locate own index. A
	// non-member keyper has nothing to register for this eon.
	return m.cfg.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		obsQueries := obskeyper.New(tx)
		coreQueries := corekeyperdb.New(tx)

		keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, eon.KeyperConfigIndex)
		if err != nil {
			return errors.Wrapf(err, "fetch keyper set %d", eon.KeyperConfigIndex)
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
			return nil
		}

		// `crypto/ecies.PrivateKey` embeds an ECDSA public key on its
		// `PublicKey.ExportECDSA()` accessor. We need raw secp256k1 public
		// key bytes (the on-chain format) for the registry.
		ecdsaPub := m.cfg.ECIESPrivateKey.ExportECDSA().PublicKey
		pubKey := ethcrypto.FromECDSAPub(&ecdsaPub)

		abi, err := contract.ECIESKeyRegistryMetaData.GetAbi()
		if err != nil {
			return errors.Wrap(err, "load ECIESKeyRegistry ABI")
		}
		data, err := abi.Pack(
			"registerKey",
			uint64(eon.KeyperConfigIndex),
			ownIndex,
			pubKey,
		)
		if err != nil {
			return errors.Wrap(err, "pack registerKey calldata")
		}
		label := fmt.Sprintf("registerKey ksi=%d", eon.KeyperConfigIndex)
		outboxID, err := txsender.EnqueueTx(ctx, tx, m.cfg.ECIESRegistryAddr, data, nil, label)
		if err != nil {
			return errors.Wrap(err, "enqueue registerKey tx")
		}
		log.Info().
			Int64("keyper-config-index", eon.KeyperConfigIndex).
			Uint64("keyper-index", ownIndex).
			Int64("tx-outbox-id", outboxID).
			Msg("enqueued ECIES key registration")
		return nil
	})
}

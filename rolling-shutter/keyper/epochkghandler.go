package keyper

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type epochKGHandler struct {
	config Config
	db     *kprdb.Queries
}

func (h *epochKGHandler) handleDecryptionTrigger(
	ctx context.Context, msg *p2pmsg.DecryptionTrigger,
) ([]p2pmsg.P2PMessage, error) {
	log.Info().Str("message", msg.String()).Msg("received decryption trigger")
	epochID, err := epochid.BytesToEpochID(msg.EpochID)
	if err != nil {
		return nil, err
	}
	return h.sendDecryptionKeyShare(ctx, epochID, int64(msg.BlockNumber))
}

func (h *epochKGHandler) sendDecryptionKeyShare(
	ctx context.Context, epochID epochid.EpochID, blockNumber int64,
) ([]p2pmsg.P2PMessage, error) {
	eon, err := h.db.GetEonForBlockNumber(ctx, blockNumber)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get eon for block %d from db", blockNumber)
	}
	batchConfig, err := h.db.GetBatchConfig(ctx, int32(eon.KeyperConfigIndex))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config %d from db", eon.KeyperConfigIndex)
	}

	// get our keyper index (and check that we in fact are a keyper)
	encodedAddress := shdb.EncodeAddress(h.config.Address())
	keyperIndex := int64(-1)
	for i, address := range batchConfig.Keypers {
		if address == encodedAddress {
			keyperIndex = int64(i)
			break
		}
	}
	if keyperIndex == -1 {
		log.Info().Str("epoch-id", epochID.Hex()).Msg("ignoring decryption trigger: we are not a keyper")
		return nil, nil
	}

	// check if we already computed (and therefore most likely sent) our key share
	shareExists, err := h.db.ExistsDecryptionKeyShare(ctx, kprdb.ExistsDecryptionKeyShareParams{
		EpochID:     epochID.Bytes(),
		KeyperIndex: keyperIndex,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get decryption key share for epoch %s from db", epochID)
	}
	if shareExists {
		return nil, nil // we already sent our share
	}

	// fetch dkg result from db
	dkgResultDB, err := h.db.GetDKGResult(ctx, eon.Eon)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dkg result for eon %d from db", eon.Eon)
	}
	if !dkgResultDB.Success {
		log.Info().Int64("eon", eon.Eon).Msg("ignoring decryption trigger: eon key generation failed")
		return nil, nil
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return nil, err
	}

	// compute the key share
	epochKG := epochkg.NewEpochKG(pureDKGResult)
	share := epochKG.ComputeEpochSecretKeyShare(epochID)

	msg := &p2pmsg.DecryptionKeyShare{
		InstanceID:  h.config.InstanceID,
		Eon:         uint64(eon.Eon),
		EpochID:     epochID.Bytes(),
		KeyperIndex: uint64(keyperIndex),
		Share:       share.Marshal(),
	}
	err = h.db.InsertDecryptionKeyShareMsg(ctx, msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert decryption key share")
	}
	log.Info().Str("epoch-id", epochID.Hex()).Int64("block-number", blockNumber).
		Msg("sending decryption key share")
	return []p2pmsg.P2PMessage{msg}, nil
}

func (h *epochKGHandler) aggregateDecryptionKeySharesFromDB(
	ctx context.Context,
	pureDKGResult *puredkg.Result,
	epochID epochid.EpochID,
) (*epochkg.EpochKG, error) {
	shares, err := h.db.SelectDecryptionKeyShares(ctx, kprdb.SelectDecryptionKeySharesParams{
		Eon:     int64(pureDKGResult.Eon),
		EpochID: epochID.Bytes(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get decryption key shares for epoch %s from db", epochID)
	}

	epochKG := epochkg.NewEpochKG(pureDKGResult)
	// For simplicity, we aggregate shares even if we don't have enough of them yet.
	for _, share := range shares {
		epochID, err := epochid.BytesToEpochID(share.EpochID)
		if err != nil {
			return nil, errors.Wrap(err, "invalid epoch id in db")
		}
		shareDecoded, err := shdb.DecodeEpochSecretKeyShare(share.DecryptionKeyShare)
		if err != nil {
			log.Warn().Str("epoch-id", epochID.Hex()).Int64("keyper-index", share.KeyperIndex).
				Msg("invalid decryption key share in DB")
			continue
		}
		err = epochKG.HandleEpochSecretKeyShare(&epochkg.EpochSecretKeyShare{
			Eon:    pureDKGResult.Eon,
			Epoch:  epochID,
			Sender: uint64(share.KeyperIndex),
			Share:  shareDecoded,
		})
		if err != nil {
			log.Info().Str("epoch-id", epochID.Hex()).Int64("keyper-index", share.KeyperIndex).
				Msg("failed to process decryption key share")
			continue
		}
	}

	return epochKG, nil
}

func (h *epochKGHandler) handleDecryptionKeyShare(ctx context.Context, msg *p2pmsg.DecryptionKeyShare) ([]p2pmsg.P2PMessage, error) {
	// Insert the share into the db. We assume that it's valid as it already passed the libp2p
	// validator.
	if err := h.db.InsertDecryptionKeyShareMsg(ctx, msg); err != nil {
		return nil, err
	}

	// Check that we don't know the decryption key yet
	epochID, err := epochid.BytesToEpochID(msg.EpochID)
	if err != nil {
		return nil, err
	}
	keyExists, err := h.db.ExistsDecryptionKey(ctx, kprdb.ExistsDecryptionKeyParams{
		Eon:     int64(msg.Eon),
		EpochID: epochID.Bytes(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query decryption key for epoch %s", epochID)
	}
	if keyExists {
		return nil, nil
	}

	// fetch dkg result from db
	dkgResultDB, err := h.db.GetDKGResult(ctx, int64(msg.Eon))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dkg result for eon %d from db", msg.Eon)
	}
	if !dkgResultDB.Success {
		log.Info().Uint64("eon", msg.Eon).
			Msg("ignoring decryption trigger: eon key generation failed")
		return nil, nil
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return nil, err
	}

	// aggregate epoch secret key
	epochKG, err := h.aggregateDecryptionKeySharesFromDB(ctx, pureDKGResult, epochID)
	if err != nil {
		return nil, err
	}
	decryptionKey, ok := epochKG.SecretKeys[epochID]
	if !ok {
		numShares := uint64(len(epochKG.SecretShares))
		if numShares < pureDKGResult.Threshold {
			// not enough shares yet
			return nil, nil
		}
		return nil, errors.Errorf(
			"failed to generate decryption key for epoch %s even though we have enough shares",
			epochID,
		)
	}
	message := &p2pmsg.DecryptionKey{
		InstanceID: h.config.InstanceID,
		Eon:        msg.Eon,
		EpochID:    epochID.Bytes(),
		Key:        decryptionKey.Marshal(),
	}
	err = h.db.InsertDecryptionKeyMsg(ctx, message)
	if err != nil {
		return nil, err
	}
	log.Info().Str("epoch-id", epochID.Hex()).Str("message", message.String()).
		Msg("broadcasting decryption key")
	return []p2pmsg.P2PMessage{message}, nil
}

func (h *epochKGHandler) handleDecryptionKey(ctx context.Context, msg *p2pmsg.DecryptionKey) ([]p2pmsg.P2PMessage, error) {
	// Insert the key into the db. We assume that it's valid as it already passed the libp2p
	// validator.
	return nil, h.db.InsertDecryptionKeyMsg(ctx, msg)
}

package keyper

import (
	"context"
	"log"

	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shlib/puredkg"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type epochKGHandler struct {
	config Config
	db     *kprdb.Queries
}

func (h *epochKGHandler) handleDecryptionTrigger(ctx context.Context, msg *shmsg.DecryptionTrigger) ([]shmsg.P2PMessage, error) {
	log.Printf("received decryption trigger for epoch %d, sending decryption key share now.", msg.EpochID)
	epochID, err := epochid.BytesToEpochID(msg.EpochID)
	if err != nil {
		return nil, err
	}
	return h.sendDecryptionKeyShare(ctx, epochID, int64(msg.BlockNumber))
}

func (h *epochKGHandler) sendDecryptionKeyShare(ctx context.Context, epochID epochid.EpochID, blockNumber int64) ([]shmsg.P2PMessage, error) {
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
		log.Printf("ignoring decryption trigger for epoch %s as we are not a keyper", epochID)
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
		log.Printf("ignoring decryption trigger for eon %d as eon key generation failed", eon.Eon)
		return nil, nil
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return nil, err
	}

	// compute the key share
	epochKG := epochkg.NewEpochKG(pureDKGResult)
	share := epochKG.ComputeEpochSecretKeyShare(epochID)
	encodedShare := share.Marshal()
	if err != nil {
		return nil, err
	}

	// store share in db and sent it
	err = h.db.InsertDecryptionKeyShare(ctx, kprdb.InsertDecryptionKeyShareParams{
		Eon:                eon.Eon,
		EpochID:            epochID.Bytes(),
		KeyperIndex:        keyperIndex,
		DecryptionKeyShare: encodedShare,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert decryption key share")
	}
	log.Printf("sending decryption key share for epoch %s", epochID)
	return []shmsg.P2PMessage{
		&shmsg.DecryptionKeyShare{
			InstanceID:  h.config.InstanceID,
			Eon:         uint64(eon.Eon),
			EpochID:     epochID.Bytes(),
			KeyperIndex: uint64(keyperIndex),
			Share:       encodedShare,
		},
	}, nil
}

func (h *epochKGHandler) insertDecryptionKeyShare(ctx context.Context, msg *shmsg.DecryptionKeyShare) error {
	err := h.db.InsertDecryptionKeyShare(ctx, kprdb.InsertDecryptionKeyShareParams{
		Eon:                int64(msg.Eon),
		EpochID:            msg.EpochID,
		KeyperIndex:        int64(msg.KeyperIndex),
		DecryptionKeyShare: msg.Share,
	})
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to insert decryption key share for epoch %d from keyper %d",
			msg.EpochID,
			msg.KeyperIndex,
		)
	}
	return nil
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
			log.Printf(
				"Warning: invalid decryption key share in db for epoch %s and keyper %d",
				epochID,
				share.KeyperIndex,
			)
			continue
		}
		err = epochKG.HandleEpochSecretKeyShare(&epochkg.EpochSecretKeyShare{
			Eon:    pureDKGResult.Eon,
			Epoch:  epochID.Bytes(),
			Sender: uint64(share.KeyperIndex),
			Share:  shareDecoded,
		})
		if err != nil {
			log.Printf(
				"error processing decryption key share for epoch %s of keyper %d: %s",
				epochID, share.KeyperIndex, err,
			)
			continue
		}
	}

	return epochKG, nil
}

func (h *epochKGHandler) handleDecryptionKeyShare(ctx context.Context, msg *shmsg.DecryptionKeyShare) ([]shmsg.P2PMessage, error) {
	// Insert the share into the db. We assume that it's valid as it already passed the libp2p
	// validator.
	if err := h.insertDecryptionKeyShare(ctx, msg); err != nil {
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
		log.Printf("ignoring decryption trigger for eon %d as eon key generation failed", msg.Eon)
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
	decryptionKeyEncoded := decryptionKey.Marshal()

	// send decryption key
	tag, err := h.db.InsertDecryptionKey(ctx, kprdb.InsertDecryptionKeyParams{
		Eon:           int64(msg.Eon),
		EpochID:       epochID.Bytes(),
		DecryptionKey: decryptionKeyEncoded,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to store decryption key for epoch %s in db", epochID)
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to insert decryption key for epoch %s, but it already exists", epochID)
		return nil, nil
	}
	log.Printf("broadcasting decryption key for epoch %s", epochID)
	return []shmsg.P2PMessage{
		&shmsg.DecryptionKey{
			InstanceID: h.config.InstanceID,
			Eon:        msg.Eon,
			EpochID:    epochID.Bytes(),
			Key:        decryptionKeyEncoded,
		},
	}, nil
}

func (h *epochKGHandler) handleDecryptionKey(ctx context.Context, msg *shmsg.DecryptionKey) ([]shmsg.P2PMessage, error) {
	// Insert the key into the db. We assume that it's valid as it already passed the libp2p
	// validator.
	epochID, err := epochid.BytesToEpochID(msg.EpochID)
	if err != nil {
		return nil, err
	}
	tag, err := h.db.InsertDecryptionKey(ctx, kprdb.InsertDecryptionKeyParams{
		Eon:           int64(msg.Eon),
		EpochID:       epochID.Bytes(),
		DecryptionKey: msg.Key,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to insert decryption key for epoch %s", epochID)
	}
	if tag.RowsAffected() == 0 {
		log.Printf("attempted to insert decryption key for epoch %s, but it already exists", epochID)
	}
	return nil, nil
}

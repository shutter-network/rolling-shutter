package keyper

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func (kpr *keyper) validateDecryptionKey(ctx context.Context, key *shmsg.DecryptionKey) (bool, error) {
	if key.GetInstanceID() != kpr.config.InstanceID {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", kpr.config.InstanceID, key.GetInstanceID())
	}

	activationBlockNumber := epochid.BlockNumber(key.EpochID)
	dkgResultDB, err := kpr.db.GetDKGResultForBlockNumber(ctx, int64(activationBlockNumber))
	if err == pgx.ErrNoRows {
		return false, errors.Errorf("no DKG result found for epoch %d", key.EpochID)
	}
	if err != nil {
		return false, errors.Wrapf(err, "failed to get dkg result for epoch %d from db", key.EpochID)
	}
	if !dkgResultDB.Success {
		return false, errors.Errorf("no successful DKG result found for epoch %d", key.EpochID)
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return false, errors.Wrapf(err, "error while decoding pure DKG result for epoch %d", key.EpochID)
	}
	epochSecretKey, err := key.GetEpochSecretKey()
	if err != nil {
		return false, err
	}

	ok, err := shcrypto.VerifyEpochSecretKey(epochSecretKey, pureDKGResult.PublicKey, key.EpochID)
	if err != nil {
		return false, errors.Wrapf(err, "error while checking epoch secret key for epoch %v", key.EpochID)
	}
	return ok, nil
}

func (kpr *keyper) validateDecryptionKeyShare(ctx context.Context, keyShare *shmsg.DecryptionKeyShare) (bool, error) {
	if keyShare.GetInstanceID() != kpr.config.InstanceID {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", kpr.config.InstanceID, keyShare.GetInstanceID())
	}

	activationBlockNumber := epochid.BlockNumber(keyShare.EpochID)
	dkgResultDB, err := kpr.db.GetDKGResultForBlockNumber(ctx, int64(activationBlockNumber))
	if err == pgx.ErrNoRows {
		return false, errors.Errorf("no DKG result found for epoch %d", keyShare.EpochID)
	}
	if err != nil {
		return false, errors.Errorf("failed to get dkg result for epoch %d from db", keyShare.EpochID)
	}
	if !dkgResultDB.Success {
		return false, errors.Errorf("no successful DKG result found for epoch %d", keyShare.EpochID)
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return false, errors.Errorf("error while decoding pure DKG result for epoch %d", keyShare.EpochID)
	}
	epochSecretKeyShare, err := keyShare.GetEpochSecretKeyShare()
	if err != nil {
		return false, err
	}
	return shcrypto.VerifyEpochSecretKeyShare(
		epochSecretKeyShare,
		pureDKGResult.PublicKeyShares[keyShare.KeyperIndex],
		shcrypto.ComputeEpochID(keyShare.EpochID),
	), nil
}

func (kpr *keyper) validateEonPublicKey(_ context.Context, key *shmsg.EonPublicKey) (bool, error) {
	if key.GetInstanceID() != kpr.config.InstanceID {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", kpr.config.InstanceID, key.GetInstanceID())
	}
	return true, nil
}

func (kpr *keyper) validateDecryptionTrigger(ctx context.Context, trigger *shmsg.DecryptionTrigger) (bool, error) {
	if trigger.GetInstanceID() != kpr.config.InstanceID {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", kpr.config.InstanceID, trigger.GetInstanceID())
	}

	blk := epochid.BlockNumber(trigger.EpochID)
	chainCollator, err := kpr.db.GetChainCollator(ctx, int64(blk))
	if err == pgx.ErrNoRows {
		return false, errors.Errorf("got decryption trigger with no collator for given block number: %d", blk)
	}
	if err != nil {
		return false, errors.Wrapf(err, "error while getting collator from db for block number: %d", blk)
	}

	collator, err := shdb.DecodeAddress(chainCollator.Collator)
	if err != nil {
		return false, errors.Wrapf(err, "error while converting collator from string to address: %s", chainCollator.Collator)
	}

	signatureValid, err := trigger.VerifySignature(collator)
	if err != nil {
		return false, errors.Wrapf(err, "error while verifying decryption trigger signature for epoch: %d", trigger.EpochID)
	}
	return signatureValid, nil
}

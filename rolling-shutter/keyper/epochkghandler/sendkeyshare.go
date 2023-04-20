package epochkghandler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type Config interface {
	GetAddress() common.Address
	GetInstanceID() uint64
}

func SendDecryptionKeyShare(
	ctx context.Context,
	config Config,
	db *kprdb.Queries,
	epochID epochid.EpochID,
	blockNumber int64,
) ([]p2pmsg.Message, error) {
	eon, err := db.GetEonForBlockNumber(ctx, blockNumber)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get eon for block %d from db", blockNumber)
	}
	batchConfig, err := db.GetBatchConfig(ctx, int32(eon.KeyperConfigIndex))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config %d from db", eon.KeyperConfigIndex)
	}

	// get our keyper index (and check that we in fact are a keyper)
	encodedAddress := shdb.EncodeAddress(config.GetAddress())
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
	shareExists, err := db.ExistsDecryptionKeyShare(ctx, kprdb.ExistsDecryptionKeyShareParams{
		Eon:         eon.Eon,
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
	dkgResultDB, err := db.GetDKGResult(ctx, eon.Eon)
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
		InstanceID:  config.GetInstanceID(),
		Eon:         uint64(eon.Eon),
		EpochID:     epochID.Bytes(),
		KeyperIndex: uint64(keyperIndex),
		Share:       share.Marshal(),
	}
	err = db.InsertDecryptionKeyShareMsg(ctx, msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert decryption key share")
	}
	log.Info().Str("epoch-id", epochID.Hex()).Int64("block-number", blockNumber).
		Msg("sending decryption key share")
	return []p2pmsg.Message{msg}, nil
}

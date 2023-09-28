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

type DecryptionTrigger struct {
	BlockNumber uint64
	EpochIDs    []epochid.EpochID
}

func ConstructDecryptionKeyShare(
	ctx context.Context,
	config Config,
	db *kprdb.Queries,
	eon kprdb.Eon,
	epochIDs []epochid.EpochID,
) (*p2pmsg.DecryptionKeyShares, error) {
	if len(epochIDs) == 0 {
		return nil, errors.New("cannot generate empty decryption key share")
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
		log.Info().Msg("ignoring decryption trigger: we are not a keyper")
		return nil, nil
	}

	// check if we already computed (and therefore most likely sent) our key share
	// XXX this only works when we sent the share for exactly one epoch.
	shareExists, err := db.ExistsDecryptionKeyShare(ctx, kprdb.ExistsDecryptionKeyShareParams{
		Eon:         eon.Eon,
		EpochID:     epochIDs[0].Bytes(),
		KeyperIndex: keyperIndex,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get decryption key share for epoch from db")
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

	var shares []*p2pmsg.KeyShare
	// compute the key share
	epochKG := epochkg.NewEpochKG(pureDKGResult)

	for _, epochID := range epochIDs {
		share := epochKG.ComputeEpochSecretKeyShare(epochID)

		shares = append(shares, &p2pmsg.KeyShare{
			EpochID: epochID.Bytes(),
			Share:   share.Marshal(),
		})
	}

	// TODO safe typecase
	msg := &p2pmsg.DecryptionKeyShares{
		InstanceID:  config.GetInstanceID(),
		Eon:         uint64(eon.Eon),
		KeyperIndex: uint64(keyperIndex),
		Shares:      shares,
	}
	err = db.InsertDecryptionKeySharesMsg(ctx, msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert decryption key share")
	}
	metricsEpochKGDecryptionKeySharesSent.Inc()
	return msg, nil
}

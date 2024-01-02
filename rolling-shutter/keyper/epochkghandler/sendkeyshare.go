package epochkghandler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type Config interface {
	GetAddress() common.Address
	GetInstanceID() uint64
}

type DecryptionTrigger struct {
	BlockNumber       uint64
	IdentityPreimages []identitypreimage.IdentityPreimage
}

func (ksh *KeyShareHandler) getEonForBlockNumber(ctx context.Context, blockNumber uint64) (database.Eon, error) {
	var (
		eon database.Eon
		err error
	)
	db := database.New(ksh.DBPool)
	block, err := medley.Uint64ToInt64Safe(blockNumber)
	if err != nil {
		return eon, errors.Wrap(err, "invalid blocknumber")
	}
	eon, err = db.GetEonForBlockNumber(ctx, block)
	return eon, errors.Wrap(err, "failed to retrieve eon from db")
}

func (ksh *KeyShareHandler) ConstructDecryptionKeyShare(
	ctx context.Context,
	eon database.Eon,
	identityPreimages []identitypreimage.IdentityPreimage,
) (*p2pmsg.DecryptionKeyShares, error) {
	if len(identityPreimages) == 0 {
		return nil, errors.New("cannot generate empty decryption key share")
	}
	db := database.New(ksh.DBPool)
	batchConfig, err := db.GetBatchConfig(ctx, int32(eon.KeyperConfigIndex))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config %d from db", eon.KeyperConfigIndex)
	}

	// get our keyper index (and check that we in fact are a keyper)
	encodedAddress := shdb.EncodeAddress(ksh.KeyperAddress)
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
	shareExists, err := db.ExistsDecryptionKeyShare(ctx, database.ExistsDecryptionKeyShareParams{
		Eon:         eon.Eon,
		EpochID:     identityPreimages[0].Bytes(),
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

	for _, identityPreimage := range identityPreimages {
		share := epochKG.ComputeEpochSecretKeyShare(identityPreimage)

		shares = append(shares, &p2pmsg.KeyShare{
			EpochID: identityPreimage.Bytes(),
			Share:   share.Marshal(),
		})
	}

	eonUint, err := medley.Int64ToUint64Safe(eon.Eon)
	if err != nil {
		return nil, err
	}
	keyperIndexUint, err := medley.Int64ToUint64Safe(keyperIndex)
	if err != nil {
		return nil, err
	}
	msg := &p2pmsg.DecryptionKeyShares{
		InstanceID:  ksh.InstanceID,
		Eon:         eonUint,
		KeyperIndex: keyperIndexUint,
		Shares:      shares,
	}
	err = db.InsertDecryptionKeySharesMsg(ctx, msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert decryption key share")
	}
	return msg, nil
}

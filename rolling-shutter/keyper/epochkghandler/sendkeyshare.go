package epochkghandler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

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

var (
	ErrIgnoreDecryptionRequest = errors.New("ignoring decryption request")
	ErrNotAKeyper              = errors.New("we are not a keyper")
	ErrEonDKGFailed            = errors.New("eon key generation failed")
	ErrSharesAlreadySent       = errors.New("shares exist already")
)

//nolint:gocyclo
func (ksh *KeyShareHandler) ConstructDecryptionKeyShares(
	ctx context.Context,
	eon database.Eon,
	identityPreimages []identitypreimage.IdentityPreimage,
) (*p2pmsg.DecryptionKeyShares, error) {
	if len(identityPreimages) == 0 {
		return nil, errors.New("cannot generate empty decryption key share")
	}
	if len(identityPreimages) > MaxNumKeysPerMessage {
		return nil, errors.Errorf("too many decryption key shares for message (%d > %d)", len(identityPreimages), MaxNumKeysPerMessage)
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
		return nil, errors.Wrap(ErrNotAKeyper, ErrIgnoreDecryptionRequest.Error())
	}

	// check if we already computed (and therefore most likely sent) our key shares
	allSharesExist := true
	for _, identityPreimage := range identityPreimages {
		shareExists, err := db.ExistsDecryptionKeyShare(ctx, database.ExistsDecryptionKeyShareParams{
			Eon:         eon.Eon,
			EpochID:     identityPreimage.Bytes(),
			KeyperIndex: keyperIndex,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to get decryption key share for epoch from db")
		}
		if !shareExists {
			allSharesExist = false
			break
		}
	}
	if allSharesExist {
		// we already sent our shares
		return nil, errors.Wrap(ErrSharesAlreadySent, ErrIgnoreDecryptionRequest.Error())
	}

	// fetch dkg result from db
	dkgResultDB, err := db.GetDKGResult(ctx, eon.Eon)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dkg result for eon %d from db", eon.Eon)
	}
	if !dkgResultDB.Success {
		return nil, errors.Wrap(ErrEonDKGFailed, ErrIgnoreDecryptionRequest.Error())
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return nil, err
	}

	var shares []*p2pmsg.KeyShare
	// compute the key shares
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

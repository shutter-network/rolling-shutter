package keyper

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shuttermint/keyper/epochkg"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

type kgstate struct {
	config Config
	db     *kprdb.Queries
}

func (s *kgstate) handleDecryptionTrigger(ctx context.Context, msg *decryptionTrigger) ([]shmsg.P2PMessage, error) {
	eon, err := s.db.GetEonForEpoch(ctx, shdb.EncodeUint64(msg.EpochID))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get eon for epoch %d from db", msg.EpochID)
	}
	batchConfig, err := s.db.GetBatchConfig(ctx, int32(eon.ConfigIndex))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config %d from db", eon.ConfigIndex)
	}

	// get our keyper index (and check that we in fact are a keyper)
	encodedAddress := shdb.EncodeAddress(s.config.Address())
	keyperIndex := int64(-1)
	for i, address := range batchConfig.Keypers {
		if address == encodedAddress {
			keyperIndex = int64(i)
			break
		}
	}
	if keyperIndex == -1 {
		log.Printf("ignoring decryption trigger for epoch %d as we are not a keyper", msg.EpochID)
		return nil, nil
	}

	// check if we already computed (and therefore most likely sent) our key share
	_, err = s.db.GetDecryptionKeyShare(ctx, kprdb.GetDecryptionKeyShareParams{
		EpochID:     shdb.EncodeUint64(msg.EpochID),
		KeyperIndex: keyperIndex,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, errors.Wrapf(err, "failed to get decryption key share for epoch %d from db", msg.EpochID)
	}
	if err != pgx.ErrNoRows {
		return nil, nil // we already sent our share
	}

	// fetch dkg result from db
	dkgResultDB, err := s.db.GetDKGResult(ctx, eon.Eon)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dkg result for eon %d from db", eon.Eon)
	}
	if !dkgResultDB.Success {
		log.Printf("ignoring decryption trigger for epoch %d as eon key generation failed", msg.EpochID)
		return nil, nil
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return nil, err
	}

	// compute the key share
	epochKG := epochkg.NewEpochKG(pureDKGResult)
	share := epochKG.ComputeEpochSecretKeyShare(msg.EpochID)
	encodedShare, err := share.GobEncode()
	if err != nil {
		return nil, err
	}

	// store share in db and sent it
	err = s.db.InsertDecryptionKeyShare(ctx, kprdb.InsertDecryptionKeyShareParams{
		EpochID:            shdb.EncodeUint64(msg.EpochID),
		KeyperIndex:        keyperIndex,
		DecryptionKeyShare: encodedShare,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert decryption key share")
	}
	log.Printf("sending decryption key share for epoch %d", msg.EpochID)
	return []shmsg.P2PMessage{
		&shmsg.DecryptionKeyShare{
			InstanceID:  s.config.InstanceID,
			EpochID:     msg.EpochID,
			KeyperIndex: uint64(keyperIndex),
			Share:       encodedShare,
		},
	}, nil
}

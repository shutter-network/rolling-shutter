package database

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// GetKeyperIndex returns the index of the keyper with the given address in the
// keyper set with the given keyperConfigIndex (which equals the eon for the
// keyper set). The keyper set is read from the chainobserver tables.
func (q *Queries) GetKeyperIndex(ctx context.Context, keyperConfigIndex int64, addr common.Address) (int64, bool, error) {
	ks, err := obskeyper.New(q.db).GetKeyperSetByKeyperConfigIndex(ctx, keyperConfigIndex)
	if errors.Is(err, pgx.ErrNoRows) {
		return -1, false, nil
	}
	if err != nil {
		return -1, false, errors.Wrapf(err, "failed to get keyper set %d from db", keyperConfigIndex)
	}
	encodedAddress := shdb.EncodeAddress(addr)
	for i, address := range ks.Keypers {
		if address == encodedAddress {
			return int64(i), true, nil
		}
	}
	return -1, false, nil
}

func (q *Queries) InsertDecryptionKeysMsg(ctx context.Context, msg *p2pmsg.DecryptionKeys) error {
	for _, key := range msg.Keys {
		identityPreimage := identitypreimage.IdentityPreimage(key.IdentityPreimage)
		tag, err := q.InsertDecryptionKey(ctx, InsertDecryptionKeyParams{
			Eon:           int64(msg.Eon),
			EpochID:       identityPreimage.Bytes(),
			DecryptionKey: key.Key,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to insert decryption key for identity %s", identityPreimage)
		}
		if tag.RowsAffected() == 0 {
			log.Debug().Str("identity", identityPreimage.Hex()).
				Msg("attempted to insert decryption key in db, but it already exists")
		}
	}
	return nil
}

func (q *Queries) InsertDecryptionKeySharesMsg(ctx context.Context, msg *p2pmsg.DecryptionKeyShares) error {
	for _, share := range msg.GetShares() {
		err := q.InsertDecryptionKeyShare(ctx, InsertDecryptionKeyShareParams{
			Eon:                int64(msg.Eon),
			EpochID:            share.IdentityPreimage,
			KeyperIndex:        int64(msg.KeyperIndex),
			DecryptionKeyShare: share.Share,
		})
		if err != nil {
			return errors.Wrapf(
				err,
				"failed to insert decryption key share from keyper %d",
				msg.KeyperIndex,
			)
		}
	}
	return nil
}

package kprdb

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func GetKeyperIndex(addr common.Address, keypers []string) (uint64, bool) {
	hexaddr := shdb.EncodeAddress(addr)
	for i, a := range keypers {
		if a == hexaddr {
			return uint64(i), true
		}
	}
	return uint64(0), false
}

func (bc *TendermintBatchConfig) KeyperIndex(addr common.Address) (uint64, bool) {
	return GetKeyperIndex(addr, bc.Keypers)
}

func (q *Queries) InsertDecryptionKeyMsg(ctx context.Context, msg *p2pmsg.DecryptionKey) error {
	epochID, err := epochid.BytesToEpochID(msg.EpochID)
	if err != nil {
		return err
	}
	tag, err := q.InsertDecryptionKey(ctx, InsertDecryptionKeyParams{
		Eon:           int64(msg.Eon),
		EpochID:       epochID.Bytes(),
		DecryptionKey: msg.Key,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to insert decryption key for epoch %s", epochID)
	}
	if tag.RowsAffected() == 0 {
		log.Info().Str("epoch-id", epochID.Hex()).
			Msg("attempted to insert decryption key in db, but it already exists")
	}
	return nil
}

func (q *Queries) InsertDecryptionKeyShareMsg(ctx context.Context, msg *p2pmsg.DecryptionKeyShare) error {
	err := q.InsertDecryptionKeyShare(ctx, InsertDecryptionKeyShareParams{
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

func (q *Queries) ScheduleShutterMessage(
	ctx context.Context,
	description string,
	msg *shmsg.Message,
) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgid, err := q.ScheduleSerializedShutterMessage(ctx, ScheduleSerializedShutterMessageParams{
		Description: description,
		Msg:         data,
	})
	if err != nil {
		return err
	}
	log.Info().Int32("id", msgid).Str("description", description).
		Msg("scheduled shuttermint message")
	return nil
}

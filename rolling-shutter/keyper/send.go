package keyper

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/fx"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func SendShutterMessages(
	ctx context.Context, queries *kprdb.Queries, messageSender fx.MessageSender,
) error {
	for {
		outgoing, err := queries.GetNextShutterMessage(ctx)
		if err == pgx.ErrNoRows {
			return nil
		}

		msg := &shmsg.Message{}
		err = proto.Unmarshal(outgoing.Msg, msg)
		if err != nil {
			return err
		}
		err = messageSender.SendMessage(ctx, msg)
		if err != nil {
			return err // XXX retry
		}
		log.Info().Int32("id", outgoing.ID).Str("description", outgoing.Description).
			Msg("send shuttermint message")
		err = queries.DeleteShutterMessage(ctx, outgoing.ID)
		if err != nil {
			return err
		}
	}
}

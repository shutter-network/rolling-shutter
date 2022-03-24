package keyper

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shuttermint/keyper/fx"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
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
		log.Printf("Send shuttermint message: id=%d %s", outgoing.ID, outgoing.Description)
		err = queries.DeleteShutterMessage(ctx, outgoing.ID)
		if err != nil {
			return err
		}
	}
}

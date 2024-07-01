package fx

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

// SendShutterMessages fetches shuttermint messages from the database and sends them to shuttermint
// via the given MesssageSender.
func SendShutterMessages(
	ctx context.Context, queries *database.Queries, messageSender MessageSender,
) error {
	for {
		outgoing, err := queries.GetNextShutterMessage(ctx)
		if err == pgx.ErrNoRows {
			return nil
		} else if err != nil {
			return err
		}

		msg := &shmsg.Message{}
		err = proto.Unmarshal(outgoing.Msg, msg)
		if err != nil {
			return err
		}
		err = messageSender.SendMessage(ctx, msg)
		if err != nil {
			if !isRetrieable(msg) {
				log.Err(err).Str("msg", msg.String()).Msg("sending non-retrieable msg failed")
				return err
			} else {
				log.Info().Str("msg", msg.String()).Msg("msg not accepted, will be retried")
				return nil
			}
		}
		log.Info().Int32("id", outgoing.ID).
			Str("description", outgoing.Description).
			Msg("send shuttermint message")
		err = queries.DeleteShutterMessage(ctx, outgoing.ID)
		if err != nil {
			return err
		}
	}
}

// isRetrieable is a no-op so far
func isRetrieable(msg *shmsg.Message) bool {
	return true
}

package shutterservice

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type MessagingMiddleware struct {
	config    *Config
	messaging p2p.Messaging
	dbpool    *pgxpool.Pool
}

func NewMessagingMiddleware(messaging p2p.Messaging, dbpool *pgxpool.Pool, config *Config) *MessagingMiddleware {
	return &MessagingMiddleware{messaging: messaging, dbpool: dbpool, config: config}
}

func (i *MessagingMiddleware) SendMessage(ctx context.Context, msg p2pmsg.Message, opts ...retry.Option) error {
	//TODO: needs to be implemented
	return nil
}

func (i *MessagingMiddleware) AddValidator(ctx p2p.ValidatorFunc, protos ...p2pmsg.Message) {
	//TODO: needs to be implemented
}

func (i *MessagingMiddleware) AddMessageHandler(mhs ...p2p.MessageHandler) {
	//TODO: needs to be implemented
}

func (i *MessagingMiddleware) Start(_ context.Context, runner service.Runner) error {
	return runner.StartService(i.messaging)
}

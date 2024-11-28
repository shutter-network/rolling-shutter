package shutterservice

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type DecryptionKeySharesHandler struct {
	dbpool *pgxpool.Pool
}

// TODO: both handlers need to be implemented
func (h *DecryptionKeySharesHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeyShares{}}
}

type DecryptionKeysHandler struct {
	dbpool *pgxpool.Pool
}

func (h *DecryptionKeysHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeys{}}
}

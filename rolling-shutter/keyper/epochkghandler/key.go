package epochkghandler

import (
	"context"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func NewDecryptionKeyHandler(config Config, dbpool *pgxpool.Pool) p2p.MessageHandler {
	return &DecryptionKeyHandler{config: config, dbpool: dbpool}
}

type DecryptionKeyHandler struct {
	config Config
	dbpool *pgxpool.Pool
}

func (*DecryptionKeyHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKey{}}
}

func (handler *DecryptionKeyHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (bool, error) {
	key := msg.(*p2pmsg.DecryptionKey)
	if key.GetInstanceID() != handler.config.GetInstanceID() {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.GetInstanceID(), key.GetInstanceID())
	}
	if key.Eon > math.MaxInt64 {
		return false, errors.Errorf("eon %d overflows int64", key.Eon)
	}

	queries := database.New(handler.dbpool)

	isKeyper, err := queries.GetKeyperStateForEon(ctx, database.GetKeyperStateForEonParams{
		KeyperAddress: []string{handler.config.GetAddress().String()},
		Eon:           int64(key.Eon),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, errors.Errorf("eon %d does not exist", key.Eon)
	}
	if err != nil {
		return false, errors.Errorf("failed to get keyper state for eon %d from db", key.Eon)
	}
	if !isKeyper {
		log.Debug().Uint64("eon", key.Eon).Msg("Ignoring decryptionKey for eon; we're not a Keyper")
		return false, nil
	}

	dkgResultDB, err := queries.GetDKGResult(ctx, int64(key.Eon))
	if err == pgx.ErrNoRows {
		return false, errors.Errorf("no DKG result found for eon %d", key.Eon)
	}
	if err != nil {
		return false, errors.Wrapf(err, "failed to get dkg result for eon %d from db", key.Eon)
	}
	if !dkgResultDB.Success {
		return false, errors.Errorf("no successful DKG result found for eon %d", key.Eon)
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return false, errors.Wrapf(err, "error while decoding pure DKG result for eon %d", key.Eon)
	}
	epochSecretKey, err := key.GetEpochSecretKey()
	if err != nil {
		return false, err
	}

	ok, err := shcrypto.VerifyEpochSecretKey(epochSecretKey, pureDKGResult.PublicKey, key.EpochID)
	if err != nil {
		return false, errors.Wrapf(err, "error while checking epoch secret key for epoch %v", key.EpochID)
	}
	return ok, nil
}

func (handler *DecryptionKeyHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	metricsEpochKGDecryptionKeysReceived.Inc()
	key := msg.(*p2pmsg.DecryptionKey)
	// Insert the key into the db. We assume that it's valid as it already passed the libp2p
	// validator.
	return nil, database.New(handler.dbpool).InsertDecryptionKeyMsg(ctx, key)
}

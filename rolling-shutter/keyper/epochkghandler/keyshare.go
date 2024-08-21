package epochkghandler

import (
	"bytes"
	"context"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func NewDecryptionKeyShareHandler(config Config, dbpool *pgxpool.Pool) p2p.MessageHandler {
	return &DecryptionKeyShareHandler{config: config, dbpool: dbpool}
}

type DecryptionKeyShareHandler struct {
	config Config
	dbpool *pgxpool.Pool
}

func (*DecryptionKeyShareHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeyShares{}}
}

func (handler *DecryptionKeyShareHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	keyShare := msg.(*p2pmsg.DecryptionKeyShares)
	if keyShare.GetInstanceId() != handler.config.GetInstanceID() {
		return pubsub.ValidationReject,
			errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.GetInstanceID(), keyShare.GetInstanceId())
	}
	if keyShare.Eon > math.MaxInt64 {
		return pubsub.ValidationReject, errors.Errorf("eon %d overflows int64", keyShare.Eon)
	}

	queries := database.New(handler.dbpool)
	_, isKeyper, err := queries.GetKeyperIndex(ctx, int64(keyShare.Eon), handler.config.GetAddress())
	if err != nil {
		return pubsub.ValidationReject, err
	}
	if !isKeyper {
		log.Debug().Uint64("eon", keyShare.Eon).Msg("Ignoring decryptionKey for eon; we're not a Keyper")
		return pubsub.ValidationReject, nil
	}

	dkgResultDB, err := database.New(handler.dbpool).GetDKGResultForKeyperConfigIndex(ctx, int64(keyShare.Eon))
	if err == pgx.ErrNoRows {
		return pubsub.ValidationReject, errors.Errorf("no DKG result found for eon %d", keyShare.Eon)
	}
	if err != nil {
		return pubsub.ValidationReject, errors.Errorf("failed to get dkg result for eon %d from db", keyShare.Eon)
	}
	if !dkgResultDB.Success {
		return pubsub.ValidationReject, errors.Errorf("no successful DKG result found for eon %d", keyShare.Eon)
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return pubsub.ValidationReject, errors.Errorf("error while decoding pure DKG result for eon %d", keyShare.Eon)
	}
	if len(keyShare.Shares) == 0 {
		return pubsub.ValidationReject, errors.New("no key shares in message")
	}
	if len(keyShare.Shares) > int(handler.config.GetMaxNumKeysPerMessage()) {
		return pubsub.ValidationReject, errors.Errorf(
			"too many key shares in message (%d > %d)",
			len(keyShare.Shares),
			handler.config.GetMaxNumKeysPerMessage(),
		)
	}

	validationResult, err := checkKeyShares(keyShare, pureDKGResult)
	return validationResult, err
}

func checkKeyShares(keyShare *p2pmsg.DecryptionKeyShares, pureDKGResult *puredkg.Result) (pubsub.ValidationResult, error) {
	shares := keyShare.GetShares()
	for i, share := range shares {
		epochSecretKeyShare, err := share.GetEpochSecretKeyShare()
		if err != nil {
			return pubsub.ValidationReject, err
		}
		if !shcrypto.VerifyEpochSecretKeyShare(
			epochSecretKeyShare,
			pureDKGResult.PublicKeyShares[keyShare.KeyperIndex],
			shcrypto.ComputeEpochID(share.EpochId),
		) {
			return pubsub.ValidationReject, errors.Errorf("cannot verify secret key share")
		}

		if i > 0 && bytes.Compare(share.EpochId, shares[i-1].EpochId) < 0 {
			return pubsub.ValidationReject, errors.Errorf("keyshares not ordered")
		}
	}
	return pubsub.ValidationAccept, nil
}

func (handler *DecryptionKeyShareHandler) HandleMessage(ctx context.Context, m p2pmsg.Message) ([]p2pmsg.Message, error) {
	metricsEpochKGDecryptionKeySharesReceived.Inc()
	msg := m.(*p2pmsg.DecryptionKeyShares)
	// Insert the shares into the db. We assume that it's valid as it already passed the libp2p
	// validator.
	db := database.New(handler.dbpool)
	if err := db.InsertDecryptionKeySharesMsg(ctx, msg); err != nil {
		return nil, err
	}

	// Check that we don't know the decryption keys yet
	eon, err := medley.Uint64ToInt64Safe(msg.Eon)
	if err != nil {
		return nil, err
	}
	allKeysExist := true
	for _, share := range msg.GetShares() {
		identityPreimage := identitypreimage.IdentityPreimage(share.EpochId)
		keyExists, err := db.ExistsDecryptionKey(ctx, database.ExistsDecryptionKeyParams{
			Eon:     eon,
			EpochID: identityPreimage.Bytes(),
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to query decryption key for epoch %s", identityPreimage)
		}
		if !keyExists {
			allKeysExist = false
			break
		}
	}
	if allKeysExist {
		return nil, nil
	}

	// fetch dkg result from db
	dkgResultDB, err := db.GetDKGResultForKeyperConfigIndex(ctx, int64(msg.Eon))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dkg result for eon %d from db", msg.Eon)
	}
	if !dkgResultDB.Success {
		log.Info().Uint64("eon", msg.Eon).
			Msg("ignoring decryption key share: eon key generation failed")
		return nil, nil
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return nil, err
	}

	// aggregate epoch secret keys
	keys := []*p2pmsg.Key{}
	for _, share := range msg.GetShares() {
		identityPreimage := identitypreimage.IdentityPreimage(share.EpochId)

		epochKG, err := handler.aggregateDecryptionKeySharesFromDB(
			ctx,
			int64(msg.Eon),
			pureDKGResult,
			identityPreimage,
		)
		if err != nil {
			return nil, err
		}
		decryptionKey, ok := epochKG.SecretKeys[identityPreimage.Hex()]
		if !ok {
			numShares := uint64(len(epochKG.SecretShares))
			if numShares < pureDKGResult.Threshold {
				// not enough shares yet for at least on identity
				return nil, nil
			}
			return nil, errors.Errorf(
				"failed to generate decryption key for epoch %s even though we have enough shares",
				identityPreimage,
			)
		}

		keys = append(keys, &p2pmsg.Key{
			Identity: identityPreimage.Bytes(),
			Key:      decryptionKey.Marshal(),
		})
	}
	message := &p2pmsg.DecryptionKeys{
		InstanceId: handler.config.GetInstanceID(),
		Eon:        msg.Eon,
		Keys:       keys,
	}
	err = db.InsertDecryptionKeysMsg(ctx, message)
	if err != nil {
		return nil, err
	}
	metricsEpochKGDecryptionKeysGenerated.Inc()
	return []p2pmsg.Message{message}, nil
}

func (handler *DecryptionKeyShareHandler) aggregateDecryptionKeySharesFromDB(
	ctx context.Context,
	keyperConfigIndex int64,
	pureDKGResult *puredkg.Result,
	identityPreimage identitypreimage.IdentityPreimage,
) (*epochkg.EpochKG, error) {
	db := database.New(handler.dbpool)
	shares, err := db.SelectDecryptionKeyShares(ctx, database.SelectDecryptionKeySharesParams{
		Eon:     keyperConfigIndex,
		EpochID: identityPreimage.Bytes(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get decryption key shares for epoch %s from db", identityPreimage)
	}

	epochKG := epochkg.NewEpochKG(pureDKGResult)
	// For simplicity, we aggregate shares even if we don't have enough of them yet.
	for _, share := range shares {
		identityPreimage := identitypreimage.IdentityPreimage(share.EpochID)

		shareDecoded, err := shdb.DecodeEpochSecretKeyShare(share.DecryptionKeyShare)
		if err != nil {
			log.Warn().Str("epoch-id", identityPreimage.Hex()).Int64("keyper-index", share.KeyperIndex).
				Msg("invalid decryption key share in DB")
			continue
		}
		err = epochKG.HandleEpochSecretKeyShare(&epochkg.EpochSecretKeyShare{
			Eon:              pureDKGResult.Eon,
			IdentityPreimage: identityPreimage,
			Sender:           uint64(share.KeyperIndex),
			Share:            shareDecoded,
		})
		if err != nil {
			log.Info().Str("epoch-id", identityPreimage.Hex()).Int64("keyper-index", share.KeyperIndex).
				Msg("failed to process decryption key share")
			continue
		}
	}

	return epochKG, nil
}

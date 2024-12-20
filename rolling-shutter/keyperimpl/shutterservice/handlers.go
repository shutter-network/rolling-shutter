package shutterservice

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/serviceztypes"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type DecryptionKeySharesHandler struct {
	dbpool *pgxpool.Pool
}

// TODO: problem in using direct api is it has decryption trigger api and shutdown api??

func (h *DecryptionKeySharesHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeyShares{}}
}

func (h *DecryptionKeySharesHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	keyShares := msg.(*p2pmsg.DecryptionKeyShares)
	extra, ok := keyShares.Extra.(*p2pmsg.DecryptionKeyShares_Service)
	if !ok {
		return pubsub.ValidationReject, errors.Errorf("unexpected extra type %T, expected service", keyShares.Extra)
	}
	if extra.Service == nil {
		return pubsub.ValidationReject, errors.New("missing extra service data")
	}

	obsKeyperDB := obskeyperdatabase.New(h.dbpool)
	keyperSet, err := obsKeyperDB.GetKeyperSetByKeyperConfigIndex(ctx, int64(keyShares.Eon))
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err,
			"failed to get keyper set from database for keyper set index %d",
			keyShares.Eon,
		)
	}
	if keyShares.KeyperIndex >= uint64(len(keyperSet.Keypers)) {
		return pubsub.ValidationReject, errors.Errorf(
			"keyper index %d out of range for keyper set %d",
			keyShares.KeyperIndex,
			keyShares.Eon,
		)
	}
	keyperAddressStr := keyperSet.Keypers[keyShares.KeyperIndex]
	keyperAddress, err := shdb.DecodeAddress(keyperAddressStr)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "failed to decode keyper address from database")
	}

	identityPreimages := []identitypreimage.IdentityPreimage{}
	for _, share := range keyShares.Shares {
		identityPreimage := identitypreimage.IdentityPreimage(share.IdentityPreimage)
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	signatureData, err := serviceztypes.NewDecryptionSignatureData(keyShares.InstanceId, keyShares.Eon, identityPreimages)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "failed to create decryption signature data object")
	}
	valid, err := signatureData.CheckSignature(extra.Service.Signature, keyperAddress)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "failed to check decryption signature")
	}
	if !valid {
		return pubsub.ValidationReject, errors.New("decryption signature invalid")
	}

	return pubsub.ValidationAccept, nil
}

func (h *DecryptionKeySharesHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	keyShares := msg.(*p2pmsg.DecryptionKeyShares)
	extra := keyShares.Extra.(*p2pmsg.DecryptionKeyShares_Service).Service

	serviceDB := database.New(h.dbpool)
	keyperCoreDB := corekeyperdatabase.New(h.dbpool)
	obsKeyperDB := obskeyperdatabase.New(h.dbpool)

	identitiesHash := computeIdentitiesHashFromShares(keyShares.Shares)
	err := serviceDB.InsertDecryptionSignature(ctx, database.InsertDecryptionSignatureParams{
		Eon:            int64(keyShares.Eon),
		KeyperIndex:    int64(keyShares.KeyperIndex),
		IdentitiesHash: identitiesHash,
		Signature:      extra.Signature,
	})
	if err != nil {
		return []p2pmsg.Message{}, errors.Wrap(err, "failed to insert vote")
	}

	keyperSet, err := obsKeyperDB.GetKeyperSetByKeyperConfigIndex(ctx, int64(keyShares.Eon))
	if err != nil {
		return []p2pmsg.Message{}, errors.Wrapf(err, "failed to get keyper set from database for index %d", keyShares.Eon)
	}
	signatures, err := serviceDB.GetDecryptionSignatures(ctx, database.GetDecryptionSignaturesParams{
		Eon:            int64(keyShares.Eon),
		IdentitiesHash: identitiesHash,
		Limit:          keyperSet.Threshold,
	})
	if err != nil {
		return []p2pmsg.Message{}, errors.Wrap(err, "failed to count decryption signatures")
	}

	// send a keys message if we have reached the required number of both the signatures and the key shares
	if len(signatures) >= int(keyperSet.Threshold) {
		keys := []*p2pmsg.Key{}
		for _, share := range keyShares.GetShares() {
			decryptionKeyDB, err := keyperCoreDB.GetDecryptionKey(ctx, corekeyperdatabase.GetDecryptionKeyParams{
				Eon:     int64(keyShares.Eon),
				EpochID: share.IdentityPreimage,
			})
			if err == pgx.ErrNoRows {
				return []p2pmsg.Message{}, nil
			}
			key := &p2pmsg.Key{
				IdentityPreimage: share.IdentityPreimage,
				Key:              decryptionKeyDB.DecryptionKey,
			}
			keys = append(keys, key)
		}
		signerIndices := []uint64{}
		signaturesCum := [][]byte{}
		for _, signature := range signatures {
			signerIndices = append(signerIndices, uint64(signature.KeyperIndex))
			signaturesCum = append(signaturesCum, signature.Signature)
		}
		decryptionKeysMsg := &p2pmsg.DecryptionKeys{
			InstanceId: keyShares.InstanceId,
			Eon:        keyShares.Eon,
			Keys:       keys,
			Extra: &p2pmsg.DecryptionKeys_Service{
				Service: &p2pmsg.ShutterServiceDecryptionKeysExtra{
					SignerIndices: signerIndices,
					Signature:     signaturesCum,
				},
			},
		}
		return []p2pmsg.Message{decryptionKeysMsg}, nil
	}
	return []p2pmsg.Message{}, nil
}

type DecryptionKeysHandler struct {
	dbpool *pgxpool.Pool
}

func (h *DecryptionKeysHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeys{}}
}

func validateSignerIndices(extra *p2pmsg.ShutterServiceDecryptionKeysExtra, n int) (pubsub.ValidationResult, error) {
	for i, signerIndex := range extra.SignerIndices {
		if i >= 1 {
			prevSignerIndex := extra.SignerIndices[i-1]
			if signerIndex == prevSignerIndex {
				return pubsub.ValidationReject, errors.New("duplicate signer index found")
			}
			if signerIndex < prevSignerIndex {
				return pubsub.ValidationReject, errors.New("signer indices not ordered")
			}
		}
		if signerIndex >= uint64(n) {
			return pubsub.ValidationReject, errors.New("signer index out of range")
		}
	}
	return pubsub.ValidationAccept, nil
}

func ValidateDecryptionKeysSignatures(
	keys *p2pmsg.DecryptionKeys,
	extra *p2pmsg.ShutterServiceDecryptionKeysExtra,
	keyperSet *obskeyperdatabase.KeyperSet,
) (pubsub.ValidationResult, error) {
	if int32(len(extra.SignerIndices)) != keyperSet.Threshold {
		return pubsub.ValidationReject, errors.Errorf("expected %d signers, got %d", keyperSet.Threshold, len(extra.SignerIndices))
	}
	res, err := validateSignerIndices(extra, len(keyperSet.Keypers))
	if res != pubsub.ValidationAccept {
		return res, err
	}
	signers, err := keyperSet.GetSubset(extra.SignerIndices)
	if err != nil {
		return pubsub.ValidationReject, err
	}
	identityPreimages := []identitypreimage.IdentityPreimage{}
	for _, key := range keys.Keys {
		identityPreimage := identitypreimage.IdentityPreimage(key.IdentityPreimage)
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	sigData, err := serviceztypes.NewDecryptionSignatureData(keys.InstanceId, keys.Eon, identityPreimages)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "failed to create decryption signature data object")
	}

	for signatureIndex := 0; signatureIndex < len(extra.Signature); signatureIndex++ {
		signature := extra.Signature[signatureIndex]
		signer := signers[signatureIndex]
		signatureValid, err := sigData.CheckSignature(signature, signer)
		if err != nil {
			return pubsub.ValidationReject, errors.Wrap(err, "failed to check decryption signature")
		}
		if !signatureValid {
			return pubsub.ValidationReject, errors.New("decryption signature invalid")
		}
	}

	return pubsub.ValidationAccept, nil
}

func (h *DecryptionKeysHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	keys := msg.(*p2pmsg.DecryptionKeys)
	extra, ok := keys.Extra.(*p2pmsg.DecryptionKeys_Service)
	if !ok {
		return pubsub.ValidationReject, errors.Errorf("unexpected extra type %T, expected Service", keys.Extra)
	}
	if extra.Service == nil {
		return pubsub.ValidationReject, errors.New("missing extra Service data")
	}

	obsKeyperDB := obskeyperdatabase.New(h.dbpool)
	keyperSet, err := obsKeyperDB.GetKeyperSetByKeyperConfigIndex(ctx, int64(keys.Eon))
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err, "failed to get keyper set from database for eon %d", keys.Eon)
	}

	res, err := ValidateDecryptionKeysSignatures(keys, extra.Service, &keyperSet)
	if res != pubsub.ValidationAccept || err != nil {
		return res, err
	}

	return pubsub.ValidationAccept, nil
}

func (h *DecryptionKeysHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	keys := msg.(*p2pmsg.DecryptionKeys)
	extra := keys.Extra.(*p2pmsg.DecryptionKeys_Service).Service
	serviceDB := database.New(h.dbpool)

	identityPreimages := []identitypreimage.IdentityPreimage{}
	for _, key := range keys.Keys {
		identityPreimage := identitypreimage.IdentityPreimage(key.IdentityPreimage)
		identityPreimages = append(identityPreimages, identityPreimage)
	}
	identitiesHash := computeIdentitiesHash(identityPreimages)
	for i, keyperIndex := range extra.SignerIndices {
		err := serviceDB.InsertDecryptionSignature(ctx, database.InsertDecryptionSignatureParams{
			Eon:            int64(keys.Eon),
			KeyperIndex:    int64(keyperIndex),
			IdentitiesHash: identitiesHash,
			Signature:      extra.Signature[i],
		})
		if err != nil {
			return []p2pmsg.Message{}, errors.Wrap(err, "failed to insert decryption signature")
		}
	}

	err := updateEventFlag(ctx, serviceDB, keys)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("failed to update events for decryption keys released")
	}

	return []p2pmsg.Message{}, nil
}

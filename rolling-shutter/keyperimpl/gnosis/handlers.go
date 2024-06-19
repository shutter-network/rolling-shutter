package gnosis

import (
	"context"
	"database/sql"
	"math"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/gnosisssztypes"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type DecryptionKeySharesHandler struct {
	dbpool *pgxpool.Pool
}

func (h *DecryptionKeySharesHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeyShares{}}
}

func (h *DecryptionKeySharesHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	keyShares := msg.(*p2pmsg.DecryptionKeyShares)
	extra, ok := keyShares.Extra.(*p2pmsg.DecryptionKeyShares_Gnosis)
	if !ok {
		return pubsub.ValidationReject, errors.Errorf("unexpected extra type %T, expected Gnosis", keyShares.Extra)
	}
	if extra.Gnosis == nil {
		return pubsub.ValidationReject, errors.New("missing extra Gnosis data")
	}

	if extra.Gnosis.Slot > math.MaxInt64 {
		return pubsub.ValidationReject, errors.New("slot number too large")
	}
	if extra.Gnosis.TxPointer > math.MaxInt64 {
		return pubsub.ValidationReject, errors.New("tx pointer too large")
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
		identityPreimage := identitypreimage.IdentityPreimage(share.EpochID)
		identityPreimages = append(identityPreimages, identityPreimage)
	}
	slotDecryptionSignatureData, err := gnosisssztypes.NewSlotDecryptionSignatureData(
		keyShares.InstanceID,
		keyShares.Eon,
		extra.Gnosis.Slot,
		extra.Gnosis.TxPointer,
		identityPreimages,
	)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "failed to create slot decryption signature data object")
	}
	signatureValid, err := slotDecryptionSignatureData.CheckSignature(extra.Gnosis.Signature, keyperAddress)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "failed to check slot decryption signature")
	}
	if !signatureValid {
		return pubsub.ValidationReject, errors.New("slot decryption signature invalid")
	}

	return pubsub.ValidationAccept, nil
}

func (h *DecryptionKeySharesHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	keyShares := msg.(*p2pmsg.DecryptionKeyShares)
	extra := keyShares.Extra.(*p2pmsg.DecryptionKeyShares_Gnosis).Gnosis

	gnosisDB := database.New(h.dbpool)
	keyperCoreDB := corekeyperdatabase.New(h.dbpool)
	obsKeyperDB := obskeyperdatabase.New(h.dbpool)

	identitiesHash := computeIdentitiesHashFromShares(keyShares.Shares)
	err := gnosisDB.InsertSlotDecryptionSignature(ctx, database.InsertSlotDecryptionSignatureParams{
		Eon:            int64(keyShares.Eon),
		Slot:           int64(extra.Slot),
		KeyperIndex:    int64(keyShares.KeyperIndex),
		TxPointer:      int64(extra.TxPointer),
		IdentitiesHash: identitiesHash,
		Signature:      extra.Signature,
	})
	if err != nil {
		return []p2pmsg.Message{}, errors.Wrap(err, "failed to insert tx pointer vote")
	}

	keyperSet, err := obsKeyperDB.GetKeyperSetByKeyperConfigIndex(ctx, int64(keyShares.Eon))
	if err != nil {
		return []p2pmsg.Message{}, errors.Wrapf(err, "failed to get keyper set from database for index %d", keyShares.Eon)
	}

	signaturesDB, err := gnosisDB.GetSlotDecryptionSignatures(ctx, database.GetSlotDecryptionSignaturesParams{
		Eon:            int64(keyShares.Eon),
		Slot:           int64(extra.Slot),
		TxPointer:      int64(extra.TxPointer),
		IdentitiesHash: identitiesHash,
		Limit:          keyperSet.Threshold,
	})
	if err != nil {
		return []p2pmsg.Message{}, errors.Wrap(err, "failed to count slot decryption signatures")
	}

	// send a keys message if we have reached the required number of both the signatures and the key shares
	if len(signaturesDB) >= int(keyperSet.Threshold) {
		keys := []*p2pmsg.Key{}
		for _, share := range keyShares.GetShares() {
			decryptionKeyDB, err := keyperCoreDB.GetDecryptionKey(ctx, corekeyperdatabase.GetDecryptionKeyParams{
				Eon:     int64(keyShares.Eon),
				EpochID: share.EpochID,
			})
			if err == pgx.ErrNoRows {
				return []p2pmsg.Message{}, nil
			}
			key := &p2pmsg.Key{
				Identity: share.EpochID,
				Key:      decryptionKeyDB.DecryptionKey,
			}
			keys = append(keys, key)
		}
		signerIndices := []uint64{}
		signatures := [][]byte{}
		for _, signature := range signaturesDB {
			signerIndices = append(signerIndices, uint64(signature.KeyperIndex))
			signatures = append(signatures, signature.Signature)
		}
		decryptionKeysMsg := &p2pmsg.DecryptionKeys{
			InstanceID: keyShares.InstanceID,
			Eon:        keyShares.Eon,
			Keys:       keys,
			Extra: &p2pmsg.DecryptionKeys_Gnosis{
				Gnosis: &p2pmsg.GnosisDecryptionKeysExtra{
					Slot:          extra.Slot,
					TxPointer:     extra.TxPointer,
					SignerIndices: signerIndices,
					Signatures:    signatures,
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

func validateSignerIndices(extra *p2pmsg.DecryptionKeys_Gnosis, n int) (pubsub.ValidationResult, error) {
	for i, signerIndex := range extra.Gnosis.SignerIndices {
		if i >= 1 {
			prevSignerIndex := extra.Gnosis.SignerIndices[i-1]
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

func (h *DecryptionKeysHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	keys := msg.(*p2pmsg.DecryptionKeys)
	extra, ok := keys.Extra.(*p2pmsg.DecryptionKeys_Gnosis)
	if !ok {
		return pubsub.ValidationReject, errors.Errorf("unexpected extra type %T, expected Gnosis", keys.Extra)
	}
	if extra.Gnosis == nil {
		return pubsub.ValidationReject, errors.New("missing extra Gnosis data")
	}

	if extra.Gnosis.Slot > math.MaxInt64 {
		return pubsub.ValidationReject, errors.New("slot number too large")
	}
	if extra.Gnosis.TxPointer > math.MaxInt32 { // the pointer will have to be incremented
		return pubsub.ValidationReject, errors.New("tx pointer too large")
	}
	if len(keys.Keys) == 0 {
		return pubsub.ValidationReject, errors.New("msg does not contain any keys")
	}

	obsKeyperDB := obskeyperdatabase.New(h.dbpool)
	keyperSet, err := obsKeyperDB.GetKeyperSetByKeyperConfigIndex(ctx, int64(keys.Eon))
	if err != nil {
		return pubsub.ValidationReject, errors.Wrapf(err, "failed to get keyper set from database for eon %d", keys.Eon)
	}

	if int32(len(extra.Gnosis.SignerIndices)) != keyperSet.Threshold {
		return pubsub.ValidationReject, errors.Errorf("expected %d signers, got %d", keyperSet.Threshold, len(extra.Gnosis.SignerIndices))
	}

	res, err := validateSignerIndices(extra, len(keyperSet.Keypers))
	if res != pubsub.ValidationAccept {
		return res, err
	}
	signers, err := keyperSet.GetSubset(extra.Gnosis.SignerIndices)
	if err != nil {
		return pubsub.ValidationReject, err
	}

	identityPreimages := []identitypreimage.IdentityPreimage{}
	for _, key := range keys.Keys {
		identityPreimage := identitypreimage.IdentityPreimage(key.Identity)
		identityPreimages = append(identityPreimages, identityPreimage)
	}
	slotDecryptionSignatureData, err := gnosisssztypes.NewSlotDecryptionSignatureData(
		keys.InstanceID,
		keys.Eon,
		extra.Gnosis.Slot,
		extra.Gnosis.TxPointer,
		identityPreimages,
	)
	if err != nil {
		return pubsub.ValidationReject, errors.Wrap(err, "failed to create slot decryption signature data object")
	}
	for signatureIndex := 0; signatureIndex < len(extra.Gnosis.Signatures); signatureIndex++ {
		signature := extra.Gnosis.Signatures[signatureIndex]
		signer := signers[signatureIndex]
		signatureValid, err := slotDecryptionSignatureData.CheckSignature(signature, signer)
		if err != nil {
			return pubsub.ValidationReject, errors.Wrap(err, "failed to check slot decryption signature")
		}
		if !signatureValid {
			return pubsub.ValidationReject, errors.New("slot decryption signature invalid")
		}
	}

	return pubsub.ValidationAccept, nil
}

func (h *DecryptionKeysHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	keys := msg.(*p2pmsg.DecryptionKeys)
	extra := keys.Extra.(*p2pmsg.DecryptionKeys_Gnosis).Gnosis
	gnosisDB := database.New(h.dbpool)
	// the first key is the block key, only the rest are tx keys, so subtract 1
	newTxPointer := int64(extra.TxPointer) + int64(len(keys.Keys)) - 1
	log.Debug().
		Uint64("eon", keys.Eon).
		Uint64("slot", extra.Slot).
		Uint64("tx-pointer-msg", extra.TxPointer).
		Int("num-keys", len(keys.Keys)).
		Int64("tx-pointer-updated", newTxPointer).
		Msg("updating tx pointer")
	err := gnosisDB.SetTxPointer(ctx, database.SetTxPointerParams{
		Eon: int64(keys.Eon),
		Age: sql.NullInt64{
			Int64: 0,
			Valid: true,
		},
		Value: newTxPointer,
	})
	if err != nil {
		return []p2pmsg.Message{}, errors.Wrap(err, "failed to set tx pointer")
	}
	return []p2pmsg.Message{}, nil
}

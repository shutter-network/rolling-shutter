package epochkghandler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type Config interface {
	GetAddress() common.Address
	GetInstanceID() uint64
}

type EpochKGHandler struct {
	config Config
	dbpool *pgxpool.Pool
}

func New(config Config, dbpool *pgxpool.Pool) *EpochKGHandler {
	return &EpochKGHandler{
		config: config,
		dbpool: dbpool,
	}
}

func (h *EpochKGHandler) SetupP2p(p2phandler *p2p.P2PHandler) {
	p2p.AddValidator(p2phandler, h.validateDecryptionKey)
	p2p.AddValidator(p2phandler, h.validateDecryptionKeyShare)
	p2p.AddValidator(p2phandler, h.validateEonPublicKey)
	p2p.AddValidator(p2phandler, h.validateDecryptionTrigger)

	p2p.AddHandlerFunc(p2phandler, h.handleDecryptionTrigger)
	p2p.AddHandlerFunc(p2phandler, h.handleDecryptionKeyShare)
	p2p.AddHandlerFunc(p2phandler, h.handleDecryptionKey)
}

func (h *EpochKGHandler) handleDecryptionTrigger(
	ctx context.Context, msg *p2pmsg.DecryptionTrigger,
) ([]p2pmsg.Message, error) {
	log.Info().Str("message", msg.String()).Msg("received decryption trigger")
	epochID, err := epochid.BytesToEpochID(msg.EpochID)
	if err != nil {
		return nil, err
	}
	return h.SendDecryptionKeyShare(ctx, epochID, int64(msg.BlockNumber))
}

func (h *EpochKGHandler) SendDecryptionKeyShare(
	ctx context.Context, epochID epochid.EpochID, blockNumber int64,
) ([]p2pmsg.Message, error) {
	db := kprdb.New(h.dbpool)
	eon, err := db.GetEonForBlockNumber(ctx, blockNumber)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get eon for block %d from db", blockNumber)
	}
	batchConfig, err := db.GetBatchConfig(ctx, int32(eon.KeyperConfigIndex))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config %d from db", eon.KeyperConfigIndex)
	}

	// get our keyper index (and check that we in fact are a keyper)
	encodedAddress := shdb.EncodeAddress(h.config.GetAddress())
	keyperIndex := int64(-1)
	for i, address := range batchConfig.Keypers {
		if address == encodedAddress {
			keyperIndex = int64(i)
			break
		}
	}
	if keyperIndex == -1 {
		log.Info().Str("epoch-id", epochID.Hex()).Msg("ignoring decryption trigger: we are not a keyper")
		return nil, nil
	}

	// check if we already computed (and therefore most likely sent) our key share
	shareExists, err := db.ExistsDecryptionKeyShare(ctx, kprdb.ExistsDecryptionKeyShareParams{
		EpochID:     epochID.Bytes(),
		KeyperIndex: keyperIndex,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get decryption key share for epoch %s from db", epochID)
	}
	if shareExists {
		return nil, nil // we already sent our share
	}

	// fetch dkg result from db
	dkgResultDB, err := db.GetDKGResult(ctx, eon.Eon)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dkg result for eon %d from db", eon.Eon)
	}
	if !dkgResultDB.Success {
		log.Info().Int64("eon", eon.Eon).Msg("ignoring decryption trigger: eon key generation failed")
		return nil, nil
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return nil, err
	}

	// compute the key share
	epochKG := epochkg.NewEpochKG(pureDKGResult)
	share := epochKG.ComputeEpochSecretKeyShare(epochID)

	msg := &p2pmsg.DecryptionKeyShare{
		InstanceID:  h.config.GetInstanceID(),
		Eon:         uint64(eon.Eon),
		EpochID:     epochID.Bytes(),
		KeyperIndex: uint64(keyperIndex),
		Share:       share.Marshal(),
	}
	err = db.InsertDecryptionKeyShareMsg(ctx, msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert decryption key share")
	}
	log.Info().Str("epoch-id", epochID.Hex()).Int64("block-number", blockNumber).
		Msg("sending decryption key share")
	return []p2pmsg.Message{msg}, nil
}

func (h *EpochKGHandler) aggregateDecryptionKeySharesFromDB(
	ctx context.Context,
	pureDKGResult *puredkg.Result,
	epochID epochid.EpochID,
) (*epochkg.EpochKG, error) {
	db := kprdb.New(h.dbpool)
	shares, err := db.SelectDecryptionKeyShares(ctx, kprdb.SelectDecryptionKeySharesParams{
		Eon:     int64(pureDKGResult.Eon),
		EpochID: epochID.Bytes(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get decryption key shares for epoch %s from db", epochID)
	}

	epochKG := epochkg.NewEpochKG(pureDKGResult)
	// For simplicity, we aggregate shares even if we don't have enough of them yet.
	for _, share := range shares {
		epochID, err := epochid.BytesToEpochID(share.EpochID)
		if err != nil {
			return nil, errors.Wrap(err, "invalid epoch id in db")
		}
		shareDecoded, err := shdb.DecodeEpochSecretKeyShare(share.DecryptionKeyShare)
		if err != nil {
			log.Warn().Str("epoch-id", epochID.Hex()).Int64("keyper-index", share.KeyperIndex).
				Msg("invalid decryption key share in DB")
			continue
		}
		err = epochKG.HandleEpochSecretKeyShare(&epochkg.EpochSecretKeyShare{
			Eon:    pureDKGResult.Eon,
			Epoch:  epochID,
			Sender: uint64(share.KeyperIndex),
			Share:  shareDecoded,
		})
		if err != nil {
			log.Info().Str("epoch-id", epochID.Hex()).Int64("keyper-index", share.KeyperIndex).
				Msg("failed to process decryption key share")
			continue
		}
	}

	return epochKG, nil
}

func (h *EpochKGHandler) handleDecryptionKeyShare(ctx context.Context, msg *p2pmsg.DecryptionKeyShare) ([]p2pmsg.Message, error) {
	// Insert the share into the db. We assume that it's valid as it already passed the libp2p
	// validator.
	db := kprdb.New(h.dbpool)

	if err := db.InsertDecryptionKeyShareMsg(ctx, msg); err != nil {
		return nil, err
	}

	// Check that we don't know the decryption key yet
	epochID, err := epochid.BytesToEpochID(msg.EpochID)
	if err != nil {
		return nil, err
	}
	keyExists, err := db.ExistsDecryptionKey(ctx, kprdb.ExistsDecryptionKeyParams{
		Eon:     int64(msg.Eon),
		EpochID: epochID.Bytes(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query decryption key for epoch %s", epochID)
	}
	if keyExists {
		return nil, nil
	}

	// fetch dkg result from db
	dkgResultDB, err := db.GetDKGResult(ctx, int64(msg.Eon))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dkg result for eon %d from db", msg.Eon)
	}
	if !dkgResultDB.Success {
		log.Info().Uint64("eon", msg.Eon).
			Msg("ignoring decryption trigger: eon key generation failed")
		return nil, nil
	}
	pureDKGResult, err := shdb.DecodePureDKGResult(dkgResultDB.PureResult)
	if err != nil {
		return nil, err
	}

	// aggregate epoch secret key
	epochKG, err := h.aggregateDecryptionKeySharesFromDB(ctx, pureDKGResult, epochID)
	if err != nil {
		return nil, err
	}
	decryptionKey, ok := epochKG.SecretKeys[epochID]
	if !ok {
		numShares := uint64(len(epochKG.SecretShares))
		if numShares < pureDKGResult.Threshold {
			// not enough shares yet
			return nil, nil
		}
		return nil, errors.Errorf(
			"failed to generate decryption key for epoch %s even though we have enough shares",
			epochID,
		)
	}
	message := &p2pmsg.DecryptionKey{
		InstanceID: h.config.GetInstanceID(),
		Eon:        msg.Eon,
		EpochID:    epochID.Bytes(),
		Key:        decryptionKey.Marshal(),
	}
	err = db.InsertDecryptionKeyMsg(ctx, message)
	if err != nil {
		return nil, err
	}
	log.Info().Str("epoch-id", epochID.Hex()).Str("message", message.String()).
		Msg("broadcasting decryption key")
	return []p2pmsg.Message{message}, nil
}

func (h *EpochKGHandler) handleDecryptionKey(ctx context.Context, msg *p2pmsg.DecryptionKey) ([]p2pmsg.Message, error) {
	// Insert the key into the db. We assume that it's valid as it already passed the libp2p
	// validator.
	return nil, kprdb.New(h.dbpool).InsertDecryptionKeyMsg(ctx, msg)
}

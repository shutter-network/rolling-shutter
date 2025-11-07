package primev

import (
	"bytes"
	"context"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	corekeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/primev/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type PrimevCommitmentHandler struct {
	config                   *Config
	decryptionTriggerChannel chan *broker.Event[*epochkghandler.DecryptionTrigger]
	dbpool                   *pgxpool.Pool
}

func (h *PrimevCommitmentHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.Commitment{}}
}

func (h *PrimevCommitmentHandler) ValidateMessage(_ context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	commitment, ok := msg.(*p2pmsg.Commitment)
	if !ok {
		return pubsub.ValidationReject, errors.Errorf("received message of unexpected type %s", msg.ProtoReflect().Descriptor().FullName())
	}
	if len(commitment.Identities) != len(commitment.TxHashes) {
		return pubsub.ValidationReject, errors.Errorf("number of identities (%d) does not match number of tx hashes (%d)",
			len(commitment.Identities), len(commitment.TxHashes))
	}
	if commitment.GetInstanceId() != h.config.InstanceID {
		return pubsub.ValidationReject, errors.Errorf("instance ID mismatch (want=%d, have=%d)",
			h.config.InstanceID, commitment.GetInstanceId())
	}
	return pubsub.ValidationAccept, nil
	// TODO: more validations need to be done here
}

func (h *PrimevCommitmentHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	log.Info().Msg("received commitment")
	commitment, ok := msg.(*p2pmsg.Commitment)
	if !ok {
		return nil, errors.Errorf("received message of unexpected type %s", msg.ProtoReflect().Descriptor().FullName())
	}

	hLog := log.With().
		Str("provider_address", commitment.ProviderAddress).
		Strs("identities_prefixes", commitment.Identities).
		Logger()

	bidderNodeAddress, err := getBidderNodeAddress(commitment.ReceivedBidDigest, commitment.ReceivedBidSignature)
	if err != nil {
		hLog.Error().Err(err).Msg("failed to get bidder node address")
		return nil, err
	}

	identityPreimages := make([]identitypreimage.IdentityPreimage, 0, len(commitment.Identities))
	identityPreimagesHex := make([]string, 0, len(commitment.Identities))
	for _, identityPrefix := range commitment.Identities {
		identityPrefixBytes, err := hex.DecodeString(identityPrefix)
		if err != nil {
			hLog.Error().Err(err).Msg("failed to decode identity prefix")
			return nil, err
		}
		identityPreimage := computeIdentity(identityPrefixBytes, bidderNodeAddress.Bytes())
		identityPreimageTyped := identitypreimage.IdentityPreimage(identityPreimage)
		identityPreimages = append(identityPreimages, identityPreimageTyped)
		identityPreimagesHex = append(identityPreimagesHex, identityPreimageTyped.Hex())
	}

	blockNumbers := make([]int64, 0, len(commitment.Identities))
	eons := make([]int64, 0, len(commitment.Identities))

	obsKeyperDB := corekeyperdatabase.New(h.dbpool)
	eon, err := obsKeyperDB.GetEonForBlockNumber(ctx, commitment.BlockNumber)
	if err != nil {
		hLog.Error().Err(err).Msg("failed to get eon for block number")
		return nil, err
	}
	for range commitment.Identities {
		blockNumbers = append(blockNumbers, commitment.BlockNumber)
		eons = append(eons, eon.Eon)
	}
	db := database.New(h.dbpool)
	err = db.InsertMultipleTransactionsAndUpsertCommitment(ctx, database.InsertMultipleTransactionsAndUpsertCommitmentParams{
		Eons:                 eons,
		IdentityPreimages:    identityPreimagesHex,
		BlockNumbers:         blockNumbers,
		TxHashes:             commitment.TxHashes,
		IdentityPrefixes:     commitment.Identities,
		ProviderAddress:      commitment.ProviderAddress,
		CommitmentSignature:  commitment.CommitmentSignature,
		CommitmentDigest:     commitment.CommitmentDigest,
		BlockNumber:          commitment.BlockNumber,
		ReceivedBidDigest:    commitment.ReceivedBidDigest,
		ReceivedBidSignature: commitment.ReceivedBidSignature,
		BidderNodeAddress:    bidderNodeAddress.Hex(),
	})
	if err != nil {
		hLog.Error().Err(err).Msg("failed to insert multiple transactions and upsert commitment")
		return nil, err
	}

	blockNumberUint64, err := medley.Int64ToUint64Safe(commitment.BlockNumber)
	if err != nil {
		hLog.Error().Err(err).Msg("failed to convert block number to uint64")
		return nil, err
	}

	// TODO: before sending the dec trigger, we need to check if majority of providers have generated commitments

	decryptionTrigger := &epochkghandler.DecryptionTrigger{
		BlockNumber:       blockNumberUint64,
		IdentityPreimages: identityPreimages,
	}
	h.decryptionTriggerChannel <- broker.NewEvent(decryptionTrigger)

	hLog.Info().Msg("sent decryption trigger")

	return nil, nil
}

func getBidderNodeAddress(digest, signature string) (*common.Address, error) {
	digestBytes := common.FromHex(digest)
	signatureBytes := common.FromHex(signature)

	if signatureBytes[64] == 27 || signatureBytes[64] == 28 {
		signatureBytes[64] -= 27 // Transform V from 27/28 to 0/1
	}

	pubKey, err := crypto.SigToPub(digestBytes, signatureBytes)
	if err != nil {
		return nil, err
	}
	bidderNodeAddress := crypto.PubkeyToAddress(*pubKey)
	return &bidderNodeAddress, nil
}

func computeIdentity(identityPrefix, sender []byte) []byte {
	var buf bytes.Buffer
	buf.Write(identityPrefix)
	buf.Write(sender)
	return crypto.Keccak256(buf.Bytes())
}

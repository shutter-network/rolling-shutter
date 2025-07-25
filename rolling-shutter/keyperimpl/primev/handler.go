package primev

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/primev/database"
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

func (h *PrimevCommitmentHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	commitment := msg.(*p2pmsg.Commitment)
	if commitment.GetInstanceId() != h.config.InstanceID {
		return pubsub.ValidationReject, errors.Errorf("instance ID mismatch (want=%d, have=%d)", h.config.InstanceID, commitment.GetInstanceId())
	}
	return pubsub.ValidationAccept, nil
	// TODO: more validations need to be done here
}

func (h *PrimevCommitmentHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	commitment := msg.(*p2pmsg.Commitment)

	// TODO: need to validate on identity preimage
	identityPreimages := make([]identitypreimage.IdentityPreimage, 0, len(commitment.TxHashes))
	for _, txHash := range commitment.TxHashes {
		identityPreimage, err := identitypreimage.HexToIdentityPreimage(txHash)
		if err != nil {
			return nil, err
		}
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	decryptionTrigger := &epochkghandler.DecryptionTrigger{
		BlockNumber:       uint64(commitment.BlockNumber),
		IdentityPreimages: identityPreimages,
	}

	blockNumbers := make([]int64, 0, len(commitment.TxHashes))
	eons := make([]int64, 0, len(commitment.TxHashes))
	for range commitment.TxHashes {
		blockNumbers = append(blockNumbers, int64(commitment.BlockNumber))
		eons = append(eons, int64(1)) // TODO: need to get eon from the commitment
	}

	db := database.New(h.dbpool)
	err := db.InsertMultipleTransactionsAndUpsertCommitment(ctx, database.InsertMultipleTransactionsAndUpsertCommitmentParams{
		Column1:             eons,
		Column2:             commitment.TxHashes, // right now identityPreimage is txHash
		Column3:             blockNumbers,
		Column4:             commitment.TxHashes,
		ProviderAddress:     commitment.ProviderAddress,
		CommitmentSignature: commitment.CommitmentSignature,
		CommitmentDigest:    commitment.CommitmentDigest,
		BlockNumber:         int64(commitment.BlockNumber),
	})
	if err != nil {
		return nil, err
	}

	h.decryptionTriggerChannel <- broker.NewEvent(decryptionTrigger)

	return nil, nil
}

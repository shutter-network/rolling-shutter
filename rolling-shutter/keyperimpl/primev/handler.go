package primev

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"

	corekeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
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
	if len(commitment.Identities) != len(commitment.TxHashes) {
		return pubsub.ValidationReject, errors.Errorf("number of identities (%d) does not match number of tx hashes (%d)", len(commitment.Identities), len(commitment.TxHashes))
	}
	if commitment.GetInstanceId() != h.config.InstanceID {
		return pubsub.ValidationReject, errors.Errorf("instance ID mismatch (want=%d, have=%d)", h.config.InstanceID, commitment.GetInstanceId())
	}
	return pubsub.ValidationAccept, nil
	// TODO: more validations need to be done here
}

func (h *PrimevCommitmentHandler) HandleMessage(ctx context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	commitment := msg.(*p2pmsg.Commitment)

	identityPreimages := make([]identitypreimage.IdentityPreimage, 0, len(commitment.Identities))
	for _, identity := range commitment.Identities {
		identityPreimage, err := identitypreimage.HexToIdentityPreimage(identity)
		if err != nil {
			return nil, err
		}
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	decryptionTrigger := &epochkghandler.DecryptionTrigger{
		BlockNumber:       uint64(commitment.BlockNumber),
		IdentityPreimages: identityPreimages,
	}

	blockNumbers := make([]int64, 0, len(commitment.Identities))
	eons := make([]int64, 0, len(commitment.Identities))

	obsKeyperDB := corekeyperdatabase.New(h.dbpool)
	eon, err := obsKeyperDB.GetEonForBlockNumber(ctx, int64(commitment.BlockNumber))
	if err != nil {
		return nil, err
	}
	for range commitment.Identities {
		blockNumbers = append(blockNumbers, int64(commitment.BlockNumber))
		eons = append(eons, eon.Eon)
	}

	db := database.New(h.dbpool)
	err = db.InsertMultipleTransactionsAndUpsertCommitment(ctx, database.InsertMultipleTransactionsAndUpsertCommitmentParams{
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

	//TODO: before sending the dec trigger, we need to check if majority of providers have generated commitments

	h.decryptionTriggerChannel <- broker.NewEvent(decryptionTrigger)

	return nil, nil
}

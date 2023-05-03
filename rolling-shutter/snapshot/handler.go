package snapshot

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/snpdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func NewTimedEpochHandler(config Config, snapshot *Snapshot) p2p.MessageHandler {
	return &TimedEpochHandler{config: config, snapshot: snapshot}
}

type TimedEpochHandler struct {
	config   Config
	snapshot *Snapshot
}

func NewDecryptionKeyHandler(config Config, snapshot *Snapshot) p2p.MessageHandler {
	return &DecryptionKeyHandler{config: config, snapshot: snapshot}
}

type DecryptionKeyHandler struct {
	config   Config
	snapshot *Snapshot
}

func NewEonPublicKeyHandler(config Config, snapshot *Snapshot) p2p.MessageHandler {
	return &EonPublicKeyHandler{config: config, snapshot: snapshot}
}

type EonPublicKeyHandler struct {
	config   Config
	snapshot *Snapshot
	dbpool   *pgxpool.Pool
}

func (*TimedEpochHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionTrigger{}}
}

func (*DecryptionKeyHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionTrigger{}}
}

func (*EonPublicKeyHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionTrigger{}}
}

func (handler *DecryptionKeyHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (bool, error) {
	decryptionKeyMsg := msg.(*p2pmsg.DecryptionKey)
	//FIXME: check snapshot business logic for decryptionKeyMsg validation
	if decryptionKeyMsg.GetInstanceID() != handler.config.GetInstanceID() {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.GetInstanceID(), decryptionKeyMsg.GetInstanceID())
	}

	key, err := decryptionKeyMsg.GetEpochSecretKey()

	if err != nil {
		return false, errors.Wrapf(err, "error getting epochSecretKey at epoch: %d", decryptionKeyMsg.EpochID)
	}

	// FIXME: unnecessary GobEncode?
	_, err = key.GobEncode()
	if err != nil {
		return false, errors.Wrap(err, "failed to encode decryption key")
	}

	return true, nil
}

func (handler *TimedEpochHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (bool, error) {
	//FIXME: add TimedEpoch validation
	timedEpochMsg := msg.(*p2pmsg.TimedEpoch)
	if timedEpochMsg.GetInstanceID() != handler.config.GetInstanceID() {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.GetInstanceID(), timedEpochMsg.GetInstanceID())
	}
	return true, nil
}

func (handler *EonPublicKeyHandler) ValidateMessage(ctx context.Context, msg p2pmsg.Message) (bool, error) {
	eonKeyMsg := msg.(*p2pmsg.EonPublicKey)
	if eonKeyMsg.GetInstanceID() != handler.config.GetInstanceID() {
		return false, errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.GetInstanceID(), eonKeyMsg.GetInstanceID())
	}
	eon := eonKeyMsg.GetEon()
	if eon == 0 {
		return false, errors.Errorf("failed to get eon public key from P2P message")
	}
	return true, nil
}

func (handler *DecryptionKeyHandler) HandleMessage(ctx context.Context, m p2pmsg.Message) ([]p2pmsg.Message, error) {
	var result []p2pmsg.Message
	key := m.(*p2pmsg.DecryptionKey)
	_, seen := seenProposals[string(key.EpochID)]
	if seen {
		return result, nil
	}
	log.Printf("Sending key %X for proposal %X to hub", key.Key, key.EpochID)

	metricKeysGenerated.Inc()

	err := handler.snapshot.hubapi.SubmitProposalKey(key.EpochID, key.Key)
	if err != nil {
		return result, err
	}
	// FIXME: Apart from needing to be in DB we need to keep track of the proposals better
	seenProposals[string(key.EpochID)] = struct{}{}
	return result, nil
}

func (handler *EonPublicKeyHandler) HandleMessage(ctx context.Context, m p2pmsg.Message) ([]p2pmsg.Message, error) {
	eonPubKeyMsg := m.(*p2pmsg.EonPublicKey)

	eonId := eonPubKeyMsg.GetEon()
	key := eonPubKeyMsg.GetPublicKey()
	db := snpdb.New(handler.dbpool)
	err := db.InsertEonPublicKey(
		ctx, snpdb.InsertEonPublicKeyParams{
			EonID:        int64(eonId),
			EonPublicKey: key,
		},
	)
	if err != nil {
		return nil, err
	}
	_, seen := seenEons[eonId]
	if seen {
		return nil, nil
	}

	metricEons.Inc()

	log.Printf("Sending Eon %d public key to hub", eonId)
	err = handler.snapshot.hubapi.SubmitEonKey(eonId, key)
	if err != nil {
		return nil, err
	}
	seenEons[eonId] = struct{}{}

	return nil, nil
}

func (handler *TimedEpochHandler) HandleMessage(ctx context.Context, m p2pmsg.Message) ([]p2pmsg.Message, error) {
	var result []p2pmsg.Message
	//FIXME: add TimedEpoch handling logic
	return result, nil
}

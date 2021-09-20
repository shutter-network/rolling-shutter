package keyper

import (
	"context"
	"crypto/ed25519"
	"log"
	"reflect"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/keyper/shutterevents"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

// ShuttermintState contains our view of the remote shutter state. Strictly speaking everything is
// stored in the database, and what we have here is kind of a cache.
type ShuttermintState struct {
	config       Config
	synchronized bool // are we synchronized
	isKeyper     bool
}

func NewShuttermintState(config Config) *ShuttermintState {
	return &ShuttermintState{config: config}
}

// Invalidate invalidates the current state. This is being called, when an error happens.
func (st *ShuttermintState) Invalidate() {
	*st = *NewShuttermintState(st.config)
}

func (st *ShuttermintState) LoadAppState(ctx context.Context, queries *kprdb.Queries) error {
	if st.synchronized {
		return nil
	}
	numBatchConfigs, err := queries.CountBatchConfigs(ctx)
	if err != nil {
		return err
	}
	st.isKeyper = numBatchConfigs > 0
	st.synchronized = true
	return nil
}

func (*ShuttermintState) StoreAppState(ctx context.Context, queries *kprdb.Queries) error {
	_ = ctx
	_ = queries
	return nil
}

func (st *ShuttermintState) sendShuttermintMessage(description string, msg *shmsg.Message) {
	// TODO
	_ = description
	_ = msg
	log.Printf("SEND SHUTTERMINT MESSAGE: %s", description)
}

func (st *ShuttermintState) handleBatchConfig(ctx context.Context, queries *kprdb.Queries, e *shutterevents.BatchConfig) error {
	if !st.isKeyper {
		if !e.IsKeyper(st.config.Address()) {
			return nil
		}
		st.isKeyper = true
		st.sendShuttermintMessage(
			"check-in",
			shmsg.NewCheckIn(
				st.config.ValidatorKey.Public().(ed25519.PublicKey),
				&st.config.EncryptionKey.PublicKey,
			),
		)
	}
	keypers := []string{}
	for _, k := range e.Keypers {
		keypers = append(keypers, k.Hex())
	}
	return queries.InsertBatchConfig(
		ctx,
		kprdb.InsertBatchConfigParams{
			ConfigIndex: int32(e.ConfigIndex),
			Height:      e.Height,
			Threshold:   int32(e.Threshold),
			Keypers:     keypers,
		},
	)
}

func (st *ShuttermintState) handleEonStarted(ctx context.Context, queries *kprdb.Queries, e *shutterevents.EonStarted) error {
	if !st.isKeyper {
		return nil
	}
	bc, err := queries.GetBatchConfig(ctx, int32(e.ConfigIndex))
	if err != nil {
		return err
	}
	keyperIndex, ok := bc.KeyperIndex(st.config.Address())
	if !ok {
		return nil
	}
	_ = puredkg.NewPureDKG(e.Eon, uint64(len(bc.Keypers)), uint64(bc.Threshold), keyperIndex)
	return nil
}

func (st *ShuttermintState) HandleEvent(ctx context.Context, queries *kprdb.Queries, event shutterevents.IEvent) error {
	var err error
	switch e := event.(type) {
	// case *shutterevents.CheckIn:
	//	// err = shutter.applyCheckIn(*e)

	case *shutterevents.BatchConfig:
		err = st.handleBatchConfig(ctx, queries, e)
	// case *shutterevents.DecryptionSignature:
	//	//err = shutter.applyDecryptionSignature(*e)
	case *shutterevents.EonStarted:
		err = st.handleEonStarted(ctx, queries, e)
		// err = shutter.applyEonStarted(*e)
	// case *shutterevents.PolyCommitment:
	//	//err = shutter.applyPolyCommitment(*e)
	// case *shutterevents.PolyEval:
	//	//err = shutter.applyPolyEval(*e)
	// case *shutterevents.Accusation:
	//	//err = shutter.applyAccusation(*e)
	// case *shutterevents.Apology:
	//	//err = shutter.applyApology(*e)
	// case *shutterevents.EpochSecretKeyShare:
	//	//err = shutter.applyEpochSecretKeyShare(*e)

	default:
		log.Printf("storeEvent not yet implemented for %s", reflect.TypeOf(event))
	}
	return err
}

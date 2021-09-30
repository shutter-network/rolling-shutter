package keyper

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/keyper/shutterevents"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

type ActiveDKG struct {
	pure        *puredkg.PureDKG
	startHeight int64
	dirty       bool
	keypers     []common.Address
}

// ShuttermintState contains our view of the remote shutter state. Strictly speaking everything is
// stored in the database, and what we have here is kind of a cache.
type ShuttermintState struct {
	config         Config
	synchronized   bool // are we synchronized
	isKeyper       bool
	encryptionKeys map[common.Address]*ecies.PublicKey
	dkg            map[uint64]*ActiveDKG
	phaseLength    PhaseLength
}

func NewShuttermintState(config Config) *ShuttermintState {
	return &ShuttermintState{
		config:         config,
		encryptionKeys: make(map[common.Address]*ecies.PublicKey),
		dkg:            make(map[uint64]*ActiveDKG),
		phaseLength:    NewConstantPhaseLength(int64(config.DKGPhaseLength)),
	}
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
	err = st.loadEncryptionKeys(ctx, queries)
	if err != nil {
		return err
	}
	st.synchronized = true
	return nil
}

func (st *ShuttermintState) loadEncryptionKeys(ctx context.Context, queries *kprdb.Queries) error {
	keys, err := queries.GetEncryptionKeys(ctx)
	if err != nil {
		return err
	}
	for _, k := range keys {
		addr, err := shdb.DecodeAddress(k.Address)
		if err != nil {
			return err
		}
		p, err := shdb.DecodeEciesPublicKey(k.EncryptionPublicKey)
		if err != nil {
			return err
		}
		st.encryptionKeys[addr] = p
	}
	return nil
}

func (st *ShuttermintState) sendPolyEvals(ctx context.Context, queries *kprdb.Queries) error {
	evals, err := queries.PolyEvalsWithEncryptionKeys(ctx)
	if err != nil {
		return err
	}
	for _, eval := range evals {
		fmt.Printf("SEND POLY EVALS: %#v", eval)
	}
	return nil
}

func (st *ShuttermintState) StoreAppState(ctx context.Context, queries *kprdb.Queries) error {
	err := st.sendPolyEvals(ctx, queries)
	if err != nil {
		return err
	}
	for eon, a := range st.dkg {
		if !a.dirty {
			continue
		}
		log.Printf("Storing dirty puredkg for eon %d", eon)
		pureBytes, err := shdb.EncodePureDKG(a.pure)
		if err != nil {
			return err
		}

		err = queries.InsertPureDKG(ctx, kprdb.InsertPureDKGParams{
			Eon:     int64(eon),
			Puredkg: pureBytes,
		})
		if err != nil {
			return err
		}
		a.dirty = false
	}
	return nil
}

func (st *ShuttermintState) scheduleShutterMessage(
	ctx context.Context,
	queries *kprdb.Queries,
	description string,
	msg *shmsg.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgid, err := queries.ScheduleShutterMessage(ctx, kprdb.ScheduleShutterMessageParams{
		Description: description,
		Msg:         data,
	})
	if err != nil {
		return err
	}
	log.Printf("scheduled shuttermint message: id=%d %s", msgid, description)
	return nil
}

func (st *ShuttermintState) handleBatchConfig(
	ctx context.Context, queries *kprdb.Queries, e *shutterevents.BatchConfig) error {
	if !st.isKeyper {
		if !e.IsKeyper(st.config.Address()) {
			return nil
		}
		st.isKeyper = true
		err := st.scheduleShutterMessage(
			ctx,
			queries,
			"check-in",
			shmsg.NewCheckIn(
				st.config.ValidatorKey.Public().(ed25519.PublicKey),
				&st.config.EncryptionKey.PublicKey,
			),
		)
		if err != nil {
			return err
		}
	}
	keypers := []string{}
	for _, k := range e.Keypers {
		keypers = append(keypers, shdb.EncodeAddress(k))
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
	err := queries.InsertEon(ctx, kprdb.InsertEonParams{
		Eon:         int64(e.Eon),
		Height:      e.Height,
		BatchIndex:  shdb.EncodeUint64(e.BatchIndex),
		ConfigIndex: int64(e.ConfigIndex),
	})
	if err != nil {
		return err
	}
	batchConfig, err := queries.GetBatchConfig(ctx, int32(e.ConfigIndex))
	if err != nil {
		return err
	}

	keypers := []common.Address{}
	for _, k := range batchConfig.Keypers {
		a, err := shdb.DecodeAddress(k)
		if err != nil {
			return err
		}
		keypers = append(keypers, a)
	}

	keyperIndex, err := medley.FindAddressIndex(keypers, st.config.Address())
	if err != nil {
		return nil
	}

	lastCommittedHeight, err := queries.GetLastCommittedHeight(ctx)
	if err != nil {
		return err
	}

	phase := st.phaseLength.getPhaseAtHeight(lastCommittedHeight+1, e.Height)
	if phase > puredkg.Dealing {
		log.Printf("Missed the dealing phase of eon %d", e.Eon)
		return nil
	}
	if phase == puredkg.Off {
		panic("phase is off")
	}

	pure := puredkg.NewPureDKG(e.Eon, uint64(len(batchConfig.Keypers)), uint64(batchConfig.Threshold), uint64(keyperIndex))
	st.dkg[e.Eon] = &ActiveDKG{
		pure:        &pure,
		dirty:       true,
		startHeight: e.Height,
		keypers:     keypers,
	}
	commitment, polyEvals, err := pure.StartPhase1Dealing()
	if err != nil {
		log.Fatalf("Aborting due to unexpected error: %+v", err)
	}

	err = st.scheduleShutterMessage(
		ctx,
		queries,
		fmt.Sprintf("poly commitment, eon=%d", e.Eon),
		shmsg.NewPolyCommitment(e.Eon, commitment.Gammas),
	)
	if err != nil {
		return err
	}

	for _, eval := range polyEvals {
		err = queries.InsertPolyEval(ctx, kprdb.InsertPolyEvalParams{
			Eon:             int64(e.Eon),
			ReceiverAddress: batchConfig.Keypers[eval.Receiver],
			Eval:            shdb.EncodeBigint(eval.Eval),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *ShuttermintState) handleCheckIn(ctx context.Context, queries *kprdb.Queries, e *shutterevents.CheckIn) error {
	st.encryptionKeys[e.Sender] = e.EncryptionPublicKey
	err := queries.InsertEncryptionKey(ctx, kprdb.InsertEncryptionKeyParams{
		Address:             shdb.EncodeAddress(e.Sender),
		EncryptionPublicKey: shdb.EncodeEciesPublicKey(e.EncryptionPublicKey),
	})
	return err
}

func (st *ShuttermintState) handlePolyCommitment(_ context.Context, _ *kprdb.Queries, e *shutterevents.PolyCommitment) error {
	dkg, ok := st.dkg[e.Eon]
	if !ok {
		log.Printf("PolyCommitment for non existent eon received: eon=%d commitment=%#v", e.Eon, e)
		return nil
	}
	senderIndex, err := medley.FindAddressIndex(dkg.keypers, e.Sender)
	if err != nil {
		log.Printf(
			"Received PolyCommitment from non keyper address: eon=%d sender=%s",
			e.Eon,
			e.Sender.Hex(),
		)
		return nil
	}

	err = dkg.pure.HandlePolyCommitmentMsg(puredkg.PolyCommitmentMsg{
		Eon:    e.Eon,
		Sender: uint64(senderIndex),
		Gammas: e.Gammas,
	})
	if err != nil {
		log.Printf("Error handling PolyCommitment: %s", err)
		return nil
	}
	dkg.dirty = true
	return nil
}

func (st *ShuttermintState) HandleEvent(ctx context.Context, queries *kprdb.Queries, event shutterevents.IEvent) error {
	var err error
	switch e := event.(type) {
	case *shutterevents.CheckIn:
		err = st.handleCheckIn(ctx, queries, e)
	case *shutterevents.BatchConfig:
		err = st.handleBatchConfig(ctx, queries, e)
	// case *shutterevents.DecryptionSignature:
	//	//err = shutter.applyDecryptionSignature(*e)
	case *shutterevents.EonStarted:
		err = st.handleEonStarted(ctx, queries, e)
	case *shutterevents.PolyCommitment:
		err = st.handlePolyCommitment(ctx, queries, e)
	// case *shutterevents.PolyEval:
	//	//err = shutter.applyPolyEval(*e)
	// case *shutterevents.Accusation:
	//	//err = shutter.applyAccusation(*e)
	// case *shutterevents.Apology:
	//	//err = shutter.applyApology(*e)
	// case *shutterevents.EpochSecretKeyShare:
	//	//err = shutter.applyEpochSecretKeyShare(*e)

	default:
		log.Printf("handleEvent not yet implemented for %s", reflect.TypeOf(event))
	}
	return err
}

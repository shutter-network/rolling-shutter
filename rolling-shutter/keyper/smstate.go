package keyper

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
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

func (dkg *ActiveDKG) markDirty() {
	dkg.dirty = true
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

	var receivers []common.Address
	var encryptedEvals [][]byte
	currentEon := int64(-1)

	send := func() error {
		if len(encryptedEvals) == 0 {
			return nil
		}
		defer func() {
			receivers = nil
			encryptedEvals = nil
		}()
		return st.scheduleShutterMessage(
			ctx,
			queries,
			fmt.Sprintf("poly eval eon=%d", currentEon),
			shmsg.NewPolyEval(uint64(currentEon), receivers, encryptedEvals),
		)
	}

	for _, eval := range evals {
		if eval.Eon != currentEon {
			err = send()
			if err != nil {
				return err
			}
			currentEon = eval.Eon
		}

		fmt.Printf("SEND POLY EVALS: eon=%d receiver=%s\n", eval.Eon, eval.ReceiverAddress)
		receiver, err := shdb.DecodeAddress(eval.ReceiverAddress)
		if err != nil {
			return err
		}
		pubkey, ok := st.encryptionKeys[receiver]
		if !ok {
			panic("key not loaded into ShuttermintState")
		}
		encrypted, err := ecies.Encrypt(rand.Reader, pubkey, eval.Eval, nil, nil)
		if err != nil {
			return err
		}
		receivers = append(receivers, receiver)
		encryptedEvals = append(encryptedEvals, encrypted)

		err = queries.DeletePolyEval(ctx, kprdb.DeletePolyEvalParams{
			Eon:             eval.Eon,
			ReceiverAddress: eval.ReceiverAddress,
		})
		if err != nil {
			return err
		}
	}
	return send()
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

func (st *ShuttermintState) handleEonStarted(
	ctx context.Context, queries *kprdb.Queries, e *shutterevents.EonStarted) error {
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

	pure := puredkg.NewPureDKG(
		e.Eon,
		uint64(len(batchConfig.Keypers)),
		uint64(batchConfig.Threshold),
		uint64(keyperIndex),
	)
	st.dkg[e.Eon] = &ActiveDKG{
		pure:        &pure,
		dirty:       true,
		startHeight: e.Height,
		keypers:     keypers,
	}
	return st.shiftPhase(ctx, queries, e.Height, e.Eon, st.dkg[e.Eon])
}

func (st *ShuttermintState) startPhase1Dealing(
	ctx context.Context, queries *kprdb.Queries, eon uint64, dkg *ActiveDKG) error {
	pure := dkg.pure
	commitment, polyEvals, err := pure.StartPhase1Dealing()
	if err != nil {
		log.Fatalf("Aborting due to unexpected error: %+v", err)
	}
	dkg.markDirty()
	err = st.scheduleShutterMessage(
		ctx,
		queries,
		fmt.Sprintf("poly commitment, eon=%d", eon),
		shmsg.NewPolyCommitment(eon, commitment.Gammas),
	)
	if err != nil {
		return err
	}

	for _, eval := range polyEvals {
		err = queries.InsertPolyEval(ctx, kprdb.InsertPolyEvalParams{
			Eon:             int64(eon),
			ReceiverAddress: shdb.EncodeAddress(dkg.keypers[eval.Receiver]),
			Eval:            shdb.EncodeBigint(eval.Eval),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *ShuttermintState) startPhase2Accusing(
	ctx context.Context, queries *kprdb.Queries, eon uint64, dkg *ActiveDKG) error {
	accusations := dkg.pure.StartPhase2Accusing()
	dkg.markDirty()
	if len(accusations) > 0 {
		var accused []common.Address
		for _, a := range accusations {
			accused = append(accused, dkg.keypers[a.Accused])
		}
		err := st.scheduleShutterMessage(
			ctx,
			queries,
			fmt.Sprintf("accusations, eon=%d, count=%d", eon, len(accusations)),
			shmsg.NewAccusation(eon, accused),
		)
		if err != nil {
			return err
		}
	} else {
		log.Printf("No one to accuse in eon %d", eon)
	}

	return nil
}

func (st *ShuttermintState) startPhase3Apologizing(
	ctx context.Context, queries *kprdb.Queries, eon uint64, dkg *ActiveDKG) error {
	apologies := dkg.pure.StartPhase3Apologizing()
	dkg.markDirty()
	if len(apologies) > 0 {
		var accusers []common.Address
		var polyEvals []*big.Int

		for _, a := range apologies {
			accusers = append(accusers, dkg.keypers[a.Accuser])
			polyEvals = append(polyEvals, a.Eval)
		}

		err := st.scheduleShutterMessage(
			ctx, queries,
			fmt.Sprintf("apologies, eon=%d, count=%d", eon, len(apologies)),
			shmsg.NewApology(eon, accusers, polyEvals),
		)
		if err != nil {
			return err
		}
	} else {
		log.Printf("No apologies needed in eon %d", eon)
	}

	return nil
}

func (st *ShuttermintState) finalizeDKG(
	ctx context.Context, queries *kprdb.Queries, eon uint64, dkg *ActiveDKG) error {
	dkg.pure.Finalize()
	dkg.markDirty()

	var dkgerror sql.NullString
	var pureResult []byte

	dkgresult, err := dkg.pure.ComputeResult()
	if err != nil {
		log.Printf("Error: DKG process failed for eon %d: %s", eon, err)
		dkgerror = sql.NullString{String: err.Error(), Valid: true}
		// st.scheduleShutterMessage(
		//	ctx, queries,
		//	"requesting DKG restart",
		//	// shmsg.NewEonStartVote(dkg.StartBatchIndex),
		//	shmsg.NewEonStartVote(dkg.StartBatchIndex),
		// )
	} else {
		log.Printf("Success: DKG process succeeded for eon %d", eon)
		pureResult, err = shdb.EncodePureDKGResult(&dkgresult)
		if err != nil {
			return err
		}
	}

	return queries.InsertDKGResult(ctx, kprdb.InsertDKGResultParams{
		Eon:        int64(eon),
		Success:    pureResult != nil,
		Error:      dkgerror,
		PureResult: pureResult,
	})
}

func (st *ShuttermintState) shiftPhase(
	ctx context.Context, queries *kprdb.Queries, height int64, eon uint64, dkg *ActiveDKG) error {
	phase := st.phaseLength.getPhaseAtHeight(height, dkg.startHeight)
	for currentPhase := dkg.pure.Phase; currentPhase < phase; currentPhase = dkg.pure.Phase {
		log.Printf(
			"Phase transition eon=%d, height=%d %s->%s",
			eon, height, currentPhase, currentPhase+1,
		)

		var err error
		switch currentPhase {
		case puredkg.Off:
			err = st.startPhase1Dealing(ctx, queries, eon, dkg)
		case puredkg.Dealing:
			err = st.startPhase2Accusing(ctx, queries, eon, dkg)
		case puredkg.Accusing:
			err = st.startPhase3Apologizing(ctx, queries, eon, dkg)
		case puredkg.Apologizing:
			err = st.finalizeDKG(ctx, queries, eon, dkg)
		case puredkg.Finalized:
			panic("internal error, currentPhase is Finalized")
		}
		if err != nil {
			return err
		}
		if dkg.pure.Phase == currentPhase {
			panic("phase did not change")
		}
	}
	return nil
}

func (st *ShuttermintState) shiftPhases(
	ctx context.Context, queries *kprdb.Queries, height int64) error {
	for eon, dkg := range st.dkg {
		err := st.shiftPhase(ctx, queries, height, eon, dkg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *ShuttermintState) handleCheckIn(
	ctx context.Context, queries *kprdb.Queries, e *shutterevents.CheckIn) error {
	st.encryptionKeys[e.Sender] = e.EncryptionPublicKey
	err := queries.InsertEncryptionKey(ctx, kprdb.InsertEncryptionKeyParams{
		Address:             shdb.EncodeAddress(e.Sender),
		EncryptionPublicKey: shdb.EncodeEciesPublicKey(e.EncryptionPublicKey),
	})
	return err
}

func (st *ShuttermintState) handlePolyCommitment(
	_ context.Context, _ *kprdb.Queries, e *shutterevents.PolyCommitment) error { //nolint:unparam
	dkg, ok := st.dkg[e.Eon]
	if !ok {
		log.Printf("PolyCommitment for non existent eon received: eon=%d commitment=%#v",
			e.Eon, e)
		return nil
	}
	senderIndex, err := medley.FindAddressIndex(dkg.keypers, e.Sender)
	if err != nil {
		log.Printf("Received PolyCommitment from non keyper address: eon=%d sender=%s",
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
	dkg.markDirty()
	return nil
}

func (st *ShuttermintState) decryptPolyEval(encrypted []byte) ([]byte, error) {
	return st.config.EncryptionKey.Decrypt(encrypted, []byte(""), []byte(""))
}

func (st *ShuttermintState) handlePolyEval(
	_ context.Context, _ *kprdb.Queries, e *shutterevents.PolyEval) error {
	myAddress := st.config.Address()
	if e.Sender == myAddress {
		return nil
	}

	dkg, ok := st.dkg[e.Eon]
	if !ok {
		log.Printf("PolyEval for non existent eon received: %s", e)
		return nil
	}
	sender, err := medley.FindAddressIndex(dkg.keypers, e.Sender)
	if err != nil {
		return nil
	}
	keyperIndex, err := medley.FindAddressIndex(dkg.keypers, myAddress)
	if err != nil {
		return err
	}

	myIndex, err := medley.FindAddressIndex(e.Receivers, myAddress)
	if err != nil {
		return nil
	}
	encrypted := e.EncryptedEvals[myIndex]

	evalBytes, err := st.decryptPolyEval(encrypted)
	if err != nil {
		log.Printf("Could not decrypt poly eval: %s", err)
		return nil
	}

	err = dkg.pure.HandlePolyEvalMsg(
		puredkg.PolyEvalMsg{
			Eon:      e.Eon,
			Sender:   uint64(sender),
			Receiver: uint64(keyperIndex),
			Eval:     new(big.Int).SetBytes(evalBytes),
		},
	)
	if err != nil {
		log.Printf("HandlePolyEvalMsg failed: %s", err)
		return nil
	}
	log.Printf("Got poly eval message: eon=%d, keyper=#%d (%s)", e.Eon, sender, e.Sender)
	dkg.markDirty()
	return nil
}

func (st *ShuttermintState) handleAccusation(
	_ context.Context, _ *kprdb.Queries, e *shutterevents.Accusation) error { //nolint:unparam
	dkg, ok := st.dkg[e.Eon]
	if !ok {
		log.Printf("Accusation for non existent eon received: %s", e)
		return nil
	}

	if dkg.pure.Phase != puredkg.Accusing {
		log.Printf("Warning: received accusation in wrong phase %s: %+v", dkg.pure.Phase, e)
		return nil
	}
	sender, err := medley.FindAddressIndex(dkg.keypers, e.Sender)
	if err != nil {
		log.Printf("Error: cannot handle accusation. bad sender: %s", e.Sender)
		return nil
	}

	for _, accused := range e.Accused {
		accusedIndex, err := medley.FindAddressIndex(dkg.keypers, accused)
		if err != nil {
			log.Printf("Error: accused address is not a keyper: %s", accused)
			continue
		}
		err = dkg.pure.HandleAccusationMsg(
			puredkg.AccusationMsg{
				Eon:     e.Eon,
				Accuser: uint64(sender),
				Accused: uint64(accusedIndex),
			},
		)
		if err != nil {
			log.Printf("Error: cannot handle accusation: %+v", err)
		}
	}
	dkg.markDirty()
	return nil
}

func (st *ShuttermintState) handleApology(
	_ context.Context, _ *kprdb.Queries, e *shutterevents.Apology) error { //nolint:unparam
	dkg, ok := st.dkg[e.Eon]
	if !ok {
		log.Printf("Apology for non existent eon received: %s", e)
		return nil
	}
	if dkg.pure.Phase != puredkg.Apologizing {
		log.Printf("Warning: received apology in wrong phase %s: %+v", dkg.pure.Phase, e)
		return nil
	}
	sender, err := medley.FindAddressIndex(dkg.keypers, e.Sender)
	if err != nil {
		log.Printf("Error: cannot handle apology. bad sender: %s", e.Sender)
		return nil
	}

	for j, accuser := range e.Accusers {
		accuserIndex, err := medley.FindAddressIndex(dkg.keypers, accuser)
		if err != nil {
			log.Printf("Error in syncApologies: %+v", err)
			continue
		}
		err = dkg.pure.HandleApologyMsg(
			puredkg.ApologyMsg{
				Eon:     e.Eon,
				Accuser: uint64(accuserIndex),
				Accused: uint64(sender),
				Eval:    e.PolyEval[j],
			})
		if err != nil {
			log.Printf("Error: cannot handle apology: %+v", err)
		}
	}
	dkg.markDirty()
	return nil
}

func (st *ShuttermintState) HandleEvent(
	ctx context.Context, queries *kprdb.Queries, event shutterevents.IEvent) error {
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
	case *shutterevents.PolyEval:
		err = st.handlePolyEval(ctx, queries, e)
	case *shutterevents.Accusation:
		err = st.handleAccusation(ctx, queries, e)
	case *shutterevents.Apology:
		err = st.handleApology(ctx, queries, e)
	default:
		log.Printf("handleEvent not yet implemented for %s: %s",
			reflect.TypeOf(event), event)
	}

	return err
}

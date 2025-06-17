package smobserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/icza/gog"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/dkgphase"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/keypermetrics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/shutterevents"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type Config interface {
	GetAddress() common.Address
	GetDKGPhaseLength() *dkgphase.PhaseLength
	GetValidatorPublicKey() ed25519.PublicKey
	GetEncryptionKey() *ecies.PrivateKey
}

type ActiveDKG struct {
	pure        *puredkg.PureDKG
	startHeight int64
	dirty       bool
	keypers     []common.Address
}

func (dkg *ActiveDKG) markDirty() {
	dkg.dirty = true
}

type PhaseLength interface {
	GetPhaseAtHeight(height int64, eonStartHeight int64) puredkg.Phase
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
		phaseLength:    config.GetDKGPhaseLength(),
	}
}

// Invalidate invalidates the current state. This is being called, when an error happens.
func (st *ShuttermintState) Invalidate() {
	*st = *NewShuttermintState(st.config)
}

func (st *ShuttermintState) Load(ctx context.Context, queries *database.Queries) error {
	if st.synchronized {
		return nil
	}
	numBatchConfigs, err := queries.CountBatchConfigs(ctx)
	if err != nil {
		return err
	}
	st.isKeyper = numBatchConfigs > 0 // XXX need to look
	err = st.loadEncryptionKeys(ctx, queries)
	if err != nil {
		return err
	}
	err = st.loadDKG(ctx, queries)
	if err != nil {
		return err
	}

	st.synchronized = true
	return nil
}

func (st *ShuttermintState) loadDKG(ctx context.Context, queries *database.Queries) error {
	dkgs, err := queries.SelectPureDKG(ctx)
	if err != nil {
		return err
	}
	for _, dkg := range dkgs {
		pure, err := shdb.DecodePureDKG(dkg.Puredkg)
		if err != nil {
			return err
		}

		keyperEon, err := queries.GetEon(ctx, dkg.Eon)
		if err != nil {
			return err
		}

		batchConfig, err := queries.GetBatchConfig(ctx, int32(keyperEon.KeyperConfigIndex))
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

		st.dkg[uint64(dkg.Eon)] = &ActiveDKG{
			pure:        pure,
			startHeight: keyperEon.Height,
			dirty:       false,
			keypers:     keypers,
		}
	}
	return nil
}

func (st *ShuttermintState) loadEncryptionKeys(ctx context.Context, queries *database.Queries) error {
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

func (st *ShuttermintState) BeforeSaveHook(ctx context.Context, queries *database.Queries) error {
	return st.sendPolyEvals(ctx, queries)
}

func (st *ShuttermintState) Save(ctx context.Context, queries *database.Queries) error {
	for eon, a := range st.dkg {
		if !a.dirty {
			continue
		}
		pureBytes, err := shdb.EncodePureDKG(a.pure)
		if err != nil {
			return err
		}

		err = queries.InsertPureDKG(ctx, database.InsertPureDKGParams{
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

func (st *ShuttermintState) sendPolyEvals(ctx context.Context, queries *database.Queries) error {
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
		return queries.ScheduleShutterMessage(
			ctx,
			fmt.Sprintf("poly eval (eon=%d)", currentEon),
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

		err = queries.DeletePolyEval(ctx, database.DeletePolyEvalParams{
			Eon:             eval.Eon,
			ReceiverAddress: eval.ReceiverAddress,
		})
		if err != nil {
			return err
		}
	}
	return send()
}

func (st *ShuttermintState) handleBatchConfig(
	ctx context.Context, queries *database.Queries, e *shutterevents.BatchConfig,
) error {
	if !e.IsKeyper(st.config.GetAddress()) {
		keypermetrics.MetricsKeyperIsKeyper.WithLabelValues(strconv.FormatUint(e.KeyperConfigIndex, 10)).Set(0)
	}
	if e.IsKeyper(st.config.GetAddress()) {
		// In case we transition to a superset of the current Keyper set or this node was a Keyper before in an older set
		// the check-in message will be a duplicate, but this isn't a problem, it will be ignored.
		st.isKeyper = true
		pubKey := st.config.GetValidatorPublicKey()
		err := queries.ScheduleShutterMessage(
			ctx,
			fmt.Sprintf("check-in (validator-pub-key=%s)", hex.EncodeToString(pubKey)),
			shmsg.NewCheckIn(
				pubKey,
				&st.config.GetEncryptionKey().PublicKey,
			),
		)
		if err != nil {
			return err
		}
		keypermetrics.MetricsKeyperIsKeyper.WithLabelValues(strconv.FormatUint(e.KeyperConfigIndex, 10)).Set(1)
	}
	keypers := []string{}
	for _, k := range e.Keypers {
		keypers = append(keypers, shdb.EncodeAddress(k))
	}
	keypermetrics.MetricsKeyperBatchConfigInfo.WithLabelValues(strconv.FormatUint(e.KeyperConfigIndex, 10), strings.Join(keypers, ",")).Set(1)
	if err := queries.InsertBatchConfig(
		ctx,
		database.InsertBatchConfigParams{
			KeyperConfigIndex:     int32(e.KeyperConfigIndex),
			Height:                e.Height,
			Threshold:             int32(e.Threshold),
			Keypers:               keypers,
			Started:               e.Started,
			ActivationBlockNumber: int64(e.ActivationBlockNumber),
		},
	); err != nil {
		return err
	}

	return queries.DeleteShutterMessageByDesc(ctx, fmt.Sprintf("new batch config (activation-block-number=%d, config-index=%d)",
		e.ActivationBlockNumber, e.KeyperConfigIndex))
}

func (st *ShuttermintState) handleBatchConfigStarted(
	ctx context.Context,
	queries *database.Queries,
	e *shutterevents.BatchConfigStarted,
) error {
	eon, err := queries.GetLatestEonForKeyperConfig(ctx, int64(e.KeyperConfigIndex))
	if err != nil {
		log.Warn().Uint64("keyperConfig", e.KeyperConfigIndex).Err(err).Msg("Couldn't get latest eon for keyper config index")
	} else {
		keypermetrics.MetricsKeyperCurrentEon.Set(float64(eon))
	}
	keypermetrics.MetricsKeyperCurrentBatchConfigIndex.Set(float64(e.KeyperConfigIndex))
	return queries.SetBatchConfigStarted(ctx, int32(e.KeyperConfigIndex))
}

func (st *ShuttermintState) handleEonStarted(
	ctx context.Context, queries *database.Queries, e *shutterevents.EonStarted,
) error {
	if e.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf("activation block number %d of eon start would overflow int64", e.ActivationBlockNumber)
	}
	err := queries.InsertEon(ctx, database.InsertEonParams{
		Eon:                   int64(e.Eon),
		Height:                e.Height,
		ActivationBlockNumber: int64(e.ActivationBlockNumber),
		KeyperConfigIndex:     int64(e.KeyperConfigIndex),
	})
	if err != nil {
		return err
	}

	if !st.isKeyper {
		return nil
	}

	batchConfig, err := queries.GetBatchConfig(ctx, int32(e.KeyperConfigIndex))
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

	keyperIndex, err := medley.FindAddressIndex(keypers, st.config.GetAddress())
	if err != nil {
		return nil
	}

	keypermetrics.MetricsKeyperEonStartBlock.WithLabelValues(strconv.FormatUint(e.Eon, 10)).Set(float64(e.ActivationBlockNumber))

	lastCommittedHeight, err := queries.GetLastCommittedHeight(ctx)
	if err != nil {
		return err
	}

	phase := st.phaseLength.GetPhaseAtHeight(lastCommittedHeight+1, e.Height)
	if phase > puredkg.Dealing {
		log.Info().Uint64("eon", e.Eon).Msg("missed the dealing phase")
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
	dkg := &ActiveDKG{
		pure:        &pure,
		dirty:       true,
		startHeight: e.Height,
		keypers:     keypers,
	}
	st.dkg[e.Eon] = dkg
	return st.shiftPhase(ctx, queries, e.Height, e.Eon, dkg)
}

func (st *ShuttermintState) startPhase1Dealing(
	ctx context.Context, queries *database.Queries, eon uint64, dkg *ActiveDKG,
) error {
	pure := dkg.pure
	commitment, polyEvals, err := pure.StartPhase1Dealing()
	if err != nil {
		log.Fatal().Err(err).Msg("aborting due to unexpected error")
	}
	dkg.markDirty()
	err = queries.ScheduleShutterMessage(
		ctx,
		fmt.Sprintf("poly commitment (eon=%d)", eon),
		shmsg.NewPolyCommitment(eon, commitment.Gammas),
	)
	if err != nil {
		return err
	}

	for _, eval := range polyEvals {
		err = queries.InsertPolyEval(ctx, database.InsertPolyEvalParams{
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
	ctx context.Context, queries *database.Queries, eon uint64, dkg *ActiveDKG,
) error {
	accusations := dkg.pure.StartPhase2Accusing()
	dkg.markDirty()
	if len(accusations) > 0 {
		var accused []common.Address
		for _, a := range accusations {
			accused = append(accused, dkg.keypers[a.Accused])
		}
		err := queries.ScheduleShutterMessage(
			ctx,
			fmt.Sprintf("accusations (eon=%d, count=%d)", eon, len(accusations)),
			shmsg.NewAccusation(eon, accused),
		)
		if err != nil {
			return err
		}
	} else {
		log.Info().Uint64("eon", eon).Msg("no one to accuse")
	}

	return nil
}

func (st *ShuttermintState) startPhase3Apologizing(
	ctx context.Context, queries *database.Queries, eon uint64, dkg *ActiveDKG,
) error {
	apologies := dkg.pure.StartPhase3Apologizing()
	dkg.markDirty()
	if len(apologies) > 0 {
		var accusers []common.Address
		var polyEvals []*big.Int

		for _, a := range apologies {
			accusers = append(accusers, dkg.keypers[a.Accuser])
			polyEvals = append(polyEvals, a.Eval)
		}

		err := queries.ScheduleShutterMessage(
			ctx,
			fmt.Sprintf("apologies (eon=%d, count=%d)", eon, len(apologies)),
			shmsg.NewApology(eon, accusers, polyEvals),
		)
		if err != nil {
			return err
		}
	} else {
		log.Info().Uint64("eon", eon).Msg("no apologies needed")
	}

	return nil
}

func (st *ShuttermintState) finalizeDKG(
	ctx context.Context, queries *database.Queries, eon uint64, dkg *ActiveDKG,
) error {
	dkg.pure.Finalize()
	// There's no need to call dkg.markDirty() here, since we now remove the object from memory
	// and the database:
	delete(st.dkg, eon)
	err := queries.DeletePureDKG(ctx, int64(eon))
	if err != nil {
		return err
	}

	// There's really no need to keep them around now. We could even get rid of them earlier.
	tag, err := queries.DeletePolyEvalByEon(ctx, int64(eon))
	if err != nil {
		return err
	}
	log.Info().Int64("count", tag.RowsAffected()).Msg("deleted poly evals")

	var dkgerror sql.NullString
	var pureResult []byte

	dkgresult, err := dkg.pure.ComputeResult()

	dkgresultmsg := shmsg.NewDKGResult(eon, err == nil)

	if err != nil {
		keypermetrics.MetricsKeyperSuccessfullDKG.WithLabelValues(strconv.FormatInt(int64(eon), 10)).Set(0)
		log.Error().Err(err).Uint64("eon", eon).Bool("success", false).
			Msg("DKG process failed")
		dkgerror = sql.NullString{String: err.Error(), Valid: true}
		_, err := queries.GetEon(ctx, int64(eon))
		if err != nil {
			return err
		}
	} else {
		keypermetrics.MetricsKeyperSuccessfullDKG.WithLabelValues(strconv.FormatInt(int64(eon), 10)).Set(1)
		log.Info().Uint64("eon", eon).Bool("success", true).Msg("DKG process succeeded")
		pureResult, err = shdb.EncodePureDKGResult(&dkgresult)
		if err != nil {
			return err
		}
		publicKeyBytes, _ := dkgresult.PublicKey.GobEncode()
		err := queries.InsertEonPublicKey(ctx, database.InsertEonPublicKeyParams{EonPublicKey: publicKeyBytes, Eon: int64(dkgresult.Eon)})
		if err != nil {
			return err
		}
	}

	err = queries.ScheduleShutterMessage(
		ctx,
		fmt.Sprintf("reporting DKG result (eon=%d)", dkgresult.Eon),
		dkgresultmsg,
	)
	if err != nil {
		return err
	}

	return queries.InsertDKGResult(ctx, database.InsertDKGResultParams{
		Eon:        int64(eon),
		Success:    pureResult != nil,
		Error:      dkgerror,
		PureResult: pureResult,
	})
}

func (st *ShuttermintState) shiftPhase(
	ctx context.Context, queries *database.Queries, height int64, eon uint64, dkg *ActiveDKG,
) error {
	phase := st.phaseLength.GetPhaseAtHeight(height, dkg.startHeight)
	for currentPhase := dkg.pure.Phase; currentPhase < phase; currentPhase = dkg.pure.Phase {
		log.Info().
			Uint64("eon", eon).
			Int64("height", height).
			Str("phaseAtHeight", phase.String()).
			Str("current-phase", currentPhase.String()).
			Str("next-phase", (currentPhase + 1).String()).
			Msg("phase transition")

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
		for i := 0; i <= int(puredkg.Finalized); i++ { // <- WTF Go
			keypermetrics.MetricsKeyperCurrentPhase.
				WithLabelValues(
					strconv.FormatUint(eon, 10),
					fmt.Sprintf("%d-%s", i, puredkg.Phase(i).String())).
				Set(gog.If[float64](int(dkg.pure.Phase) == i, 1, 0))
		}
	}
	return nil
}

func (st *ShuttermintState) shiftPhases(
	ctx context.Context, queries *database.Queries, height int64,
) error {
	for eon, dkg := range st.dkg {
		err := st.shiftPhase(ctx, queries, height, eon, dkg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *ShuttermintState) handleCheckIn(
	ctx context.Context, queries *database.Queries, e *shutterevents.CheckIn,
) error {
	st.encryptionKeys[e.Sender] = e.EncryptionPublicKey
	err := queries.InsertEncryptionKey(ctx, database.InsertEncryptionKeyParams{
		Address:             shdb.EncodeAddress(e.Sender),
		EncryptionPublicKey: shdb.EncodeEciesPublicKey(e.EncryptionPublicKey),
	})
	return err
}

func (st *ShuttermintState) handlePolyCommitment(
	_ context.Context, _ *database.Queries, e *shutterevents.PolyCommitment,
) error { //nolint:unparam
	dkg, ok := st.dkg[e.Eon]
	if !ok {
		log.Info().Str("event", e.String()).
			Msg("event for non existent eon received")
		return nil
	}
	senderIndex, err := medley.FindAddressIndex(dkg.keypers, e.Sender)
	if err != nil {
		log.Info().Str("event", e.String()).
			Msg("Received PolyCommitment from non keyper address")
		return nil
	}

	err = dkg.pure.HandlePolyCommitmentMsg(puredkg.PolyCommitmentMsg{
		Eon:    e.Eon,
		Sender: uint64(senderIndex),
		Gammas: e.Gammas,
	})
	if err != nil {
		log.Info().Str("event", e.String()).Err(err).
			Msg("failed to handle PolyCommitment")
		return nil
	}
	dkg.markDirty()
	return nil
}

func (st *ShuttermintState) decryptPolyEval(encrypted []byte) ([]byte, error) {
	return st.config.GetEncryptionKey().Decrypt(encrypted, []byte(""), []byte(""))
}

func (st *ShuttermintState) handlePolyEval(
	_ context.Context, _ *database.Queries, e *shutterevents.PolyEval,
) error {
	myAddress := st.config.GetAddress()
	if e.Sender == myAddress {
		return nil
	}

	dkg, ok := st.dkg[e.Eon]
	if !ok {
		log.Info().Str("event", e.String()).
			Msg("event for non existent eon received")
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
		log.Info().Str("event", e.String()).Err(err).
			Msg("could not decrypt poly eval")
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
		log.Info().Err(err).Msg("failed to handle PolyEvalMsg")
		return nil
	}
	log.Info().Str("event", e.String()).Int("keyper", sender).
		Msg("got poly eval message")
	dkg.markDirty()
	return nil
}

func (st *ShuttermintState) handleAccusation(
	_ context.Context, _ *database.Queries, e *shutterevents.Accusation,
) error { //nolint:unparam
	dkg, ok := st.dkg[e.Eon]
	if !ok {
		log.Info().Str("event", e.String()).
			Msg("event for non existent eon received")
		return nil
	}

	if dkg.pure.Phase != puredkg.Accusing {
		log.Warn().Str("event", e.String()).Str("phase", dkg.pure.Phase.String()).
			Msg("received accusation in wrong phase")
		return nil
	}
	sender, err := medley.FindAddressIndex(dkg.keypers, e.Sender)
	if err != nil {
		log.Info().Str("event", e.String()).
			Msg("cannot handle accusation. bad sender")
		return nil
	}

	for _, accused := range e.Accused {
		accusedIndex, err := medley.FindAddressIndex(dkg.keypers, accused)
		if err != nil {
			log.Info().Str("event", e.String()).Str("address", accused.Hex()).
				Msg("accused address is not a keyper")
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
			log.Info().Str("event", e.String()).Err(err).
				Msg("cannot handle accusation")
		}
	}
	dkg.markDirty()
	return nil
}

func (st *ShuttermintState) handleApology(
	_ context.Context, _ *database.Queries, e *shutterevents.Apology,
) error { //nolint:unparam
	dkg, ok := st.dkg[e.Eon]
	if !ok {
		log.Info().Str("event", e.String()).
			Msg("event for non existent eon received")
		return nil
	}
	if dkg.pure.Phase != puredkg.Apologizing {
		log.Warn().Str("phase", dkg.pure.Phase.String()).
			Msg("Warning: received apology in wrong phase")
		return nil
	}
	sender, err := medley.FindAddressIndex(dkg.keypers, e.Sender)
	if err != nil {
		log.Info().Str("event", e.String()).Msg("failed to handle apology. bad sender")
		return nil
	}

	for j, accuser := range e.Accusers {
		accuserIndex, err := medley.FindAddressIndex(dkg.keypers, accuser)
		if err != nil {
			log.Info().Str("event", e.String()).Str("address", accuser.Hex()).
				Msg("accuser address is not a keyper")
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
			log.Info().Str("event", e.String()).Err(err).Msg("failed to handle apology")
		}
	}
	dkg.markDirty()
	return nil
}

func (st *ShuttermintState) HandleEvent(
	ctx context.Context, queries *database.Queries, event shutterevents.IEvent,
) error {
	var err error
	log.Info().Str("event", event.String()).Msg("handle shuttermint event")
	switch e := event.(type) {
	case *shutterevents.CheckIn:
		err = st.handleCheckIn(ctx, queries, e)
	case *shutterevents.BatchConfig:
		err = st.handleBatchConfig(ctx, queries, e)
	case *shutterevents.BatchConfigStarted:
		err = st.handleBatchConfigStarted(ctx, queries, e)
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
		log.Warn().Str("type", reflect.TypeOf(event).String()).Interface("event", event).
			Msg("HandleEvent not yet implemented for event type")
	}

	return err
}

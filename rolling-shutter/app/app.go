package app

import (
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/go-amino"
	abcitypes "github.com/tendermint/tendermint/abci/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/shutterevents"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

var (
	// PersistMinDuration is the minimum duration between two calls to persistToDisk
	// TODO we should probably increase the default here and we should have a way to collect
	// garbage to keep the persisted state small enough.
	// The variable is declared here, because we do not want to persist it as part of the
	// application. The same could be said about the Gobpath field though, which we persist as
	// part of the application.
	// If we set this to zero, the state will get saved on every call to Commit.
	PersistMinDuration time.Duration = 30 * time.Second

	// NonExistentValidator is an artificial key used to replace the voting power of validators
	// that haven't sent their CheckIn message yet.
	NonExistentValidator ValidatorPubkey
)

func init() {
	gob.Register(crypto.S256()) // Allow gob to serialize ecsda.PrivateKey

	var err error
	k := [32]byte{'n', 'o', 'v', 'a', 'l', 'i', 'd', 'a', 't', 'o', 'r'}
	NonExistentValidator, err = NewValidatorPubkey(k[:])
	if err != nil {
		panic(err)
	}
}

// Visit https://github.com/tendermint/spec/blob/master/spec/abci/abci.md for more information on
// the application interface we're implementing here.
// https://docs.tendermint.com/master/spec/abci/apps.html also provides some useful information

// CheckTx checks if a transaction is valid. If return Code != 0, it will be rejected from the
// mempool and hence not broadcasted to other peers and not included in a proposal block.
func (app *ShutterApp) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	signer, msg, err := app.decodeTx(req.Tx)
	if err != nil {
		return abcitypes.ResponseCheckTx{Code: 1, Log: "cannot decode message"}
	}
	if string(msg.ChainId) != app.ChainID {
		return abcitypes.ResponseCheckTx{Code: 1, Log: "wrong chain"}
	}

	// Check that the message's nonce has not been used by the sender yet. Note that
	// `app.NonceTracker` keeps track of nonces of transactions included in the chain, while
	// `app.CheckTxState` keeps track of all transactions in the mempool checked since the last
	// `Commit`.
	if !app.NonceTracker.Check(signer, msg.RandomNonce) {
		return abcitypes.ResponseCheckTx{Code: 1, Log: "nonce already used"}
	}
	if !app.CheckTxState.AddTx(signer, msg) {
		return abcitypes.ResponseCheckTx{Code: 1, Log: "not a keyper set member"}
	}
	return abcitypes.ResponseCheckTx{Code: 0, GasWanted: 1}
}

// NewShutterApp creates a new ShutterApp.
func NewShutterApp() *ShutterApp {
	return &ShutterApp{
		Configs:      []*BatchConfig{{}},
		DKGMap:       make(map[uint64]*DKGInstance),
		ConfigVoting: NewConfigVoting(),
		Identities:   make(map[common.Address]ValidatorPubkey),
		BlocksSeen:   make(map[common.Address]uint64),
		CheckTxState: NewCheckTxState(),
		NonceTracker: NewNonceTracker(),
		ChainID:      "", // will be set in InitChain
	}
}

// LoadShutterAppFromFile loads a shutter app from a file.
func LoadShutterAppFromFile(gobpath string) (ShutterApp, error) {
	var shapp ShutterApp
	gobfile, err := os.Open(gobpath)
	if os.IsNotExist(err) {
		shapp = *NewShutterApp()
	} else if err != nil {
		return shapp, err
	} else {
		defer gobfile.Close()
		dec := gob.NewDecoder(gobfile)
		err = dec.Decode(&shapp)
		if err != nil {
			return shapp, err
		}
		log.Info().
			Str("file", gobpath).
			Time("last-saved", shapp.LastSaved).
			Int64("last-block-height", shapp.LastBlockHeight).
			Bool("devmode", shapp.DevMode).
			Msg("Loaded shutter app from file")
	}

	shapp.Gobpath = gobpath
	shapp.LastSaved = time.Now() // Do not persist immediately after starting
	return shapp, nil
}

// checkConfig checks if the given BatchConfig could be added.
func (app *ShutterApp) checkConfig(cfg BatchConfig) error {
	err := cfg.EnsureValid()
	if err != nil {
		return err
	}
	lastConfig := app.LastConfig()
	if cfg.ActivationBlockNumber < lastConfig.ActivationBlockNumber {
		return errors.Errorf(
			"start activation block number of next config (%d) lower than current one (%d)",
			cfg.ActivationBlockNumber,
			lastConfig.ActivationBlockNumber,
		)
	}
	if cfg.KeyperConfigIndex <= lastConfig.KeyperConfigIndex {
		return errors.Errorf(
			"config index of next config (%d) not greater than current one (%d)",
			cfg.KeyperConfigIndex,
			lastConfig.KeyperConfigIndex,
		)
	}
	return nil
}

func (app *ShutterApp) addConfig(cfg BatchConfig) error {
	err := app.checkConfig(cfg)
	if err != nil {
		return err
	}
	log.Info().Uint64("config-index", cfg.KeyperConfigIndex).
		Msg("adding keyper config")
	app.Configs = append(app.Configs, &cfg)
	app.updateCheckTxMembers()
	return nil
}

// updateCheckTxMembers sets the member set of the check tx state to the set of known keypers.
// This should be called whenever a new config is added.
func (app *ShutterApp) updateCheckTxMembers() {
	// This potentially double counts some keypers, but that's ok as CheckTxState.SetMembers
	// ignores duplicates.
	members := []common.Address{}
	for _, c := range app.Configs {
		members = append(members, c.Keypers...)
	}
	app.CheckTxState.SetMembers(members)
}

func (app *ShutterApp) Query(_ abcitypes.RequestQuery) abcitypes.ResponseQuery {
	return abcitypes.ResponseQuery{
		Code: 1,
		Log:  "query not implemented",
	}
}

// Info should return the latest committed state of the app. On startup, tendermint calls the Info
// method and will replay blocks that are not yet committed.
// See https://github.com/tendermint/spec/blob/master/spec/abci/apps.md#crash-recovery
func (app *ShutterApp) Info(_ abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{
		LastBlockHeight:  app.LastBlockHeight,
		LastBlockAppHash: []byte(""),
	}
}

func (ShutterApp) ListSnapshots(abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

func (ShutterApp) LoadSnapshotChunk(abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (ShutterApp) ApplySnapshotChunk(abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}

func (ShutterApp) OfferSnapshot(abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

/*
	BlockExecution

The first time a new blockchain is started, Tendermint calls InitChain. From then on, the following
sequence of methods is executed for each block:

BeginBlock, [DeliverTx], EndBlock, Commit

where one DeliverTx is called for each transaction in the block. The result is an updated
application state. Cryptographic commitments to the results of DeliverTx, EndBlock, and Commit are
included in the header of the next block.
*/
func (app *ShutterApp) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	genesisState := GenesisAppState{}
	err := amino.NewCodec().UnmarshalJSON(req.AppStateBytes, &genesisState)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot unmarshal genesis app state")
	}

	bc := BatchConfig{
		ActivationBlockNumber: 0,
		Keypers:               genesisState.GetKeypers(),
		Threshold:             genesisState.Threshold,
	}
	err = bc.EnsureValid()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid genesis app state")
	}

	if len(app.Configs) == 1 && len(app.Configs[0].Keypers) == 0 {
		log.Info().
			Uint64("initial-eon", genesisState.InitialEon).
			Str("chain-id", req.ChainId).
			Msg("initializing new chain")
		for i, k := range genesisState.Keypers {
			log.Info().Int("index", i).Str("keyper", k.String()).Msg("initial keyper")
		}
		validators, err := MakePowermap(req.Validators)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot handle validator keys")
		}
		app.Validators = validators
		app.Configs = []*BatchConfig{&bc}
		app.EONCounter = genesisState.InitialEon
		app.CheckTxState = NewCheckTxState()
		app.updateCheckTxMembers()
	} else {
		// XXX This else block is not executed anymore. Maybe we should remove it.
		// Ensure that our app state matches the genesis config
		if !reflect.DeepEqual(bc, *app.Configs[0]) {
			log.Fatal().
				Interface("initial-state", bc).
				Interface("stored-state", app.Configs[0]).
				Msg("mismatch between stored app state and initial app state")
		}
		if app.EONCounter < genesisState.InitialEon {
			log.Fatal().
				Uint64("EonCounter", app.EONCounter).
				Uint64("InitialEon", genesisState.InitialEon).
				Msg("mismatch between stored app state and initial app state")
		}
	}

	app.ChainID = req.ChainId

	return abcitypes.ResponseInitChain{}
}

func (app *ShutterApp) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	var events []abcitypes.Event
	if req.Header.Height == 1 {
		events = append(events, app.Configs[0].MakeABCIEvent())
	}
	return abcitypes.ResponseBeginBlock{Events: events}
}

func (app *ShutterApp) PrepareProposal(req abcitypes.RequestPrepareProposal) abcitypes.ResponsePrepareProposal {
	txs := make([][]byte, 0, len(req.Txs))
	var totalBytes int64
	for _, tx := range req.Txs {
		totalBytes += int64(len(tx))
		if totalBytes > req.MaxTxBytes {
			break
		}
		txs = append(txs, tx)
	}
	return abcitypes.ResponsePrepareProposal{Txs: txs}
}

func (app *ShutterApp) ProcessProposal(_ abcitypes.RequestProcessProposal) abcitypes.ResponseProcessProposal {
	return abcitypes.ResponseProcessProposal{
		Status: abcitypes.ResponseProcessProposal_ACCEPT,
	}
}

// decodeTx decodes the given transaction.  It's kind of strange that we have do URL decode the
// message outselves instead of tendermint doing it for us.
func (ShutterApp) decodeTx(tx []byte) (signer common.Address, msg *shmsg.MessageWithNonce, err error) {
	var signedMsg []byte
	signedMsg, err = base64.RawURLEncoding.DecodeString(string(tx))
	if err != nil {
		return
	}
	signer, err = shmsg.GetSigner(signedMsg)
	if err != nil {
		return
	}

	msg, err = shmsg.GetMessage(signedMsg)
	if err != nil {
		return
	}
	return
}

func (app *ShutterApp) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	signer, msg, err := app.decodeTx(req.Tx)
	if err != nil {
		msg := fmt.Sprintf("Error while decoding transaction: %s", err)
		log.Print(msg)
		return makeErrorResponse(msg)
	}
	if string(msg.ChainId) != app.ChainID {
		return makeErrorResponse(fmt.Sprintf("wrong chain id (expected %s, got %s)", app.ChainID, msg.ChainId))
	}
	if !app.NonceTracker.Check(signer, msg.RandomNonce) {
		msg := fmt.Sprintf("Nonce %d of %s already used", msg.RandomNonce, signer.Hex())
		return makeErrorResponse(msg)
	}
	app.NonceTracker.Add(signer, msg.RandomNonce)
	return app.deliverMessage(msg.Msg, signer)
}

func makeErrorResponse(msg string) abcitypes.ResponseDeliverTx {
	return abcitypes.ResponseDeliverTx{
		Code:   1,
		Log:    msg,
		Events: []abcitypes.Event{},
	}
}

func notAKeyper(sender common.Address) abcitypes.ResponseDeliverTx {
	return makeErrorResponse(fmt.Sprintf(
		"sender %s is not a keyper", sender.Hex()))
}

func (app *ShutterApp) allowedToVoteOnConfigChanges(sender common.Address) bool {
	lastConfig := app.LastConfig()
	_, ok := lastConfig.KeyperIndex(sender)
	return ok
}

func (app *ShutterApp) deliverBatchConfig(msg *shmsg.BatchConfig, sender common.Address) abcitypes.ResponseDeliverTx {
	bc, err := shutterevents.BatchConfigFromMessage(msg)
	if err != nil {
		return makeErrorResponse(fmt.Sprintf("Malformed BatchConfig message: %s", err))
	}

	if reflect.DeepEqual(*app.LastConfig(), bc) {
		// The config has already been accepted. So, let's just return success
		// XXX We do not check if we're allowed to vote on config changes here
		log.Info().Uint64("config-index", bc.KeyperConfigIndex).Msg("keyper config already accepted")
		return abcitypes.ResponseDeliverTx{
			Code: 0,
		}
	}
	err = app.checkConfig(bc)
	if err != nil {
		return makeErrorResponse(fmt.Sprintf("checkConfig: %s", err))
	}

	if !app.allowedToVoteOnConfigChanges(sender) {
		return makeErrorResponse("not allowed to vote on config changes")
	}

	var events []abcitypes.Event

	err = app.ConfigVoting.AddVote(sender, bc)
	if err != nil {
		return makeErrorResponse(fmt.Sprintf("Error adding vote: %s", err))
	}

	_, ok := app.ConfigVoting.Outcome(int(app.LastConfig().Threshold))
	if ok {
		app.ConfigVoting = NewConfigVoting()
		err = app.addConfig(bc)
		if err != nil {
			return makeErrorResponse(fmt.Sprintf("Error in addConfig: %s", err))
		}

		events = append(events, bc.MakeABCIEvent())
		dkg := app.StartDKG(bc)
		events = append(events, shutterevents.EonStarted{
			Eon:                   dkg.Eon,
			ActivationBlockNumber: bc.ActivationBlockNumber,
			KeyperConfigIndex:     bc.KeyperConfigIndex,
		}.MakeABCIEvent())
	}

	return abcitypes.ResponseDeliverTx{
		Code:   0,
		Events: events,
	}
}

// isKeyper checks if the given address is a keyper in any config (current and previous ones).
func (app *ShutterApp) isKeyper(a common.Address) bool {
	for _, cfg := range app.Configs {
		_, ok := cfg.KeyperIndex(a)
		if ok {
			return true
		}
	}
	return false
}

func (app *ShutterApp) deliverCheckIn(msg *shmsg.CheckIn, sender common.Address) abcitypes.ResponseDeliverTx {
	_, ok := app.Identities[sender]
	if ok {
		return makeErrorResponse(fmt.Sprintf(
			"sender %s already checked in", sender.Hex()))
	}
	if !app.isKeyper(sender) {
		return notAKeyper(sender)
	}

	validatorPublicKey, err := NewValidatorPubkey(msg.ValidatorPublicKey)
	if err != nil {
		return makeErrorResponse(fmt.Sprintf(
			"malformed validator public key: %s", err))
	}
	encryptionPublicKeyECDSA, err := crypto.DecompressPubkey(msg.EncryptionPublicKey)
	if err != nil {
		return makeErrorResponse(fmt.Sprintf("malformed encryption public key: %s", err))
	}
	encryptionPublicKey := ecies.ImportECDSAPublic(encryptionPublicKeyECDSA)

	app.Identities[sender] = validatorPublicKey
	return abcitypes.ResponseDeliverTx{
		Code: 0,
		Events: []abcitypes.Event{
			shutterevents.CheckIn{
				Sender:              sender,
				EncryptionPublicKey: encryptionPublicKey,
			}.MakeABCIEvent(),
		},
	}
}

func (app *ShutterApp) deliverBlockSeen(
	msg *shmsg.BlockSeen,
	sender common.Address,
) abcitypes.ResponseDeliverTx {
	if msg.BlockNumber > app.BlocksSeen[sender] {
		app.BlocksSeen[sender] = msg.BlockNumber
	}
	return abcitypes.ResponseDeliverTx{
		Code:   0,
		Events: []abcitypes.Event{},
	}
}

func (app *ShutterApp) deliverDKGResult(msg *shmsg.DKGResult, sender common.Address) abcitypes.ResponseDeliverTx {
	dkginstance, ok := app.DKGMap[msg.Eon]
	if !ok {
		return makeErrorResponse(fmt.Sprintf(
			"cannot handle DKGResult message for eon %d", msg.Eon),
		)
	}
	config := dkginstance.Config
	if !config.IsKeyper(sender) {
		return notAKeyper(sender)
	}

	err := dkginstance.SuccessVoting.AddVote(sender, msg.Success)
	if err != nil {
		return makeErrorResponse("already voted on dkg result")
	}

	dkg, started := app.maybeStartEon(msg.Eon)
	if !started {
		return abcitypes.ResponseDeliverTx{
			Code:   0,
			Events: []abcitypes.Event{},
		}
	}
	return abcitypes.ResponseDeliverTx{
		Code: 0,
		Events: []abcitypes.Event{
			shutterevents.EonStarted{
				Eon:                   dkg.Eon,
				ActivationBlockNumber: config.ActivationBlockNumber,
				KeyperConfigIndex:     config.KeyperConfigIndex,
			}.MakeABCIEvent(),
		},
	}
}

func (app *ShutterApp) maybeStartEon(eon uint64) (*DKGInstance, bool) {
	dkg, ok := app.DKGMap[eon]
	if !ok {
		return nil, false
	}

	threshold := int(dkg.Config.Threshold)
	success, ok := dkg.SuccessVoting.Outcome(threshold)
	// dismiss votes for Eon that was voted on successfully already
	outdatedEon := app.EONCounter > eon
	if !ok || success || outdatedEon {
		return nil, false
	}
	return app.StartDKG(dkg.Config), true
}

func (app *ShutterApp) handlePolyEvalMsg(msg *shmsg.PolyEval, sender common.Address) abcitypes.ResponseDeliverTx {
	appMsg, err := ParsePolyEvalMsg(msg, sender)
	if err != nil {
		msg := fmt.Sprintf("Error: Failed to parse PolyEval message: %+v", err)
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	dkg := app.DKGMap[appMsg.Eon]
	if dkg == nil {
		msg := "Error: Received PolyEval message while DKG is not active"
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	err = dkg.RegisterPolyEvalMsg(*appMsg)
	if err != nil {
		msg := fmt.Sprintf("Error: Failed to register PolyEval message: %+v", err)
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	event := appMsg.MakeABCIEvent()
	return abcitypes.ResponseDeliverTx{
		Code:   0,
		Events: []abcitypes.Event{event},
	}
}

func (app *ShutterApp) handlePolyCommitmentMsg(msg *shmsg.PolyCommitment, sender common.Address) abcitypes.ResponseDeliverTx {
	appMsg, err := ParsePolyCommitmentMsg(msg, sender)
	if err != nil {
		msg := fmt.Sprintf("Error: Failed to parse PolyCommitment message: %+v", err)
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	dkg := app.DKGMap[appMsg.Eon]
	if dkg == nil {
		msg := "Error: Received PolyCommitment message while DKG is not active"
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	err = dkg.RegisterPolyCommitmentMsg(*appMsg)
	if err != nil {
		msg := fmt.Sprintf("Error: Failed to register PolyCommitment message: %+v", err)
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	event := appMsg.MakeABCIEvent()
	return abcitypes.ResponseDeliverTx{
		Code:   0,
		Events: []abcitypes.Event{event},
	}
}

func (app *ShutterApp) handleAccusationMsg(msg *shmsg.Accusation, sender common.Address) abcitypes.ResponseDeliverTx {
	appMsg, err := ParseAccusationMsg(msg, sender)
	if err != nil {
		msg := fmt.Sprintf("Error: Failed to parse Accusation message: %+v", err)
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	dkg := app.DKGMap[appMsg.Eon]
	if dkg == nil {
		msg := "Error: Received Accusation message while DKG is not active"
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	err = dkg.RegisterAccusationMsg(*appMsg)
	if err != nil {
		msg := fmt.Sprintf("Error: Failed to register Accusation message: %+v", err)
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	event := appMsg.MakeABCIEvent()
	return abcitypes.ResponseDeliverTx{
		Code:   0,
		Events: []abcitypes.Event{event},
	}
}

func (app *ShutterApp) handleApologyMsg(msg *shmsg.Apology, sender common.Address) abcitypes.ResponseDeliverTx {
	appMsg, err := ParseApologyMsg(msg, sender)
	if err != nil {
		msg := fmt.Sprintf("Error: Failed to parse Apology message: %+v", err)
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	dkg := app.DKGMap[appMsg.Eon]
	if dkg == nil {
		msg := "Error: Received Apology message while DKG is not active"
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	err = dkg.RegisterApologyMsg(*appMsg)
	if err != nil {
		msg := fmt.Sprintf("Error: Failed to register Apology message: %+v", err)
		log.Print(msg)
		return makeErrorResponse(msg)
	}

	event := appMsg.MakeABCIEvent()
	return abcitypes.ResponseDeliverTx{
		Code:   0,
		Events: []abcitypes.Event{event},
	}
}

func (app *ShutterApp) deliverMessage(msg *shmsg.Message, sender common.Address) abcitypes.ResponseDeliverTx {
	if msg.GetBatchConfig() != nil {
		return app.deliverBatchConfig(msg.GetBatchConfig(), sender)
	}
	if msg.GetBlockSeen() != nil {
		return app.deliverBlockSeen(msg.GetBlockSeen(), sender)
	}
	if msg.GetCheckIn() != nil {
		return app.deliverCheckIn(msg.GetCheckIn(), sender)
	}
	if msg.GetDkgResult() != nil {
		return app.deliverDKGResult(msg.GetDkgResult(), sender)
	}

	if msg.GetPolyEval() != nil {
		return app.handlePolyEvalMsg(msg.GetPolyEval(), sender)
	}
	if msg.GetPolyCommitment() != nil {
		return app.handlePolyCommitmentMsg(msg.GetPolyCommitment(), sender)
	}
	if msg.GetAccusation() != nil {
		return app.handleAccusationMsg(msg.GetAccusation(), sender)
	}
	if msg.GetApology() != nil {
		return app.handleApologyMsg(msg.GetApology(), sender)
	}
	log.Print("Error: cannot deliver messsage: ", msg)
	return makeErrorResponse("cannot deliver message")
}

func (app *ShutterApp) StartDKG(config BatchConfig) *DKGInstance {
	app.EONCounter++
	dkg := NewDKGInstance(config, app.EONCounter)
	app.DKGMap[dkg.Eon] = &dkg
	return &dkg
}

// LastConfig returns the config with the highest known index.
func (app *ShutterApp) LastConfig() *BatchConfig {
	if len(app.Configs) == 0 {
		panic("internal error: app.Configs is empty")
	}
	return app.Configs[len(app.Configs)-1]
}

// makePowermap creates a power map for the given slice of keypers. The voting power of each keyper
// that hasn't registered yet, is given to the NonExistentValidator key.
func (app *ShutterApp) makePowermap(keypers []common.Address) Powermap {
	pm := make(Powermap)
	for _, k := range keypers {
		pubkey, ok := app.Identities[k]
		if ok {
			pm[pubkey] += 10
		} else {
			pm[NonExistentValidator] += 10
		}
	}
	return pm
}

// CurrentValidators returns a powermap of current validators.
func (app *ShutterApp) CurrentValidators() Powermap {
	for i := len(app.Configs) - 1; i >= 0; i-- {
		if app.Configs[i].Started && app.Configs[i].ValidatorsUpdated {
			return app.makePowermap(app.Configs[i].Keypers)
		}
	}
	return app.Validators
}

// countCheckedInKeypers counts the number of keypers that have already checked in in the given slice.
func (app *ShutterApp) countCheckedInKeypers(keypers []common.Address) uint64 {
	var numCheckedIn uint64
	for _, k := range keypers {
		_, ok := app.Identities[k]
		if ok {
			numCheckedIn++
		}
	}
	return numCheckedIn
}

func (app *ShutterApp) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	var events []abcitypes.Event

	for i, config := range app.Configs {
		if !config.Started {
			var allowanceConfigIndex int
			if i > 0 {
				allowanceConfigIndex = i - 1
			}
			numRequiredVotes := app.Configs[allowanceConfigIndex].Threshold

			var numVotes uint64
			for _, k := range app.Configs[allowanceConfigIndex].Keypers {
				b, ok := app.BlocksSeen[k]
				if ok && b >= config.ActivationBlockNumber {
					numVotes++
				}
			}
			if numVotes >= numRequiredVotes {
				log.Info().Uint64("config-index", config.KeyperConfigIndex).Msg("starting keyper config")
				config.Started = true
				events = append(events, shutterevents.BatchConfigStarted{
					KeyperConfigIndex: config.KeyperConfigIndex,
				}.MakeABCIEvent())
			}
		}
		if config.Started && !config.ValidatorsUpdated && app.countCheckedInKeypers(config.Keypers) >= numRequiredTransitionValidators(config) {
			config.ValidatorsUpdated = true
		}
	}

	newValidators := app.CurrentValidators()
	validatorUpdates := DiffPowermaps(app.Validators, newValidators).ValidatorUpdates()
	app.Validators = newValidators
	app.LastBlockHeight = req.Height
	if app.DevMode {
		if len(validatorUpdates) > 0 {
			log.Info().Int("count", len(validatorUpdates)).Msg("ignoring validator updates in dev mode")
		}
		return abcitypes.ResponseEndBlock{Events: events}
	}
	if len(validatorUpdates) > 0 {
		log.Info().Int("count", len(validatorUpdates)).Interface("validator-updates", validatorUpdates).
			Msg("applying validator updates")
	}
	return abcitypes.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
		Events:           events,
	}
}

// numRequiredTransitionValidators returns the number of validators required to be online before
// transitioning to a new config. This number is either the threshold value in the config or 2/3
// of the validator set, whatever is greater. This makes sure that both the security assumption
// defined by the threshold is met and that the chain is able to make progress.
func numRequiredTransitionValidators(config *BatchConfig) uint64 {
	n := len(config.Keypers)
	if n == 0 {
		// this case doesn't make much sense, but the normal path would return 1 which makes even
		// less sense
		return 0
	}
	defenders := uint64(n - (n+2)/3 + 1)
	if config.Threshold >= defenders {
		return config.Threshold
	}
	return defenders
}

// persistToDisk stores the ShutterApp on disk. This method first writes to a temporary file and
// renames the file later. Most probably this will not work on windows!
func (app *ShutterApp) PersistToDisk() error {
	log.Info().Int64("height", app.LastBlockHeight).Msg("persisting state to disk")
	tmppath := app.Gobpath + ".tmp"
	file, err := os.Create(tmppath)
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Error().Err(err).Str("path", tmppath).Msg("failed to close file")
			return
		}
	}()

	app.LastSaved = time.Now()
	enc := gob.NewEncoder(file)
	err = enc.Encode(app)
	if err != nil {
		return err
	}
	err = file.Sync()
	if err != nil {
		return err
	}
	err = os.Rename(tmppath, app.Gobpath)
	return err
}

func (app *ShutterApp) maybePersistToDisk() error {
	if app.Gobpath == "" {
		return nil
	}
	if time.Since(app.LastSaved) <= PersistMinDuration {
		return nil
	}
	return app.PersistToDisk()
}

func (app *ShutterApp) Commit() abcitypes.ResponseCommit {
	app.CheckTxState.Reset()

	err := app.maybePersistToDisk()
	if err != nil {
		log.Error().Err(err).Msg("cannot persist state to disk")
	}

	return abcitypes.ResponseCommit{}
}

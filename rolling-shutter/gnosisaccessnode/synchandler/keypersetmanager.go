package synchandler

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/shop-contracts/bindings"

	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/gnosisaccessnode/storage"
	keypersetsync "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/synchandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func init() {
	var err error
	KeyperSetManagerContractABI, err = bindings.KeyperSetManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

var KeyperSetManagerContractABI *abi.ABI

func NewKeyperSetAdded(
	ethClient client.Client,
	store *storage.Memory,
	contractAddress common.Address,
) (syncer.ContractEventHandler, error) {
	// we need access to an additional contract here in oder to pull in some more
	// required information about the keyper sets:
	ksm, err := bindings.NewKeyperSetManager(contractAddress, ethClient)
	if err != nil {
		return nil, err
	}
	return syncer.WrapHandler(&KeyperSetAdded{
		storage:                 store,
		evABI:                   KeyperSetManagerContractABI,
		keyperSetManagerAddress: contractAddress,
		keyperSetManager:        ksm,
		ethClient:               ethClient,
	})
}

type KeyperSetAdded struct {
	storage                 *storage.Memory
	evABI                   *abi.ABI
	keyperSetManagerAddress common.Address
	ethClient               client.Client
	keyperSetManager        *bindings.KeyperSetManager
}

func (handler *KeyperSetAdded) Address() common.Address {
	return handler.keyperSetManagerAddress
}

func (*KeyperSetAdded) Event() string {
	return "KeyperSetAdded"
}

func (handler *KeyperSetAdded) ABI() abi.ABI {
	return *handler.evABI
}

func (handler *KeyperSetAdded) Accept(
	_ context.Context,
	_ types.Header,
	_ bindings.KeyperSetManagerKeyperSetAdded,
) (bool, error) {
	return true, nil
}

func (handler *KeyperSetAdded) Handle(
	ctx context.Context,
	update syncer.ChainUpdateContext,
	events []bindings.KeyperSetManagerKeyperSetAdded,
) error {
	// TODO: handle reorgs here
	_ = update

	for _, ev := range events {
		keyperSet, err := keypersetsync.QueryFullKeyperSetFromKeyperSetAddedEvent(
			ctx,
			handler.ethClient,
			ev,
			handler.keyperSetManager,
		)
		if err != nil {
			log.Error().Err(err).Msg("KeyperSetAdded event, error querying keyperset-data")
		}
		// FIXME: integer overflow protection
		obsKeyperSet := obskeyperdatabase.KeyperSet{
			KeyperConfigIndex:     int64(keyperSet.Eon),
			ActivationBlockNumber: int64(keyperSet.ActivationBlock),
			Keypers:               shdb.EncodeAddresses(keyperSet.Members),
			Threshold:             int32(keyperSet.Threshold),
		}
		log.Info().
			Uint64("keyper-config-index", keyperSet.Eon).
			Uint64("activation-block-number", keyperSet.ActivationBlock).
			Int("num-keypers", len(keyperSet.Members)).
			Uint64("threshold", keyperSet.Threshold).
			Msg("adding keyper set")
		handler.storage.AddKeyperSet(keyperSet.Eon, &obsKeyperSet)
	}
	return nil
}

package synchandler

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/shop-contracts/bindings"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/gnosisaccessnode/storage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
)

func init() {
	var err error
	KeyBroadcastContractContractABI, err = bindings.KeyBroadcastContractMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

var KeyBroadcastContractContractABI *abi.ABI

func NewEonKeyBroadcast(
	store *storage.Memory,
	contractAddress common.Address,
) (syncer.ContractEventHandler, error) {
	return syncer.WrapHandler(&EonKeyBroadcast{
		evABI:               KeyBroadcastContractContractABI,
		keyBroadcastAddress: contractAddress,
		storage:             store,
	})
}

type EonKeyBroadcast struct {
	storage             *storage.Memory
	evABI               *abi.ABI
	keyBroadcastAddress common.Address
}

func (handler *EonKeyBroadcast) Address() common.Address {
	return handler.keyBroadcastAddress
}

func (*EonKeyBroadcast) Event() string {
	// TODO: look this up that his is correct
	return "EonKeyBroadcast"
}

func (handler *EonKeyBroadcast) ABI() abi.ABI {
	return *handler.evABI
}

func (handler *EonKeyBroadcast) Accept(
	_ context.Context,
	_ types.Header,
	_ bindings.KeyBroadcastContractEonKeyBroadcast,
) (bool, error) {
	return true, nil
}

func (handler *EonKeyBroadcast) Handle(
	_ context.Context,
	_ syncer.ChainUpdateContext,
	events []bindings.KeyBroadcastContractEonKeyBroadcast,
) error {
	for _, ev := range events {
		key := new(shcrypto.EonPublicKey)
		err := key.Unmarshal(ev.Key)
		if err != nil {
			log.Error().
				Err(err).
				Hex("key", ev.Key).
				Int("keyper-config-index", int(ev.Eon)).
				Msg("received invalid eon key")
			return nil
		}
		log.Info().
			Int("keyper-config-index", int(ev.Eon)).
			Hex("key", ev.Key).
			Msg("adding eon key")
		handler.storage.AddEonKey(ev.Eon, key)
	}
	return nil
}

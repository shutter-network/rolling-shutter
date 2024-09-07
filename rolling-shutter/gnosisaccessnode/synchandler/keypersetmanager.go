package synchandler

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/gnosisaccessnode/storage"
	keypersetsync "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/synchandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/shop-contracts/bindings"
)

var (
	ErrParseKeyperSet = errors.New("can't parse KeyperSet")

	errNilCallOpts        = errors.New("nil call-opts")
	errNilOptsBlockNumber = errors.New("opts block-number is nil, but 'latest' not allowed")
	errLatestBlock        = errors.New("'nil' latest block")
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
	client client.Client,
	storage *storage.Memory,
	contractAddress common.Address,
) (syncer.ContractEventHandler, error) {
	ksm, err := bindings.NewKeyperSetManager(contractAddress, client)
	if err != nil {
		return nil, err
	}
	return syncer.WrapHandler(&KeyperSetAdded{
		// TODO: log
		// log:                     log,
		evABI:                   KeyperSetManagerContractABI,
		keyperSetManagerAddress: contractAddress,
		keyperSetManager:        ksm,
		ethClient:               client,
	})
}

type KeyperSetAdded struct {
	storage                 *storage.Memory
	evABI                   *abi.ABI
	keyperSetManagerAddress common.Address
	dbpool                  *pgxpool.Pool

	// we only need this because we have to poll
	// additional data from the contract
	ethClient        client.Client
	keyperSetManager *bindings.KeyperSetManager
}

func (handler *KeyperSetAdded) Address() common.Address {
	return handler.keyperSetManagerAddress
}

func (_ *KeyperSetAdded) Event() string {
	return "KeyperSetAdded"
}

func (handler *KeyperSetAdded) ABI() abi.ABI {
	return *handler.evABI
}
func (handler *KeyperSetAdded) Accept(
	ctx context.Context,
	header types.Header,
	ev bindings.KeyperSetManagerKeyperSetAdded,
) (bool, error) {
	return true, nil
}
func (handler *KeyperSetAdded) Handle(
	ctx context.Context,
	query syncer.QueryContext,
	events []bindings.KeyperSetManagerKeyperSetAdded,
) error {
	// TODO: handle reorgs here

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
		return nil
	}
	return nil
}

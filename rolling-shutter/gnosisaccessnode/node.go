package gnosisaccessnode

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	chainsync "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/legacychainsync"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/legacychainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type GnosisAccessNode struct {
	config  *Config
	storage *Storage
}

func New(config *Config) *GnosisAccessNode {
	return &GnosisAccessNode{
		config:  config,
		storage: NewStorage(),
	}
}

func (node *GnosisAccessNode) Start(ctx context.Context, runner service.Runner) error {
	services := []service.Service{}

	messageSender, err := p2p.New(node.config.P2P)
	if err != nil {
		return errors.Wrap(err, "failed to initialize p2p messaging")
	}
	messageSender.AddMessageHandler(NewDecryptionKeysHandler(node.config, node.storage))
	services = append(services, messageSender)

	chainSyncClient, err := chainsync.NewClient(
		ctx,
		chainsync.WithClientURL(node.config.GnosisNode.EthereumURL),
		chainsync.WithKeyperSetManager(node.config.Contracts.KeyperSetManager),
		chainsync.WithKeyBroadcastContract(node.config.Contracts.KeyBroadcastContract),
		chainsync.WithSyncNewKeyperSet(node.onNewKeyperSet),
		chainsync.WithSyncNewEonKey(node.onNewEonKey),
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize chain sync client")
	}
	services = append(services, chainSyncClient)

	if node.config.Metrics.Enabled {
		metricsServer := metricsserver.New(node.config.Metrics)
		services = append(services, metricsServer)
	}

	return runner.StartService(services...)
}

func (node *GnosisAccessNode) onNewKeyperSet(_ context.Context, keyperSet *syncevent.KeyperSet) error {
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
	node.storage.AddKeyperSet(keyperSet.Eon, &obsKeyperSet)
	return nil
}

func (node *GnosisAccessNode) onNewEonKey(_ context.Context, eonKey *syncevent.EonPublicKey) error {
	key := new(shcrypto.EonPublicKey)
	err := key.Unmarshal(eonKey.Key)
	if err != nil {
		log.Error().
			Err(err).
			Hex("key", eonKey.Key).
			Int("keyper-config-index", int(eonKey.Eon)).
			Msg("received invalid eon key")
		return nil
	}
	log.Info().
		Int("keyper-config-index", int(eonKey.Eon)).
		Hex("key", eonKey.Key).
		Msg("adding eon key")
	node.storage.AddEonKey(eonKey.Eon, key)
	return nil
}

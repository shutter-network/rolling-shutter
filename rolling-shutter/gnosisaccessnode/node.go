package gnosisaccessnode

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/gnosisaccessnode/storage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/gnosisaccessnode/synchandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

type GnosisAccessNode struct {
	config  *Config
	storage *storage.Memory
}

func New(config *Config) *GnosisAccessNode {
	return &GnosisAccessNode{
		config:  config,
		storage: storage.NewMemory(),
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

	ethClient, err := ethclient.DialContext(ctx, node.config.GnosisNode.EthereumURL)
	if err != nil {
		return err
	}
	keyperSetAdded, err := synchandler.NewKeyperSetAdded(
		ethClient,
		node.storage,
		node.config.Contracts.KeyperSetManager,
	)
	if err != nil {
		return err
	}
	eonKeyBroadcast, err := synchandler.NewEonKeyBroadcast(
		node.storage,
		node.config.Contracts.KeyBroadcastContract,
	)
	if err != nil {
		return err
	}
	chainsyncOpts := []chainsync.Option{
		chainsync.WithClient(ethClient),
		chainsync.WithContractEventHandler(keyperSetAdded),
		chainsync.WithContractEventHandler(eonKeyBroadcast),
	}
	chainsyncer, err := chainsync.New(chainsyncOpts...)
	if err != nil {
		return fmt.Errorf("can't instantiate chainsync: %w", err)
	}
	services = append(services, chainsyncer)
	if node.config.Metrics.Enabled {
		metricsServer := metricsserver.New(node.config.Metrics)
		services = append(services, metricsServer)
	}

	return runner.StartService(services...)
}

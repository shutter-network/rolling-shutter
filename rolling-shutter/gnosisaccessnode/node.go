package gnosisaccessnode

import (
	"context"

	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

type GnosisAccessNode struct {
	config *Config
}

func New(config *Config) *GnosisAccessNode {
	return &GnosisAccessNode{
		config: config,
	}
}

func (node *GnosisAccessNode) Start(ctx context.Context, runner service.Runner) error {
	messageSender, err := p2p.New(node.config.P2P)
	if err != nil {
		return errors.Wrap(err, "failed to initialize p2p messaging")
	}
	messageSender.AddMessageHandler(NewDecryptionKeysHandler(node.config))
	return runner.StartService(messageSender)
}

package gnosisaccessnode

import (
	"context"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
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
	return nil
}

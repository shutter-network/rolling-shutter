package snapshot

import (
	"context"

	"github.com/rs/zerolog/log"
)

type Snapshot struct {
	Config Config
}

func New(config Config) *Snapshot {
	return &Snapshot{
		Config: config,
	}
}

func (d *Snapshot) Run(ctx context.Context) error {
	log.Info().Msg("starting Snapshot Hub interface")
	return nil
}

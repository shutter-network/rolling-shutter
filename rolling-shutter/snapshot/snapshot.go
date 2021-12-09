package snapshot

import (
	"context"
	"log"
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
	log.Printf(
		"starting Snapshot Hub interface",
	)
	return nil
}

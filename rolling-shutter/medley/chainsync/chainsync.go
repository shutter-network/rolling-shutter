package chainsync

import (
	"context"
	"fmt"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type Chainsync struct {
	options *options
	fetcher *syncer.Fetcher
}

func New(options ...Option) (*Chainsync, error) {
	opts := defaultOptions()
	for _, o := range options {
		if err := o(opts); err != nil {
			return nil, fmt.Errorf("error applying option to Chainsync: %w", err)
		}
	}

	if err := opts.verify(); err != nil {
		return nil, fmt.Errorf("error verifying options to Chainsync: %w", err)
	}
	return &Chainsync{
		options: opts,
		// will be initialized during Start()
		fetcher: nil,
	}, nil
}

func (c *Chainsync) Start(ctx context.Context, runner service.Runner) error {
	var err error
	c.fetcher, err = c.options.initFetcher(ctx)
	if err != nil {
		return fmt.Errorf("error initializing Chainsync: %w", err)
	}
	return c.fetcher.Start(ctx, runner)
}

package gnosiskeyperwatcher

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	keyper "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type BlocksWatcher struct {
	config *keyper.Config
}

func NewBlocksWatcher(config *keyper.Config) *BlocksWatcher {
	return &BlocksWatcher{
		config: config,
	}
}

func (w *BlocksWatcher) Start(ctx context.Context, runner service.Runner) error {
	runner.Go(func() error {
		ethClient, err := ethclient.Dial(w.config.Gnosis.EthereumURL)
		if err != nil {
			return err
		}

		newHeads := make(chan *types.Header)
		sub, err := ethClient.SubscribeNewHead(ctx, newHeads)
		if err != nil {
			return err
		}
		defer sub.Unsubscribe()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case head := <-newHeads:
				w.logNewHead(head)
			case err := <-sub.Err():
				return err
			}
		}
	})
	return nil
}

func (w *BlocksWatcher) logNewHead(head *types.Header) {
	log.Info().
		Int64("number", head.Number.Int64()).
		Hex("hash", head.Hash().Bytes()).
		Msg("new head")
}

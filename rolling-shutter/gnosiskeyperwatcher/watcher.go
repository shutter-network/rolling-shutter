package gnosiskeyperwatcher

import (
	"context"

	keyper "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type Watcher struct {
	config *keyper.Config
}

func New(config *keyper.Config) *Watcher {
	return &Watcher{
		config: config,
	}
}

func (w *Watcher) Start(_ context.Context, runner service.Runner) error {
	blocksChannel := make(chan *BlockReceivedEvent)
	blocksWatcher := NewBlocksWatcher(w.config, blocksChannel)
	keysWatcher := NewKeysWatcher(w.config, blocksChannel)
	return runner.StartService(blocksWatcher, keysWatcher)
}

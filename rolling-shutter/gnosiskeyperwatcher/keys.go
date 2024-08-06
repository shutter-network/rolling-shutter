package gnosiskeyperwatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/rs/zerolog/log"

	keyper "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type KeysWatcher struct {
	config        *keyper.Config
	blocksChannel chan *BlockReceivedEvent

	recentBlocksMux sync.Mutex
	recentBlocks    map[uint64]*BlockReceivedEvent
	mostRecentBlock uint64
}

func NewKeysWatcher(config *keyper.Config, blocksChannel chan *BlockReceivedEvent) *KeysWatcher {
	return &KeysWatcher{
		config:          config,
		blocksChannel:   blocksChannel,
		recentBlocksMux: sync.Mutex{},
		recentBlocks:    make(map[uint64]*BlockReceivedEvent),
		mostRecentBlock: 0,
	}
}

func (w *KeysWatcher) Start(ctx context.Context, runner service.Runner) error {
	p2pService, err := p2p.New(w.config.P2P)
	if err != nil {
		return err
	}
	p2pService.AddMessageHandler(w)

	runner.Go(func() error { return w.insertBlocks(ctx) })

	return runner.StartService(p2pService)
}

func (w *KeysWatcher) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{
		&p2pmsg.DecryptionKeys{},
	}
}

func (w *KeysWatcher) ValidateMessage(_ context.Context, _ p2pmsg.Message) (pubsub.ValidationResult, error) {
	return pubsub.ValidationAccept, nil
}

func (w *KeysWatcher) HandleMessage(_ context.Context, msgUntyped p2pmsg.Message) ([]p2pmsg.Message, error) {
	t := time.Now()
	msg := msgUntyped.(*p2pmsg.DecryptionKeys)
	extra := msg.Extra.(*p2pmsg.DecryptionKeys_Gnosis).Gnosis

	ev, ok := w.getRecentBlock(extra.Slot)
	if !ok {
		log.Warn().
			Uint64("keys-block", extra.Slot).
			Uint64("most-recent-block", w.mostRecentBlock).
			Msg("received keys for unknown block")
		return []p2pmsg.Message{}, nil
	}

	dt := t.Sub(ev.Time)
	log.Info().
		Uint64("block", extra.Slot).
		Int("num-keys", len(msg.Keys)).
		Str("latency", fmt.Sprintf("%.2fs", dt.Seconds())).
		Msg("new keys")
	return []p2pmsg.Message{}, nil
}

func (w *KeysWatcher) insertBlocks(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-w.blocksChannel:
			if !ok {
				return nil
			}
			w.insertBlock(ev)
			w.clearOldBlocks(ev)
		}
	}
}

func (w *KeysWatcher) insertBlock(ev *BlockReceivedEvent) {
	w.recentBlocksMux.Lock()
	defer w.recentBlocksMux.Unlock()
	w.recentBlocks[ev.Header.Number.Uint64()] = ev
	if ev.Header.Number.Uint64() > w.mostRecentBlock {
		w.mostRecentBlock = ev.Header.Number.Uint64()
	}
}

func (w *KeysWatcher) clearOldBlocks(latestEv *BlockReceivedEvent) {
	w.recentBlocksMux.Lock()
	defer w.recentBlocksMux.Unlock()

	tooOld := []uint64{}
	for block := range w.recentBlocks {
		if block < latestEv.Header.Number.Uint64()-100 {
			tooOld = append(tooOld, block)
		}
	}
	for _, block := range tooOld {
		delete(w.recentBlocks, block)
	}
}

func (w *KeysWatcher) getRecentBlock(blockNumber uint64) (*BlockReceivedEvent, bool) {
	w.recentBlocksMux.Lock()
	defer w.recentBlocksMux.Unlock()
	ev, ok := w.recentBlocks[blockNumber]
	return ev, ok
}

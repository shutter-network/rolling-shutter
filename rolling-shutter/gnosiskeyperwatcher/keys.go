package gnosiskeyperwatcher

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/rs/zerolog/log"

	keyper "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type KeysWatcher struct {
	config *keyper.Config
}

func NewKeysWatcher(config *keyper.Config) *KeysWatcher {
	return &KeysWatcher{
		config: config,
	}
}

func (w *KeysWatcher) Start(_ context.Context, runner service.Runner) error {
	p2pService, err := p2p.New(w.config.P2P)
	if err != nil {
		return err
	}
	p2pService.AddMessageHandler(w)

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
	msg := msgUntyped.(*p2pmsg.DecryptionKeys)
	extra := msg.Extra.(*p2pmsg.DecryptionKeys_Gnosis).Gnosis
	log.Info().
		Uint64("block", extra.Slot).
		Int("num-keys", len(msg.Keys)).
		Msg("new keys")
	return []p2pmsg.Message{}, nil
}

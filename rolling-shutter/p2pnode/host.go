package p2pnode

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type P2PNode struct {
	p2p *p2p.P2PHandler
}

func New(config Config) *P2PNode {
	p2pHandler := p2p.New(
		p2p.Config{
			ListenAddrs:     config.ListenAddresses,
			BootstrapPeers:  config.CustomBootstrapAddresses,
			PrivKey:         config.PrivateKey,
			Environment:     config.Environment,
			IsBootstrapNode: true,
		})
	return &P2PNode{
		p2p: p2pHandler,
	}
}

func (p *P2PNode) validateDecryptionTrigger(_ context.Context, _ *p2pmsg.DecryptionTrigger) (bool, error) {
	return true, nil
}

func (p *P2PNode) handleDecryptionTrigger(_ context.Context, msg *p2pmsg.DecryptionTrigger) ([]p2pmsg.Message, error) {
	msgs := []p2pmsg.Message{}
	log.Info().Str("message", msg.String()).Msg("received message")
	return msgs, nil
}

func (p *P2PNode) validateDecryptionKeyShare(_ context.Context, _ *p2pmsg.DecryptionKeyShare) (bool, error) {
	return true, nil
}

func (p *P2PNode) handleDecryptionKeyShare(_ context.Context, msg *p2pmsg.DecryptionKeyShare) ([]p2pmsg.Message, error) {
	msgs := []p2pmsg.Message{}
	log.Info().Str("message", msg.String()).Msg("received message")
	return msgs, nil
}

func (p *P2PNode) validateEonPublicKey(_ context.Context, _ *p2pmsg.EonPublicKey) (bool, error) {
	return true, nil
}

func (p *P2PNode) handleEonPublicKey(_ context.Context, msg *p2pmsg.EonPublicKey) ([]p2pmsg.Message, error) {
	msgs := []p2pmsg.Message{}
	log.Info().Str("message", msg.String()).Msg("received message")
	return msgs, nil
}

func (p *P2PNode) validateDecryptionKey(_ context.Context, _ *p2pmsg.DecryptionKey) (bool, error) {
	return true, nil
}

func (p *P2PNode) handleDecryptionKey(_ context.Context, msg *p2pmsg.DecryptionKey) ([]p2pmsg.Message, error) {
	msgs := []p2pmsg.Message{}
	log.Info().Str("message", msg.String()).Msg("received message")
	return msgs, nil
}

func (p *P2PNode) Start(ctx context.Context, runner service.Runner) error {
	p2p.AddValidator(p.p2p, p.validateDecryptionKey)
	p2p.AddHandlerFunc(p.p2p, p.handleDecryptionKey)

	p2p.AddValidator(p.p2p, p.validateDecryptionTrigger)
	p2p.AddHandlerFunc(p.p2p, p.handleDecryptionTrigger)

	p2p.AddValidator(p.p2p, p.validateDecryptionKeyShare)
	p2p.AddHandlerFunc(p.p2p, p.handleDecryptionKeyShare)

	p2p.AddValidator(p.p2p, p.validateEonPublicKey)
	p2p.AddHandlerFunc(p.p2p, p.handleEonPublicKey)
	return p.p2p.Start(ctx, runner)
}

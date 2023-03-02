package p2pnode

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
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

func (p *P2PNode) validateDecryptionTrigger(_ context.Context, _ *shmsg.DecryptionTrigger) (bool, error) {
	return true, nil
}

func (p *P2PNode) handleDecryptionTrigger(_ context.Context, msg *shmsg.DecryptionTrigger) ([]shmsg.P2PMessage, error) {
	msgs := []shmsg.P2PMessage{}
	log.Info().Str("message", msg.String()).Msg("received message")
	return msgs, nil
}

func (p *P2PNode) validateDecryptionKeyShare(_ context.Context, _ *shmsg.DecryptionKeyShare) (bool, error) {
	return true, nil
}

func (p *P2PNode) handleDecryptionKeyShare(_ context.Context, msg *shmsg.DecryptionKeyShare) ([]shmsg.P2PMessage, error) {
	msgs := []shmsg.P2PMessage{}
	log.Info().Str("message", msg.String()).Msg("received message")
	return msgs, nil
}

func (p *P2PNode) validateEonPublicKey(_ context.Context, _ *shmsg.EonPublicKey) (bool, error) {
	return true, nil
}

func (p *P2PNode) handleEonPublicKey(_ context.Context, msg *shmsg.EonPublicKey) ([]shmsg.P2PMessage, error) {
	msgs := []shmsg.P2PMessage{}
	log.Info().Str("message", msg.String()).Msg("received message")
	return msgs, nil
}

func (p *P2PNode) validateDecryptionKey(_ context.Context, _ *shmsg.DecryptionKey) (bool, error) {
	return true, nil
}

func (p *P2PNode) handleDecryptionKey(_ context.Context, msg *shmsg.DecryptionKey) ([]shmsg.P2PMessage, error) {
	msgs := []shmsg.P2PMessage{}
	log.Info().Str("message", msg.String()).Msg("received message")
	return msgs, nil
}

func (p *P2PNode) Run(ctx context.Context) error {
	p2p.AddValidator(p.p2p, p.validateDecryptionKey)
	p2p.AddHandlerFunc(p.p2p, p.handleDecryptionKey)

	p2p.AddValidator(p.p2p, p.validateDecryptionTrigger)
	p2p.AddHandlerFunc(p.p2p, p.handleDecryptionTrigger)

	p2p.AddValidator(p.p2p, p.validateDecryptionKeyShare)
	p2p.AddHandlerFunc(p.p2p, p.handleDecryptionKeyShare)

	p2p.AddValidator(p.p2p, p.validateEonPublicKey)
	p2p.AddHandlerFunc(p.p2p, p.handleEonPublicKey)
	return p.p2p.Run(ctx)
}

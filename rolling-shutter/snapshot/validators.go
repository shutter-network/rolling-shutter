package snapshot

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot/snptopics"
)

func (snp *Snapshot) makeMessagesValidators() map[string]pubsub.Validator {
	return map[string]pubsub.Validator{
		snptopics.DecryptionKey: snp.validateDecryptionKey,
		snptopics.EonPublicKey:  snp.validateEonPublicKey,
	}
}

func (snp *Snapshot) validateDecryptionKey(ctx context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	// TODO: Implement
	return true
}

func (snp *Snapshot) validateEonPublicKey(_ context.Context, _ peer.ID, libp2pMessage *pubsub.Message) bool {
	// TODO: Implement
	return true
}

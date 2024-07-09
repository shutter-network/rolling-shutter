package roguenode

import (
	"context"
	"math"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/rand"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func (node *RogueNode) sendMessages(ctx context.Context, incomingMessageCh chan p2pmsg.Message) error {
	ticker := time.NewTicker(time.Duration(node.config.SendInterval) * time.Millisecond)
	defer ticker.Stop()

	var latestMessage p2pmsg.Message

	for {
		select {
		case msg := <-incomingMessageCh:
			latestMessage = msg
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := node.sendAlteredMessage(ctx, latestMessage)
			if err != nil {
				return err
			}
		}
	}
}

func (node *RogueNode) sendAlteredMessage(ctx context.Context, template p2pmsg.Message) error {
	if template == nil {
		log.Info().Msg("no keys message received yet, sending empty message")
		return node.messaging.SendMessage(ctx, &p2pmsg.DecryptionKeys{})
	}
	untypedMessage := proto.Clone(template)
	msg := untypedMessage.(*p2pmsg.DecryptionKeys)
	alteredMsg := alterMessage(msg)
	return node.messaging.SendMessage(ctx, alteredMsg)
}

func alterMessage(msg *p2pmsg.DecryptionKeys) *p2pmsg.DecryptionKeys {
	i := rand.Intn(6)
	switch i {
	case 0:
		return alterSlot(msg)
	case 1:
		return alterTxPointer(msg)
	case 2:
		return alterSigners(msg)
	case 3:
		return alterInstanceID(msg)
	case 4:
		return alterEon(msg)
	case 5:
		return alterKeys(msg)
	default:
		panic("missing case")
	}
}

func alterSlot(msg *p2pmsg.DecryptionKeys) *p2pmsg.DecryptionKeys {
	extra := msg.GetGnosis()
	i := rand.Intn(5)
	switch i {
	case 0:
		log.Debug().
			Str("field", "slot").
			Str("modification", "incrementing").
			Msg("altering message")
		extra.Slot++
	case 1:
		log.Debug().
			Str("field", "slot").
			Str("modification", "decrementing").
			Msg("altering message")
		extra.Slot--
	case 2:
		log.Debug().
			Str("field", "slot").
			Str("modification", "setting to zero").
			Msg("altering message")
		extra.Slot = 0
	case 3:
		log.Debug().
			Str("field", "slot").
			Str("modification", "setting to max").
			Msg("altering message")
		extra.Slot = math.MaxUint64
	case 4:
		log.Debug().
			Str("field", "slot").
			Str("modification", "randomizing").
			Msg("altering message")
		extra.Slot = rand.Uint64()
	default:
		panic("missing case")
	}
	return msg
}

func alterTxPointer(msg *p2pmsg.DecryptionKeys) *p2pmsg.DecryptionKeys {
	extra := msg.GetGnosis()
	i := rand.Intn(5)
	switch i {
	case 0:
		log.Debug().
			Str("field", "tx pointer").
			Str("modification", "incrementing").
			Msg("altering message")
		extra.TxPointer++
	case 1:
		log.Debug().
			Str("field", "tx pointer").
			Str("modification", "decrementing").
			Msg("altering message")
		extra.TxPointer--
	case 2:
		log.Debug().
			Str("field", "tx pointer").
			Str("modification", "setting to zero").
			Msg("altering message")
		extra.TxPointer = 0
	case 3:
		log.Debug().
			Str("field", "tx pointer").
			Str("modification", "setting to max").
			Msg("altering message")
		extra.TxPointer = math.MaxUint64
	case 4:
		log.Debug().
			Str("field", "tx pointer").
			Str("modification", "randomizing").
			Msg("altering message")
		extra.TxPointer = rand.Uint64()
	default:
		panic("missing case")
	}
	return msg
}

func alterSigners(msg *p2pmsg.DecryptionKeys) *p2pmsg.DecryptionKeys {
	extra := msg.GetGnosis()
	i := rand.Intn(6)
	switch i {
	case 0:
		log.Debug().
			Str("field", "signatures").
			Str("modification", "removing last").
			Msg("altering message")
		n := len(extra.SignerIndices) - 1
		extra.SignerIndices = extra.SignerIndices[:n]
		extra.Signatures = extra.Signatures[:n]
	case 1:
		log.Debug().
			Str("field", "signatures").
			Str("modification", "making first invalid").
			Msg("altering message")
		extra.Signatures[0] = make([]byte, len(extra.Signatures[0]))
	case 2:
		log.Debug().
			Str("field", "signatures").
			Str("modification", "reordering").
			Msg("altering message")
		slices.Reverse(extra.SignerIndices)
		slices.Reverse(extra.Signatures)
	case 3:
		log.Debug().
			Str("field", "signatures").
			Str("modification", "add duplicate").
			Msg("altering message")
		n := len(extra.SignerIndices)
		extra.SignerIndices = append(extra.SignerIndices, extra.SignerIndices[n-1])
		extra.Signatures = append(extra.Signatures, extra.Signatures[n-1])
	case 4:
		log.Debug().
			Str("field", "signatures").
			Str("modification", "remove signature, keeping index").
			Msg("altering message")
		n := len(extra.Signatures) - 1
		extra.Signatures = extra.Signatures[:n]
	case 5:
		log.Debug().
			Str("field", "signatures").
			Str("modification", "remove index, keeping signature").
			Msg("altering message")
		n := len(extra.SignerIndices) - 1
		extra.SignerIndices = extra.SignerIndices[:n]
	default:
		panic("missing case")
	}
	return msg
}

func alterInstanceID(msg *p2pmsg.DecryptionKeys) *p2pmsg.DecryptionKeys {
	i := rand.Intn(5)
	switch i {
	case 0:
		log.Debug().
			Str("field", "instanceID").
			Str("modification", "incrementing").
			Msg("altering message")
		msg.InstanceID++
	case 1:
		log.Debug().
			Str("field", "instanceID").
			Str("modification", "decrementing").
			Msg("altering message")
		msg.InstanceID--
	case 2:
		log.Debug().
			Str("field", "instanceID").
			Str("modification", "setting to zero").
			Msg("altering message")
		msg.InstanceID = 0
	case 3:
		log.Debug().
			Str("field", "instanceID").
			Str("modification", "setting to max").
			Msg("altering message")
		msg.InstanceID = math.MaxUint64
	case 4:
		log.Debug().
			Str("field", "instanceID").
			Str("modification", "randomizing").
			Msg("altering message")
		msg.InstanceID = rand.Uint64()
	default:
		panic("missing case")
	}
	return msg
}

func alterEon(msg *p2pmsg.DecryptionKeys) *p2pmsg.DecryptionKeys {
	i := rand.Intn(5)
	switch i {
	case 0:
		log.Debug().
			Str("field", "eon").
			Str("modification", "incrementing").
			Msg("altering message")
		msg.Eon++
	case 1:
		log.Debug().
			Str("field", "eon").
			Str("modification", "decrementing").
			Msg("altering message")
		msg.Eon--
	case 2:
		log.Debug().
			Str("field", "eon").
			Str("modification", "setting to zero").
			Msg("altering message")
		msg.Eon = 0
	case 3:
		log.Debug().
			Str("field", "eon").
			Str("modification", "setting to max").
			Msg("altering message")
		msg.Eon = math.MaxUint64
	case 4:
		log.Debug().
			Str("field", "eon").
			Str("modification", "incrementing").
			Msg("altering message")
		msg.Eon = rand.Uint64()
	default:
		panic("missing case")
	}
	return msg
}

func alterKeys(msg *p2pmsg.DecryptionKeys) *p2pmsg.DecryptionKeys {
	i := rand.Intn(7)
	switch i {
	case 0:
		log.Debug().
			Str("field", "keys").
			Str("modification", "removing last").
			Msg("altering message")
		n := len(msg.Keys) - 1
		msg.Keys = msg.Keys[:n]
	case 1:
		log.Debug().
			Str("field", "keys").
			Str("modification", "making key invalid").
			Msg("altering message")
		msg.Keys[0].Key = make([]byte, len(msg.Keys[0].Key))
	case 2:
		log.Debug().
			Str("field", "keys").
			Str("modification", "setting identity to zero").
			Msg("altering message")
		msg.Keys[0].Identity = make([]byte, 52)
	case 3:
		log.Debug().
			Str("field", "keys").
			Str("modification", "reversing order").
			Msg("altering message")
		slices.Reverse(msg.Keys)
	case 4:
		log.Debug().
			Str("field", "keys").
			Str("modification", "adding duplicate key").
			Msg("altering message")
		n := len(msg.Keys)
		msg.Keys = append(msg.Keys, msg.Keys[n-1])
	case 5:
		log.Debug().
			Str("field", "keys").
			Str("modification", "adding short identity").
			Msg("altering message")
		key := &p2pmsg.Key{
			Identity: make([]byte, 51),
			Key:      msg.Keys[0].Key,
		}
		msg.Keys = append(msg.Keys, key)
	case 6:
		log.Debug().
			Str("field", "keys").
			Str("modification", "adding long identity").
			Msg("altering message")
		key := &p2pmsg.Key{
			Identity: make([]byte, 53),
			Key:      msg.Keys[0].Key,
		}
		msg.Keys = append(msg.Keys, key)
	default:
		panic("missing case")
	}
	return msg
}

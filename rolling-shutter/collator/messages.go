package collator

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/shutter-network/shutter/shuttermint/collator/cltrtopics"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
	"google.golang.org/protobuf/proto"
)

type timedEpoch shmsg.TimedEpoch

type message interface {
	implementsMessage()
	GetInstanceID() uint64
}

func (*timedEpoch) implementsMessage()       {}
func (te *timedEpoch) GetInstanceID() uint64 { return te.InstanceID }

func unmarshalP2PMessage(msg *p2p.Message) (message, error) {
	if msg == nil {
		return nil, nil
	}
	switch msg.Topic {
	case cltrtopics.TimedEpoch:
		return unmarshalTimedEpoch(msg)
	default:
		return nil, errors.New("unhandled topic from P2P message")
	}
}

func unmarshalTimedEpoch(msg *p2p.Message) (message, error) {
	timedEpochMsg := shmsg.TimedEpoch{}
	if err := proto.Unmarshal(msg.Message, &timedEpochMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal timed epoch P2P message")
	}
	return (*timedEpoch)(&timedEpochMsg), nil
}

type unhandledTopicError struct {
	topic string
	msg   string
}

func (e *unhandledTopicError) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.topic)
}

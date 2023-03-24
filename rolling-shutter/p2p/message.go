package p2p

import (
	"fmt"
	"reflect"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

// Message envelopes the serialized protobuf bytes with additional topic and sender info.
type Message struct {
	Topic        string
	Message      []byte
	Sender       peer.ID
	ReceivedFrom peer.ID
	ID           string
}

func (msg Message) Unmarshal() (p2pmsg.Message, *p2pmsg.TraceContext, error) {
	var err error

	unmshl, traceContext, err := p2pmsg.Unmarshal(msg.Message)
	if err != nil {
		return nil, traceContext, errors.Wrap(err, "failed to unmarshal message")
	}

	err = unmshl.Validate()
	if err != nil {
		return nil, traceContext, errors.Wrap(err, fmt.Sprintf("verification failed <%s>", reflect.TypeOf(unmshl).String()))
	}
	return unmshl, traceContext, nil
}

package p2p

import (
	"fmt"
	"reflect"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func UnmarshalPubsubMessage(msg *pubsub.Message) (p2pmsg.Message, *p2pmsg.TraceContext, error) {
	var err error

	unmshl, traceContext, err := p2pmsg.Unmarshal(msg.GetData())
	if err != nil {
		return nil, traceContext, errors.Wrap(err, "failed to unmarshal message")
	}

	err = unmshl.Validate()
	if err != nil {
		return nil, traceContext, errors.Wrap(err, fmt.Sprintf("verification failed <%s>", reflect.TypeOf(unmshl).String()))
	}
	return unmshl, traceContext, nil
}

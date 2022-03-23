package p2p

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

// Message envelopes the serialized protobuf bytes with additional topic and sender info.
type Message struct {
	Topic    string
	Message  []byte
	SenderID string
}

func (msg Message) Unmarshal() (shmsg.P2PMessage, error) {
	var err error

	unmshl, err := shmsg.NewP2PMessageFromTopic(msg.Topic)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve deserialisation type")
	}

	if err = proto.Unmarshal(msg.Message, unmshl); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal protobuf <%s>", reflect.TypeOf(unmshl).String()))
	}

	err = unmshl.Validate()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("verification failed <%s>", reflect.TypeOf(unmshl).String()))
	}
	return unmshl, nil
}

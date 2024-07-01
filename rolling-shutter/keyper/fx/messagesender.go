package fx

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/shutterevents/shtxresp"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

// IRetriable is an error that knows if it makes sense to retry an action.
type IRetriable interface {
	error
	IsRetriable() bool
}

// RemoteError is raised for shuttermint messages that return with a result code != 0, i.e. where
// the shuttermint app generated an error.
type RemoteError struct {
	msg string
}

var _ IRetriable = &RemoteError{}

func (remoteError *RemoteError) Error() string {
	return fmt.Sprintf("remote error: %s", remoteError.msg)
}

func (remoteError *RemoteError) IsRetriable() bool {
	return false
}

// MessageSender defines the interface of sending messages to shuttermint.
type MessageSender interface {
	SendMessage(context.Context, *shmsg.Message) error
}

// RPCMessageSender signs messages and sends them via RPC to shuttermint.
type RPCMessageSender struct {
	rpcclient     client.Client
	chainID       string
	signingKey    *ecdsa.PrivateKey
	AllowedToSend bool
}

var _ MessageSender = &RPCMessageSender{}

// MockMessageSender sends all messages to a channel so that they can be checked for testing.
type MockMessageSender struct {
	Msgs chan *shmsg.Message
}

var _ MessageSender = &MockMessageSender{}

var mockMessageSenderBufferSize = 0x10000

// NewRPCMessageSender creates a new RPCMessageSender.
func NewRPCMessageSender(cl client.Client, signingKey *ecdsa.PrivateKey) RPCMessageSender {
	return RPCMessageSender{
		rpcclient:     cl,
		chainID:       "",
		signingKey:    signingKey,
		AllowedToSend: false,
	}
}

// SendMessage signs the given shmsg.Message and sends the message to shuttermint.
func (ms *RPCMessageSender) SendMessage(ctx context.Context, msg *shmsg.Message) error {
	if !ms.AllowedToSend {
		log.Info().Str("msg", msg.String()).Msg("not allowed to send")
		return nil
	}
	if err := ms.maybeFetchChainID(ctx); err != nil {
		return err
	}

	msgWithNonce := ms.addNonceAndChainID(msg)
	signedMessage, err := shmsg.SignMessage(msgWithNonce, ms.signingKey)
	if err != nil {
		return err
	}
	tx := tmtypes.Tx(base64.RawURLEncoding.EncodeToString(signedMessage))
	res, err := ms.rpcclient.BroadcastTxCommit(ctx, tx)
	if err != nil {
		return err
	}

	if res.CheckTx.Code != 0 {
		return &RemoteError{
			msg: fmt.Sprintf("checktx: %s", res.CheckTx.Log),
		}
	}

	switch res.DeliverTx.Code {
	case shtxresp.Ok:
		return nil
	case shtxresp.Seen:
		log.Warn().Str("tx", res.DeliverTx.Log).Msg("delivertx: message already seen, ignoring")
		return nil
	case shtxresp.Error:
		return &RemoteError{
			msg: fmt.Sprintf("delivertx: %s", res.DeliverTx.Log),
		}
	}
	return nil
}

func (ms *RPCMessageSender) addNonceAndChainID(msg *shmsg.Message) *shmsg.MessageWithNonce {
	return &shmsg.MessageWithNonce{
		ChainId:     []byte(ms.chainID),
		RandomNonce: randomNonce(),
		Msg:         msg,
	}
}

func (ms *RPCMessageSender) maybeFetchChainID(ctx context.Context) error {
	if ms.chainID != "" {
		return nil
	}

	info, err := ms.rpcclient.BlockchainInfo(ctx, 0, 0)
	if err != nil {
		return err
	}
	if len(info.BlockMetas) == 0 {
		return errors.Errorf("failed to fetch block meta to check chain id")
	}

	ms.chainID = info.BlockMetas[0].Header.ChainID
	return nil
}

func randomNonce() uint64 {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic("Failed to read random bytes for nonce.")
	}
	return binary.LittleEndian.Uint64(bytes[:])
}

// NewMockMessageSender creates a new MockMessageSender. We use a buffered channel with a rather
// large size in order to simplify writing our tests.
func NewMockMessageSender() MockMessageSender {
	return MockMessageSender{
		Msgs: make(chan *shmsg.Message, mockMessageSenderBufferSize),
	}
}

func (ms *MockMessageSender) SendMessage(_ context.Context, msg *shmsg.Message) error {
	ms.Msgs <- msg
	return nil
}

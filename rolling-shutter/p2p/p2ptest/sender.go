package p2ptest

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type Messaging interface {
	service.Service
	SendMessage(context.Context, p2pmsg.Message, ...retry.Option) error
	AddValidator(valFunc p2p.ValidatorFunc, protos ...p2pmsg.Message)
	AddMessageHandler(mhs ...p2p.MessageHandler)
}

var ErrSendTimeout = errors.New("message send timed out")

type SentMessage struct {
	Time    time.Time
	Message p2pmsg.Message
}

func NewTestMessaging() (*TestMessaging, error) {
	// HACK: set the p2p node config to example values.
	// This is just to initialize it, we won't be starting it's
	// background service.
	p2pConfig := p2p.NewConfig()
	err := p2pConfig.SetExampleValues()
	if err != nil {
		return nil, err
	}
	p2pHandler, err := p2p.New(p2pConfig)
	if err != nil {
		return nil, err
	}
	receiveChan := make(chan p2pmsg.Message)
	return &TestMessaging{
		P2PMessaging: p2pHandler,
		messageIn:    receiveChan,
		MessageIn:    receiveChan,
		SentMessages: []*SentMessage{},
	}, nil
}

type TestMessaging struct {
	*p2p.P2PMessaging
	messageIn <-chan p2pmsg.Message

	MessageIn    chan p2pmsg.Message
	SentMessages []*SentMessage
}

func (ts *TestMessaging) AddHandlerFunc(handlerFunc p2p.HandlerFunc, protos ...p2pmsg.Message) {
	ts.P2PMessaging.AddHandlerFunc(handlerFunc, protos...)
}

func (ts *TestMessaging) AddValidator(valFunc p2p.ValidatorFunc, protos ...p2pmsg.Message) {
	ts.P2PMessaging.AddValidator(valFunc, protos...)
}

func (ts *TestMessaging) AddMessageHandler(mhs ...p2p.MessageHandler) {
	ts.P2PMessaging.AddMessageHandler(mhs...)
}

func (ts *TestMessaging) AddGossipTopic(topic string) {
	ts.P2PMessaging.AddGossipTopic(topic)
}

func (ts *TestMessaging) Start(
	ctx context.Context,
	runner service.Runner,
) error { //nolint:unparam
	runner.Go(
		func() error {
			return ts.runHandleMessages(ctx)
		})
	return nil
}

func (ts *TestMessaging) SendMessage(
	_ context.Context,
	msg p2pmsg.Message,
	retryOpts ...retry.Option,
) error { //nolint:unparam
	_ = retryOpts
	sent := &SentMessage{
		Time:    time.Now(),
		Message: msg,
	}
	ts.SentMessages = append(ts.SentMessages, sent)
	return nil
}

func (ts *TestMessaging) runHandleMessages(ctx context.Context) error {
	// This will consume incoming messages and dispatch them to the registered handler functions
	// If the handler returns messages, then they will be sent to the broadcast
	for {
		select {
		case msg, ok := <-ts.messageIn:
			if !ok {
				return nil
			}
			msgsOut, err := ts.P2PMessaging.Handle(ctx, msg)
			if errors.Is(err, p2p.ErrNoMessageHandler) {
				log.Info().
					Str("topic", msg.Topic()).
					Msg("no message handler found")
				continue
			}
			if err != nil {
				log.Info().
					Err(err).
					Str("topic", msg.Topic()).
					Msg("failed to handle message")
			}
			for _, msgOut := range msgsOut {
				err := ts.SendMessage(ctx, msgOut)
				log.Info().Err(err).Str("message", msgOut.LogInfo()).Str("topic", msgOut.Topic()).
					Msg("failed to send message")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (ts *TestMessaging) StopReceive() {
	close(ts.MessageIn)
}

func (ts *TestMessaging) PushMessage(ctx context.Context, msg p2pmsg.Message, timeout time.Duration) error {
	t := time.NewTimer(timeout)
	select {
	case ts.MessageIn <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return ErrSendTimeout
	}
}

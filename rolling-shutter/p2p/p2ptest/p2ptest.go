// p2ptest contains code for testing code implementing a p2p2.MessageHandler.
package p2ptest

import (
	"context"
	"testing"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/rs/zerolog/log"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

// MustValidateMessageResult calls the handlers ValidateMessage method and ensures it returns the
// expected result.
func MustValidateMessageResult(
	t *testing.T,
	expectedResult pubsub.ValidationResult,
	handler p2p.MessageHandler,
	ctx context.Context, //nolint:revive
	msg p2pmsg.Message,
) {
	t.Helper()
	validationResult, err := handler.ValidateMessage(ctx, msg)
	accepted := validationResult == pubsub.ValidationAccept
	log.Debug().
		Interface("msg", msg).
		Int("result", int(validationResult)).
		Int("expected", int(expectedResult)).Err(err).Msg("ValidateMessage")
	if accepted {
		assert.NilError(t, err, "validation returned error")
	}
	assert.Equal(t, expectedResult, validationResult, "validation did not validate with expected result ")
}

// MustHandleMessage makes sure the handler validates and handles the given message without errors.
func MustHandleMessage(
	t *testing.T,
	handler p2p.MessageHandler,
	ctx context.Context, //nolint:revive
	msg p2pmsg.Message,
) []p2pmsg.Message {
	t.Helper()
	MustValidateMessageResult(t, pubsub.ValidationAccept, handler, ctx, msg)
	msgs, err := handler.HandleMessage(ctx, msg)
	assert.NilError(t, err)
	return msgs
}

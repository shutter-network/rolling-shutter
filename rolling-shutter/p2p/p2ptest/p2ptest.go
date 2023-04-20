// p2ptest contains code for testing code implementing a p2p2.MessageHandler.
package p2ptest

import (
	"context"
	"testing"

	"github.com/rs/zerolog/log"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

// MustValidateMessageResult calls the handlers ValidateMessage method and ensures it returns the
// expected result.
func MustValidateMessageResult(
	t *testing.T,
	expectedResult bool,
	handler p2p.MessageHandler,
	ctx context.Context, //nolint:revive
	msg p2pmsg.Message,
) {
	t.Helper()
	ok, err := handler.ValidateMessage(ctx, msg)
	log.Debug().Interface("msg", msg).Bool("ok?", ok).Bool("expect", expectedResult).Err(err).Msg("ValidateMessage")
	if expectedResult {
		assert.NilError(t, err, "validation returned error")
		assert.Assert(t, ok, "validation failed")
	} else {
		assert.Assert(t, !ok, "validation unexpectedly succeeded")
	}
}

// MustHandleMessage makes sure the handler validates and handles the given message without errors.
func MustHandleMessage(
	t *testing.T,
	handler p2p.MessageHandler,
	ctx context.Context, //nolint:revive
	msg p2pmsg.Message,
) []p2pmsg.Message {
	t.Helper()
	MustValidateMessageResult(t, true, handler, ctx, msg)
	msgs, err := handler.HandleMessage(ctx, msg)
	assert.NilError(t, err)
	return msgs
}

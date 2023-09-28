package jrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/bitwurx/jrpc2"
)

type NodeInfoJRPC struct {
	Server *jrpc2.Server

	httpServer *http.Server
	// getDecryptionKeyCallback func(ctx context.Context, epochID []byte) error
	// requestEonKeyCallback    func(ctx context.Context) error
}

// This struct is used for unmarshaling the method params
type AddParams struct {
	X *float64 `json:"x"`
	Y *float64 `json:"y"`
}

// Each params struct must implement the FromPositional method.
// This method will be passed an array of interfaces if positional parameters
// are passed in the rpc call
func (ap *AddParams) FromPositional(params []interface{}) error {
	if len(params) != 2 {
		return errors.New("exactly two integers are required")
	}

	x := params[0].(float64)
	y := params[1].(float64)
	ap.X = &x
	ap.Y = &y

	return nil
}

func (jrpc *NodeInfoJRPC) GetStatus(ctx context.Context, params json.RawMessage) (
	interface{},
	*jrpc2.ErrorObject,
) {
	gdkParams := new(AddParams)
	if err := jrpc2.ParseParams(params, gdkParams); err != nil {
		return nil, err
	}

	if gdkParams.EonID == nil || gdkParams.EpochID == nil {
		return nil, &jrpc2.ErrorObject{
			Code:    jrpc2.InvalidParamsCode,
			Message: jrpc2.InvalidParamsMsg,
			Data:    "Two parameters required",
		}
	}

	err := snpjrpc.getDecryptionKeyCallback(ctx, *gdkParams.EpochID)
	if err != nil {
		return nil, &jrpc2.ErrorObject{
			Code:    jrpc2.InternalErrorCode,
			Message: jrpc2.InternalErrorMsg,
			Data: fmt.Sprintf(
				"Error requesting decryption key for proposal %s: %v",
				*gdkParams.EpochID,
				err,
			),
		}
	}

	return true, nil
}

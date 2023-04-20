package snpjrpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bitwurx/jrpc2"
	"github.com/pkg/errors"
)

type SnpJRPC struct {
	Server *jrpc2.Server

	getDecryptionKeyCallback func(ctx context.Context, epochId []byte) error
	requestEonKeyCallback    func(ctx context.Context) error
}

type HexEncodedByteArray []byte

type GetDecryptionKeyParams struct {
	EonId   *uint64              `json:"eon_id,string"`
	EpochId *HexEncodedByteArray `json:"proposal"`
}

func (b HexEncodedByteArray) MarshalJSON() ([]byte, error) {
	hexString := hex.EncodeToString(b)
	return json.Marshal(hexString)
}

func (b *HexEncodedByteArray) UnmarshalJSON(data []byte) (err error) {
	var hexString string
	if err = json.Unmarshal(data, &hexString); err != nil {
		return
	}
	*b, err = hex.DecodeString(hexString)
	return
}

func (gdkp *GetDecryptionKeyParams) FromPositional(params []interface{}) error {
	if len(params) != 2 {
		return errors.Errorf("Two parameters required")
	}
	eonId, err := strconv.ParseUint(params[0].(string), 10, 64)
	if err != nil {
		return err
	}
	var epochId HexEncodedByteArray
	epochId, err = hex.DecodeString(params[1].(string))
	if err != nil {
		return err
	}
	gdkp.EonId = &eonId
	gdkp.EpochId = &epochId

	return nil
}

func (snpjrpc *SnpJRPC) GetDecryptionKey(ctx context.Context, params json.RawMessage) (
	interface{},
	*jrpc2.ErrorObject,
) {
	gdkParams := new(GetDecryptionKeyParams)
	if err := jrpc2.ParseParams(params, gdkParams); err != nil {
		return nil, err
	}

	if gdkParams.EonId == nil || gdkParams.EpochId == nil {
		return nil, &jrpc2.ErrorObject{
			Code:    jrpc2.InvalidParamsCode,
			Message: jrpc2.InvalidParamsMsg,
			Data:    "Two parameters required",
		}
	}

	err := snpjrpc.getDecryptionKeyCallback(ctx, *gdkParams.EpochId)
	if err != nil {
		return nil, &jrpc2.ErrorObject{
			Code:    jrpc2.InternalErrorCode,
			Message: jrpc2.InternalErrorMsg,
			Data: fmt.Sprintf(
				"Error requesting decryption key for proposal %s: %v",
				*gdkParams.EpochId,
				err,
			),
		}
	}

	return true, nil
}

func (snpjrpc *SnpJRPC) RequestEonKey(ctx context.Context, params json.RawMessage) (
	interface{},
	*jrpc2.ErrorObject,
) {
	err := snpjrpc.requestEonKeyCallback(ctx)
	if err != nil {
		return nil, &jrpc2.ErrorObject{
			Code:    jrpc2.InternalErrorCode,
			Message: jrpc2.InternalErrorMsg,
			Data: fmt.Sprintf(
				"Error requesting eon key %v",
				err,
			),
		}
	}
	return true, nil
}

func New(
	jsonrpcHost string,
	jsonrpcPort uint16,
	getDecryptionKeyCallback func(ctx context.Context, epochId []byte) error,
	requestEonKeyCallback func(ctx context.Context) error,
) *SnpJRPC {
	host := fmt.Sprintf("%s:%d", jsonrpcHost, jsonrpcPort)
	server := jrpc2.NewServer(host, "/api/v1/rpc", nil)

	jrpc := SnpJRPC{
		Server: server,

		getDecryptionKeyCallback: getDecryptionKeyCallback,
		requestEonKeyCallback:    requestEonKeyCallback,
	}

	server.RegisterWithContext(
		"get_decryption_key",
		jrpc2.MethodWithContext{Method: jrpc.GetDecryptionKey},
	)
	server.RegisterWithContext(
		"request_eon_key",
		jrpc2.MethodWithContext{Method: jrpc.RequestEonKey},
	)

	return &jrpc
}

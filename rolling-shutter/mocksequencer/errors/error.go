package rpcerrors

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
)

const (
	errcodeDefault             = -32000
	errcodeParseError          = -32700
	errcodeInternalError       = -32603
	errcodeTransactionRejected = -32003
)

func Default(err error) error {
	if err == nil {
		return nil
	}
	return &errWithCode{error: err, code: errcodeDefault}
}

func InternalServerError(err error) error {
	if err == nil {
		return nil
	}
	return &errWithCode{error: err, code: errcodeInternalError, rpcMessage: "internal server error"}
}

func ParseError(err error) error {
	if err == nil {
		return nil
	}
	return &errWithCode{error: err, code: errcodeParseError}
}

func TransactionRejected(err error) error {
	if err == nil {
		return nil
	}
	return &errWithCode{error: err, code: errcodeTransactionRejected}
}

type errWithCode struct {
	error
	code       int
	rpcMessage string
}

func (w *errWithCode) ErrorCode() int   { return w.code }
func (w *errWithCode) Error() string    { return w.error.Error() }
func (w *errWithCode) Cause() error     { return w.error }
func (w *errWithCode) RPCError() string { return w.rpcMessage }

type rpcError interface {
	rpc.Error
	RPCError() string
}

type causer interface {
	Cause() error
}

func ExtractRPCError(err error) rpc.Error {
	var errorCode int
	var rpcMessage string

	// retrieve the error code.
	// the outer-most error code is the relevant
	// one, since it was set higher in the call stack
	e := err
	for e != nil {
		rpcErr, ok := e.(rpcError)
		if ok {
			errorCode = rpcErr.ErrorCode()
			rpcMessage = rpcErr.RPCError()
			break
		}
		eWithCause, ok := e.(causer)
		if !ok {
			// don't expose the message to the
			// user, when no explicit rpc-error
			// was set
			// just return a generic error message
			errorCode = errcodeInternalError
			rpcMessage = "internal server error"
			break
		}
		e = eWithCause.Cause()
	}

	if rpcMessage != "" {
		// if the RPCError() returns a non-null string,
		// the error string will be overwritten
		return &errWithCode{error: errors.New(rpcMessage), code: errorCode}
	}
	// otherwise use the outmost error and
	return &errWithCode{error: err, code: errorCode}
}

//nolint:revive
package logger

import "github.com/ethereum/go-ethereum/log"

type NoopLogHandler struct{}

func (nh *NoopLogHandler) Log(r *log.Record) error { return nil }

type NoopLogger struct{}

func (n *NoopLogger) New(ctx ...interface{}) log.Logger    { return &NoopLogger{} }
func (n *NoopLogger) GetHandler() log.Handler              { return &NoopLogHandler{} }
func (n *NoopLogger) SetHandler(h log.Handler)             {}
func (n *NoopLogger) Trace(msg string, ctx ...interface{}) {}
func (n *NoopLogger) Debug(msg string, ctx ...interface{}) {}
func (n *NoopLogger) Info(msg string, ctx ...interface{})  {}
func (n *NoopLogger) Warn(msg string, ctx ...interface{})  {}
func (n *NoopLogger) Error(msg string, ctx ...interface{}) {}
func (n *NoopLogger) Crit(msg string, ctx ...interface{})  {}

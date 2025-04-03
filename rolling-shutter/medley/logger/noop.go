//nolint:revive
package logger

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/log"
)

type NoopLogger struct{}

func (n *NoopLogger) With(ctx ...interface{}) log.Logger                   { return n }
func (n *NoopLogger) New(ctx ...interface{}) log.Logger                    { return &NoopLogger{} }
func (n *NoopLogger) Log(level slog.Level, msg string, ctx ...interface{}) {}
func (n *NoopLogger) Trace(msg string, ctx ...interface{})                 {}
func (n *NoopLogger) Debug(msg string, ctx ...interface{})                 {}
func (n *NoopLogger) Info(msg string, ctx ...interface{})                  {}
func (n *NoopLogger) Warn(msg string, ctx ...interface{})                  {}
func (n *NoopLogger) Error(msg string, ctx ...interface{})                 {}
func (n *NoopLogger) Crit(msg string, ctx ...interface{})                  {}
func (n *NoopLogger) Write(level slog.Level, msg string, attrs ...any)     {}
func (n *NoopLogger) Enabled(ctx context.Context, level slog.Level) bool   { return false }
func (n *NoopLogger) Handler() slog.Handler                                { return slog.Default().Handler() }

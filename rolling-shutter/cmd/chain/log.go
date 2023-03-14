package chain

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tenderlog "github.com/tendermint/tendermint/libs/log"
)

type tendermintLogger struct {
	zerolog.Logger
}

func (l tendermintLogger) Info(msg string, keyVals ...interface{}) {
	l.Logger.Info().Fields(keyVals).Msg(msg)
}

func (l tendermintLogger) Error(msg string, keyVals ...interface{}) {
	e := l.Logger.Error()
	e.Fields(keyVals).Msg(msg)
}

func (l tendermintLogger) Debug(msg string, keyVals ...interface{}) {
	l.Logger.Debug().Fields(keyVals).Msg(msg)
}

func (l tendermintLogger) With(keyVals ...interface{}) tenderlog.Logger {
	return tendermintLogger{
		Logger: l.Logger.With().Fields(keyVals).Logger(),
	}
}

// newLogger sets up a new tendermint logger adapted from the global logger. This will be
// compatible with the logger we configure in cmd/root.go. We use this as a replacement for
// tenderlog.NewDefaultLogger.
func newLogger(level string) (tenderlog.Logger, error) {
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level (%s): %w", level, err)
	}
	logger := log.Logger.Level(logLevel)
	logger = logger.With().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + 1).Logger()
	return tendermintLogger{Logger: logger}, nil
}

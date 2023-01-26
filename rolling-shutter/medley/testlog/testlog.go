// Package testlog configures zerolog for use in tests.
package testlog

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var once sync.Once

// colorize returns the string s wrapped in ANSI code c, unless disabled is true.
func colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

func setup() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	pathsep := string(os.PathSeparator)
	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		return fmt.Sprintf("%s:%d", file[1+strings.LastIndex(file, pathsep):], line)
	}

	zerolog.TimeFieldFormat = "    " // hack to indent log messages
	nocolor := os.Getenv("NOCOLOR") == "1"
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		NoColor:    nocolor,
		Out:        os.Stderr,
		TimeFormat: zerolog.TimeFieldFormat,
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.CallerFieldName,
			zerolog.MessageFieldName,
		},
		FormatCaller: func(i interface{}) string {
			return colorize(fmt.Sprintf("[%20s]", i), 1, nocolor)
		},
	}).With().Caller().Timestamp().Logger()
}

// Setup sets up zerolog for use in tests.
func Setup() {
	once.Do(setup)
}

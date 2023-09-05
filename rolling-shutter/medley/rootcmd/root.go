// Package rootcmd implements a cobra command with logging command line switches
package rootcmd

import (
	"fmt"
	"os"
	"strings"

	golog "github.com/ipfs/go-log/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
)

var (
	logNoColorArg    bool
	ArgNameNocolor   string = "no-color"
	logFormatArg     string
	ArgNameLogformat string = "logformat"
	logLevelArg      string
	ArgNameLoglevel  string = "loglevel"
)

func configureCaller(l zerolog.Logger, short bool) zerolog.Logger {
	if short {
		pathsep := string(os.PathSeparator)
		// default is long filename
		zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
			return fmt.Sprintf("%s:%d", file[1+strings.LastIndex(file, pathsep):], line)
		}
	}
	return l.With().Caller().Logger()
}

func configureTime(l zerolog.Logger) zerolog.Logger {
	zerolog.TimeFieldFormat = "2006/01/02 15:04:05.000000"
	return l.With().Timestamp().Logger()
}

// colorize returns the string s wrapped in ANSI code c, unless disabled is true.
func colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

func setupLogging() (zerolog.Logger, error) {
	// create a basic logger with stdout writer
	// we will change the writer later

	l := zerolog.New(os.Stdout)
	exclude := []string{}
	// change the "message" field name, so that
	// it doesn't collide with e.g. logging of
	// shutter "message"
	zerolog.MessageFieldName = "log"

	logFormat := viper.GetString(ArgNameLogformat)
	switch logFormat {
	case "max", "long":
		l = configureTime(l)
		l = configureCaller(l, true)
	case "short":
		// no time/date logging
		l = configureCaller(l, true)
		exclude = []string{
			zerolog.TimestampFieldName,
		}
	case "min":
		// no time/date logging
		// no caller logging
		exclude = []string{
			zerolog.TimestampFieldName,
			zerolog.CallerFieldName,
		}
	default:
		return l, errors.Errorf("flag '%s' value '%s' not recognized", ArgNameLogformat, logFormat)
	}

	logLevel := viper.GetString(ArgNameLoglevel)
	switch logLevel {
	case "":
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		golog.SetAllLoggers(golog.LevelInfo)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		golog.SetAllLoggers(golog.LevelWarn)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		golog.SetAllLoggers(golog.LevelDebug)
	default:
		return l, errors.Errorf("flag '%s' value '%s' not recognized", ArgNameLoglevel, logLevel)
	}

	// reset the writer
	l = l.Output(zerolog.ConsoleWriter{
		NoColor:    logNoColorArg,
		Out:        os.Stderr,
		TimeFormat: zerolog.TimeFieldFormat,
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.CallerFieldName,
			zerolog.MessageFieldName,
		},
		PartsExclude: exclude,
		FormatCaller: func(i interface{}) string {
			return colorize(fmt.Sprintf("[%20s]", i), 1, logNoColorArg)
		},
	})

	return l, nil
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Version:           shversion.Version(),
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := medley.BindFlags(cmd)
			if err != nil {
				return err
			}

			logger, err := setupLogging()
			if err != nil {
				return errors.Wrap(err, "failed to setup logging")
			}
			log.Logger = logger
			return nil
		},
	}
	viper.BindEnv("NOCOLOR")
	cmd.PersistentFlags().BoolVar(
		&logNoColorArg,
		ArgNameNocolor,
		viper.GetBool("NOCOLOR"),
		"do not write colored logs")

	cmd.PersistentFlags().StringVar(
		&logFormatArg,
		ArgNameLogformat,
		"long",
		"set log format, possible values:  min, short, long, max",
	)
	cmd.PersistentFlags().StringVar(
		&logLevelArg,
		ArgNameLoglevel,
		"info",
		"set log level, possible values:  warn, info, debug",
	)
	return cmd
}

func Main(c *cobra.Command) {
	status := 0
	if err := c.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		status = 1
	}
	os.Exit(status)
}

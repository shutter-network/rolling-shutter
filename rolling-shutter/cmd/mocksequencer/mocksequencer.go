package mocksequencer

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/rpc"
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		// TODO  long usage
		command.Usage(
			"Run a Shutter mock sequencer",
			"",
		),
		command.WithGenerateConfigSubcommand(),
	)
	return builder.Command()
}

func main(config *mocksequencer.Config) error {
	if config.Debug {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Msgf("Starting mock sequencer version %s", shversion.Version())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-termChan
		log.Info().Str("signal", sig.String()).Msg("Received signal, shutting down")
		cancel()
	}()

	l1PollInterval := time.Duration(config.EthereumPollInterval) * time.Second
	sequencer := mocksequencer.New(
		new(big.Int).SetUint64(config.ChainID),
		config.HTTPListenAddress,
		config.EthereumURL,
		l1PollInterval,
		config.MaxBlockDeviation,
	)

	services := []mocksequencer.RPCService{
		&rpc.EthService{},
		&rpc.ShutterService{},
	}
	if config.Admin {
		services = append(services, &rpc.AdminService{})
	}
	log.Info().Str("listen-on", config.HTTPListenAddress).Msg("Serving JSON-RPC")
	err := sequencer.ListenAndServe(
		ctx,
		services...,
	)
	if err == context.Canceled {
		log.Info().Msg("Bye.")
		return nil
	}
	return err
}

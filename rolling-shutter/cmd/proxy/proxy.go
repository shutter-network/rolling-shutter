package proxy

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/proxy"
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		// TODO  long usage
		command.Usage(
			"Run a Ethereum JSON RPC proxy",
			"",
		),
		command.WithGenerateConfigSubcommand(),
	)
	return builder.Command()
}

func main(cfg *proxy.Config) error {
	log.Info().Msgf("Starting shutter proxy version %s", shversion.Version())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-termChan
		log.Info().Str("signal", sig.String()).Msg("Received signal, shutting down")
		cancel()
	}()

	err := proxy.Run(ctx, cfg)
	if err == context.Canceled {
		log.Info().Msg("Bye.")
		return nil
	}
	return err
}

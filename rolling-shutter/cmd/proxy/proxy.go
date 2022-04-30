package proxy

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/proxy"
)

var cfgFile string

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Run a json rpc proxy",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return proxyMain()
		},
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	cmd.MarkPersistentFlagRequired("config")
	cmd.MarkPersistentFlagFilename("config")
	return cmd
}

func readConfig() (proxy.Config, error) {
	config := proxy.Config{}
	viper.AddConfigPath("$HOME/.config/shutter")
	viper.SetConfigName("proxy")
	viper.SetConfigType("toml")
	viper.SetConfigFile(cfgFile)
	var err error
	err = viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found
		if cfgFile != "" {
			return config, err
		}
	} else if err != nil {
		return config, err // Config file was found but another error was produced
	}
	err = config.Unmarshal(viper.GetViper())
	return config, err
}

func proxyMain() error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	config, err := readConfig()
	if err != nil {
		return err
	}

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

	err = proxy.Run(ctx, config)
	if err == context.Canceled {
		log.Info().Msg("Bye.")
		return nil
	}
	return err
}

package mocksequencer

import (
	"bytes"
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
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/rpc"
)

var (
	outputFile string
	cfgFile    string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock-sequencer",
		Short: "Run a node that pretends to be a sequencer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := readConfig()
			if err != nil {
				return err
			}
			return mockSequencerMain(&config)
		},
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	cmd.AddCommand(generateConfigCmd())
	return cmd
}

func exampleConfig() *mocksequencer.Config {
	return &mocksequencer.Config{
		HTTPListenAddress:    ":8555",
		EthereumURL:          "http://localhost:8545",
		ChainID:              42,
		MaxBlockDeviation:    5,
		EthereumPollInterval: 1,
		Admin:                true,
		Debug:                true,
	}
}

func generateConfig() error {
	config := exampleConfig()
	buf := &bytes.Buffer{}
	err := config.WriteTOML(buf)
	if err != nil {
		return err
	}
	return medley.SecureSpit(outputFile, buf.Bytes())
}

func generateConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-config",
		Short: "Generate a mock sequencer configuration file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateConfig()
		},
	}
	cmd.PersistentFlags().StringVar(&outputFile, "output", "", "output file")
	cmd.MarkPersistentFlagRequired("output")

	return cmd
}

func readConfig() (mocksequencer.Config, error) {
	viper.SetEnvPrefix("SEQUENCER")
	viper.BindEnv("EthereumURL")

	viper.SetDefault("HTTPListenAddress", "localhost:8555")
	viper.SetDefault("EthereumURL", "http://localhost:8545")

	viper.SetDefault("ChainID", uint64(42))
	viper.SetDefault("MaxBlockDeviation", uint64(5))
	viper.SetDefault("EthereumPollInterval", uint64(1))
	viper.SetDefault("Admin", true)
	viper.SetDefault("Debug", true)

	config := mocksequencer.Config{}

	viper.AddConfigPath("$HOME/.config/shutter")
	viper.SetConfigName("mocksequencer")
	viper.SetConfigType("toml")
	viper.SetConfigFile(cfgFile)

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found
		if cfgFile != "" {
			return config, err
		}
	} else if err != nil {
		return config, err // Config file was found but another error was produced
	}

	err = config.Unmarshal(viper.GetViper())
	if err != nil {
		return config, err
	}

	return config, nil
}

func mockSequencerMain(config *mocksequencer.Config) error {
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

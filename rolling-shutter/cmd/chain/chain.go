package chain

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abciclient "github.com/tendermint/tendermint/abci/client"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/app"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

var cfgFile string

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain",
		Short: "Run a node for Shutter's Tendermint chain",
		Long:  `This command runs a node that will connect to Shutter's Tendermint chain.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			chainMain()
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "config file (required)")
	cmd.MarkFlagRequired("config")
	cmd.AddCommand(initCmd())
	return cmd
}

func chainMain() {
	log.Info().Str("version", shversion.Version()).Msg("starting shuttermint")
	config, err := readConfig(cfgFile)
	if err != nil {
		log.Error().Err(err).Msg("could not read config file")
		os.Exit(2)
	}

	err = service.RunWithSighandler(context.Background(), &appService{config: config})
	if err != nil {
		log.Error().Err(err).Msg("service failed")
		os.Exit(1)
	}
}

type appService struct {
	config *cfg.Config
}

func (as *appService) Start(ctx context.Context, runner service.Runner) error {
	logger, err := newLogger(as.config.LogLevel)
	if err != nil {
		return errors.Wrap(err, "failed to create tendermint logger")
	}

	nodeid, err := as.config.LoadNodeKeyID()
	if err != nil {
		return err
	}
	log.Info().Str("node-id", string(nodeid)).Msg("loaded node-id")

	shapp, err := app.LoadShutterAppFromFile(
		filepath.Join(as.config.DBDir(), "shutter.gob"),
	)
	if err != nil {
		return err
	}

	tmNode, err := node.New(
		as.config,
		logger,
		abciclient.NewLocalCreator(&shapp),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create new Tendermint node")
	}
	err = tmNode.Start()
	if err != nil {
		return errors.Wrap(err, "failed to start Tendermint node")
	}
	runner.Go(func() error {
		tmNode.Wait()
		log.Debug().Msg("Node stopped")
		return nil
	})
	runner.Go(func() error {
		<-ctx.Done()
		log.Debug().Msg("Stopping node")
		err := tmNode.Stop()
		if err != nil {
			log.Error().Err(err).Msg("failed to stop Tendermint node")
		}
		return nil
	})
	return nil
}

func readConfig(configFile string) (*cfg.Config, error) {
	config := cfg.DefaultConfig()
	config.RootDir = filepath.Dir(filepath.Dir(configFile))
	config.SetRoot(config.RootDir)
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "viper failed to read config file")
	}
	if err := viper.Unmarshal(config); err != nil {
		return nil, errors.Wrap(err, "viper failed to unmarshal config")
	}
	if err := config.ValidateBasic(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}
	return config, nil
}

package chain

import (
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abciclient "github.com/tendermint/tendermint/abci/client"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/service"
	"github.com/tendermint/tendermint/node"

	"github.com/shutter-network/shutter/shuttermint/app"
	"github.com/shutter-network/shutter/shuttermint/cmd/shversion"
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
	stdlog.Printf("Starting shuttermint version %s", shversion.Version())

	node, err := newTendermint(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
	err = node.Start()
	if err != nil {
		panic(err)
	}
	defer func() {
		err = node.Stop()
		if err != nil {
			panic(err)
		}
		node.Wait()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	stdlog.Printf("Got signal '%s'. Exiting.", sig)
	// Previously we had an os.Exit(0) call here, but now we do wait until the defer function
	// above is done
}

func newTendermint(configFile string) (service.Service, error) {
	// read config
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
	nodeid, err := config.LoadNodeKeyID()
	if err != nil {
		return nil, err
	}
	stdlog.Printf("Id: %s", nodeid)
	logger, err := log.NewDefaultLogger(config.LogFormat, config.LogLevel, false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create default logger")
	}

	shapp, err := app.LoadShutterAppFromFile(
		filepath.Join(config.DBDir(), "shutter.gob"))
	if err != nil {
		return nil, err
	}

	srvc, err := node.New(
		config,
		logger,
		abciclient.NewLocalCreator(&shapp),
		nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new Tendermint node")
	}
	return srvc, nil
}

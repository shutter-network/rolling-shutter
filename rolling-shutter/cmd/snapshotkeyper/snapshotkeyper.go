package snapshotkeyper

import (
	"bytes"
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshotkeyper"
)

var (
	cfgFile    string
	outputFile string
)

func initDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initdb",
		Short: "Initialize the database of the snapshot keyper",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initDB()
		},
	}
	return cmd
}

func generateConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-config",
		Short: "Generate a snapshot keyper configuration file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateConfig()
		},
	}
	cmd.PersistentFlags().StringVar(&outputFile, "output", "", "output file")
	cmd.MarkPersistentFlagRequired("output")
	return cmd
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshotkeyper",
		Short: "Run a Shutter snapshotkeyper node",
		Long: `This command runs a keyper node. It will connect to both an Ethereum and a
Shuttermint node which have to be started separately in advance.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return keyperMain()
		},
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	cmd.AddCommand(initDBCmd())
	cmd.AddCommand(generateConfigCmd())
	return cmd
}

func readKeyperConfig() (snapshotkeyper.Config, error) {
	viper.SetEnvPrefix("KEYPER")
	viper.BindEnv("ShuttermintURL")
	viper.BindEnv("EthereumURL")
	viper.BindEnv("ContractsURL")
	viper.BindEnv("DeploymentDir")
	viper.BindEnv("SigningKey")
	viper.BindEnv("ValidatorSeed")
	viper.BindEnv("EncryptionKey")
	viper.BindEnv("DKGPhaseLength")
	viper.BindEnv("DatabaseURL")
	viper.BindEnv("ListenAddress")
	viper.BindEnv("CustomBootstrapAddresses")

	viper.SetDefault("ShuttermintURL", "http://localhost:26657")
	defaultListenAddress, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/2000")
	viper.SetDefault("ListenAddress", defaultListenAddress)
	viper.SetDefault("CustomBootstrapAddresses", []peer.AddrInfo{})

	defer func() {
		if viper.ConfigFileUsed() != "" {
			log.Info().Str("config", viper.ConfigFileUsed()).Msg("read config")
		}
	}()
	var err error
	config := snapshotkeyper.Config{}

	viper.AddConfigPath("$HOME/.config/shutter")
	viper.SetConfigName("keyper")
	viper.SetConfigType("toml")
	viper.SetConfigFile(cfgFile)

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

	if err != nil {
		return config, err
	}

	return config, err
}

func keyperMain() error {
	config, err := readKeyperConfig()
	if err != nil {
		return errors.WithMessage(err, "Please check your configuration")
	}

	log.Info().
		Str("version", shversion.Version()).
		Str("address", config.GetAddress().Hex()).
		Str("shuttermint", config.Shuttermint.ShuttermintURL).
		Msg("starting snapshotkeyper")

	return service.RunWithSighandler(context.Background(), snapshotkeyper.New(config))
}

func initDB() error {
	ctx := context.Background()

	kc, err := readKeyperConfig()
	if err != nil {
		return errors.WithMessage(err, "Please check your configuration")
	}

	dbpool, err := pgxpool.Connect(ctx, kc.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()

	err = kprdb.ValidateKeyperDB(ctx, dbpool)
	if err == nil {
		shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database already exists")
		return nil
	} else if errors.Is(err, metadb.ErrSchemaMismatch) {
		return err
	}

	// initialize the db
	err = kprdb.InitDB(ctx, dbpool)
	if err != nil {
		return err
	}
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database initialized")
	return nil
}

func exampleConfig() (*snapshotkeyper.Config, error) {
	cfg := &snapshotkeyper.Config{
		ShuttermintURL:     "http://localhost:26657",
		EthereumURL:        "http://127.0.0.1:8545/",
		ContractsURL:       "http://127.0.0.1:8555/",
		DeploymentDir:      "./deployments/localhost/",
		DKGPhaseLength:     30,
		DKGStartBlockDelta: 12000,
		ListenAddresses: []multiaddr.Multiaddr{
			p2p.MustMultiaddr("/ip4/127.0.0.1/tcp/2000"),
		},
		CustomBootstrapAddresses: []peer.AddrInfo{
			p2p.MustAddrInfo("/ip4/127.0.0.1/tcp/2001/p2p/QmdfBeR6odD1pRKendUjWejhMd9wybivDq5RjixhRhiERg"),
			p2p.MustAddrInfo("/ip4/127.0.0.1/tcp/2002/p2p/QmV9YbMDLDi736vTzy97jn54p43o74fLxc5DnLUrcmK6WP"),
		},
		InstanceID: 0,

		HTTPEnabled:       false,
		HTTPListenAddress: ":3000",
	}
	err := cfg.GenerateNewKeys()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func generateConfig() error {
	cfg, err := exampleConfig()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err = cfg.WriteTOML(buf); err != nil {
		return errors.Wrap(err, "failed to write keyper config file")
	}
	return medley.SecureSpit(outputFile, buf.Bytes())
}

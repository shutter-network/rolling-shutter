package collator

import (
	"bytes"
	"context"
	"crypto/rand"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4/pgxpool"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

var (
	outputFile string
	cfgFile    string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collator",
		Short: "Run a collator node",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return main()
		},
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	cmd.AddCommand(initDBCmd())
	cmd.AddCommand(generateConfigCmd())
	return cmd
}

func initDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initdb",
		Short: "Initialize the database of the collator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initDB()
		},
	}
	return cmd
}

func initDB() error {
	ctx := context.Background()

	cfg, err := readConfig()
	if err != nil {
		return err
	}

	dbpool, err := pgxpool.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()

	err = cltrdb.ValidateDB(ctx, dbpool)
	if err == nil {
		shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database already exists")
		return nil
	} else if errors.Is(err, metadb.ErrSchemaMismatch) {
		return err
	}

	// initialize the db
	err = cltrdb.InitDB(ctx, dbpool)
	if err != nil {
		return err
	}
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database initialized")
	return nil
}

func generateConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-config",
		Short: "Generate a collator configuration file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateConfig()
		},
	}
	cmd.PersistentFlags().StringVar(&outputFile, "output", "", "output file")
	cmd.MarkPersistentFlagRequired("output")

	return cmd
}

func readConfig() (config.Config, error) {
	viper.SetEnvPrefix("COLLATOR")
	viper.BindEnv("EthereumURL")
	viper.BindEnv("SequencerURL")
	viper.BindEnv("DeploymentDir")
	viper.BindEnv("ListenAddress")
	viper.BindEnv("CustomBootstrapAddresses")
	viper.BindEnv("EthereumKey")
	viper.BindEnv("EpochDuration")
	defaultListenAddress, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/2000")
	viper.SetDefault("ListenAddress", defaultListenAddress)
	viper.SetDefault("CustomBootstrapAddresses", []peer.AddrInfo{})
	viper.SetDefault("EpochDuration", "5s")

	cfg := config.Config{}

	viper.AddConfigPath("$HOME/.config/shutter")
	viper.SetConfigName("collator")
	viper.SetConfigType("toml")
	viper.SetConfigFile(cfgFile)

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found
		if cfgFile != "" {
			return cfg, err
		}
	} else if err != nil {
		return cfg, err // Config file was found but another error was produced
	}

	err = cfg.Unmarshal(viper.GetViper())
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func exampleConfig() (*config.Config, error) {
	ethereumKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	p2pkey, _, err := p2pcrypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &config.Config{
		EthereumURL:   "http://127.0.0.1:8545/",
		ContractsURL:  "http://127.0.0.1:8545/",
		DeploymentDir: "./deployments/localhost/",
		ListenAddresses: []multiaddr.Multiaddr{
			p2p.MustMultiaddr("/ip4/127.0.0.1/tcp/2000"),
		},
		CustomBootstrapAddresses: []peer.AddrInfo{},
		DatabaseURL:              "",
		SequencerURL:             "http://127.0.0.1:9545/",
		HTTPListenAddress:        ":3000",
		EthereumKey:              ethereumKey,
		P2PKey:                   p2pkey,
		EpochDuration:            5 * time.Second,
		ExecutionBlockDelay:      5,
	}, nil
}

func generateConfig() error {
	cfg, err := exampleConfig()
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	err = cfg.WriteTOML(buf)
	if err != nil {
		return err
	}

	return medley.SecureSpit(afero.NewOsFs(), outputFile, buf.Bytes())
}

func main() error {
	cfg, err := readConfig()
	if err != nil {
		return err
	}
	return service.RunWithSighandler(context.Background(), collator.New(cfg))
}

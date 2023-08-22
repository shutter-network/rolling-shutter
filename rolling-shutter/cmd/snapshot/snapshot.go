package snapshot

import (
	"bytes"
	"context"
	"crypto/rand"

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

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/snpdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot"
)

var (
	cfgFile    string
	outputFile string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Run the Snapshot Hub communication module",
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
		Short: "Initialize the database of the snapshot module",
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
		Short: "Generate a snapshot configuration file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateConfig()
		},
	}
	cmd.PersistentFlags().StringVar(&outputFile, "output", "", "output file")
	err := cmd.MarkPersistentFlagRequired("output")
	if err != nil {
		return nil
	}

	return cmd
}

func initDB() error {
	ctx := context.Background()

	config, err := readConfig()
	if err != nil {
		return err
	}

	dbpool, err := pgxpool.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()

	err = snpdb.ValidateSnapshotDB(ctx, dbpool)
	if err == nil {
		shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database already exists")
		return nil
	} else if errors.Is(err, metadb.ErrSchemaMismatch) {
		return err
	}

	// initialize the db
	err = snpdb.InitDB(ctx, dbpool)
	if err != nil {
		return err
	}
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database initialized")

	return nil
}

func readConfig() (snapshot.Config, error) {
	viper.SetEnvPrefix("SNAPSHOT")
	viper.BindEnv("EthereumURL")
	viper.BindEnv("ListenAddresses")
	viper.BindEnv("CustomBootstrapAddresses")
	viper.BindEnv("SnapshotHubURL")

	defaultListenAddress, _ := multiaddr.NewMultiaddr("/ip6/::1/tcp/2000")
	viper.SetDefault("ListenAddresses", []multiaddr.Multiaddr{defaultListenAddress})
	viper.SetDefault("CustomBootstrapAddresses", []peer.AddrInfo{})

	config := snapshot.Config{}

	viper.AddConfigPath("$HOME/.config/shutter")
	viper.SetConfigName("snapshot")
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

func exampleConfig() (*snapshot.Config, error) {
	ethereumKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	p2pkey, _, err := p2pcrypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &snapshot.Config{
		EthereumURL: "http://[::1]:8545/",
		ListenAddresses: []multiaddr.Multiaddr{
			p2p.MustMultiaddr("/ip4/127.0.0.1/tcp/2000"),
		},
		CustomBootstrapAddresses: []peer.AddrInfo{
			p2p.MustAddrInfo(
				"/ip4/127.0.0.1/tcp/2001/p2p/QmdfBeR6odD1pRKendUjWejhMd9wybivDq5RjixhRhiERg",
			),
			p2p.MustAddrInfo(
				"/ip4/127.0.0.1/tcp/2002/p2p/QmV9YbMDLDi736vTzy97jn54p43o74fLxc5DnLUrcmK6WP",
			),
		},
		DatabaseURL: "postgres://localhost:5432/shutter_snapshot",

		SnapshotHubURL: "",

		MetricsEnabled: false,
		MetricsHost:    "127.0.0.1",
		MetricsPort:    9191,

		EthereumKey: ethereumKey,
		P2PKey:      p2pkey,
	}, nil
}

func generateConfig() error {
	config, err := exampleConfig()
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	err = config.WriteTOML(buf)
	if err != nil {
		return err
	}
	return medley.SecureSpit(afero.NewOsFs(), outputFile, buf.Bytes())
}

func main() error {
	config, err := readConfig()
	if err != nil {
		return err
	}
	return service.RunWithSighandler(
		context.Background(),
		snapshot.New(config),
	)
}

package snapshot

import (
	"bytes"
	"context"
	"crypto/rand"
	"os"
	"os/signal"
	"syscall"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4/pgxpool"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot/snpdb"
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
	viper.SetDefault("ListenAddress", defaultListenAddress)
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
			p2p.MustMultiaddr("/ip6/::1/tcp/2000"),
		},
		CustomBootstrapAddresses: []peer.AddrInfo{},
		DatabaseURL:              "",
		SnapshotHubURL:           "",

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
	return medley.SecureSpit(outputFile, buf.Bytes())
}

func main() error {
	config, err := readConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-termChan
		log.Info().Str("signal", sig.String()).Msg("received  OS signal, shutting down")
		cancel()
	}()

	d := snapshot.New(config)
	err = d.Run(ctx)
	if err == context.Canceled {
		log.Info().Msg("bye")
		return nil
	}
	return err
}

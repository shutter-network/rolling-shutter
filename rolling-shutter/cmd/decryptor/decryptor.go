package decryptor

import (
	"bytes"
	"context"
	"crypto/rand"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/libp2p/go-libp2p-core/crypto"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/decryptor"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
)

var (
	outputFile string
	cfgFile    string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decryptor",
		Short: "Run a decryptor node",
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
		Short: "Initialize the database of the decryptor",
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
		Short: "Generate a decryptor configuration file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateConfig()
		},
	}
	cmd.PersistentFlags().StringVar(&outputFile, "output", "", "output file")
	cmd.MarkPersistentFlagRequired("output")

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
	err = dcrdb.InitDecryptorDB(ctx, dbpool)
	if err != nil {
		return err
	}
	log.Printf("Database initialized (%s)", shdb.ConnectionInfo(dbpool))

	return nil
}

func readConfig() (decryptor.Config, error) {
	viper.SetEnvPrefix("DECRYPTOR")
	viper.BindEnv("ListenAddress")
	viper.BindEnv("PeerMultiaddrs")
	viper.SetDefault("ShuttermintURL", "http://localhost:26657")
	defaultListenAddress, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/2000")
	viper.SetDefault("ListenAddress", defaultListenAddress)
	viper.SetDefault("PeerMultiaddrs", make([]multiaddr.Multiaddr, 0))

	config := decryptor.Config{}

	viper.AddConfigPath("$HOME/.config/shutter")
	viper.SetConfigName("decryptor")
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

func exampleConfig() (*decryptor.Config, error) {
	p2pkey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	signingKey, _, err := shbls.RandomKeyPair(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &decryptor.Config{
		ListenAddress:  p2p.MustMultiaddr("/ip4/127.0.0.1/tcp/2000"),
		PeerMultiaddrs: []multiaddr.Multiaddr{},
		DatabaseURL:    "",
		P2PKey:         p2pkey,
		SigningKey:     signingKey,
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
	ctx := context.Background()

	config, err := readConfig()
	if err != nil {
		return err
	}

	d := decryptor.New(config)
	return d.Run(ctx)
}

package cmd

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/p2p"
)

var gossipTopicNames = [3]string{"cipherBatch", "decryptionKey", "signature"}

var decryptorCmd = &cobra.Command{
	Use:   "decryptor",
	Short: "Run a decryptor node",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return decryptorMain()
	},
}

var initDecryptorDBCmd = &cobra.Command{
	Use:   "initdb",
	Short: "Initialize the database of the decryptor",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initDecryptorDB()
	},
}

type DecryptorConfig struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
	DatabaseURL    string
}

func init() {
	decryptorCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	decryptorCmd.AddCommand(initDecryptorDBCmd)
}

func initDecryptorDB() error {
	ctx := context.Background()

	config, err := readDecryptorConfig()
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
	log.Println("database successfully initialized")

	return nil
}

func readDecryptorConfig() (DecryptorConfig, error) {
	viper.SetEnvPrefix("DECRYPTOR")
	viper.BindEnv("ListenAddress")
	viper.BindEnv("PeerMultiaddrs")
	viper.SetDefault("ShuttermintURL", "http://localhost:26657")
	defaultListenAddress, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/2000")
	viper.SetDefault("ListenAddress", defaultListenAddress)
	viper.SetDefault("PeerMultiaddrs", make([]multiaddr.Multiaddr, 0))

	config := DecryptorConfig{}

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

	err = viper.Unmarshal(&config, viper.DecodeHook(medley.MultiaddrHook()))
	if err != nil {
		return config, err
	}

	return config, nil
}

func decryptorMain() error {
	ctx := context.Background()

	config, err := readDecryptorConfig()
	if err != nil {
		return err
	}

	dbpool, err := pgxpool.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	err = dcrdb.ValidateDecryptorDB(ctx, dbpool)
	if err != nil {
		return err
	}

	p := p2p.NewP2p()
	if err := p.CreateHost(ctx, config.ListenAddress); err != nil {
		return err
	}
	if err := p.JoinTopics(ctx, gossipTopicNames[:]); err != nil {
		return err
	}
	if err := p.ConnectToPeers(ctx, config.PeerMultiaddrs); err != nil {
		return err
	}

	return nil
}

package decryptor

import (
	"bytes"
	"context"
	"crypto/rand"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mitchellh/mapstructure"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/p2p"
)

var (
	outputFile       string
	cfgFile          string
	gossipTopicNames = [3]string{"cipherBatch", "decryptionKey", "signature"}
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

type DecryptorConfig struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
	DatabaseURL    string
	P2PKey         crypto.PrivKey
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
	log.Println("database successfully initialized")

	return nil
}

func readConfig() (DecryptorConfig, error) {
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

	err = viper.Unmarshal(
		&config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				medley.MultiaddrHook,
				medley.P2PKeyHook,
			),
		),
	)
	if err != nil {
		return config, err
	}

	return config, nil
}

var decryptorTemplate = medley.MustBuildTemplate(
	"decryptor",
	`# Shutter decryptor config for /p2p/{{ .P2PKey | P2PKeyPublic}}

# DatabaseURL looks like postgres://username:password@localhost:5432/database_name
# It it's empty, we use the standard PG* environment variables
DatabaseURL     = "{{ .DatabaseURL }}"

# p2p configuration
ListenAddress   = "{{ .ListenAddress }}"
PeerMultiaddrs  = [{{ .PeerMultiaddrs | QuoteList}}]

# Secret Keys
P2PKey          = "{{ .P2PKey | P2PKey}}"
`)

func mustMultiaddr(s string) multiaddr.Multiaddr {
	a, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		panic(err)
	}
	return a
}

func generateConfig() error {
	p2pkey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return err
	}

	config := DecryptorConfig{
		ListenAddress:  mustMultiaddr("/ip4/127.0.0.1/tcp/2000"),
		PeerMultiaddrs: nil,
		DatabaseURL:    "",
		P2PKey:         p2pkey,
	}
	buf := &bytes.Buffer{}
	err = decryptorTemplate.Execute(buf, config)
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

	dbpool, err := pgxpool.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	err = dcrdb.ValidateDecryptorDB(ctx, dbpool)
	if err != nil {
		return err
	}

	p := p2p.NewP2PWithKey(config.P2PKey)
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

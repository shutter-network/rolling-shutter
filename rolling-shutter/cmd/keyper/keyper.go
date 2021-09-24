package keyper

import (
	"bytes"
	"context"
	"log"
	"path/filepath"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/client/http"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/shutter/shuttermint/cmd/shversion"
	"github.com/shutter-network/shutter/shuttermint/keyper"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/p2p"
)

var (
	cfgFile    string
	outputFile string
)

func initDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initdb",
		Short: "Initialize the database of the keyper",
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
		Short: "Generate a keyper configuration file",
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
		Use:   "keyper",
		Short: "Run a Shutter keyper node",
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

func readKeyperConfig() (keyper.Config, error) {
	viper.SetEnvPrefix("KEYPER")
	viper.BindEnv("ShuttermintURL")
	viper.BindEnv("SigningKey")
	viper.BindEnv("ValidatorSeed")
	viper.BindEnv("EncryptionKey")
	viper.BindEnv("DKGPhaseLength")
	viper.BindEnv("DatabaseURL")
	viper.BindEnv("ListenAddress")
	viper.BindEnv("PeerMultiaddrs")

	viper.SetDefault("ShuttermintURL", "http://localhost:26657")
	defaultListenAddress, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/2000")
	viper.SetDefault("ListenAddress", defaultListenAddress)
	viper.SetDefault("PeerMultiaddrs", make([]multiaddr.Multiaddr, 0))

	defer func() {
		if viper.ConfigFileUsed() != "" {
			log.Printf("Read config from %s", viper.ConfigFileUsed())
		}
	}()
	var err error
	config := keyper.Config{}

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

	if !filepath.IsAbs(config.DBDir) {
		r := filepath.Dir(viper.ConfigFileUsed())
		dbdir, err := filepath.Abs(filepath.Join(r, config.DBDir))
		if err != nil {
			return config, err
		}
		config.DBDir = dbdir
	}

	return config, err
}

func keyperMain() error {
	ctx := context.Background()

	kc, err := readKeyperConfig()
	if err != nil {
		return errors.WithMessage(err, "Please check your configuration")
	}

	log.Printf(
		"Starting keyper version %s with signing key %s, using %s for Shuttermint",
		shversion.Version(),
		kc.Address().Hex(),
		kc.ShuttermintURL,
	)

	dbpool, err := pgxpool.Connect(ctx, kc.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()

	var cl client.Client
	cl, err = http.New(kc.ShuttermintURL, "/websocket")
	if err != nil {
		return err
	}
	err = keyper.SyncAppWithDB(ctx, kc, cl, dbpool)
	if err != nil {
		return err
	}

	if err := kprdb.ValidateKeyperDB(ctx, dbpool); err != nil {
		return err
	}

	p := p2p.New(p2p.Config{
		ListenAddr:     kc.ListenAddress,
		PeerMultiaddrs: kc.PeerMultiaddrs,
		PrivKey:        kc.P2PKey,
	})

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return p.Run(ctx, keyper.GossipTopicNames)
	})
	return group.Wait()
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

	// initialize the db
	err = kprdb.InitKeyperDB(ctx, dbpool)
	if err != nil {
		return err
	}
	log.Println("database successfully initialized")

	return nil
}

func exampleConfig() (*keyper.Config, error) {
	cfg := &keyper.Config{
		ShuttermintURL: "http://localhost:26657",
		DKGPhaseLength: 30,
		ListenAddress:  p2p.MustMultiaddr("/ip4/127.0.0.1/tcp/2000"),
		PeerMultiaddrs: []multiaddr.Multiaddr{
			p2p.MustMultiaddr("/ip4/127.0.0.1/tcp/2001/p2p/QmdfBeR6odD1pRKendUjWejhMd9wybivDq5RjixhRhiERg"),
			p2p.MustMultiaddr("/ip4/127.0.0.1/tcp/2002/p2p/QmV9YbMDLDi736vTzy97jn54p43o74fLxc5DnLUrcmK6WP"),
		},
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

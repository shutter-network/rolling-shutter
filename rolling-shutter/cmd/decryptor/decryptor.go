package decryptor

import (
	"bytes"
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4/pgxpool"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/decryptor"
	"github.com/shutter-network/shutter/shuttermint/decryptor/blsregistry"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/medley/txbatch"
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
	cmd.AddCommand(registerIdentityCmd())
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

func registerIdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-identity",
		Short: "Register the decryptors identity on chain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return registerIdentity()
		},
	}

	return cmd
}

func registerIdentity() error {
	ctx := context.Background()

	config, err := readConfig()
	if err != nil {
		return err
	}

	ethereumClient, err := ethclient.DialContext(ctx, config.EthereumURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(ethereumClient, config.DeploymentDir)
	if err != nil {
		return err
	}

	batch, err := txbatch.New(ctx, ethereumClient, config.EthereumKey)
	if err != nil {
		return err
	}

	/*
	   If we don't set the gas price here, we run into
	   https://github.com/ethereum/go-ethereum/issues/23479
	*/
	batch.TransactOpts.GasPrice = big.NewInt(2 * 1000000000)

	err = blsregistry.Register(batch, contracts, config.SigningKey)
	if err != nil {
		return err
	}
	_, err = batch.WaitMined(ctx)
	return err
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
	err = dcrdb.InitDB(ctx, dbpool)
	if err != nil {
		return err
	}
	log.Printf("Database initialized (%s)", shdb.ConnectionInfo(dbpool))

	return nil
}

func readConfig() (decryptor.Config, error) {
	viper.SetEnvPrefix("DECRYPTOR")
	viper.BindEnv("EthereumURL")
	viper.BindEnv("DeploymentDir")
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
	ethereumKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	p2pkey, _, err := p2pcrypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	signingKey, _, err := shbls.RandomKeyPair(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &decryptor.Config{
		EthereumURL:    "http://127.0.0.1:8545/",
		DeploymentDir:  "./deployments/localhost/",
		ListenAddress:  p2p.MustMultiaddr("/ip4/127.0.0.1/tcp/2000"),
		PeerMultiaddrs: []multiaddr.Multiaddr{},
		DatabaseURL:    "",

		EthereumKey: ethereumKey,
		P2PKey:      p2pkey,
		SigningKey:  signingKey,
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
		log.Printf("Received %s signal, shutting down", sig)
		cancel()
	}()

	d := decryptor.New(config)
	err = d.Run(ctx)
	if err == context.Canceled {
		log.Printf("Bye.")
		return nil
	}
	return err
}

package cmd

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shuttermint/cmd/shversion"
	"github.com/shutter-network/shutter/shuttermint/keyper"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
)

var outputFile string

// keyperCmd represents the keyper command.
var keyperCmd = &cobra.Command{
	Use:   "keyper",
	Short: "Run a Shutter keyper node",
	Long: `This command runs a keyper node. It will connect to both an Ethereum and a
Shuttermint node which have to be started separately in advance.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return keyperMain()
	},
}

var initKeyperDBCmd = &cobra.Command{
	Use:   "initdb",
	Short: "Initialize the database of the keyper",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initKeyperDB()
	},
}

var generateConfigCmd = &cobra.Command{
	Use:   "generate-config",
	Short: "Generate a keyper configuration file",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateKeyperConfig()
	},
}

func init() {
	keyperCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	keyperCmd.AddCommand(initKeyperDBCmd)
	generateConfigCmd.PersistentFlags().StringVar(&outputFile, "output", "", "output file")
	generateConfigCmd.MarkPersistentFlagRequired("output")
	keyperCmd.AddCommand(generateConfigCmd)
}

func readKeyperConfig() (keyper.Config, error) {
	viper.SetEnvPrefix("KEYPER")
	viper.BindEnv("ShuttermintURL")
	viper.BindEnv("SigningKey")
	viper.BindEnv("ValidatorSeed")
	viper.BindEnv("EncryptionKey")
	viper.BindEnv("DKGPhaseLength")
	viper.BindEnv("DatabaseURL")

	viper.SetDefault("ShuttermintURL", "http://localhost:26657")

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

	if err := kprdb.ValidateKeyperDB(ctx, dbpool); err != nil {
		return err
	}

	return errors.Errorf("keyper command not implemented")
}

func initKeyperDB() error {
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

func generateKeyperConfig() error {
	cfg := keyper.Config{
		ShuttermintURL: "http://localhost:26657",
		DKGPhaseLength: 30,
	}
	err := cfg.GenerateNewKeys()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return errors.Wrap(err, "failed to create keyper config file")
	}
	if err = cfg.WriteTOML(file); err != nil {
		return errors.Wrap(err, "failed to write keyper config file")
	}
	if err = file.Close(); err != nil {
		return errors.Wrap(err, "failed to close keyper config file")
	}
	return nil
}

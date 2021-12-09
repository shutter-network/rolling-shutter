package snapshot

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/snapshot"
	"github.com/shutter-network/shutter/shuttermint/snapshot/snpdb"
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
	err = snpdb.InitSnapshotDB(ctx, dbpool)
	if err != nil {
		return err
	}
	log.Printf("Database initialized (%s)", shdb.ConnectionInfo(dbpool))

	return nil
}

func readConfig() (snapshot.Config, error) {
	viper.SetEnvPrefix("SNAPSHOT")

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
	return &snapshot.Config{}, nil
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

	d := snapshot.New(config)
	err = d.Run(ctx)
	if err == context.Canceled {
		log.Printf("Bye.")
		return nil
	}
	return err
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var decryptorCmd = &cobra.Command{
	Use:   "decryptor",
	Short: "Run a decryptor node",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return decryptorMain()
	},
}

type DecryptorConfig struct {
}

func init() {
	decryptorCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func decryptorMain() error {
	config, err := readDecryptorConfig()
	if err != nil {
		return err
	}
	fmt.Printf("%+v", config)
	return nil
}

func readDecryptorConfig() (DecryptorConfig, error) {
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

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}

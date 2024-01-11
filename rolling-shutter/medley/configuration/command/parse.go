package command

import (
	"bytes"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
)

func CommandAddOutputFileFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String("output", "", "output file")
	cmd.MarkPersistentFlagRequired("output")
}

func CommandAddConfigFileFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String("config", "", "config file")
	cmd.MarkPersistentFlagFilename("config")
}

func WriteConfig(fs afero.Fs, config configuration.Config, outPath string) error {
	buf := &bytes.Buffer{}
	if err := configuration.WriteTOML(buf, config); err != nil {
		return errors.Wrap(err, "failed to write config file")
	}
	return medley.SecureSpit(fs, outPath, buf.Bytes())
}

// ParseCLI reads in the CLI argument context from the
// cobra.Command instance and unmarshals the configuration.
// For this to work, all field-types defined on the config
// have to be native types or implement the encoding.TextUnmarshaler
// interface.
func ParseCLI(v *viper.Viper, cmd *cobra.Command, config configuration.Config) error {
	if v == nil {
		// get the global viper instance
		v = viper.GetViper()
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	v.AddConfigPath("$HOME/.config/shutter")
	v.SetConfigName(config.Name())
	v.SetConfigType("toml")
	v.SetConfigFile(configPath)
	defer func() {
		if v.ConfigFileUsed() != "" {
			log.Info().Str("config", v.ConfigFileUsed()).Msg("read config")
		}
	}()

	envVars := configuration.GetEnvironmentVarsRecursive(config)
	for key, vars := range envVars {
		args := []string{key}
		args = append(args, vars...)
		err := v.BindEnv(args...)
		if err != nil {
			return err
		}
	}

	err = v.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found
		if configPath != "" {
			return err
		}
	} else if err != nil {
		return err // Config file was found but another error was produced
	}

	return ParseViper(v, config)
}

// Parse reads in the CLI argument context from the
// cobra.Command instance and unmarshals the configuration.
// For this to work, all field-types defined on the config
// have to be native types or implement the encoding.TextUnmarshaler
// interface.
func ParseViper(v *viper.Viper, config configuration.Config) error {
	// This filtering is here because the AllKeys() also returns keys with value nil,
	// although it doesn't say so in the docstring
	keysSetByUser := []string{}
	for _, k := range v.AllKeys() {
		value := v.Get(k)
		if value == nil {
			// should not happen, since AllKeys() returns only keys holding a value,
			// check just in case anything changes
			continue
		}
		keysSetByUser = append(keysSetByUser, k)
	}

	// set the default values recursively for all configuration options
	// the user did not provide by any means
	err := configuration.SetDefaultValuesRecursive(config, keysSetByUser)
	if err != nil {
		return err
	}
	// pre-validation: check that the user set
	// values for all non-defaults
	required := configuration.GetRequiredRecursive(config)
	notSet := []string{}
	for varPath := range required {
		if !v.IsSet(varPath) {
			notSet = append(notSet, varPath)
		}
	}
	if len(notSet) > 0 {
		return errors.Errorf(
			"missing required configuration values at configuration paths: %v",
			notSet,
		)
	}

	err = v.Unmarshal(
		config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				medley.TextUnmarshalerHook,
				mapstructure.StringToSliceHookFunc(","),
			),
		),
	)
	if err != nil {
		return err
	}
	return config.Validate()
}

func LogConfig(ev *zerolog.Event, config configuration.Config, msg string) error {
	s := configuration.GetSensitiveRecursive(config)
	redactedPaths := []string{}
	for redacted := range s {
		redactedPaths = append(redactedPaths, redacted)
	}
	dc, err := configuration.ToDict(config, redactedPaths)
	if err != nil {
		return err
	}
	ev.Interface("config", dc).Msg(msg)
	return nil
}

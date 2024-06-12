package command

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
)

type (
	Option                                   func(*commandBuilderConfig)
	ConfigurableFunc[T configuration.Config] func(cfg T) error
)

// CommandBuilder is a factory for easily generating a
// CLI based on config structs implementing `medley.configuration.Config`.
type CommandBuilder[T configuration.Config] struct {
	cobraCommand  *cobra.Command
	config        T
	builderConfig *commandBuilderConfig
}

type commandBuilderConfig struct {
	dumpConfig     bool
	generateConfig bool
	name           string
	shortUsage     string
	longUsage      string
	filesystem     afero.Fs
	initDBCommand  bool
}

func newCommandBuilder(name string) *commandBuilderConfig {
	return &commandBuilderConfig{
		name:           name,
		generateConfig: false,
		dumpConfig:     false,
		shortUsage:     fmt.Sprintf("start the '%s'", name),
		longUsage:      "",
		filesystem:     afero.NewOsFs(),
	}
}

func NewConfigForFunc[T configuration.Config](fn ConfigurableFunc[T]) T {
	typ := reflect.TypeOf(fn).In(0).Elem()
	nw, ok := reflect.New(typ).Interface().(T)
	if !ok {
		panic("type error during instantiation of new config")
	}
	return nw
}

// Build builds the full cobra.Command executing the passed in
// "main" function.
// The configuration parsing is inferred by main's only argument, which
// has to be a configuration struct that complies with the configuration.Config
// interface.
func Build[T configuration.Config](
	main ConfigurableFunc[T],
	options ...Option,
) *CommandBuilder[T] {
	cfg := NewConfigForFunc(main)
	cfg.Init()
	builder := newCommandBuilder(strings.ToLower(cfg.Name()))
	for _, opt := range options {
		opt(builder)
	}

	cb := &CommandBuilder[T]{
		cobraCommand:  &cobra.Command{},
		builderConfig: builder,
	}

	cb.cobraCommand = &cobra.Command{
		Use:   builder.name,
		Short: builder.shortUsage,
		Long:  builder.longUsage,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := NewConfigForFunc(main)
			cfg.Init()
			v := viper.GetViper()
			v.SetFs(builder.filesystem)
			err := ParseCLI(v, cmd, cfg)
			if err != nil {
				return errors.Wrap(err, "unable to parse configuration")
			}
			log.Debug().
				Interface("config", cfg).
				Msg("got config")
			return main(cfg)
		},
	}
	cb.cobraCommand.PersistentFlags().String("config", "", "config file")
	cb.cobraCommand.MarkPersistentFlagFilename("config")

	if builder.generateConfig {
		genConfigCmd := &cobra.Command{
			Use:   "generate-config",
			Short: fmt.Sprintf("Generate a '%s' configuration file", builder.name),
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg := NewConfigForFunc(main)
				cfg.Init()
				err := configuration.SetExampleValuesRecursive(cfg)
				if err != nil {
					return err
				}
				outPath, err := cmd.Flags().GetString("output")
				if err != nil {
					return err
				}
				return WriteConfig(builder.filesystem, cfg, outPath)
			},
		}
		genConfigCmd.PersistentFlags().String("output", "", "output file")
		genConfigCmd.MarkPersistentFlagRequired("output")
		cb.cobraCommand.AddCommand(genConfigCmd)
	}
	if builder.dumpConfig {
		dumpConfigCmd := &cobra.Command{
			Use:   "dump-config",
			Short: fmt.Sprintf("Dump a '%s' configuration file, based on given config and env vars", builder.name),
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg := NewConfigForFunc(main)
				cfg.Init()
				v := viper.GetViper()
				v.SetFs(builder.filesystem)
				err := ParseCLI(v, cmd, cfg)
				if err != nil {
					return errors.Wrap(err, "unable to parse configuration")
				}
				outPath, err := cmd.Flags().GetString("output")
				if err != nil {
					return err
				}
				log.Debug().
					Interface("config", cfg).
					Msg("dumping config")
				return WriteConfig(builder.filesystem, cfg, outPath)
			},
		}
		dumpConfigCmd.PersistentFlags().String("output", "", "output file")
		dumpConfigCmd.MarkPersistentFlagRequired("output")
		dumpConfigCmd.PersistentFlags().String("config", "", "config file")
		dumpConfigCmd.MarkPersistentFlagFilename("config")
		cb.cobraCommand.AddCommand(dumpConfigCmd)
	}
	return cb
}

type CobraRunE func(cmd *cobra.Command, args []string) error

// AddFunctionSubcommand attaches an additional subcommand to the command initially built by the
// Build method. The command executes the given function and takes the given arguments.
func (cb *CommandBuilder[T]) AddFunctionSubcommand(
	fnc ConfigurableFunc[T],
	use, short string,
	args cobra.PositionalArgs,
) {
	cb.cobraCommand.AddCommand(&cobra.Command{
		Use:   use,
		Short: short,
		Args:  args,
		RunE:  cb.WrapFuncParseConfig(fnc),
	})
}

func (cb *CommandBuilder[T]) WrapFuncParseConfig(fnc ConfigurableFunc[T]) CobraRunE {
	return func(cmd *cobra.Command, args []string) error {
		cfg := NewConfigForFunc(fnc)
		cfg.Init()
		v := viper.GetViper()
		v.SetFs(cb.builderConfig.filesystem)
		err := ParseCLI(v, cmd, cfg)
		if err != nil {
			return errors.WithMessage(err, "Please check your configuration")
		}
		log.Debug().
			Interface("config", cfg).
			Msg("got config")
		return fnc(cfg)
	}
}

// AddInitDBCommand attaches an additional subcommand
// 'initdb' to the command initially built by the Build method.
// The initDB function argument is structured in the same way than the "main"
// function passed in to the Build method.
func (cb *CommandBuilder[T]) AddInitDBCommand(initDB ConfigurableFunc[T]) {
	cb.AddFunctionSubcommand(
		initDB,
		"initdb",
		fmt.Sprintf("Initialize the database of the '%s'", cb.builderConfig.name),
		cobra.NoArgs,
	)
}

func (cb *CommandBuilder[_]) Command() *cobra.Command {
	return cb.cobraCommand
}

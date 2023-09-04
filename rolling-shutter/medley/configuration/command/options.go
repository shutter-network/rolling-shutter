package command

import "github.com/spf13/afero"

// CommandName overwrites the CLI invokation name of the built command.
// If this option is not provided, the name returned by the Config struct's Name()
// method will be used.
func CommandName(name string) Option {
	return func(c *commandBuilderConfig) {
		c.name = name
	}
}

// Usage sets the short and long usage strings that are shown in the CLI help messages.
func Usage(short, long string) Option {
	return func(c *commandBuilderConfig) {
		c.shortUsage = short
		c.longUsage = long
	}
}

// WithGenerateConfigSubcommand attaches an additional subcommand
// 'generate-config' to the command.
// This allows to generate an example configuration file based on the
// example configuration parameters provided by the Config structs.
func WithGenerateConfigSubcommand() Option {
	return func(c *commandBuilderConfig) {
		c.generateConfig = true
	}
}

// WithDumpConfigSubcommand attaches an additional subcommand
// 'dump-config' to the command.
// This allows to parse the given configuration (file, env-var),
// and write out all accumulated values in a configuration file.
func WithDumpConfigSubcommand() Option {
	return func(c *commandBuilderConfig) {
		c.dumpConfig = true
	}
}

// WithFileSystem overwrites overwrite the `afero` Filesystem wrapper used
// for reading and writing configuration files.
// This is mainly helpful for tests, where an in-memory filesystem
// should be used.
// If this option is not given, the OS filesystem will be used.
func WithFileSystem(fs afero.Fs) Option {
	return func(c *commandBuilderConfig) {
		c.filesystem = fs
	}
}

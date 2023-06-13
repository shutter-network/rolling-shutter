package configuration

import (
	"io"

	"github.com/pelletier/go-toml/v2"
)

// Config provides an Interface for a general configuration structure that facilitates the generation of command line interfaces.
// It promotes code reuse for configuring commonly used subpackages such as p2p and the ethereum-node.
// The Config interface defines the initialization, naming, validation, and default/required values for the configuration structure.
// It also specifies how to generate example values and additional headers for config file generation.
// To define configuration parameters, the struct implementing the Config interface
// must define the parameter as a public field. The field's type must either implement the encodeable.TextEncodeable interface,
// be a primitive type or be a pointer to a struct implementing the Config interface itself for nested configurations.
// The parameter's user exposed name is derived from the field name and the path of nested configuration structs.
// The interface works in conjunction with a tree traversal algorithm that uses reflection on the config's fields and derives metadata
// of the configuration from the interface methods.
// Nested 'child' configurations can define their own metadata in a commonly used Confgig implementation,
// and are not required to be specified in parent configurations.
// However, it is possible to override child behavior from a parent config, such as when default values of nested configs
// should differ across different subcommands.
//
// It is important to perform a compile-time check to ensure that a config implements the Config interface. For example:
//
//	var (
//	    _ configuration.Config = &ShuttermintConfig{}
//	    _ configuration.Config = &Config{}
//	)
//
// This check is particularly crucial for nested configs because they are only distinguished from other types through
// reflection and whether they implement the Config interface.
// The config tree traversal algorithms, such as validation and configuration generation, will skip configuration
// subtrees if they do not implement the Config interface.
type Config interface {
	// Init must initialize all pointer field values implementing encodeable.TextEncodeable.
	// Additionally Init must initialize field values of nested configuration structs
	// implementation Config.
	// Init is NOT called recursively on all nested config values, and thus
	// must establish a recursively fully initialized struct.
	// Therefore nested config initialization has to be implemented in the
	// root Init method manually, or the nested config's Init method has to be called explicitly.
	Init()

	// Name must return the name of the configuration.
	// The return value is used for configuration parsing:
	// - it is used for CLI subcommand name, when the Config implementation is
	//   at the root (subcommand) level.
	// - it is used for default configuration file names
	// - it is used for legacy environment variable prefixes.
	Name() string

	// Validate may implement logic that may be used for validating a configuration after parsing
	// for additional specific constraints on the configuration values.
	// An invalid config must cause Validate to return an error.
	// Validate is NOT called recursively on all nested config values but only for the root config.
	// Therefore if values of nested configs should be validated, this has to be implemented in the
	// root Validate method manually, or the nested config's Validate method has to be called explicitly.
	// If no validation is required, the method must return nil.
	Validate() error

	// SetDefaultValues sets values on the config instances fields that should be used as default values.
	// Default values will then be used when the user does not provide a value for the field's associated
	// configuration parameter.
	// SetDefaultValues will be called recursively for all nested config fields and thus a parent config
	// struct does not have to set the nested default values in it's SetDefaultValues.
	// Nested default values CAN be overwritten from an ancestor config's SetDefaultValues, if for example
	// the commonly used default values are not suited for that subcommand.
	SetDefaultValues() error

	// SetExampleValues sets values on the config instances fields that should be used as example values for
	// a generated configuration file.
	// SetExampleValues will be called recursively for all nested config fields and thus a parent config
	// struct does not have to set the nested example values in it's SetExampleValues.
	// Nested example values CAN be overwritten from an ancestor config's SetExampleValues, if for example
	// the commonly used example values are not suited for that subcommand.
	SetExampleValues() error

	// TOMLWriteHeader provides the option to inject custom text to the generated example
	// configuration file.
	// TOMLWriteHeader will be called recursively for the root and all child config fields.
	// The tree of nested configs is traversed depth-first and ancester headers will be written
	// after their child headers.
	// The written output must comply to the TOML specification and should mainly
	// be used to add comment headers to the configuration file.
	TOMLWriteHeader(w io.Writer) (int, error)
}

// WriteTOML writes a toml configuration file with the given config.
func WriteTOML(w io.Writer, config Config) error {
	err := writeTOMLHeadersRecursive(config, w)
	if err != nil {
		return err
	}
	enc := toml.NewEncoder(w)
	err = enc.Encode(config)
	if err != nil {
		return err
	}
	return nil
}

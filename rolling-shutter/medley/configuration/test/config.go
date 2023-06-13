package test

import (
	"errors"
	"io"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/keys"
)

var (
	_ configuration.Config = &NestedConfig{}
	_ configuration.Config = &Config{}

	libp2pKey  string = "CAESQOwP/OT9r6+HApehrPIN5hA/zeKfw2HWucEuRm3mVXfAEUE5d/MkEYCyakQwIbirosHMoSC7MH3p+vic4yrLyVc="
	ecdsaKey   string = "8a1e84264bd35f11ec867ebc8328cf0b29e427cc1d38b63919339e291b6863ac"
	ed25519Key string = "500418849232a0c918b6333450b4afab775101eb1504fd94fe46e1e6fd53a2c2"
)

func NewConfig() *Config {
	c := &Config{}
	c.Init()
	return c
}

func (c *Config) Init() {
	c.NestedConfig = NewNestedConfig()
}

type Config struct {
	UInt64 uint64 `shconfig:",required"`

	Bool   bool
	String string

	NestedConfig *NestedConfig
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Name() string {
	return "test"
}

func (c *Config) SetDefaultValues() error {
	c.Bool = true
	c.String = "barfoo"
	// check that the parent can overwrite child default values
	c.NestedConfig.NestedDefaultString = "parentfoobar"
	return nil
}

func (c *Config) SetExampleValues() error {
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}
	c.UInt64 = 123
	return nil
}

func (c Config) TOMLWriteHeader(w io.Writer) (int, error) {
	return w.Write([]byte("# second header\n"))
}

func NewNestedConfig() *NestedConfig {
	c := &NestedConfig{}
	c.Init()
	return c
}

type NestedConfig struct {
	NestedDefaultInt     int
	NestedDefaultString  string
	NestedEd25519Public  *keys.Ed25519Public  `shconfig:",required"`
	NestedEd25519Private *keys.Ed25519Private `shconfig:",required"`

	NestedECDSAPublic  *keys.ECDSAPublic  `shconfig:",required"`
	NestedECDSAPrivate *keys.ECDSAPrivate `shconfig:",required"`

	NestedLibp2pPrivate *keys.Libp2pPrivate `shconfig:",required"`
	NestedLibp2pPublic  *keys.Libp2pPublic  `shconfig:",required"`
}

func (c *NestedConfig) Init() {
	c.NestedECDSAPrivate = &keys.ECDSAPrivate{}
	c.NestedECDSAPublic = &keys.ECDSAPublic{}

	c.NestedEd25519Private = &keys.Ed25519Private{}
	c.NestedEd25519Public = &keys.Ed25519Public{}

	c.NestedLibp2pPrivate = &keys.Libp2pPrivate{}
	c.NestedLibp2pPublic = &keys.Libp2pPublic{}
}

func (c *NestedConfig) Name() string {
	return "nestedtest"
}

func (c *NestedConfig) Validate() error {
	return nil
}

func (c *NestedConfig) SetDefaultValues() error {
	c.NestedDefaultString = "foobar"
	c.NestedDefaultInt = 42
	return nil
}

func (c *NestedConfig) SetExampleValues() error {
	var ok bool
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}

	c.NestedEd25519Private = &keys.Ed25519Private{}

	err = c.NestedEd25519Private.UnmarshalText([]byte(ed25519Key))
	if err != nil {
		return err
	}

	c.NestedEd25519Public, ok = c.NestedEd25519Private.Public().(*keys.Ed25519Public)
	if !ok {
		return errors.New("keys.Ed25519Private.Public() methodreturned wrong type")
	}

	c.NestedECDSAPrivate = &keys.ECDSAPrivate{}
	err = c.NestedECDSAPrivate.UnmarshalText([]byte(ecdsaKey))
	if err != nil {
		return err
	}

	c.NestedECDSAPublic, ok = c.NestedECDSAPrivate.Public().(*keys.ECDSAPublic)
	if !ok {
		return errors.New("keys.ECDSAPrivate.Public() method returned wrong type")
	}

	c.NestedLibp2pPrivate = &keys.Libp2pPrivate{}
	err = c.NestedLibp2pPrivate.UnmarshalText([]byte(libp2pKey))
	if err != nil {
		return err
	}

	c.NestedLibp2pPublic, ok = c.NestedLibp2pPrivate.Public().(*keys.Libp2pPublic)
	if !ok {
		return errors.New("keys.Libp2pPrivate.Public() method returned wrong type")
	}

	return nil
}

func (c NestedConfig) TOMLWriteHeader(w io.Writer) (int, error) {
	return w.Write([]byte("# first header\n"))
}

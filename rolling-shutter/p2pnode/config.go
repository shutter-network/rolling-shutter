package p2pnode

import (
	"io"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

var _ configuration.Config = &Config{}

func NewConfig() *Config {
	c := &Config{}
	c.Init()
	return c
}

type Config struct {
	ListenMessages bool `comment:"whether to register handlers on the messages and log them"`

	P2P *p2p.Config
}

func (c *Config) Init() {
	c.P2P = p2p.NewConfig()
}

func (c *Config) Name() string {
	return "p2pnode"
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) SetDefaultValues() error {
	c.ListenMessages = true
	return nil
}

func (c *Config) SetExampleValues() error {
	return c.SetDefaultValues()
}

func (c Config) TOMLWriteHeader(w io.Writer) (int, error) {
	return w.Write([]byte("# Peer role: bootstrap\n"))
}

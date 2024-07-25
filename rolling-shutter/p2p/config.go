package p2p

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/address"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/env"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/keys"
)

var (
	defaultListenAddrs []*address.P2PAddress
	_                  configuration.Config = &Config{}
)

func init() {
	cfg := &libp2p.Config{}
	err := libp2p.DefaultListenAddrs(cfg)
	if err != nil {
		return
	}
	defaultListenAddrs = []*address.P2PAddress{}
	for _, add := range cfg.ListenAddrs {
		defaultListenAddrs = append(defaultListenAddrs, &address.P2PAddress{Multiaddr: add})
	}
}

func NewConfig() *Config {
	c := &Config{}
	c.Init()
	return c
}

func (c *Config) Init() {
	c.P2PKey = &keys.Libp2pPrivate{}
}

type Config struct {
	P2PKey                   *keys.Libp2pPrivate `shconfig:",required"`
	ListenAddresses          []*address.P2PAddress
	AdvertiseAddresses       []*address.P2PAddress `comment:"Optional, addresses to be advertised to other peers instead of auto-detected ones."`
	CustomBootstrapAddresses []*address.P2PAddress `comment:"Overwrite p2p boostrap nodes"`
	Environment              env.Environment
	DiscoveryNamespace       string `shconfig:",required" comment:"Must be unique for each instance id."`
}

func (c *Config) Name() string {
	return "p2p"
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) SetDefaultValues() error {
	c.ListenAddresses = defaultListenAddrs
	c.Environment = env.EnvironmentProduction
	return nil
}

func (c *Config) SetExampleValues() error {
	// use the default ones for that environment when empty
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}
	c.CustomBootstrapAddresses = []*address.P2PAddress{
		address.MustP2PAddress(
			"/ip4/127.0.0.1/tcp/2001/p2p/QmdfBeR6odD1pRKendUjWejhMd9wybivDq5RjixhRhiERg",
		),
		address.MustP2PAddress(
			"/ip4/127.0.0.1/tcp/2002/p2p/QmV9YbMDLDi736vTzy97jn54p43o74fLxc5DnLUrcmK6WP",
		),
	}
	c.Environment = env.EnvironmentProduction
	c.DiscoveryNamespace = "shutter-42"

	p2pkey, err := keys.GenerateLibp2pPrivate(rand.Reader)
	if err != nil {
		return err
	}
	c.P2PKey = p2pkey
	return nil
}

func (c Config) TOMLWriteHeader(w io.Writer) (int, error) {
	id, err := c.P2PKey.PeerID()
	if err != nil {
		return 0, err
	}
	return fmt.Fprintf(w, "# Peer identity: /p2p/%s\n", id)
}

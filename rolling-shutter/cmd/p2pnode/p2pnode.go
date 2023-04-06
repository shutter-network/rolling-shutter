package p2pnode

import (
	"bytes"
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pnode"
)

var (
	cfgFile            string
	outputFile         string
	defaultListenAddrs []multiaddr.Multiaddr
)

func init() {
	cfg := &libp2p.Config{}
	err := libp2p.DefaultListenAddrs(cfg)
	if err != nil {
		return
	}
	defaultListenAddrs = cfg.ListenAddrs
}

func generateConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-config",
		Short: "Generate a p2pnode configuration file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateConfig()
		},
	}
	cmd.PersistentFlags().StringVar(&outputFile, "output", "", "output file")
	cmd.MarkPersistentFlagRequired("output")
	return cmd
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "p2pnode",
		Short: "Run a p2p node",
		Long:  `currently for testing only `,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := readP2PNodeConfig()
			if err != nil {
				return errors.WithMessage(err, "Please check your configuration")
			}
			log.Debug().Str("env-in-cfg", config.Environment.String()).Msg("got deserialized config")
			return main(config)
		},
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	cmd.AddCommand(generateConfigCmd())
	return cmd
}

func readP2PNodeConfig() (p2pnode.Config, error) {
	defaultListenAddress, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/2000")
	viper.SetDefault("ListenAddress", defaultListenAddress)
	viper.SetDefault("CustomBootstrapAddresses", []peer.AddrInfo{})

	defer func() {
		if viper.ConfigFileUsed() != "" {
			log.Info().Str("config", viper.ConfigFileUsed()).Msg("read config")
		}
	}()
	var err error
	config := p2pnode.Config{}

	viper.AddConfigPath("$HOME/.config/shutter")
	viper.SetConfigName("p2p")
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

	return config, err
}

func main(config p2pnode.Config) error {
	log.Info().
		Str("version", shversion.Version()).
		Msg("starting p2pnode")
	return service.RunWithSighandler(context.Background(), p2pnode.New(config))
}

func exampleConfig() (*p2pnode.Config, error) {
	cfg := &p2pnode.Config{
		ListenAddresses: defaultListenAddrs,
		// use the default ones for that environment when empty
		CustomBootstrapAddresses: []peer.AddrInfo{
			p2p.MustAddrInfo("/ip4/127.0.0.1/tcp/2001/p2p/QmdfBeR6odD1pRKendUjWejhMd9wybivDq5RjixhRhiERg"),
			p2p.MustAddrInfo("/ip4/127.0.0.1/tcp/2002/p2p/QmV9YbMDLDi736vTzy97jn54p43o74fLxc5DnLUrcmK6WP"),
		},
		Environment: p2p.Staging,
	}
	err := cfg.GenerateNewKeys()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func generateConfig() error {
	cfg, err := exampleConfig()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err = cfg.WriteTOML(buf); err != nil {
		return errors.Wrap(err, "failed to write p2pnode config file")
	}
	return medley.SecureSpit(outputFile, buf.Bytes())
}

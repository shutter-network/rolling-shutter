package mocknode

import (
	"bytes"
	"context"
	"crypto/rand"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/mocknode"
)

var (
	outputFile string
	cfgFile    string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mocknode",
		Short: "Run a mock node",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return main()
		},
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	cmd.AddCommand(generateConfigCmd())
	return cmd
}

func generateConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-config",
		Short: "Generate a mock node configuration file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateConfig()
		},
	}
	cmd.PersistentFlags().StringVar(&outputFile, "output", "", "output file")
	cmd.MarkPersistentFlagRequired("output")

	return cmd
}

func readConfig() (mocknode.Config, error) {
	viper.SetEnvPrefix("MOCKNODE")
	viper.BindEnv("ListenAddress")
	viper.BindEnv("PeerMultiaddrs")
	defaultListenAddress, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/2000")
	viper.SetDefault("ListenAddress", defaultListenAddress)
	viper.SetDefault("PeerMultiaddrs", make([]multiaddr.Multiaddr, 0))
	viper.SetDefault("Rate", 1.0)

	config := mocknode.Config{}

	viper.AddConfigPath("$HOME/.config/shutter")
	viper.SetConfigName("mocknode")
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

	err = viper.Unmarshal(
		&config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				medley.MultiaddrHook,
				medley.P2PKeyHook,
			),
		),
	)
	if err != nil {
		return config, err
	}

	return config, nil
}

var mockNodeTemplate = medley.MustBuildTemplate(
	"mocknode",
	`# Shutter mock node config for /p2p/{{ .P2PKey | P2PKeyPublic}}

# p2p configuration
ListenAddress   = "{{ .ListenAddress }}"
PeerMultiaddrs  = [{{ .PeerMultiaddrs | QuoteList}}]

# Secret Keys
P2PKey          = "{{ .P2PKey | P2PKey}}"

# Mock messages
InstanceID              = {{ .InstanceID }}
Rate                    = {{ .Rate }}
SendDecryptionTriggers  = {{ .SendDecryptionTriggers }}
SendCipherBatches       = {{ .SendCipherBatches }}
SendDecryptionKeys      = {{ .SendDecryptionKeys }}
`)

func mustMultiaddr(s string) multiaddr.Multiaddr {
	a, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		panic(err)
	}
	return a
}

func generateConfig() error {
	p2pkey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return err
	}

	config := mocknode.Config{
		ListenAddress:  mustMultiaddr("/ip4/127.0.0.1/tcp/2000"),
		PeerMultiaddrs: nil,
		P2PKey:         p2pkey,

		InstanceID:             0,
		Rate:                   1.0,
		SendDecryptionTriggers: true,
		SendCipherBatches:      true,
		SendDecryptionKeys:     true,
	}
	buf := &bytes.Buffer{}
	err = mockNodeTemplate.Execute(buf, config)
	if err != nil {
		return err
	}
	return medley.SecureSpit(outputFile, buf.Bytes())
}

func main() error {
	ctx := context.Background()

	config, err := readConfig()
	if err != nil {
		return err
	}

	mockNode := mocknode.MockNode{
		Config: config,
	}
	return mockNode.Run(ctx)
}

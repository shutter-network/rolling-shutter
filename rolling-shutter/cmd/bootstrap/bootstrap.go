package bootstrap

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/rpc/client/http"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/fx"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

type Config struct {
	ShuttermintURL    string   `mapstructure:"shuttermint-url"`
	EthereumURL       string   `mapstructure:"ethereum-url"`
	DeploymentDir     string   `mapstructure:"deployment-dir"`
	KeyperConfigIndex int      `mapstructure:"index"`
	SigningKey        string   `mapstructure:"signing-key"`
	Keypers           []string `mapstructure:"ethereum-url"`
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap Shuttermint by submitting the initial batch config",
		Long: `This command sends a batch config to the Shuttermint chain in a message signed
with the given private key. This will instruct a newly created chain to update
its validator set according to the keyper set defined in the batch config. The
private key must correspond to the initial validator address as defined in the
chain's genesis config.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &Config{}
			if err := viper.Unmarshal(config); err != nil {
				return err
			}
			return bootstrap(config)
		},
	}

	cmd.PersistentFlags().StringP(
		"ethereum-url",
		"",
		"http://localhost:8545",
		"Ethereum URL",
	)

	cmd.PersistentFlags().StringP(
		"deployment-dir",
		"",
		"./deployments/localhost",
		"Deployment directory",
	)

	cmd.PersistentFlags().StringP(
		"shuttermint-url",
		"s",
		"http://localhost:26657",
		"Shuttermint RPC URL",
	)
	cmd.PersistentFlags().IntP(
		"index",
		"i",
		1,
		"keyper config index to bootstrap with (use latest if negative)",
	)

	cmd.PersistentFlags().StringP(
		"signing-key",
		"k",
		"",
		"private key of the keyper to send the message with",
	)

	return cmd
}

func bootstrap(config *Config) error {
	ctx := context.Background()
	ethereumClient, err := ethclient.DialContext(ctx, config.EthereumURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(ethereumClient, config.DeploymentDir)
	if err != nil {
		return err
	}

	shmcl, err := http.New(config.ShuttermintURL, "/websocket")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Shuttermint node")
	}

	signingKey, err := crypto.HexToECDSA(config.SigningKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse signing key")
	}

	keyperConfigIndex := uint64(config.KeyperConfigIndex)
	cfg, err := contracts.KeypersConfigsList.KeypersConfigs(nil, big.NewInt(int64(keyperConfigIndex)))
	if err != nil {
		return err
	}
	addr, err := contracts.KeypersConfigsList.AddrsSeq(nil)
	if err != nil {
		return err
	}
	seq, err := contract.NewAddrsSeq(addr, ethereumClient)
	if err != nil {
		return err
	}
	keypers, err := seq.GetAddrs(nil, cfg.SetIndex)
	if err != nil {
		return err
	}

	log.Info().Interface("config", cfg).Interface("keypers", keypers).
		Msg("using configuration")

	threshold := cfg.Threshold

	ms := fx.NewRPCMessageSender(shmcl, signingKey)
	activationBlockNumber := cfg.ActivationBlockNumber
	batchConfigMsg := shmsg.NewBatchConfig(
		activationBlockNumber,
		keypers,
		threshold,
		keyperConfigIndex,
	)

	err = ms.SendMessage(ctx, batchConfigMsg)
	if err != nil {
		return errors.Errorf("Failed to send batch config message: %v", err)
	}

	blockSeenMsg := shmsg.NewBlockSeen(activationBlockNumber)
	err = ms.SendMessage(ctx, blockSeenMsg)
	if err != nil {
		return errors.Errorf("Failed to send start message: %v", err)
	}

	log.Info().
		Uint64("keyper-config-index", keyperConfigIndex).
		Uint64("activation-block-number", activationBlockNumber).
		Uint64("threshold", threshold).
		Int("num-keypers", len(keypers)).
		Msg("submitted bootstrapping transaction")
	return nil
}

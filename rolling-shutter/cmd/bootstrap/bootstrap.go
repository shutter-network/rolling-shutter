package bootstrap

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/rpc/client/http"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/fx"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

var bootstrapFlags struct {
	ShuttermintURL    string
	EthereumURL       string
	DeploymentDir     string
	KeyperConfigIndex int
	SigningKey        string
	Keypers           []string
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
			return bootstrap()
		},
	}

	cmd.PersistentFlags().StringVarP(
		&bootstrapFlags.EthereumURL,
		"ethereum-url",
		"",
		"http://localhost:8545",
		"Ethereum URL",
	)

	cmd.PersistentFlags().StringVarP(
		&bootstrapFlags.DeploymentDir,
		"deployment-dir",
		"",
		"./deployments/localhost",
		"Deployment directory",
	)

	cmd.PersistentFlags().StringVarP(
		&bootstrapFlags.ShuttermintURL,
		"shuttermint-url",
		"s",
		"http://localhost:26657",
		"Shuttermint RPC URL",
	)
	cmd.PersistentFlags().IntVarP(
		&bootstrapFlags.KeyperConfigIndex,
		"index",
		"i",
		1,
		"keyper config index to bootstrap with (use latest if negative)",
	)

	cmd.PersistentFlags().StringVarP(
		&bootstrapFlags.SigningKey,
		"signing-key",
		"k",
		"",
		"private key of the keyper to send the message with",
	)
	cmd.MarkPersistentFlagRequired("signing-key")

	return cmd
}

func bootstrap() error {
	ctx := context.Background()
	ethereumClient, err := ethclient.DialContext(ctx, bootstrapFlags.EthereumURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(ethereumClient, bootstrapFlags.DeploymentDir)
	if err != nil {
		return err
	}

	shmcl, err := http.New(bootstrapFlags.ShuttermintURL)
	if err != nil {
		log.Fatalf("Error connecting to Shuttermint node: %v", err)
	}

	signingKey, err := crypto.HexToECDSA(bootstrapFlags.SigningKey)
	if err != nil {
		log.Fatalf("Invalid signing key: %v", err)
	}

	keyperConfigIndex := uint64(bootstrapFlags.KeyperConfigIndex)
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

	log.Printf("using config=%+v keypers=%+v", cfg, keypers)

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

	log.Println("Submitted bootstrapping transaction")
	log.Printf("Config index: %d", keyperConfigIndex)
	log.Printf("Activation block number: %d", activationBlockNumber)
	log.Printf("Threshold: %d", threshold)
	log.Printf("Num Keypers: %d", len(keypers))
	return nil
}

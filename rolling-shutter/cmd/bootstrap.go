package cmd

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/rpc/client/http"

	"github.com/shutter-network/shutter/shuttermint/keyper/fx"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

var bootstrapFlags struct {
	ShuttermintURL   string
	BatchConfigIndex int
	ContractsPath    string
	SigningKey       string
}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap Shuttermint by submitting the initial batch config",
	Long: `This command sends a batch config to the Shuttermint chain in a message signed
with the given private key. This will instruct a newly created chain to update
its validator set according to the keyper set defined in the batch config. The
private key must correspond to the initial validator address as defined in the
chain's genesis config.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		bootstrap()
	},
}

func init() {
	bootstrapCmd.PersistentFlags().StringVarP(
		&bootstrapFlags.ShuttermintURL,
		"shuttermint-url",
		"s",
		"http://localhost:26657",
		"Shuttermint RPC URL",
	)
	bootstrapCmd.PersistentFlags().IntVarP(
		&bootstrapFlags.BatchConfigIndex,
		"index",
		"i",
		-1,
		"index of the batch config to bootstrap with (use latest if negative)",
	)

	bootstrapCmd.PersistentFlags().StringVarP(
		&bootstrapFlags.ContractsPath,
		"contracts",
		"c",
		"",
		"read config contract address from the given contracts.json file",
	)
	bootstrapCmd.MarkPersistentFlagRequired("contracts")

	bootstrapCmd.PersistentFlags().StringVarP(
		&bootstrapFlags.SigningKey,
		"signing-key",
		"k",
		"",
		"private key of the keyper to send the message with",
	)
	bootstrapCmd.MarkPersistentFlagRequired("signing-key")
}

func bootstrap() {
	shmcl, err := http.New(bootstrapFlags.ShuttermintURL, "/websocket")
	if err != nil {
		log.Fatalf("Error connecting to Shuttermint node: %v", err)
	}

	signingKey, err := crypto.HexToECDSA(bootstrapFlags.SigningKey)
	if err != nil {
		log.Fatalf("Invalid signing key: %v", err)
	}

	batchConfigIndex := uint64(bootstrapFlags.BatchConfigIndex)
	keypers := []common.Address{}
	threshold := uint64(2)

	ms := fx.NewRPCMessageSender(shmcl, signingKey)
	startBatchIndex := uint64(0)
	batchConfigMsg := shmsg.NewBatchConfig(
		startBatchIndex,
		keypers,
		threshold,
		batchConfigIndex,
		false,
		false,
	)

	err = ms.SendMessage(context.Background(), batchConfigMsg)
	if err != nil {
		log.Fatalf("Failed to send batch config message: %v", err)
	}

	batchConfigStartedMsg := shmsg.NewBatchConfigStarted(batchConfigIndex)
	err = ms.SendMessage(context.Background(), batchConfigStartedMsg)
	if err != nil {
		log.Fatalf("Failed to send start message: %v", err)
	}

	log.Println("Submitted bootstrapping transaction")
	log.Printf("Config index: %d", batchConfigIndex)
	log.Printf("StartBatchIndex: %d", startBatchIndex)
	log.Printf("Threshold: %d", threshold)
	log.Printf("Num Keypers: %d", len(keypers))
}

package bootstrap

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/rpc/client/http"

	"github.com/shutter-network/shutter/shuttermint/keyper/fx"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

var bootstrapFlags struct {
	ShuttermintURL   string
	BatchConfigIndex int
	SigningKey       string
	Keypers          []string
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
		&bootstrapFlags.ShuttermintURL,
		"shuttermint-url",
		"s",
		"http://localhost:26657",
		"Shuttermint RPC URL",
	)
	cmd.PersistentFlags().IntVarP(
		&bootstrapFlags.BatchConfigIndex,
		"index",
		"i",
		1,
		"index of the batch config to bootstrap with (use latest if negative)",
	)

	cmd.PersistentFlags().StringVarP(
		&bootstrapFlags.SigningKey,
		"signing-key",
		"k",
		"",
		"private key of the keyper to send the message with",
	)
	cmd.MarkPersistentFlagRequired("signing-key")

	cmd.PersistentFlags().StringSliceVarP(
		&bootstrapFlags.Keypers,
		"keyper",
		"K",
		nil,
		"keyper address")
	cmd.MarkPersistentFlagRequired("keyper")
	return cmd
}

func twothirds(numKeypers int) int {
	return (2*numKeypers + 2) / 3
}

func bootstrap() error {
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
	for _, a := range bootstrapFlags.Keypers {
		if !common.IsHexAddress(a) {
			return errors.Errorf("--keyper argument '%s' is not an address", a)
		}
		keypers = append(keypers, common.HexToAddress(a))
	}

	threshold := uint64(twothirds(len(keypers)))

	ms := fx.NewRPCMessageSender(shmcl, signingKey)
	activationBlockNumber := uint64(0)
	batchConfigMsg := shmsg.NewBatchConfig(
		activationBlockNumber,
		keypers,
		threshold,
		batchConfigIndex,
		false,
		false,
	)

	err = ms.SendMessage(context.Background(), batchConfigMsg)
	if err != nil {
		return errors.Errorf("Failed to send batch config message: %v", err)
	}

	batchConfigStartedMsg := shmsg.NewBatchConfigStarted(batchConfigIndex)
	err = ms.SendMessage(context.Background(), batchConfigStartedMsg)
	if err != nil {
		return errors.Errorf("Failed to send start message: %v", err)
	}

	log.Println("Submitted bootstrapping transaction")
	log.Printf("Config index: %d", batchConfigIndex)
	log.Printf("Activation block number: %d", activationBlockNumber)
	log.Printf("Threshold: %d", threshold)
	log.Printf("Num Keypers: %d", len(keypers))
	return nil
}

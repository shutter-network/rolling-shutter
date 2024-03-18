package bootstrap

import (
	"context"
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/rpc/client/http"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/fx"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func BootstrapValidators(config *Config) error {
	file, err := os.ReadFile(config.KeyperSetFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read keyper-set file")
	}
	ks := &event.KeyperSet{}

	err = json.Unmarshal(file, ks)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read parse keyper-set")
	}

	ctx := context.Background()

	shmcl, err := http.New(config.ShuttermintURL, "/websocket")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Shuttermint node")
	}

	ms := fx.NewRPCMessageSender(shmcl, config.SigningKey.Key)
	batchConfigMsg := shmsg.NewBatchConfig(
		ks.ActivationBlock,
		ks.Members,
		ks.Threshold,
		ks.Index,
	)

	err = ms.SendMessage(ctx, batchConfigMsg)
	if err != nil {
		return errors.Errorf("Failed to send batch config message: %v", err)
	}

	blockSeenMsg := shmsg.NewBlockSeen(ks.ActivationBlock)
	err = ms.SendMessage(ctx, blockSeenMsg)
	if err != nil {
		return errors.Errorf("Failed to send start message: %v", err)
	}

	return nil
}

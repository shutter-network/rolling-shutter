package bootstrap

import (
	"context"
	"encoding/json"
	"os"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
)

func GetKeyperSet(ctx context.Context, config *Config) error {
	sl2, err := sync.NewShutterL2Client(
		ctx,
		sync.WithClientURL(config.JSONRPCURL),
	)
	if err != nil {
		return err
	}
	keyperSet, err := sl2.GetKeyperSetForBlock(ctx, number.LatestBlock)
	if err != nil {
		return err
	}
	file, _ := json.MarshalIndent(keyperSet, "", " ")
	return os.WriteFile(config.KeyperSetFilePath, file, 0o644)
}

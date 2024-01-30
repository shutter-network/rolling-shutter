package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
)

func GetKeyperSet(ctx context.Context, config *Config) error {
	sl2, err := chainsync.NewClient(
		ctx,
		chainsync.WithClientURL(config.JSONRPCURL),
	)
	if err != nil {
		return err
	}
	var keyperSet *event.KeyperSet
	switch {
	case config.ByIndex == nil && config.ByActivationBlockNumber == nil:
		keyperSet, err = sl2.GetKeyperSetForBlock(ctx, number.LatestBlock)
	case config.ByIndex == nil && config.ByActivationBlockNumber != nil:
		keyperSet, err = sl2.GetKeyperSetForBlock(ctx, config.ByActivationBlockNumber)
	case config.ByIndex != nil && config.ByActivationBlockNumber == nil:
		keyperSet, err = sl2.GetKeyperSetByIndex(ctx, *config.ByIndex)
	case config.ByIndex != nil && config.ByActivationBlockNumber != nil:
		return errors.New("can only retrieve keyper set by either index or activation-blocknumber")
	default:
		return nil
	}
	if err != nil {
		return err
	}
	file, _ := json.MarshalIndent(keyperSet, "", " ")
	return os.WriteFile(config.KeyperSetFilePath, file, 0o644)
}

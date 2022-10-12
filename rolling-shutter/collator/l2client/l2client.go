// Package l2client provides some convenience functions to interact with the sequencer.
package l2client

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
)

// GetBatchIndex retrieves the current batch index from the sequencer.
func GetBatchIndex(ctx context.Context, l2Client *rpc.Client) (uint64, error) {
	var blockNumber uint64

	f := func(ctx context.Context) (*string, error) {
		var result string
		log.Debug().Msg("polling batch-index from sequencer")
		err := l2Client.CallContext(ctx, &result, "shutter_batchIndex")
		if err != nil {
			return nil, err
		}
		return &result, nil
	}

	result, err := retry.FunctionCall(ctx, f)
	if err != nil {
		return blockNumber, errors.Wrapf(err, "can't retrieve batch-index from sequencer")
	}

	blockNumber, err = hexutil.DecodeUint64(*result)
	if err != nil {
		return blockNumber, errors.Wrap(err, "can't decode batch-index")
	}
	return blockNumber, nil
}

// SendTransaction sends a transaction to the sequencer. It uses the raw rpc.Client instead of the
// usual ethclient.Client wrapper because we want to use the modified txtypes marshaling here
// instead of the one from the go-ethereum repository.
func SendTransaction(ctx context.Context, client *rpc.Client, tx *txtypes.Transaction) error {
	data, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	f := func(ctx context.Context) (string, error) {
		var result string
		//
		err := client.CallContext(ctx, &result, "shutter_submitBatch", hexutil.Encode(data))
		if err != nil {
			return result, err
		}
		return result, nil
	}
	_, err = retry.FunctionCall(ctx, f)
	if err != nil {
		return errors.Wrap(err, "can't send transaction")
	}
	return err
}

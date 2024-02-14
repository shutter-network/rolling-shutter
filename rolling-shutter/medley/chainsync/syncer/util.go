package syncer

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
)

var (
	errNilCallOpts        = errors.New("nil call-opts")
	errNilOptsBlockNumber = errors.New("opts block-number is nil, but 'latest' not allowed")
	errLatestBlock        = errors.New("'nil' latest block")
)

type ManualFilterHandler interface {
	QueryAndHandle(ctx context.Context, block uint64) error
}

func logToCallOpts(ctx context.Context, log *types.Log) *bind.CallOpts {
	block := new(big.Int)
	block.SetUint64(log.BlockNumber)
	return &bind.CallOpts{
		BlockNumber: block,
		Context:     ctx,
	}
}

func guardCallOpts(opts *bind.CallOpts, allowLatest bool) error {
	if opts == nil {
		return errNilCallOpts
	}
	if !allowLatest {
		n := number.BigToBlockNumber(opts.BlockNumber)
		if n.IsLatest() {
			return errLatestBlock
		}
	}
	return nil
}

func fixCallOpts(ctx context.Context, c client.Client, opts *bind.CallOpts) (*bind.CallOpts, *uint64, error) {
	err := guardCallOpts(opts, false)
	if err == nil {
		return opts, nil, nil
	}
	// query the current latest block and fix it
	latest, queryErr := c.BlockNumber(ctx)
	if queryErr != nil {
		return nil, nil, errors.Wrap(err, "query latest block-number")
	}
	blockNumber := number.NewBlockNumber(&latest)
	if errors.Is(err, errNilCallOpts) {
		opts = &bind.CallOpts{
			Context:     ctx,
			BlockNumber: blockNumber.Int,
		}
		return opts, &latest, nil
	}
	if errors.Is(err, errLatestBlock) {
		opts.BlockNumber = blockNumber.Int
		return opts, &latest, nil
	}
	return nil, nil, err
}

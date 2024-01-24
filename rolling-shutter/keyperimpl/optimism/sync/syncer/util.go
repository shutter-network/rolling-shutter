package syncer

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
)

var (
	errNilCallOpts = errors.New("nil call-opts")
	errLatestBlock = errors.New("'nil' latest block")
)

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
		n := number.BlockNumber{Int: opts.BlockNumber}
		if n.IsLatest() {
			return errLatestBlock
		}
	}
	return nil
}

func fixCallOpts(ctx context.Context, c client.Client, opts *bind.CallOpts) (*bind.CallOpts, error) {
	err := guardCallOpts(opts, false)
	if err != nil {
		return opts, nil
	}
	// query the current latest block and fix it
	latest, err := c.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	blockNumber := number.NewBlockNumber(&latest)
	if errors.Is(err, errNilCallOpts) {
		opts = &bind.CallOpts{
			Context:     ctx,
			BlockNumber: blockNumber.Int,
		}
		return opts, nil
	}
	if errors.Is(err, errLatestBlock) {
		opts.BlockNumber = blockNumber.Int
		return opts, nil
	}
	return nil, err
}

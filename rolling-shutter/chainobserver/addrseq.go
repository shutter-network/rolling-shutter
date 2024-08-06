package chainobserver

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
)

func RetryGetAddrs(ctx context.Context, addrsSeq *contract.AddrsSeq, n uint64) ([]common.Address, error) {
	callOpts := &bind.CallOpts{
		Pending: false,
		// We call for the current height instead of the height at which the event was emitted,
		// because the sets cannot change retroactively and we won't need an archive node.
		BlockNumber: nil,
		Context:     ctx,
	}
	addrs, err := retry.FunctionCall(ctx, func(_ context.Context) ([]common.Address, error) {
		return addrsSeq.GetAddrs(callOpts, n)
	})
	if err != nil {
		return []common.Address{}, errors.Wrapf(err, "failed to query address set from contract")
	}
	return addrs, nil
}

package syncer

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethevent "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/contracts/v2/bindings/dkgcontract"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

// dkgEventWatcher is the subset of *dkgcontract.Dkgcontract that DKGContractSyncer
// uses: the five bulletin-board event subscriptions. Pulling it out as an
// interface lets tests substitute a fake backend that emits canned events
// without touching a simulated chain.
type dkgEventWatcher interface {
	WatchDealingSubmitted(opts *bind.WatchOpts, sink chan<- *dkgcontract.DkgcontractDealingSubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (gethevent.Subscription, error)
	WatchAccusationSubmitted(opts *bind.WatchOpts, sink chan<- *dkgcontract.DkgcontractAccusationSubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (gethevent.Subscription, error)
	WatchApologySubmitted(opts *bind.WatchOpts, sink chan<- *dkgcontract.DkgcontractApologySubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (gethevent.Subscription, error)
	WatchSuccessVoteSubmitted(opts *bind.WatchOpts, sink chan<- *dkgcontract.DkgcontractSuccessVoteSubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (gethevent.Subscription, error)
	WatchDKGSucceeded(opts *bind.WatchOpts, sink chan<- *dkgcontract.DkgcontractDKGSucceeded, keyperSetIndex []uint64, retryCounter []uint64) (gethevent.Subscription, error)
}

// DKGContractSyncer watches exactly one DKG contract for the five bulletin-board
// event types (DealingSubmitted, AccusationSubmitted, ApologySubmitted,
// SuccessVoteSubmitted, DKGSucceeded) and forwards them to the shared Handler.
// It is a leaf service.Service with no knowledge of the keyper set universe;
// it is constructed and started by DKGSyncer once per unique DKG contract
// address.
type DKGContractSyncer struct {
	Contract   *dkgcontract.Dkgcontract
	Addr       common.Address
	Log        log.Logger
	Handler    event.DKGEventHandler
	StartBlock *number.BlockNumber

	// backend overrides Contract for test injection. When nil, Start binds the
	// five Watch* subscriptions to the pre-bound Contract.
	backend dkgEventWatcher
}

func (s *DKGContractSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}

	backend := s.backend
	if backend == nil {
		backend = s.Contract
	}

	startBlock := *s.StartBlock.ToUInt64Ptr()
	watchOpts := &bind.WatchOpts{
		Start:   &startBlock,
		Context: ctx,
	}

	dealingCh := make(chan *dkgcontract.DkgcontractDealingSubmitted, channelSize)
	accusationCh := make(chan *dkgcontract.DkgcontractAccusationSubmitted, channelSize)
	apologyCh := make(chan *dkgcontract.DkgcontractApologySubmitted, channelSize)
	successVoteCh := make(chan *dkgcontract.DkgcontractSuccessVoteSubmitted, channelSize)
	successCh := make(chan *dkgcontract.DkgcontractDKGSucceeded, channelSize)

	dealingSub, err := backend.WatchDealingSubmitted(watchOpts, dealingCh, nil, nil, nil)
	if err != nil {
		return errors.Wrap(err, "watch DealingSubmitted")
	}
	accusationSub, err := backend.WatchAccusationSubmitted(watchOpts, accusationCh, nil, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		return errors.Wrap(err, "watch AccusationSubmitted")
	}
	apologySub, err := backend.WatchApologySubmitted(watchOpts, apologyCh, nil, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		return errors.Wrap(err, "watch ApologySubmitted")
	}
	successVoteSub, err := backend.WatchSuccessVoteSubmitted(watchOpts, successVoteCh, nil, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		apologySub.Unsubscribe()
		return errors.Wrap(err, "watch SuccessVoteSubmitted")
	}
	successSub, err := backend.WatchDKGSucceeded(watchOpts, successCh, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		apologySub.Unsubscribe()
		successVoteSub.Unsubscribe()
		return errors.Wrap(err, "watch DKGSucceeded")
	}

	runner.Go(func() error {
		err := s.watchContractEvents(
			ctx,
			dealingCh, accusationCh, apologyCh, successVoteCh, successCh,
			dealingSub.Err(), accusationSub.Err(), apologySub.Err(), successVoteSub.Err(), successSub.Err(),
		)
		if err != nil {
			s.Log.Error("error watching DKG events", "error", err.Error(), "dkg-contract", s.Addr.Hex())
		}
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		apologySub.Unsubscribe()
		successVoteSub.Unsubscribe()
		successSub.Unsubscribe()
		return err
	})
	return nil
}

// watchContractEvents is the per-contract subscription loop. It drains the five
// event channels, forwards events to the shared Handler via deliver(), and exits
// on context cancellation or subscription error.
func (s *DKGContractSyncer) watchContractEvents(
	ctx context.Context,
	dealingCh <-chan *dkgcontract.DkgcontractDealingSubmitted,
	accusationCh <-chan *dkgcontract.DkgcontractAccusationSubmitted,
	apologyCh <-chan *dkgcontract.DkgcontractApologySubmitted,
	successVoteCh <-chan *dkgcontract.DkgcontractSuccessVoteSubmitted,
	successCh <-chan *dkgcontract.DkgcontractDKGSucceeded,
	dealingErr, accusationErr, apologyErr, successVoteErr, successErr <-chan error,
) error {
	for {
		select {
		case ev, ok := <-dealingCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.DealingEvent{
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				Commitment:     ev.Commitment,
				PolyEvals:      ev.PolyEvals,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-accusationCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.AccusationEvent{
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				AccusedIndices: ev.AccusedIndices,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-apologyCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.ApologyEvent{
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				AccuserIndices: ev.AccuserIndices,
				PolyEvalData:   ev.PolyEvalData,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-successVoteCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.SuccessVoteEvent{
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				EonPublicKey:   ev.EonPublicKey,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-successCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.SuccessEvent{
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				EonPublicKey:   ev.EonPublicKey,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case err := <-dealingErr:
			if err != nil {
				return errors.Wrapf(err, "DealingSubmitted subscription (%s)", s.Addr.Hex())
			}
		case err := <-accusationErr:
			if err != nil {
				return errors.Wrapf(err, "AccusationSubmitted subscription (%s)", s.Addr.Hex())
			}
		case err := <-apologyErr:
			if err != nil {
				return errors.Wrapf(err, "ApologySubmitted subscription (%s)", s.Addr.Hex())
			}
		case err := <-successVoteErr:
			if err != nil {
				return errors.Wrapf(err, "SuccessVoteSubmitted subscription (%s)", s.Addr.Hex())
			}
		case err := <-successErr:
			if err != nil {
				return errors.Wrapf(err, "DKGSucceeded subscription (%s)", s.Addr.Hex())
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *DKGContractSyncer) deliver(ctx context.Context, ev event.DKGEvent) {
	if err := s.Handler(ctx, ev); err != nil {
		s.Log.Error(
			"handler for DKG event errored",
			"error",
			err.Error(),
		)
	}
}

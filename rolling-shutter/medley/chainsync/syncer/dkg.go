package syncer

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

// DKGEventSyncer mirrors the structure of KeyperSetSyncer: it performs an
// initial poll for already-succeeded DKG Instances and then subscribes to the
// five bulletin-board event types emitted by the DKG Contract.
type DKGEventSyncer struct {
	Client           client.Client
	Contract         *contract.DKGContract
	KeyperSetManager *bindings.KeyperSetManager
	Log              log.Logger
	StartBlock       *number.BlockNumber
	Handler          event.DKGEventHandler

	dealingCh     chan *contract.DKGContractDealingSubmitted
	accusationCh  chan *contract.DKGContractAccusationSubmitted
	apologyCh     chan *contract.DKGContractApologySubmitted
	successVoteCh chan *contract.DKGContractSuccessVoteSubmitted
	successCh     chan *contract.DKGContractDKGSucceeded
}

func (s *DKGEventSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}

	if s.StartBlock.IsLatest() {
		latest, err := s.Client.BlockNumber(ctx)
		if err != nil {
			return err
		}
		s.StartBlock.SetUint64(latest)
	}

	watchOpts := &bind.WatchOpts{
		Start:   s.StartBlock.ToUInt64Ptr(),
		Context: ctx,
	}

	initial, err := s.getInitialSuccesses(ctx)
	if err != nil {
		return err
	}
	for _, ev := range initial {
		if err := s.Handler(ctx, ev); err != nil {
			s.Log.Error(
				"handler for initial DKG success errored",
				"error",
				err.Error(),
			)
		}
	}

	s.dealingCh = make(chan *contract.DKGContractDealingSubmitted, channelSize)
	s.accusationCh = make(chan *contract.DKGContractAccusationSubmitted, channelSize)
	s.apologyCh = make(chan *contract.DKGContractApologySubmitted, channelSize)
	s.successVoteCh = make(chan *contract.DKGContractSuccessVoteSubmitted, channelSize)
	s.successCh = make(chan *contract.DKGContractDKGSucceeded, channelSize)

	dealingSub, err := s.Contract.WatchDealingSubmitted(watchOpts, s.dealingCh, nil, nil, nil)
	if err != nil {
		return errors.Wrap(err, "watch DealingSubmitted")
	}
	accusationSub, err := s.Contract.WatchAccusationSubmitted(watchOpts, s.accusationCh, nil, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		return errors.Wrap(err, "watch AccusationSubmitted")
	}
	apologySub, err := s.Contract.WatchApologySubmitted(watchOpts, s.apologyCh, nil, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		return errors.Wrap(err, "watch ApologySubmitted")
	}
	successVoteSub, err := s.Contract.WatchSuccessVoteSubmitted(watchOpts, s.successVoteCh, nil, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		apologySub.Unsubscribe()
		return errors.Wrap(err, "watch SuccessVoteSubmitted")
	}
	successSub, err := s.Contract.WatchDKGSucceeded(watchOpts, s.successCh, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		apologySub.Unsubscribe()
		successVoteSub.Unsubscribe()
		return errors.Wrap(err, "watch DKGSucceeded")
	}

	runner.Go(func() error {
		err := s.watchEvents(
			ctx,
			dealingSub.Err(),
			accusationSub.Err(),
			apologySub.Err(),
			successVoteSub.Err(),
			successSub.Err(),
		)
		if err != nil {
			s.Log.Error("error watching DKG events", err.Error())
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

// getInitialSuccesses queries DKGContract.succeeded(k) for each known keyper
// set index. Already-succeeded instances are delivered as synthetic events so
// the local cache (e.g. dkg_result) can be populated before any live events.
func (s *DKGEventSyncer) getInitialSuccesses(ctx context.Context) ([]*event.DKGEvent, error) {
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: s.StartBlock.Int,
	}
	if err := guardCallOpts(opts, false); err != nil {
		return nil, err
	}

	numKS, err := s.KeyperSetManager.GetNumKeyperSets(opts)
	if err != nil {
		return nil, errors.Wrap(err, "get num keyper sets")
	}

	events := []*event.DKGEvent{}
	for i := uint64(0); i < numKS; i++ {
		succeeded, err := s.Contract.Succeeded(opts, i)
		if err != nil {
			return nil, errors.Wrapf(err, "query succeeded for keyper set %d", i)
		}
		if !succeeded {
			continue
		}
		events = append(events, &event.DKGEvent{
			Kind:           event.DKGEventKindSuccess,
			KeyperSetIndex: i,
			AtBlockNumber:  number.BigToBlockNumber(opts.BlockNumber),
		})
	}
	return events, nil
}

func (s *DKGEventSyncer) watchEvents(
	ctx context.Context,
	dealingErr, accusationErr, apologyErr, successVoteErr, successErr <-chan error,
) error {
	for {
		select {
		case ev, ok := <-s.dealingCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.DKGEvent{
				Kind:           event.DKGEventKindDealing,
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				Commitment:     ev.Commitment,
				PolyEvals:      ev.PolyEvals,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-s.accusationCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.DKGEvent{
				Kind:           event.DKGEventKindAccusation,
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				AccusedIndices: ev.AccusedIndices,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-s.apologyCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.DKGEvent{
				Kind:           event.DKGEventKindApology,
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				AccuserIndices: ev.AccuserIndices,
				PolyEvalData:   ev.PolyEvalData,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-s.successVoteCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.DKGEvent{
				Kind:           event.DKGEventKindSuccessVote,
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				EonPublicKey:   ev.EonPublicKey,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-s.successCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.DKGEvent{
				Kind:           event.DKGEventKindSuccess,
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				EonPublicKey:   ev.EonPublicKey,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case err := <-dealingErr:
			if err != nil {
				return errors.Wrap(err, "DealingSubmitted subscription")
			}
		case err := <-accusationErr:
			if err != nil {
				return errors.Wrap(err, "AccusationSubmitted subscription")
			}
		case err := <-apologyErr:
			if err != nil {
				return errors.Wrap(err, "ApologySubmitted subscription")
			}
		case err := <-successVoteErr:
			if err != nil {
				return errors.Wrap(err, "SuccessVoteSubmitted subscription")
			}
		case err := <-successErr:
			if err != nil {
				return errors.Wrap(err, "DKGSucceeded subscription")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *DKGEventSyncer) deliver(ctx context.Context, ev *event.DKGEvent) {
	if err := s.Handler(ctx, ev); err != nil {
		s.Log.Error(
			"handler for DKG event errored",
			"error",
			err.Error(),
			"kind",
			ev.Kind,
		)
	}
}

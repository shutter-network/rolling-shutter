package syncer

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/tee"
)

type UnsafeHeadSyncer struct {
	//  Not used when eVerify is true
	Client  client.Client
	Log     log.Logger
	Handler event.BlockHandler

	// Not used when eVerify is true.
	newLatestHeadCh chan *types.Header
}

func (s *UnsafeHeadSyncer) Start(ctx context.Context, runner service.Runner) error {
	// Extended/External/Enclave verification
	// This can be true when not running in SGX, too.
	// We currently assume the runner (separate process that contains the enclave) is already running.
	// It might be possible to start it from within the enclave, but I am not sure about that.
	eVerify := true

	if s.Handler == nil {
		return errors.New("no handler registered")
	}

	// Difference between SubscribeNewHead and the ethereum-enclave:
	// - SubscribeNewHead returns all headers (no gaps, potentially with duplicates)
	// - On a reorg: SubscribeNeHead returns all headers (starting at the reorg)
	// - ethereum-enclave (finalized) returns all headers if AFTER the start blocknum (no gaps)
	// - ethereum-enclave (finalized) has gaps BEFORE the start blocknum
	// - ethereum-enclave (finalized) will panic if there is a reorg (major consensus failure and slashing)
	// - ethereum-enclave (optimistic) has gaps
	if eVerify {
		// We don't want to trust the geth client. This provides a channel containing blocks.
		// The tee.Config allows configuring which external verifiers should be considered valid.
		// It is only relevant when we're running in SGX, too. See comments.
		// conn.Attestation() contains additional configuration like the hash the etheruem-enclave started at (genesis).
		// It should be checked or forwarded in a remote attestation to make sure we're not following the wrong chain.
		conn, err := tee.DialVerifiedChainDataChannel("127.0.0.1:8001", tee.Config{
			// // Configuration options for SGX verification:
			// SameSigner: false,		// Require that the etheruem-enclave is signed with the same key as we are
			// SignerID:   []byte{},	// Alternative: Specify the SignerID
			// MrEnclave:  []byte{},	// Hash of the code the ethereum-enclave is running
			// ProductID:  new(uint16),	// Number by the signer to identify the ethereum-enclave to distinguish different products/binaries
			// MinISVSVN:  0,			// Minimum (security) version number specified when signing the ethereum enclave

			// Currently we do not use events verified by the ethereum-enclave, so we can just tell it to not extract them.
			EventExtractionStartBlocknum: ^uint64(0),
			Contracts:                    nil, // For now: We are not interested in events
		})
		if err != nil {
			return err
		}
		runner.Go(func() error {
			err := s.watchLatestUnsafeHeadEVerify(ctx, conn)
			if err != nil {
				s.Log.Error("error watching latest unsafe head with e-verify", err.Error())
			}
			conn.Close()
			return err
		})
	} else {
		s.newLatestHeadCh = make(chan *types.Header, 1)
		subs, err := s.Client.SubscribeNewHead(ctx, s.newLatestHeadCh)
		if err != nil {
			return err
		}
		runner.Go(func() error {
			err := s.watchLatestUnsafeHead(ctx, subs.Err())
			if err != nil {
				s.Log.Error("error watching latest unsafe head", err.Error())
			}
			subs.Unsubscribe()
			return err
		})
	}

	return nil
}

func (s *UnsafeHeadSyncer) watchLatestUnsafeHeadEVerify(ctx context.Context, conn *tee.Connection) error {
	for {
		select {
		case verified, ok := <-conn.Headers():
			// We have three options:
			// - Fully trust the go-ethereum node (problematic/insecure)
			// - Use the Finalized header (about 21 minutes delay but no reorgs unless there is a major chain failure)
			// - Use the optimistic header
			//
			// I do not know for sure which one is best suited:
			// - The existing code accounts for reorgs, so the optimistic header should be fine.
			// - But on the other hand it might result in leaking/exposing/outputting key material too early (which cannot be undone).
			header := verified.Optimistic

			if !ok {
				return nil
			}
			num := big.NewInt(0)
			num.SetUint64(header.Number)

			// The ethereum enclave (external verifier) primarily uses beacon-chain data. Thus it
			// has the execution layer header data in a different format and not exactly the same data.
			// However, shutter does not need the majority of that data.
			// The fields below can be extended by other data from ExecutionPayloadHeader if needed.
			// I could not find any place in this repo that uses the other fields.
			ev := &event.LatestBlock{
				Number:    number.BigToBlockNumber(num),
				BlockHash: header.BlockHash,
				Header: &types.Header{
					ParentHash: header.ParentHash,
					Number:     num,
					Time:       header.Timestamp,
				},
			}
			err := s.Handler(ctx, ev)
			if err != nil {
				s.Log.Error(
					"handler for `NewLatestBlock` errored",
					"error",
					err.Error(),
				)
			}
		case err := <-conn.Errors():
			if err != nil {
				s.Log.Error("subscription error for watchLatestUnsafeHead", err.Error())
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *UnsafeHeadSyncer) watchLatestUnsafeHead(ctx context.Context, subsErr <-chan error) error {
	for {
		select {
		case newHeader, ok := <-s.newLatestHeadCh:
			if !ok {
				return nil
			}
			ev := &event.LatestBlock{
				Number:    number.BigToBlockNumber(newHeader.Number),
				BlockHash: newHeader.Hash(),
				Header:    newHeader,
			}
			err := s.Handler(ctx, ev)
			if err != nil {
				s.Log.Error(
					"handler for `NewLatestBlock` errored",
					"error",
					err.Error(),
				)
			}
		case err := <-subsErr:
			if err != nil {
				s.Log.Error("subscription error for watchLatestUnsafeHead", err.Error())
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

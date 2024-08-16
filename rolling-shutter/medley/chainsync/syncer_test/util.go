package tester

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/shop-contracts/bindings"
)

func init() {
	var err error
	KeyBroadcastContractABI, err = bindings.KeyBroadcastContractMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

var KeyBroadcastContractAddress = common.BigToAddress(big.NewInt(42))
var KeyBroadcastContractABI *abi.ABI

func MakeChain(start int64, startParent common.Hash, numHeader uint, seed int64) []*types.Header {
	n := numHeader
	parent := startParent
	num := big.NewInt(start)
	h := []*types.Header{}

	// change the hashes for different seeds
	mixinh := common.BigToHash(big.NewInt(seed))
	for n > 0 {
		head := &types.Header{
			ParentHash: parent,
			Number:     num,
			MixDigest:  mixinh,
		}
		h = append(h, head)
		num = new(big.Int).Add(num, big.NewInt(1))
		parent = head.Hash()
		n--
	}
	return h
}

func NewTestKeyBroadcastHandler(log log.Logger) (*TestKeyBroadcastHandler, chan syncer.QueryContext, error) {
	querySyncChan := make(chan syncer.QueryContext)
	return &TestKeyBroadcastHandler{
		log:           log,
		evABI:         KeyBroadcastContractABI,
		address:       KeyBroadcastContractAddress,
		eons:          map[common.Hash]uint64{},
		querySyncChan: querySyncChan,
	}, querySyncChan, nil
}

type TestKeyBroadcastHandler struct {
	log           log.Logger
	evABI         *abi.ABI
	address       common.Address
	eons          map[common.Hash]uint64
	querySyncChan chan syncer.QueryContext
}

func (teh *TestKeyBroadcastHandler) Address() common.Address {
	return teh.address
}

func (teh *TestKeyBroadcastHandler) Log(msg string, ctx ...any) {
	teh.log.Info(msg, ctx)
}

func (teh *TestKeyBroadcastHandler) Event() string {
	return "EonKeyBroadcast"
}

func (teh *TestKeyBroadcastHandler) ABI() abi.ABI {
	return *teh.evABI
}
func (teh *TestKeyBroadcastHandler) Accept(
	ctx context.Context,
	header types.Header,
	ev bindings.KeyBroadcastContractEonKeyBroadcast,
) (bool, error) {
	return true, nil
}
func (teh *TestKeyBroadcastHandler) Handle(
	ctx context.Context,
	qCtx syncer.QueryContext,
	evs []bindings.KeyBroadcastContractEonKeyBroadcast,
) error {
	if qCtx.Remove != nil {
		for _, h := range qCtx.Remove.Get() {
			_, ok := teh.eons[h.Hash()]
			if ok {
				delete(teh.eons, h.Hash())
			}
		}
	}
	if qCtx.Update != nil {
		for _, ev := range evs {
			teh.eons[ev.Raw.BlockHash] = ev.Eon
		}
	}
	// teh.querySyncChan <- qCtx
	return nil
}

func (teh *TestKeyBroadcastHandler) GetEons() map[uint64]struct{} {
	m := map[uint64]struct{}{}
	for _, v := range teh.eons {
		m[v] = struct{}{}
	}
	return m
}
func (teh *TestKeyBroadcastHandler) GetBlockHashes() map[common.Hash]struct{} {
	m := map[common.Hash]struct{}{}
	for hsh := range teh.eons {
		m[hsh] = struct{}{}
	}
	return m
}

func NewTestChainUpdateHandler(log log.Logger) (*TestChainUpdateHandler, chan syncer.QueryContext, error) {
	querySyncChan := make(chan syncer.QueryContext)
	return &TestChainUpdateHandler{
		log:           log,
		querySyncChan: querySyncChan,
		chainCache:    syncer.NewMemoryChainCache(100, nil),
	}, querySyncChan, nil
}

type TestChainUpdateHandler struct {
	log           log.Logger
	querySyncChan chan syncer.QueryContext
	chainCache    syncer.ChainCache
}

func (teh *TestChainUpdateHandler) Handle(
	ctx context.Context,
	qCtx syncer.QueryContext,
) error {
	err := teh.chainCache.Update(ctx, qCtx)
	teh.querySyncChan <- qCtx
	return err
}

func (teh *TestChainUpdateHandler) GetBlockHashes(ctx context.Context) (map[common.Hash]struct{}, error) {
	m := map[common.Hash]struct{}{}
	chain, err := teh.chainCache.Get(ctx)
	if err != nil {
		return m, err
	}
	for _, h := range chain.Get() {
		m[h.Hash()] = struct{}{}
	}
	return m, nil
}

func MustPackKeyBroadcast(eon uint64, key []byte, header types.Header) *types.Log {
	l, err := PackKeyBroadcast(eon, key, header)
	if err != nil {
		panic("can't pack key broadcast event")
	}
	return l

}

// This roughly emulates what the EVM does
// and packs a EonKeyBroadcast log.
func PackKeyBroadcast(eon uint64, key []byte, header types.Header) (*types.Log, error) {
	event := "EonKeyBroadcast"
	address := KeyBroadcastContractAddress
	evABI := KeyBroadcastContractABI.Events[event]

	data, err := evABI.Inputs.Pack(eon, key)
	if err != nil {
		return nil, err
	}
	topics := []common.Hash{KeyBroadcastContractABI.Events[event].ID}
	return &types.Log{
		Address:     address,
		Data:        data,
		Topics:      topics,
		BlockNumber: header.Number.Uint64(),
		BlockHash:   header.Hash(),

		// NOTE: we don't set all the values here, make
		// sure no reader relies on them when writing test handler
		// (e.g. TxHash, TxIndex, ...)
	}, nil
}

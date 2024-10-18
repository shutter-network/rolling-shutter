package tester

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
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

func NewTestKeyBroadcastHandler(logger log.Logger) (*TestKeyBroadcastHandler, error) { //nolint: unparam
	return &TestKeyBroadcastHandler{
		log:     logger,
		evABI:   KeyBroadcastContractABI,
		address: KeyBroadcastContractAddress,
		eons:    map[common.Hash]uint64{},
	}, nil
}

type TestKeyBroadcastHandler struct {
	log     log.Logger
	evABI   *abi.ABI
	address common.Address
	eons    map[common.Hash]uint64
}

func (tkbh *TestKeyBroadcastHandler) Address() common.Address {
	return tkbh.address
}

func (tkbh *TestKeyBroadcastHandler) Log(msg string, ctx ...any) {
	tkbh.log.Info(msg, ctx)
}

func (tkbh *TestKeyBroadcastHandler) Event() string {
	return "EonKeyBroadcast" //nolint: goconst
}

func (tkbh *TestKeyBroadcastHandler) ABI() abi.ABI {
	return *tkbh.evABI
}

func (tkbh *TestKeyBroadcastHandler) Accept(
	_ context.Context,
	_ types.Header,
	_ bindings.KeyBroadcastContractEonKeyBroadcast,
) (bool, error) {
	return true, nil
}

func (tkbh *TestKeyBroadcastHandler) Handle(
	_ context.Context,
	update syncer.ChainUpdateContext,
	evs []bindings.KeyBroadcastContractEonKeyBroadcast,
) error {
	if update.Remove != nil {
		for _, h := range update.Remove.Get() {
			_, ok := tkbh.eons[h.Hash()]
			if ok {
				delete(tkbh.eons, h.Hash())
			}
		}
	}
	if update.Append != nil {
		for _, ev := range evs {
			tkbh.eons[ev.Raw.BlockHash] = ev.Eon
		}
	}
	return nil
}

func (tkbh *TestKeyBroadcastHandler) GetEons() map[uint64]struct{} {
	m := map[uint64]struct{}{}
	for _, v := range tkbh.eons {
		m[v] = struct{}{}
	}
	return m
}
func (tkbh *TestKeyBroadcastHandler) GetBlockHashes() map[common.Hash]struct{} {
	m := map[common.Hash]struct{}{}
	for hsh := range tkbh.eons {
		m[hsh] = struct{}{}
	}
	return m
}

func NewTestChainUpdateHandler(logger log.Logger) (*TestChainUpdateHandler, chan syncer.ChainUpdateContext, error) { //nolint: unparam
	querySyncChan := make(chan syncer.ChainUpdateContext)
	return &TestChainUpdateHandler{
		log:           logger,
		querySyncChan: querySyncChan,
		chainCache:    syncer.NewMemoryChainCache(100, nil),
	}, querySyncChan, nil
}

type TestChainUpdateHandler struct {
	log           log.Logger
	querySyncChan chan syncer.ChainUpdateContext
	chainCache    syncer.ChainCache
}

func (tkbh *TestChainUpdateHandler) Handle(
	ctx context.Context,
	update syncer.ChainUpdateContext,
) error {
	err := tkbh.chainCache.Update(ctx, update)
	tkbh.querySyncChan <- update
	return err
}

func (tkbh *TestChainUpdateHandler) GetBlockHashes(ctx context.Context) (map[common.Hash]struct{}, error) {
	m := map[common.Hash]struct{}{}
	chain, err := tkbh.chainCache.Get(ctx)
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

package syncer

import (
	"context"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethlog "github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/logger"
)

// recordingLogger captures every Warn call for assertion in tests. All other
// log levels are inherited from NoopLogger.
type recordingLogger struct {
	*logger.NoopLogger
	mu    sync.Mutex
	warns []logEntry
}

type logEntry struct {
	msg string
	ctx []interface{}
}

func newRecordingLogger() *recordingLogger {
	return &recordingLogger{NoopLogger: &logger.NoopLogger{}}
}

func (r *recordingLogger) Warn(msg string, ctx ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.warns = append(r.warns, logEntry{msg: msg, ctx: ctx})
}

func (r *recordingLogger) warnCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.warns)
}

// fakeKeyperSetIndexer implements keyperSetIndexer for unit tests.
type fakeKeyperSetIndexer struct {
	addresses []common.Address
}

func (f *fakeKeyperSetIndexer) GetNumKeyperSets(*bind.CallOpts) (uint64, error) {
	return uint64(len(f.addresses)), nil
}

func (f *fakeKeyperSetIndexer) GetKeyperSetAddress(_ *bind.CallOpts, index uint64) (common.Address, error) {
	if index >= uint64(len(f.addresses)) {
		return common.Address{}, errors.New("index out of range")
	}
	return f.addresses[index], nil
}

// stubResolver returns a resolver that maps each KeyperSet address to a
// preconfigured DKG contract address.
func stubResolver(mapping map[common.Address]common.Address) dkgContractResolver {
	return func(_ context.Context, _ *bind.CallOpts, ksAddr common.Address) (common.Address, error) {
		addr, ok := mapping[ksAddr]
		if !ok {
			return common.Address{}, errors.Errorf("unknown keyper set %s", ksAddr.Hex())
		}
		return addr, nil
	}
}

func newSyncer(t *testing.T, resolver dkgContractResolver) (*DKGEventSyncer, *recordingLogger) {
	t.Helper()
	rec := newRecordingLogger()
	s := &DKGEventSyncer{
		Log:                rec,
		resolveDKGContract: resolver,
	}
	return s, rec
}

func TestRecordDKGContractAddsNewAddress(t *testing.T) {
	s, rec := newSyncer(t, nil)
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")

	added := s.recordDKGContract(addr, 0, common.HexToAddress("0xaa"))
	assert.Equal(t, true, added, "first insertion should be reported as newly added")
	assert.Equal(t, 0, rec.warnCount(), "non-zero address must not produce a warning")
	assert.DeepEqual(t, []common.Address{addr}, s.trackedDKGContractList())
}

func TestRecordDKGContractDeduplicates(t *testing.T) {
	s, _ := newSyncer(t, nil)
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")

	addedFirst := s.recordDKGContract(addr, 0, common.HexToAddress("0xaa"))
	addedSecond := s.recordDKGContract(addr, 1, common.HexToAddress("0xbb"))

	assert.Equal(t, true, addedFirst)
	assert.Equal(t, false, addedSecond, "second insertion of the same address must be reported as not newly added")
	assert.Equal(t, 1, len(s.trackedDKGContractList()), "tracked set must contain a single entry after dedup")
}

func TestRecordDKGContractZeroAddressWarnsAndSkips(t *testing.T) {
	s, rec := newSyncer(t, nil)

	added := s.recordDKGContract(common.Address{}, 7, common.HexToAddress("0xaa"))

	assert.Equal(t, false, added, "zero address must not be added")
	assert.Equal(t, 0, len(s.trackedDKGContractList()))
	assert.Equal(t, 1, rec.warnCount(), "zero address must produce exactly one warning")
}

func TestScanInitialDKGContractsPopulatesSet(t *testing.T) {
	ks0 := common.HexToAddress("0xaa")
	ks1 := common.HexToAddress("0xbb")
	dkg0 := common.HexToAddress("0xcc")
	dkg1 := common.HexToAddress("0xdd")

	indexer := &fakeKeyperSetIndexer{addresses: []common.Address{ks0, ks1}}
	resolver := stubResolver(map[common.Address]common.Address{ks0: dkg0, ks1: dkg1})
	s, _ := newSyncer(t, resolver)

	opts := &bind.CallOpts{Context: context.Background(), BlockNumber: big.NewInt(1)}
	err := s.scanInitialDKGContracts(context.Background(), opts, indexer)
	assert.NilError(t, err)

	got := s.trackedDKGContractList()
	assert.Equal(t, 2, len(got))
	want := map[common.Address]struct{}{dkg0: {}, dkg1: {}}
	for _, addr := range got {
		_, ok := want[addr]
		assert.Equal(t, true, ok, "unexpected DKG contract %s in tracked set", addr.Hex())
	}
}

func TestScanInitialDKGContractsDeduplicatesSharedContract(t *testing.T) {
	ks0 := common.HexToAddress("0xaa")
	ks1 := common.HexToAddress("0xbb")
	ks2 := common.HexToAddress("0xcc")
	shared := common.HexToAddress("0xdd")

	indexer := &fakeKeyperSetIndexer{addresses: []common.Address{ks0, ks1, ks2}}
	resolver := stubResolver(map[common.Address]common.Address{ks0: shared, ks1: shared, ks2: shared})
	s, _ := newSyncer(t, resolver)

	opts := &bind.CallOpts{Context: context.Background(), BlockNumber: big.NewInt(1)}
	err := s.scanInitialDKGContracts(context.Background(), opts, indexer)
	assert.NilError(t, err)

	assert.DeepEqual(t, []common.Address{shared}, s.trackedDKGContractList())
}

func TestScanInitialDKGContractsZeroAddressSkippedWithWarning(t *testing.T) {
	ks0 := common.HexToAddress("0xaa")
	ks1 := common.HexToAddress("0xbb")
	dkg1 := common.HexToAddress("0xdd")

	indexer := &fakeKeyperSetIndexer{addresses: []common.Address{ks0, ks1}}
	resolver := stubResolver(map[common.Address]common.Address{ks0: {}, ks1: dkg1})
	s, rec := newSyncer(t, resolver)

	opts := &bind.CallOpts{Context: context.Background(), BlockNumber: big.NewInt(1)}
	err := s.scanInitialDKGContracts(context.Background(), opts, indexer)
	assert.NilError(t, err)

	assert.DeepEqual(t, []common.Address{dkg1}, s.trackedDKGContractList())
	assert.Equal(t, 1, rec.warnCount(), "exactly one warning should be emitted for the zero-address keyper set")
}

func TestScanInitialDKGContractsEmptyKeyperSetManager(t *testing.T) {
	indexer := &fakeKeyperSetIndexer{}
	resolver := stubResolver(nil)
	s, rec := newSyncer(t, resolver)

	opts := &bind.CallOpts{Context: context.Background(), BlockNumber: big.NewInt(1)}
	err := s.scanInitialDKGContracts(context.Background(), opts, indexer)
	assert.NilError(t, err)

	assert.Equal(t, 0, len(s.trackedDKGContractList()))
	assert.Equal(t, 0, rec.warnCount())
}

func TestHandleKeyperSetAddedRecordsNewContract(t *testing.T) {
	ksAddr := common.HexToAddress("0xaa")
	dkgAddr := common.HexToAddress("0xbb")
	resolver := stubResolver(map[common.Address]common.Address{ksAddr: dkgAddr})
	s, _ := newSyncer(t, resolver)

	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: ksAddr,
		Eon:               3,
	}
	ev.Raw.BlockNumber = 42

	s.handleKeyperSetAdded(context.Background(), ev)

	assert.DeepEqual(t, []common.Address{dkgAddr}, s.trackedDKGContractList())
}

func TestHandleKeyperSetAddedDeduplicatesAgainstExisting(t *testing.T) {
	ks0 := common.HexToAddress("0xaa")
	ks1 := common.HexToAddress("0xbb")
	shared := common.HexToAddress("0xcc")
	resolver := stubResolver(map[common.Address]common.Address{ks0: shared, ks1: shared})
	s, _ := newSyncer(t, resolver)

	indexer := &fakeKeyperSetIndexer{addresses: []common.Address{ks0}}
	opts := &bind.CallOpts{Context: context.Background(), BlockNumber: big.NewInt(1)}
	assert.NilError(t, s.scanInitialDKGContracts(context.Background(), opts, indexer))
	assert.DeepEqual(t, []common.Address{shared}, s.trackedDKGContractList())

	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: ks1,
		Eon:               1,
	}
	ev.Raw.BlockNumber = 99
	s.handleKeyperSetAdded(context.Background(), ev)

	assert.DeepEqual(t, []common.Address{shared}, s.trackedDKGContractList())
}

func TestHandleKeyperSetAddedZeroAddressLogsWarning(t *testing.T) {
	ksAddr := common.HexToAddress("0xaa")
	resolver := stubResolver(map[common.Address]common.Address{ksAddr: {}})
	s, rec := newSyncer(t, resolver)

	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: ksAddr,
		Eon:               5,
	}
	s.handleKeyperSetAdded(context.Background(), ev)

	assert.Equal(t, 0, len(s.trackedDKGContractList()))
	assert.Equal(t, 1, rec.warnCount())
}

func TestHandleKeyperSetAddedResolverErrorIsToleratedWithoutTracking(t *testing.T) {
	resolver := func(context.Context, *bind.CallOpts, common.Address) (common.Address, error) {
		return common.Address{}, errors.New("simulated RPC failure")
	}
	s, _ := newSyncer(t, resolver)

	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: common.HexToAddress("0xaa"),
		Eon:               9,
	}
	s.handleKeyperSetAdded(context.Background(), ev)

	assert.Equal(t, 0, len(s.trackedDKGContractList()),
		"a resolver error must not result in any DKG contract being tracked")
}

// gethLoggerInterface compile-time assertion: recordingLogger must satisfy the
// log.Logger interface so it can be wired into DKGEventSyncer.Log in real
// scenarios as well as in these tests.
var _ gethlog.Logger = (*recordingLogger)(nil)

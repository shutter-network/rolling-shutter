package sequencer_test

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocknode"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/encoding"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/rpc"
)

var (
	ADDRESS1                   common.Address = common.BigToAddress(big.NewInt(1))
	sequencerL1PollInterval                   = time.Second
	sequencerMaxBlockDeviation uint64         = 5
)

var BigIntComparer gocmp.Option = gocmp.Comparer(func(x, y *big.Int) bool {
	if x == nil || y == nil {
		// this only happens when one of the fields
		// in the struct is a nil ptr and the other isn't
		// if both are nil, then the compare function isn't called
		return false
	}
	return x.Cmp(y) == 0
})

type Fixtures struct {
	Sequencer       *mocksequencer.Sequencer
	SequencerClient *client.Client
	L1Service       *L1Service

	ChainID         *big.Int
	Signer          txtypes.Signer
	KeyEnvironment  *encoding.EonKeyEnvironment
	PrivkeyCollator *ecdsa.PrivateKey
	AddressCollator common.Address
	PrivkeySenders  []*ecdsa.PrivateKey
	AddressSenders  []common.Address
}

func (fx *Fixtures) MakeShutterTx(
	nonce,
	batchIndex,
	l1BlockNumber uint64,
	payload *txtypes.ShutterPayload,
) (*txtypes.ShutterTx, error) {
	if payload == nil {
		payload = &txtypes.ShutterPayload{
			To:    &ADDRESS1,
			Data:  []byte("secret"),
			Value: big.NewInt(123456789),
		}
	}
	identityPreimage := identitypreimage.Uint64ToIdentityPreimage(batchIndex)

	encryptedPayload, err := mocknode.EncryptShutterPayload(payload, identityPreimage, fx.KeyEnvironment.EonPublicKey())
	if err != nil {
		return nil, err
	}

	shtxInner := &txtypes.ShutterTx{
		ChainID:          fx.ChainID,
		Nonce:            nonce,
		GasTipCap:        big.NewInt(4200),
		GasFeeCap:        big.NewInt(1100000000),
		Gas:              1000,
		EncryptedPayload: encryptedPayload,
		Payload:          payload,
		BatchIndex:       batchIndex,
		L1BlockNumber:    l1BlockNumber,
	}
	return shtxInner, nil
}

type L1Service struct {
	URL         string
	blockNumber uint64
	mux         sync.RWMutex
}

func NewL1Service(url string) *L1Service {
	return &L1Service{
		URL:         url,
		blockNumber: 0,
		mux:         sync.RWMutex{},
	}
}

func (s *L1Service) setBlockNumber(n uint64) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.blockNumber = n
}

func (s *L1Service) BlockNumber() hexutil.Uint64 {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return hexutil.Uint64(s.blockNumber)
}

func (s *L1Service) listenAndServe(ctx context.Context) error {
	rpcServer := ethrpc.NewServer()

	err := rpcServer.RegisterName("eth", s)
	if err != nil {
		return errors.Wrap(err, "error while trying to register RPCService")
	}
	return mocksequencer.RPCListenAndServe(ctx, rpcServer, s.URL, make(<-chan error))
}

func NewFixtures(ctx context.Context, numSenders int, serveHTTP bool) (*Fixtures, error) {
	var err error
	chainID := big.NewInt(42)
	signer := txtypes.NewLondonSigner(chainID)
	fx := &Fixtures{
		Signer:  signer,
		ChainID: chainID,
	}
	fx.KeyEnvironment, err = encoding.NewEonKeyEnvironment()
	if err != nil {
		return nil, err
	}
	fx.PrivkeyCollator, err = ethcrypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	fx.AddressCollator = ethcrypto.PubkeyToAddress(fx.PrivkeyCollator.PublicKey)

	fx.PrivkeySenders = make([]*ecdsa.PrivateKey, numSenders)
	fx.AddressSenders = make([]common.Address, numSenders)

	for i := 0; i < numSenders; i++ {
		var p *ecdsa.PrivateKey
		p, err = ethcrypto.GenerateKey()
		if err != nil {
			return nil, err
		}
		fx.PrivkeySenders[i] = p
		fx.AddressSenders[i] = ethcrypto.PubkeyToAddress(p.PublicKey)
	}
	eonPubKey, err := fx.KeyEnvironment.EonPublicKey().GobEncode()
	if err != nil {
		return nil, err
	}

	// right now the tests will fail if any of those ports
	// are already in use
	l1ServerURL := "http://localhost:8555"
	fx.Sequencer = mocksequencer.New(
		fx.ChainID,
		":8545",
		l1ServerURL,
		sequencerL1PollInterval,
		sequencerMaxBlockDeviation,
	)

	// set the sequencer internal activation-block based values
	fx.Sequencer.EonKeys.Set(eonPubKey, 0)
	fx.Sequencer.Collators.Set(fx.AddressCollator, 0)

	//nolint:godox //this is not worth to be a issue right now
	// TODO(ezdac) it would be better if we could use a httptest server
	// for the sequencer and the l1 server
	// in order to assign us unoccupied ports etc., something like:
	//
	// server = httptest.NewServer(http.HandlerFunc(mock.handle))
	// url = server.URL

	if serveHTTP {
		fx.L1Service = NewL1Service(":8555")
		errgrp, errctx := errgroup.WithContext(ctx)
		errgrp.Go(func() error {
			return fx.L1Service.listenAndServe(errctx)
		})
		errgrp.Go(func() error {
			return fx.Sequencer.ListenAndServe(
				errctx,
				&rpc.AdminService{},
				&rpc.EthService{},
				&rpc.ShutterService{},
			)
		})
		time.Sleep(1000 * time.Millisecond)
		// connect to the mocksequencer running in the background
		// via HTTP
		fx.SequencerClient, err = client.DialContext(errctx, "http://localhost:8545")
	}
	return fx, err
}

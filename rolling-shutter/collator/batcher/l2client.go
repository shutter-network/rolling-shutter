package batcher

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/l2client"
)

type AccountInfo struct {
	Balance *big.Int
	Nonce   uint64
}

type BlockInfo interface {
	Number() *big.Int
	GasLimit() uint64
	BaseFee() *big.Int
	Coinbase() common.Address
}

type L2ClientReader interface {
	GetAccountInfo(ctx context.Context, account common.Address) (AccountInfo, error)
	GetBatchIndex(ctx context.Context) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
	GetBlockInfo(ctx context.Context) (BlockInfo, error)
}

type rpcClient struct {
	client *rpc.Client
}

func (rc *rpcClient) GetAccountInfo(ctx context.Context, account common.Address) (AccountInfo, error) {
	ec := ethclient.NewClient(rc.client)
	balance, err := ec.BalanceAt(ctx, account, nil)
	if err != nil {
		return AccountInfo{}, err
	}
	nonce, err := ec.NonceAt(ctx, account, nil)
	if err != nil {
		return AccountInfo{}, err
	}
	return AccountInfo{
		Balance: balance,
		Nonce:   nonce,
	}, nil
}

func (rc *rpcClient) GetBatchIndex(ctx context.Context) (uint64, error) {
	return l2client.GetBatchIndex(ctx, rc.client)
}

func (rc *rpcClient) ChainID(ctx context.Context) (*big.Int, error) {
	return ethclient.NewClient(rc.client).ChainID(ctx)
}

func (rc *rpcClient) GetBlockInfo(ctx context.Context) (BlockInfo, error) {
	return ethclient.NewClient(rc.client).BlockByNumber(ctx, nil)
}

func NewRPCClient(ctx context.Context, url string) (L2ClientReader, error) {
	client, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	return &rpcClient{client: client}, nil
}

package client

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"
)

type Client struct {
	*ethclient.Client
	client *ethrpc.Client
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := ethrpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: ethclient.NewClient(c),
		client: c,
	}, nil
}

func (c *Client) SetBalance(ctx context.Context, address common.Address, balance *big.Int) error {
	var result int
	err := c.client.CallContext(ctx, &result, "admin_setBalance", address.Hex(), hexutil.EncodeBig(balance))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) BatchIndex(ctx context.Context) (uint64, error) {
	var result hexutil.Uint64
	err := c.client.CallContext(ctx, &result, "shutter_batchIndex")
	if err != nil {
		return uint64(result), err
	}
	return uint64(result), nil
}

func (c *Client) SubmitBatch(ctx context.Context, tx *txtypes.Transaction) (*common.Hash, error) {
	data, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	var result string
	err = c.client.CallContext(ctx, &result, "shutter_submitBatch", hexutil.Encode(data))
	// result is the TX hash if successful, otherwise the empty string
	if err != nil || result == "" {
		return nil, err
	}
	txHash := common.HexToHash(result)
	return &txHash, err
}

func (c *Client) TransactionByHash(
	ctx context.Context,
	hash common.Hash,
) (*txtypes.Transaction, bool, error) {
	var result *txtypes.TransactionData
	err := c.client.CallContext(ctx, &result, "shutter_getTransactionByHash", hash)
	if err != nil {
		return nil, false, err
	} else if result == nil {
		return nil, false, ethereum.NotFound
	} else if result.R == nil {
		return nil, false, errors.New("server returned transaction without signature")
	}
	tx := &txtypes.Transaction{}
	err = tx.FromTransactionData(result)
	if err != nil {
		return nil, false, errors.Wrap(err, "can't decode transaction")
	}

	return tx, result.BlockNumber == nil, err
}

package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/sandbox"
)

var (
	chainId = big.NewInt(452346)
	signer  = types.LatestSignerForChainID(chainId)
)

func main() {
	ctx := context.Background()

	collatorPrivateKey := sandbox.GanacheKey(7)

	unsignedBatchTx := &types.BatchTx{
		ChainID:       chainId,
		DecryptionKey: []byte{},
		BatchIndex:    0,
		L1BlockNumber: 0,
		Timestamp:     big.NewInt(0),
		Transactions:  [][]byte{},
	}
	batchTx, err := types.SignNewTx(collatorPrivateKey, signer, unsignedBatchTx)
	panicIfErr(err)
	batchTxBytes, err := batchTx.MarshalBinary()
	panicIfErr(err)

	fmt.Printf("%+v\n", batchTx.GetInner().V)

	client, err := rpc.Dial("http://localhost:8547")
	panicIfErr(err)

	var result string
	err = client.CallContext(ctx, &result, "eth_sendRawTransaction", hexutil.Encode(batchTxBytes))
	panicIfErr(err)
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

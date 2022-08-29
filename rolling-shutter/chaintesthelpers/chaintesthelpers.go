// Package chaintesthelpers provides functions to write tests that uses a blockchain node with
// deployed rolling-shutter related contracts.
// Testing functions should be skipped if `NodeURLEnv` or `DeploymentsDirEnv` env are not set.
// Test functions should look like:
//
//	func Test(t *testinng.T) {
//	  SkipChainTests(t)
//	  ctx := context.Background()
//	  contracts, cleanup := NewTestContracts(t, ctx)
//	  defer cleanup()
//	}
//
// Or if snapshots need to be taken during the test:
//
//	func Test(t *testing.T) {
//	  SkipChainTests(t)
//	  ctx := context.Background()
//	  client := GetChainClient(t, ctx)
//	  contracts := client.NewTestContracts(t)
//
//	  snapshotID := client.TakeSnapshot(t, ctx)
//	  client.RevertToSnapshot(t, ctx, snapshotID)
//	}
package chaintesthelpers

import (
	"context"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
)

const (
	NodeURLEnv        = "ROLLING_SHUTTER_TEST_NODE_URL"
	DeploymentsDirEnv = "ROLLING_SHUTTER_DEPLOYMENTS_DIR"
	HardhatFundedKey  = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
)

type ChainClient struct {
	rpcClient *rpc.Client
}

type SnapshotID string

func NewChainClient(ctx context.Context, t *testing.T, ethURL string) ChainClient {
	t.Helper()

	rpcClient, err := rpc.DialContext(ctx, ethURL)
	assert.NilError(t, err)

	return ChainClient{rpcClient}
}

func (client ChainClient) TakeSnapshot(ctx context.Context, t *testing.T) SnapshotID {
	t.Helper()
	result := ""
	err := client.rpcClient.CallContext(ctx, &result, "evm_snapshot", nil...)
	assert.NilError(t, err)
	return SnapshotID(result)
}

func (client ChainClient) GetBlockNumber(ctx context.Context, t *testing.T) uint64 {
	t.Helper()

	blk, err := ethclient.NewClient(client.rpcClient).BlockNumber(ctx)
	assert.NilError(t, err)

	return blk
}

func (client ChainClient) MineBlock(ctx context.Context, t *testing.T) {
	t.Helper()
	err := client.rpcClient.CallContext(ctx, nil, "evm_mine", nil...)
	assert.NilError(t, err)
}

func (client ChainClient) RevertToSnapshot(ctx context.Context, t *testing.T, snapshotID SnapshotID) {
	t.Helper()
	err := client.rpcClient.CallContext(ctx, nil, "evm_revert", snapshotID)
	assert.NilError(t, err)
}

func (client ChainClient) NewTestContracts(t *testing.T) *deployment.Contracts {
	t.Helper()
	ethClient := ethclient.NewClient(client.rpcClient)
	deploymentDir := os.Getenv(DeploymentsDirEnv)
	contracts, err := deployment.NewContracts(ethClient, deploymentDir)
	assert.NilError(t, err)
	return contracts
}

func NewTestContracts(ctx context.Context, t *testing.T) (*deployment.Contracts, func()) {
	t.Helper()
	deploymentDir := os.Getenv(DeploymentsDirEnv)

	client := GetChainClient(ctx, t)
	snapshotID := client.TakeSnapshot(ctx, t)
	revert := func() {
		client.RevertToSnapshot(ctx, t, snapshotID)
	}

	ethClient := ethclient.NewClient(client.rpcClient)
	contracts, err := deployment.NewContracts(ethClient, deploymentDir)
	assert.NilError(t, err)

	return contracts, revert
}

func SkipChainTests(t *testing.T) {
	t.Helper()
	if os.Getenv(NodeURLEnv) == "" || os.Getenv(DeploymentsDirEnv) == "" {
		t.Skip("skipping test using a chain node")
	}
}

func GetChainClient(ctx context.Context, t *testing.T) ChainClient {
	t.Helper()
	ethURL := os.Getenv(NodeURLEnv)
	return NewChainClient(ctx, t, ethURL)
}

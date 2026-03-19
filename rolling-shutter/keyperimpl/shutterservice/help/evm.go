package help

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"gotest.tools/assert"
)

type Setup struct {
	Backend         backends.SimulatedBackend
	Auth            *bind.TransactOpts
	Contract        *Emitter
	ContractAddress common.Address
}

func SetupBackend(t *testing.T) Setup {
	t.Helper()

	// create funded genesis account
	privateKey, err := crypto.GenerateKey()
	assert.NilError(t, err, "failed to generate private key %v", err)

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	assert.NilError(t, err, "failed to create transactor %v", err)

	balance := new(big.Int)
	balance.SetString("1000000000000000000000", 10) // 1000 ETH
	alloc := make(types.GenesisAlloc)
	alloc[auth.From] = types.Account{Balance: balance}

	// create SimulatedBackend
	b := simulated.NewBackend(alloc, simulated.WithBlockGasLimit(8000000))
	backend := backends.SimulatedBackend{
		Backend: b,
		Client:  b.Client(),
	}
	// deploy Emitter contract
	// Emitter.go is generated through `make all`
	contractAddress, _, _, err := DeployEmitter(auth, backend)
	assert.NilError(t, err, "failed to deploy contract: %v", err)
	backend.Commit()

	// bind contract
	contract, err := NewEmitter(contractAddress, backend)
	assert.NilError(t, err, "failed to bind contract instance to address %v: %v", contractAddress, err)

	return Setup{
		Backend:         backend,
		Auth:            auth,
		Contract:        contract,
		ContractAddress: contractAddress,
	}
}

func CollectLog(t *testing.T, setup Setup, tx *types.Transaction) (*types.Log, error) {
	t.Helper()
	// Commit the block to process the transaction
	setup.Backend.Commit()

	// get Receipt for Logs
	receipt, err := setup.Backend.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt: %v", err)
	}
	return receipt.Logs[0], nil
}

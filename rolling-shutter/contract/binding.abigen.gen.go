// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// CollatorConfig is an auto generated low-level Go binding around an user-defined struct.
type CollatorConfig struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}

// KeypersConfig is an auto generated low-level Go binding around an user-defined struct.
type KeypersConfig struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
	Threshold             uint64
}

// AddrsSeqMetaData contains all meta data concerning the AddrsSeq contract.
var AddrsSeqMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"i\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"newAddrs\",\"type\":\"address[]\"}],\"name\":\"Added\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"}],\"name\":\"Appended\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"newAddrs\",\"type\":\"address[]\"}],\"name\":\"add\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"append\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"i\",\"type\":\"uint64\"}],\"name\":\"at\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"}],\"name\":\"countNth\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b50338061003557604051631e4fbdf760e01b81525f600482015260240160405180910390fd5b61003e8161004c565b5061004761009b565b610170565b5f80546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080515f602080830182815283850190945292825260018054808201825591528151805192937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909201926100f492849201906100f9565b505050565b828054828255905f5260205f2090810192821561014c579160200282015b8281111561014c57825182546001600160a01b0319166001600160a01b03909116178255602090920191600190910190610117565b5061015892915061015c565b5090565b5b80821115610158575f815560010161015d565b6109078061017d5f395ff3fe608060405234801561000f575f80fd5b5060043610610085575f3560e01c80637f353d55116100585780637f353d55146100f65780638da5cb5b146100fe578063c4c1c94f1461010e578063f2fde38b14610121575f80fd5b806306661abd146100895780632a2d01f8146100ae57806335147092146100c1578063715018a6146100ec575b5f80fd5b610091610134565b6040516001600160401b0390911681526020015b60405180910390f35b6100916100bc3660046106f1565b610148565b6100d46100cf366004610711565b6101ed565b6040516001600160a01b0390911681526020016100a5565b6100f4610331565b005b6100f4610344565b5f546001600160a01b03166100d4565b6100f461011c366004610742565b610418565b6100f461012f3660046107c6565b610549565b600180545f91610143916107f3565b905090565b5f610151610134565b6001600160401b0316826001600160401b0316106101c05760405162461bcd60e51b815260206004820152602160248201527f41646472735365712e636f756e744e74683a206e206f7574206f662072616e676044820152606560f81b60648201526084015b60405180910390fd5b6001826001600160401b0316815481106101dc576101dc61081a565b5f9182526020909120015492915050565b5f6101f6610134565b6001600160401b0316836001600160401b0316106102565760405162461bcd60e51b815260206004820152601b60248201527f41646472735365712e61743a206e206f7574206f662072616e6765000000000060448201526064016101b7565b6001836001600160401b0316815481106102725761027261081a565b5f918252602090912001546001600160401b038316106102d45760405162461bcd60e51b815260206004820152601b60248201527f41646472735365712e61743a2069206f7574206f662072616e6765000000000060448201526064016101b7565b6001836001600160401b0316815481106102f0576102f061081a565b905f5260205f20015f01826001600160401b0316815481106103145761031461081a565b5f918252602090912001546001600160a01b031690505b92915050565b610339610586565b6103425f6105b2565b565b61034c610586565b61035e60016001600160401b036107f3565b6001600160401b0316600180549050106103c65760405162461bcd60e51b815260206004820152602360248201527f41646472735365712e617070656e643a20736571206578636565656473206c696044820152621b5a5d60ea1b60648201526084016101b7565b600180547f5ff9c98a1faf73c018d22371cb08c08dec1412825b68523a8e7deaa17683a6b9916103f5916107f3565b6040516001600160401b03909116815260200160405180910390a1610342610601565b610420610586565b600180545f9161042f9161082e565b90505f6001826001600160401b03168154811061044e5761044e61081a565b5f91825260208220015491505b6001600160401b038116841115610505576001836001600160401b0316815481106104885761048861081a565b905f5260205f20015f018585836001600160401b03168181106104ad576104ad61081a565b90506020020160208101906104c291906107c6565b81546001810183555f928352602090922090910180546001600160a01b0319166001600160a01b03909216919091179055806104fd81610841565b91505061045b565b507f54a93d30cc356a58fe6fe472b453c3ea842500e17a2e9972af429d866f305fbd8282868660405161053b9493929190610866565b60405180910390a150505050565b610551610586565b6001600160a01b03811661057a57604051631e4fbdf760e01b81525f60048201526024016101b7565b610583816105b2565b50565b5f546001600160a01b031633146103425760405163118cdaa760e01b81523360048201526024016101b7565b5f80546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080515f602080830182815283850190945292825260018054808201825591528151805192937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf69092019261065a928492019061065f565b505050565b828054828255905f5260205f209081019282156106b2579160200282015b828111156106b257825182546001600160a01b0319166001600160a01b0390911617825560209092019160019091019061067d565b506106be9291506106c2565b5090565b5b808211156106be575f81556001016106c3565b80356001600160401b03811681146106ec575f80fd5b919050565b5f60208284031215610701575f80fd5b61070a826106d6565b9392505050565b5f8060408385031215610722575f80fd5b61072b836106d6565b9150610739602084016106d6565b90509250929050565b5f8060208385031215610753575f80fd5b82356001600160401b0380821115610769575f80fd5b818501915085601f83011261077c575f80fd5b81358181111561078a575f80fd5b8660208260051b850101111561079e575f80fd5b60209290920196919550909350505050565b80356001600160a01b03811681146106ec575f80fd5b5f602082840312156107d6575f80fd5b61070a826107b0565b634e487b7160e01b5f52601160045260245ffd5b6001600160401b03828116828216039080821115610813576108136107df565b5092915050565b634e487b7160e01b5f52603260045260245ffd5b8181038181111561032b5761032b6107df565b5f6001600160401b0380831681810361085c5761085c6107df565b6001019392505050565b6001600160401b038581168252841660208083019190915260606040830181905282018390525f9084906080840190835b868110156108c3576001600160a01b036108b0856107b0565b1683529281019291810191600101610897565b50909897505050505050505056fea26469706673582212205e1a62d5be98221b90b384ec4b253c70f3aa6ead1b7445fa350554cfd6c2d1e564736f6c63430008160033",
}

// AddrsSeqABI is the input ABI used to generate the binding from.
// Deprecated: Use AddrsSeqMetaData.ABI instead.
var AddrsSeqABI = AddrsSeqMetaData.ABI

// AddrsSeqBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AddrsSeqMetaData.Bin instead.
var AddrsSeqBin = AddrsSeqMetaData.Bin

// DeployAddrsSeq deploys a new Ethereum contract, binding an instance of AddrsSeq to it.
func DeployAddrsSeq(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *AddrsSeq, error) {
	parsed, err := AddrsSeqMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AddrsSeqBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AddrsSeq{AddrsSeqCaller: AddrsSeqCaller{contract: contract}, AddrsSeqTransactor: AddrsSeqTransactor{contract: contract}, AddrsSeqFilterer: AddrsSeqFilterer{contract: contract}}, nil
}

// AddrsSeq is an auto generated Go binding around an Ethereum contract.
type AddrsSeq struct {
	AddrsSeqCaller     // Read-only binding to the contract
	AddrsSeqTransactor // Write-only binding to the contract
	AddrsSeqFilterer   // Log filterer for contract events
}

// AddrsSeqCaller is an auto generated read-only Go binding around an Ethereum contract.
type AddrsSeqCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddrsSeqTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AddrsSeqTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddrsSeqFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AddrsSeqFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddrsSeqSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AddrsSeqSession struct {
	Contract     *AddrsSeq         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AddrsSeqCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AddrsSeqCallerSession struct {
	Contract *AddrsSeqCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// AddrsSeqTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AddrsSeqTransactorSession struct {
	Contract     *AddrsSeqTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// AddrsSeqRaw is an auto generated low-level Go binding around an Ethereum contract.
type AddrsSeqRaw struct {
	Contract *AddrsSeq // Generic contract binding to access the raw methods on
}

// AddrsSeqCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AddrsSeqCallerRaw struct {
	Contract *AddrsSeqCaller // Generic read-only contract binding to access the raw methods on
}

// AddrsSeqTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AddrsSeqTransactorRaw struct {
	Contract *AddrsSeqTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAddrsSeq creates a new instance of AddrsSeq, bound to a specific deployed contract.
func NewAddrsSeq(address common.Address, backend bind.ContractBackend) (*AddrsSeq, error) {
	contract, err := bindAddrsSeq(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AddrsSeq{AddrsSeqCaller: AddrsSeqCaller{contract: contract}, AddrsSeqTransactor: AddrsSeqTransactor{contract: contract}, AddrsSeqFilterer: AddrsSeqFilterer{contract: contract}}, nil
}

// NewAddrsSeqCaller creates a new read-only instance of AddrsSeq, bound to a specific deployed contract.
func NewAddrsSeqCaller(address common.Address, caller bind.ContractCaller) (*AddrsSeqCaller, error) {
	contract, err := bindAddrsSeq(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AddrsSeqCaller{contract: contract}, nil
}

// NewAddrsSeqTransactor creates a new write-only instance of AddrsSeq, bound to a specific deployed contract.
func NewAddrsSeqTransactor(address common.Address, transactor bind.ContractTransactor) (*AddrsSeqTransactor, error) {
	contract, err := bindAddrsSeq(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AddrsSeqTransactor{contract: contract}, nil
}

// NewAddrsSeqFilterer creates a new log filterer instance of AddrsSeq, bound to a specific deployed contract.
func NewAddrsSeqFilterer(address common.Address, filterer bind.ContractFilterer) (*AddrsSeqFilterer, error) {
	contract, err := bindAddrsSeq(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AddrsSeqFilterer{contract: contract}, nil
}

// bindAddrsSeq binds a generic wrapper to an already deployed contract.
func bindAddrsSeq(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AddrsSeqMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AddrsSeq *AddrsSeqRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AddrsSeq.Contract.AddrsSeqCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AddrsSeq *AddrsSeqRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddrsSeq.Contract.AddrsSeqTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AddrsSeq *AddrsSeqRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AddrsSeq.Contract.AddrsSeqTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AddrsSeq *AddrsSeqCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AddrsSeq.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AddrsSeq *AddrsSeqTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddrsSeq.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AddrsSeq *AddrsSeqTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AddrsSeq.Contract.contract.Transact(opts, method, params...)
}

// At is a free data retrieval call binding the contract method 0x35147092.
//
// Solidity: function at(uint64 n, uint64 i) view returns(address)
func (_AddrsSeq *AddrsSeqCaller) At(opts *bind.CallOpts, n uint64, i uint64) (common.Address, error) {
	var out []interface{}
	err := _AddrsSeq.contract.Call(opts, &out, "at", n, i)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// At is a free data retrieval call binding the contract method 0x35147092.
//
// Solidity: function at(uint64 n, uint64 i) view returns(address)
func (_AddrsSeq *AddrsSeqSession) At(n uint64, i uint64) (common.Address, error) {
	return _AddrsSeq.Contract.At(&_AddrsSeq.CallOpts, n, i)
}

// At is a free data retrieval call binding the contract method 0x35147092.
//
// Solidity: function at(uint64 n, uint64 i) view returns(address)
func (_AddrsSeq *AddrsSeqCallerSession) At(n uint64, i uint64) (common.Address, error) {
	return _AddrsSeq.Contract.At(&_AddrsSeq.CallOpts, n, i)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint64)
func (_AddrsSeq *AddrsSeqCaller) Count(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _AddrsSeq.contract.Call(opts, &out, "count")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint64)
func (_AddrsSeq *AddrsSeqSession) Count() (uint64, error) {
	return _AddrsSeq.Contract.Count(&_AddrsSeq.CallOpts)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint64)
func (_AddrsSeq *AddrsSeqCallerSession) Count() (uint64, error) {
	return _AddrsSeq.Contract.Count(&_AddrsSeq.CallOpts)
}

// CountNth is a free data retrieval call binding the contract method 0x2a2d01f8.
//
// Solidity: function countNth(uint64 n) view returns(uint64)
func (_AddrsSeq *AddrsSeqCaller) CountNth(opts *bind.CallOpts, n uint64) (uint64, error) {
	var out []interface{}
	err := _AddrsSeq.contract.Call(opts, &out, "countNth", n)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// CountNth is a free data retrieval call binding the contract method 0x2a2d01f8.
//
// Solidity: function countNth(uint64 n) view returns(uint64)
func (_AddrsSeq *AddrsSeqSession) CountNth(n uint64) (uint64, error) {
	return _AddrsSeq.Contract.CountNth(&_AddrsSeq.CallOpts, n)
}

// CountNth is a free data retrieval call binding the contract method 0x2a2d01f8.
//
// Solidity: function countNth(uint64 n) view returns(uint64)
func (_AddrsSeq *AddrsSeqCallerSession) CountNth(n uint64) (uint64, error) {
	return _AddrsSeq.Contract.CountNth(&_AddrsSeq.CallOpts, n)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AddrsSeq *AddrsSeqCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AddrsSeq.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AddrsSeq *AddrsSeqSession) Owner() (common.Address, error) {
	return _AddrsSeq.Contract.Owner(&_AddrsSeq.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AddrsSeq *AddrsSeqCallerSession) Owner() (common.Address, error) {
	return _AddrsSeq.Contract.Owner(&_AddrsSeq.CallOpts)
}

// Add is a paid mutator transaction binding the contract method 0xc4c1c94f.
//
// Solidity: function add(address[] newAddrs) returns()
func (_AddrsSeq *AddrsSeqTransactor) Add(opts *bind.TransactOpts, newAddrs []common.Address) (*types.Transaction, error) {
	return _AddrsSeq.contract.Transact(opts, "add", newAddrs)
}

// Add is a paid mutator transaction binding the contract method 0xc4c1c94f.
//
// Solidity: function add(address[] newAddrs) returns()
func (_AddrsSeq *AddrsSeqSession) Add(newAddrs []common.Address) (*types.Transaction, error) {
	return _AddrsSeq.Contract.Add(&_AddrsSeq.TransactOpts, newAddrs)
}

// Add is a paid mutator transaction binding the contract method 0xc4c1c94f.
//
// Solidity: function add(address[] newAddrs) returns()
func (_AddrsSeq *AddrsSeqTransactorSession) Add(newAddrs []common.Address) (*types.Transaction, error) {
	return _AddrsSeq.Contract.Add(&_AddrsSeq.TransactOpts, newAddrs)
}

// Append is a paid mutator transaction binding the contract method 0x7f353d55.
//
// Solidity: function append() returns()
func (_AddrsSeq *AddrsSeqTransactor) Append(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddrsSeq.contract.Transact(opts, "append")
}

// Append is a paid mutator transaction binding the contract method 0x7f353d55.
//
// Solidity: function append() returns()
func (_AddrsSeq *AddrsSeqSession) Append() (*types.Transaction, error) {
	return _AddrsSeq.Contract.Append(&_AddrsSeq.TransactOpts)
}

// Append is a paid mutator transaction binding the contract method 0x7f353d55.
//
// Solidity: function append() returns()
func (_AddrsSeq *AddrsSeqTransactorSession) Append() (*types.Transaction, error) {
	return _AddrsSeq.Contract.Append(&_AddrsSeq.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AddrsSeq *AddrsSeqTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddrsSeq.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AddrsSeq *AddrsSeqSession) RenounceOwnership() (*types.Transaction, error) {
	return _AddrsSeq.Contract.RenounceOwnership(&_AddrsSeq.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AddrsSeq *AddrsSeqTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _AddrsSeq.Contract.RenounceOwnership(&_AddrsSeq.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AddrsSeq *AddrsSeqTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _AddrsSeq.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AddrsSeq *AddrsSeqSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AddrsSeq.Contract.TransferOwnership(&_AddrsSeq.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AddrsSeq *AddrsSeqTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AddrsSeq.Contract.TransferOwnership(&_AddrsSeq.TransactOpts, newOwner)
}

// AddrsSeqAddedIterator is returned from FilterAdded and is used to iterate over the raw logs and unpacked data for Added events raised by the AddrsSeq contract.
type AddrsSeqAddedIterator struct {
	Event *AddrsSeqAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AddrsSeqAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AddrsSeqAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AddrsSeqAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AddrsSeqAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AddrsSeqAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AddrsSeqAdded represents a Added event raised by the AddrsSeq contract.
type AddrsSeqAdded struct {
	N        uint64
	I        uint64
	NewAddrs []common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterAdded is a free log retrieval operation binding the contract event 0x54a93d30cc356a58fe6fe472b453c3ea842500e17a2e9972af429d866f305fbd.
//
// Solidity: event Added(uint64 n, uint64 i, address[] newAddrs)
func (_AddrsSeq *AddrsSeqFilterer) FilterAdded(opts *bind.FilterOpts) (*AddrsSeqAddedIterator, error) {

	logs, sub, err := _AddrsSeq.contract.FilterLogs(opts, "Added")
	if err != nil {
		return nil, err
	}
	return &AddrsSeqAddedIterator{contract: _AddrsSeq.contract, event: "Added", logs: logs, sub: sub}, nil
}

// WatchAdded is a free log subscription operation binding the contract event 0x54a93d30cc356a58fe6fe472b453c3ea842500e17a2e9972af429d866f305fbd.
//
// Solidity: event Added(uint64 n, uint64 i, address[] newAddrs)
func (_AddrsSeq *AddrsSeqFilterer) WatchAdded(opts *bind.WatchOpts, sink chan<- *AddrsSeqAdded) (event.Subscription, error) {

	logs, sub, err := _AddrsSeq.contract.WatchLogs(opts, "Added")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AddrsSeqAdded)
				if err := _AddrsSeq.contract.UnpackLog(event, "Added", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAdded is a log parse operation binding the contract event 0x54a93d30cc356a58fe6fe472b453c3ea842500e17a2e9972af429d866f305fbd.
//
// Solidity: event Added(uint64 n, uint64 i, address[] newAddrs)
func (_AddrsSeq *AddrsSeqFilterer) ParseAdded(log types.Log) (*AddrsSeqAdded, error) {
	event := new(AddrsSeqAdded)
	if err := _AddrsSeq.contract.UnpackLog(event, "Added", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AddrsSeqAppendedIterator is returned from FilterAppended and is used to iterate over the raw logs and unpacked data for Appended events raised by the AddrsSeq contract.
type AddrsSeqAppendedIterator struct {
	Event *AddrsSeqAppended // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AddrsSeqAppendedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AddrsSeqAppended)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AddrsSeqAppended)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AddrsSeqAppendedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AddrsSeqAppendedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AddrsSeqAppended represents a Appended event raised by the AddrsSeq contract.
type AddrsSeqAppended struct {
	N   uint64
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAppended is a free log retrieval operation binding the contract event 0x5ff9c98a1faf73c018d22371cb08c08dec1412825b68523a8e7deaa17683a6b9.
//
// Solidity: event Appended(uint64 n)
func (_AddrsSeq *AddrsSeqFilterer) FilterAppended(opts *bind.FilterOpts) (*AddrsSeqAppendedIterator, error) {

	logs, sub, err := _AddrsSeq.contract.FilterLogs(opts, "Appended")
	if err != nil {
		return nil, err
	}
	return &AddrsSeqAppendedIterator{contract: _AddrsSeq.contract, event: "Appended", logs: logs, sub: sub}, nil
}

// WatchAppended is a free log subscription operation binding the contract event 0x5ff9c98a1faf73c018d22371cb08c08dec1412825b68523a8e7deaa17683a6b9.
//
// Solidity: event Appended(uint64 n)
func (_AddrsSeq *AddrsSeqFilterer) WatchAppended(opts *bind.WatchOpts, sink chan<- *AddrsSeqAppended) (event.Subscription, error) {

	logs, sub, err := _AddrsSeq.contract.WatchLogs(opts, "Appended")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AddrsSeqAppended)
				if err := _AddrsSeq.contract.UnpackLog(event, "Appended", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAppended is a log parse operation binding the contract event 0x5ff9c98a1faf73c018d22371cb08c08dec1412825b68523a8e7deaa17683a6b9.
//
// Solidity: event Appended(uint64 n)
func (_AddrsSeq *AddrsSeqFilterer) ParseAppended(log types.Log) (*AddrsSeqAppended, error) {
	event := new(AddrsSeqAppended)
	if err := _AddrsSeq.contract.UnpackLog(event, "Appended", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AddrsSeqOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AddrsSeq contract.
type AddrsSeqOwnershipTransferredIterator struct {
	Event *AddrsSeqOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AddrsSeqOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AddrsSeqOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AddrsSeqOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AddrsSeqOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AddrsSeqOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AddrsSeqOwnershipTransferred represents a OwnershipTransferred event raised by the AddrsSeq contract.
type AddrsSeqOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AddrsSeq *AddrsSeqFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AddrsSeqOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AddrsSeq.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AddrsSeqOwnershipTransferredIterator{contract: _AddrsSeq.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AddrsSeq *AddrsSeqFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AddrsSeqOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AddrsSeq.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AddrsSeqOwnershipTransferred)
				if err := _AddrsSeq.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AddrsSeq *AddrsSeqFilterer) ParseOwnershipTransferred(log types.Log) (*AddrsSeqOwnershipTransferred, error) {
	event := new(AddrsSeqOwnershipTransferred)
	if err := _AddrsSeq.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchCounterMetaData contains all meta data concerning the BatchCounter contract.
var BatchCounterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"CallerNotZeroAddress\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"oldIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"newIndex\",\"type\":\"uint64\"}],\"name\":\"NewBatchIndex\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"batchIndex\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"increment\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"newBatchIndex\",\"type\":\"uint64\"}],\"name\":\"set\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b506102738061001d5f395ff3fe608060405234801561000f575f80fd5b506004361061003f575f3560e01c80631d8c311f14610043578063d09de08a14610058578063e79993f314610060575b5f80fd5b6100566100513660046101db565b610090565b005b610056610119565b5f546100739067ffffffffffffffff1681565b60405167ffffffffffffffff909116815260200160405180910390f35b33156100af57604051631448d0ef60e01b815260040160405180910390fd5b5f546040805167ffffffffffffffff928316815291831660208301527f5867f9e83f14fb505a43dd58880b1de7e3b5cddbfa99bb92a15dad48b453410b910160405180910390a15f805467ffffffffffffffff191667ffffffffffffffff92909216919091179055565b331561013857604051631448d0ef60e01b815260040160405180910390fd5b5f547f5867f9e83f14fb505a43dd58880b1de7e3b5cddbfa99bb92a15dad48b453410b9067ffffffffffffffff16610171816001610209565b6040805167ffffffffffffffff93841681529290911660208301520160405180910390a15f80546001919081906101b390849067ffffffffffffffff16610209565b92506101000a81548167ffffffffffffffff021916908367ffffffffffffffff160217905550565b5f602082840312156101eb575f80fd5b813567ffffffffffffffff81168114610202575f80fd5b9392505050565b67ffffffffffffffff81811683821601908082111561023657634e487b7160e01b5f52601160045260245ffd5b509291505056fea264697066735822122023a62814b59cce3de40766d40026c2e39a45e1ae2846251d6e02b10a2784b4ef64736f6c63430008160033",
}

// BatchCounterABI is the input ABI used to generate the binding from.
// Deprecated: Use BatchCounterMetaData.ABI instead.
var BatchCounterABI = BatchCounterMetaData.ABI

// BatchCounterBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use BatchCounterMetaData.Bin instead.
var BatchCounterBin = BatchCounterMetaData.Bin

// DeployBatchCounter deploys a new Ethereum contract, binding an instance of BatchCounter to it.
func DeployBatchCounter(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *BatchCounter, error) {
	parsed, err := BatchCounterMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(BatchCounterBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &BatchCounter{BatchCounterCaller: BatchCounterCaller{contract: contract}, BatchCounterTransactor: BatchCounterTransactor{contract: contract}, BatchCounterFilterer: BatchCounterFilterer{contract: contract}}, nil
}

// BatchCounter is an auto generated Go binding around an Ethereum contract.
type BatchCounter struct {
	BatchCounterCaller     // Read-only binding to the contract
	BatchCounterTransactor // Write-only binding to the contract
	BatchCounterFilterer   // Log filterer for contract events
}

// BatchCounterCaller is an auto generated read-only Go binding around an Ethereum contract.
type BatchCounterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchCounterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BatchCounterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchCounterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BatchCounterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchCounterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BatchCounterSession struct {
	Contract     *BatchCounter     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BatchCounterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BatchCounterCallerSession struct {
	Contract *BatchCounterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// BatchCounterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BatchCounterTransactorSession struct {
	Contract     *BatchCounterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// BatchCounterRaw is an auto generated low-level Go binding around an Ethereum contract.
type BatchCounterRaw struct {
	Contract *BatchCounter // Generic contract binding to access the raw methods on
}

// BatchCounterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BatchCounterCallerRaw struct {
	Contract *BatchCounterCaller // Generic read-only contract binding to access the raw methods on
}

// BatchCounterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BatchCounterTransactorRaw struct {
	Contract *BatchCounterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBatchCounter creates a new instance of BatchCounter, bound to a specific deployed contract.
func NewBatchCounter(address common.Address, backend bind.ContractBackend) (*BatchCounter, error) {
	contract, err := bindBatchCounter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BatchCounter{BatchCounterCaller: BatchCounterCaller{contract: contract}, BatchCounterTransactor: BatchCounterTransactor{contract: contract}, BatchCounterFilterer: BatchCounterFilterer{contract: contract}}, nil
}

// NewBatchCounterCaller creates a new read-only instance of BatchCounter, bound to a specific deployed contract.
func NewBatchCounterCaller(address common.Address, caller bind.ContractCaller) (*BatchCounterCaller, error) {
	contract, err := bindBatchCounter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BatchCounterCaller{contract: contract}, nil
}

// NewBatchCounterTransactor creates a new write-only instance of BatchCounter, bound to a specific deployed contract.
func NewBatchCounterTransactor(address common.Address, transactor bind.ContractTransactor) (*BatchCounterTransactor, error) {
	contract, err := bindBatchCounter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BatchCounterTransactor{contract: contract}, nil
}

// NewBatchCounterFilterer creates a new log filterer instance of BatchCounter, bound to a specific deployed contract.
func NewBatchCounterFilterer(address common.Address, filterer bind.ContractFilterer) (*BatchCounterFilterer, error) {
	contract, err := bindBatchCounter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BatchCounterFilterer{contract: contract}, nil
}

// bindBatchCounter binds a generic wrapper to an already deployed contract.
func bindBatchCounter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BatchCounterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BatchCounter *BatchCounterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BatchCounter.Contract.BatchCounterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BatchCounter *BatchCounterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchCounter.Contract.BatchCounterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BatchCounter *BatchCounterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BatchCounter.Contract.BatchCounterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BatchCounter *BatchCounterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BatchCounter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BatchCounter *BatchCounterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchCounter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BatchCounter *BatchCounterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BatchCounter.Contract.contract.Transact(opts, method, params...)
}

// BatchIndex is a free data retrieval call binding the contract method 0xe79993f3.
//
// Solidity: function batchIndex() view returns(uint64)
func (_BatchCounter *BatchCounterCaller) BatchIndex(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _BatchCounter.contract.Call(opts, &out, "batchIndex")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// BatchIndex is a free data retrieval call binding the contract method 0xe79993f3.
//
// Solidity: function batchIndex() view returns(uint64)
func (_BatchCounter *BatchCounterSession) BatchIndex() (uint64, error) {
	return _BatchCounter.Contract.BatchIndex(&_BatchCounter.CallOpts)
}

// BatchIndex is a free data retrieval call binding the contract method 0xe79993f3.
//
// Solidity: function batchIndex() view returns(uint64)
func (_BatchCounter *BatchCounterCallerSession) BatchIndex() (uint64, error) {
	return _BatchCounter.Contract.BatchIndex(&_BatchCounter.CallOpts)
}

// Increment is a paid mutator transaction binding the contract method 0xd09de08a.
//
// Solidity: function increment() returns()
func (_BatchCounter *BatchCounterTransactor) Increment(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchCounter.contract.Transact(opts, "increment")
}

// Increment is a paid mutator transaction binding the contract method 0xd09de08a.
//
// Solidity: function increment() returns()
func (_BatchCounter *BatchCounterSession) Increment() (*types.Transaction, error) {
	return _BatchCounter.Contract.Increment(&_BatchCounter.TransactOpts)
}

// Increment is a paid mutator transaction binding the contract method 0xd09de08a.
//
// Solidity: function increment() returns()
func (_BatchCounter *BatchCounterTransactorSession) Increment() (*types.Transaction, error) {
	return _BatchCounter.Contract.Increment(&_BatchCounter.TransactOpts)
}

// Set is a paid mutator transaction binding the contract method 0x1d8c311f.
//
// Solidity: function set(uint64 newBatchIndex) returns()
func (_BatchCounter *BatchCounterTransactor) Set(opts *bind.TransactOpts, newBatchIndex uint64) (*types.Transaction, error) {
	return _BatchCounter.contract.Transact(opts, "set", newBatchIndex)
}

// Set is a paid mutator transaction binding the contract method 0x1d8c311f.
//
// Solidity: function set(uint64 newBatchIndex) returns()
func (_BatchCounter *BatchCounterSession) Set(newBatchIndex uint64) (*types.Transaction, error) {
	return _BatchCounter.Contract.Set(&_BatchCounter.TransactOpts, newBatchIndex)
}

// Set is a paid mutator transaction binding the contract method 0x1d8c311f.
//
// Solidity: function set(uint64 newBatchIndex) returns()
func (_BatchCounter *BatchCounterTransactorSession) Set(newBatchIndex uint64) (*types.Transaction, error) {
	return _BatchCounter.Contract.Set(&_BatchCounter.TransactOpts, newBatchIndex)
}

// BatchCounterNewBatchIndexIterator is returned from FilterNewBatchIndex and is used to iterate over the raw logs and unpacked data for NewBatchIndex events raised by the BatchCounter contract.
type BatchCounterNewBatchIndexIterator struct {
	Event *BatchCounterNewBatchIndex // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BatchCounterNewBatchIndexIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchCounterNewBatchIndex)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BatchCounterNewBatchIndex)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BatchCounterNewBatchIndexIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchCounterNewBatchIndexIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchCounterNewBatchIndex represents a NewBatchIndex event raised by the BatchCounter contract.
type BatchCounterNewBatchIndex struct {
	OldIndex uint64
	NewIndex uint64
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNewBatchIndex is a free log retrieval operation binding the contract event 0x5867f9e83f14fb505a43dd58880b1de7e3b5cddbfa99bb92a15dad48b453410b.
//
// Solidity: event NewBatchIndex(uint64 oldIndex, uint64 newIndex)
func (_BatchCounter *BatchCounterFilterer) FilterNewBatchIndex(opts *bind.FilterOpts) (*BatchCounterNewBatchIndexIterator, error) {

	logs, sub, err := _BatchCounter.contract.FilterLogs(opts, "NewBatchIndex")
	if err != nil {
		return nil, err
	}
	return &BatchCounterNewBatchIndexIterator{contract: _BatchCounter.contract, event: "NewBatchIndex", logs: logs, sub: sub}, nil
}

// WatchNewBatchIndex is a free log subscription operation binding the contract event 0x5867f9e83f14fb505a43dd58880b1de7e3b5cddbfa99bb92a15dad48b453410b.
//
// Solidity: event NewBatchIndex(uint64 oldIndex, uint64 newIndex)
func (_BatchCounter *BatchCounterFilterer) WatchNewBatchIndex(opts *bind.WatchOpts, sink chan<- *BatchCounterNewBatchIndex) (event.Subscription, error) {

	logs, sub, err := _BatchCounter.contract.WatchLogs(opts, "NewBatchIndex")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchCounterNewBatchIndex)
				if err := _BatchCounter.contract.UnpackLog(event, "NewBatchIndex", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewBatchIndex is a log parse operation binding the contract event 0x5867f9e83f14fb505a43dd58880b1de7e3b5cddbfa99bb92a15dad48b453410b.
//
// Solidity: event NewBatchIndex(uint64 oldIndex, uint64 newIndex)
func (_BatchCounter *BatchCounterFilterer) ParseNewBatchIndex(log types.Log) (*BatchCounterNewBatchIndex, error) {
	event := new(BatchCounterNewBatchIndex)
	if err := _BatchCounter.contract.UnpackLog(event, "NewBatchIndex", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CollatorConfigsListMetaData contains all meta data concerning the CollatorConfigsList contract.
var CollatorConfigsListMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"_addrsSeq\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"collatorSetIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"collatorConfigIndex\",\"type\":\"uint64\"}],\"name\":\"NewConfig\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structCollatorConfig\",\"name\":\"config\",\"type\":\"tuple\"}],\"name\":\"addNewCfg\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"addrsSeq\",\"outputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"collatorConfigs\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"}],\"name\":\"getActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structCollatorConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b50604051610a0e380380610a0e83398101604081905261002e91610249565b338061005457604051631e4fbdf760e01b81525f60048201526024015b60405180910390fd5b61005d816101fa565b50600280546001600160a01b0319166001600160a01b038316908117909155604051630545a03f60e31b81525f6004820152632a2d01f890602401602060405180830381865afa1580156100b3573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906100d79190610276565b6001600160401b03161561013e5760405162461bcd60e51b815260206004820152602860248201527f4164647273536571206d757374206861766520656d707479206c697374206174604482015267020696e64657820360c41b606482015260840161004b565b6040805180820182525f808252602080830182815260018054808201825590845293517fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909401805491516001600160401b0390811668010000000000000000026001600160801b0319909316951694909417179092558251818152918201819052918101919091527ff1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d99060600160405180910390a15061029c565b5f80546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b5f60208284031215610259575f80fd5b81516001600160a01b038116811461026f575f80fd5b9392505050565b5f60208284031215610286575f80fd5b81516001600160401b038116811461026f575f80fd5b610765806102a95f395ff3fe608060405234801561000f575f80fd5b506004361061007a575f3560e01c806379f780991161005857806379f78099146100eb5780638da5cb5b146100fe578063b5351b0d1461010e578063f2fde38b14610148575f80fd5b80634d89eaaf1461007e578063715018a6146100ae57806377e18fc4146100b8575b5f80fd5b600254610091906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b6100b661015b565b005b6100cb6100c63660046105a2565b61016e565b604080516001600160401b039384168152929091166020830152016100a5565b6100b66100f93660046105b9565b6101a1565b5f546001600160a01b0316610091565b61012161011c3660046105e3565b61042f565b6040805182516001600160401b0390811682526020938401511692810192909252016100a5565b6100b6610156366004610605565b6104ea565b610163610527565b61016c5f610553565b565b6001818154811061017d575f80fd5b5f918252602090912001546001600160401b038082169250600160401b9091041682565b6101a9610527565b6101b960408201602083016105e3565b6001600160401b031660025f9054906101000a90046001600160a01b03166001600160a01b03166306661abd6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610212573d5f803e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610236919061062b565b6001600160401b0316116102b75760405162461bcd60e51b815260206004820152603a60248201527f4e6f20617070656e6465642073657420696e2073657120636f72726573706f6e60448201527f64696e6720746f20636f6e66696727732073657420696e64657800000000000060648201526084015b60405180910390fd5b6102c460208201826105e3565b6001600160401b031660018080805490506102df919061065a565b815481106102ef576102ef610673565b5f918252602090912001546001600160401b031611156103775760405162461bcd60e51b815260206004820152603860248201527f43616e6e6f7420616464206e6577207365742077697468206c6f77657220626c60448201527f6f636b206e756d626572207468616e2070726576696f7573000000000000000060648201526084016102ae565b6001805480820182555f9190915281907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6016103b38282610687565b507ff1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d990506103e460208301836105e3565b6103f460408401602085016105e3565b6001805461040291906106f3565b604080516001600160401b039485168152928416602084015292168183015290519081900360600190a150565b604080518082019091525f8082526020820152600180545f916104519161065a565b90505b826001600160401b03166001828154811061047157610471610673565b5f918252602090912001546001600160401b0316116104d8576001818154811061049d5761049d610673565b5f918252602091829020604080518082019091529101546001600160401b038082168352600160401b90910416918101919091529392505050565b806104e28161071a565b915050610454565b6104f2610527565b6001600160a01b03811661051b57604051631e4fbdf760e01b81525f60048201526024016102ae565b61052481610553565b50565b5f546001600160a01b0316331461016c5760405163118cdaa760e01b81523360048201526024016102ae565b5f80546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b5f602082840312156105b2575f80fd5b5035919050565b5f604082840312156105c9575f80fd5b50919050565b6001600160401b0381168114610524575f80fd5b5f602082840312156105f3575f80fd5b81356105fe816105cf565b9392505050565b5f60208284031215610615575f80fd5b81356001600160a01b03811681146105fe575f80fd5b5f6020828403121561063b575f80fd5b81516105fe816105cf565b634e487b7160e01b5f52601160045260245ffd5b8181038181111561066d5761066d610646565b92915050565b634e487b7160e01b5f52603260045260245ffd5b8135610692816105cf565b6001600160401b03811690508154816001600160401b0319821617835560208401356106bd816105cf565b6fffffffffffffffff00000000000000008160401b16836fffffffffffffffffffffffffffffffff198416171784555050505050565b6001600160401b0382811682821603908082111561071357610713610646565b5092915050565b5f8161072857610728610646565b505f19019056fea264697066735822122085d9e5babb5fd193a562d06e33e9444af96c8bf7627666b2095ed909445ebce164736f6c63430008160033",
}

// CollatorConfigsListABI is the input ABI used to generate the binding from.
// Deprecated: Use CollatorConfigsListMetaData.ABI instead.
var CollatorConfigsListABI = CollatorConfigsListMetaData.ABI

// CollatorConfigsListBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use CollatorConfigsListMetaData.Bin instead.
var CollatorConfigsListBin = CollatorConfigsListMetaData.Bin

// DeployCollatorConfigsList deploys a new Ethereum contract, binding an instance of CollatorConfigsList to it.
func DeployCollatorConfigsList(auth *bind.TransactOpts, backend bind.ContractBackend, _addrsSeq common.Address) (common.Address, *types.Transaction, *CollatorConfigsList, error) {
	parsed, err := CollatorConfigsListMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(CollatorConfigsListBin), backend, _addrsSeq)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CollatorConfigsList{CollatorConfigsListCaller: CollatorConfigsListCaller{contract: contract}, CollatorConfigsListTransactor: CollatorConfigsListTransactor{contract: contract}, CollatorConfigsListFilterer: CollatorConfigsListFilterer{contract: contract}}, nil
}

// CollatorConfigsList is an auto generated Go binding around an Ethereum contract.
type CollatorConfigsList struct {
	CollatorConfigsListCaller     // Read-only binding to the contract
	CollatorConfigsListTransactor // Write-only binding to the contract
	CollatorConfigsListFilterer   // Log filterer for contract events
}

// CollatorConfigsListCaller is an auto generated read-only Go binding around an Ethereum contract.
type CollatorConfigsListCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CollatorConfigsListTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CollatorConfigsListTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CollatorConfigsListFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CollatorConfigsListFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CollatorConfigsListSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CollatorConfigsListSession struct {
	Contract     *CollatorConfigsList // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// CollatorConfigsListCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CollatorConfigsListCallerSession struct {
	Contract *CollatorConfigsListCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// CollatorConfigsListTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CollatorConfigsListTransactorSession struct {
	Contract     *CollatorConfigsListTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// CollatorConfigsListRaw is an auto generated low-level Go binding around an Ethereum contract.
type CollatorConfigsListRaw struct {
	Contract *CollatorConfigsList // Generic contract binding to access the raw methods on
}

// CollatorConfigsListCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CollatorConfigsListCallerRaw struct {
	Contract *CollatorConfigsListCaller // Generic read-only contract binding to access the raw methods on
}

// CollatorConfigsListTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CollatorConfigsListTransactorRaw struct {
	Contract *CollatorConfigsListTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCollatorConfigsList creates a new instance of CollatorConfigsList, bound to a specific deployed contract.
func NewCollatorConfigsList(address common.Address, backend bind.ContractBackend) (*CollatorConfigsList, error) {
	contract, err := bindCollatorConfigsList(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CollatorConfigsList{CollatorConfigsListCaller: CollatorConfigsListCaller{contract: contract}, CollatorConfigsListTransactor: CollatorConfigsListTransactor{contract: contract}, CollatorConfigsListFilterer: CollatorConfigsListFilterer{contract: contract}}, nil
}

// NewCollatorConfigsListCaller creates a new read-only instance of CollatorConfigsList, bound to a specific deployed contract.
func NewCollatorConfigsListCaller(address common.Address, caller bind.ContractCaller) (*CollatorConfigsListCaller, error) {
	contract, err := bindCollatorConfigsList(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CollatorConfigsListCaller{contract: contract}, nil
}

// NewCollatorConfigsListTransactor creates a new write-only instance of CollatorConfigsList, bound to a specific deployed contract.
func NewCollatorConfigsListTransactor(address common.Address, transactor bind.ContractTransactor) (*CollatorConfigsListTransactor, error) {
	contract, err := bindCollatorConfigsList(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CollatorConfigsListTransactor{contract: contract}, nil
}

// NewCollatorConfigsListFilterer creates a new log filterer instance of CollatorConfigsList, bound to a specific deployed contract.
func NewCollatorConfigsListFilterer(address common.Address, filterer bind.ContractFilterer) (*CollatorConfigsListFilterer, error) {
	contract, err := bindCollatorConfigsList(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CollatorConfigsListFilterer{contract: contract}, nil
}

// bindCollatorConfigsList binds a generic wrapper to an already deployed contract.
func bindCollatorConfigsList(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CollatorConfigsListMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CollatorConfigsList *CollatorConfigsListRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CollatorConfigsList.Contract.CollatorConfigsListCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CollatorConfigsList *CollatorConfigsListRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CollatorConfigsList.Contract.CollatorConfigsListTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CollatorConfigsList *CollatorConfigsListRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CollatorConfigsList.Contract.CollatorConfigsListTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CollatorConfigsList *CollatorConfigsListCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CollatorConfigsList.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CollatorConfigsList *CollatorConfigsListTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CollatorConfigsList.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CollatorConfigsList *CollatorConfigsListTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CollatorConfigsList.Contract.contract.Transact(opts, method, params...)
}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_CollatorConfigsList *CollatorConfigsListCaller) AddrsSeq(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CollatorConfigsList.contract.Call(opts, &out, "addrsSeq")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_CollatorConfigsList *CollatorConfigsListSession) AddrsSeq() (common.Address, error) {
	return _CollatorConfigsList.Contract.AddrsSeq(&_CollatorConfigsList.CallOpts)
}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_CollatorConfigsList *CollatorConfigsListCallerSession) AddrsSeq() (common.Address, error) {
	return _CollatorConfigsList.Contract.AddrsSeq(&_CollatorConfigsList.CallOpts)
}

// CollatorConfigs is a free data retrieval call binding the contract method 0x77e18fc4.
//
// Solidity: function collatorConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex)
func (_CollatorConfigsList *CollatorConfigsListCaller) CollatorConfigs(opts *bind.CallOpts, arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}, error) {
	var out []interface{}
	err := _CollatorConfigsList.contract.Call(opts, &out, "collatorConfigs", arg0)

	outstruct := new(struct {
		ActivationBlockNumber uint64
		SetIndex              uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ActivationBlockNumber = *abi.ConvertType(out[0], new(uint64)).(*uint64)
	outstruct.SetIndex = *abi.ConvertType(out[1], new(uint64)).(*uint64)

	return *outstruct, err

}

// CollatorConfigs is a free data retrieval call binding the contract method 0x77e18fc4.
//
// Solidity: function collatorConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex)
func (_CollatorConfigsList *CollatorConfigsListSession) CollatorConfigs(arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}, error) {
	return _CollatorConfigsList.Contract.CollatorConfigs(&_CollatorConfigsList.CallOpts, arg0)
}

// CollatorConfigs is a free data retrieval call binding the contract method 0x77e18fc4.
//
// Solidity: function collatorConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex)
func (_CollatorConfigsList *CollatorConfigsListCallerSession) CollatorConfigs(arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}, error) {
	return _CollatorConfigsList.Contract.CollatorConfigs(&_CollatorConfigsList.CallOpts, arg0)
}

// GetActiveConfig is a free data retrieval call binding the contract method 0xb5351b0d.
//
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64))
func (_CollatorConfigsList *CollatorConfigsListCaller) GetActiveConfig(opts *bind.CallOpts, activationBlockNumber uint64) (CollatorConfig, error) {
	var out []interface{}
	err := _CollatorConfigsList.contract.Call(opts, &out, "getActiveConfig", activationBlockNumber)

	if err != nil {
		return *new(CollatorConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(CollatorConfig)).(*CollatorConfig)

	return out0, err

}

// GetActiveConfig is a free data retrieval call binding the contract method 0xb5351b0d.
//
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64))
func (_CollatorConfigsList *CollatorConfigsListSession) GetActiveConfig(activationBlockNumber uint64) (CollatorConfig, error) {
	return _CollatorConfigsList.Contract.GetActiveConfig(&_CollatorConfigsList.CallOpts, activationBlockNumber)
}

// GetActiveConfig is a free data retrieval call binding the contract method 0xb5351b0d.
//
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64))
func (_CollatorConfigsList *CollatorConfigsListCallerSession) GetActiveConfig(activationBlockNumber uint64) (CollatorConfig, error) {
	return _CollatorConfigsList.Contract.GetActiveConfig(&_CollatorConfigsList.CallOpts, activationBlockNumber)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CollatorConfigsList *CollatorConfigsListCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CollatorConfigsList.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CollatorConfigsList *CollatorConfigsListSession) Owner() (common.Address, error) {
	return _CollatorConfigsList.Contract.Owner(&_CollatorConfigsList.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CollatorConfigsList *CollatorConfigsListCallerSession) Owner() (common.Address, error) {
	return _CollatorConfigsList.Contract.Owner(&_CollatorConfigsList.CallOpts)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x79f78099.
//
// Solidity: function addNewCfg((uint64,uint64) config) returns()
func (_CollatorConfigsList *CollatorConfigsListTransactor) AddNewCfg(opts *bind.TransactOpts, config CollatorConfig) (*types.Transaction, error) {
	return _CollatorConfigsList.contract.Transact(opts, "addNewCfg", config)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x79f78099.
//
// Solidity: function addNewCfg((uint64,uint64) config) returns()
func (_CollatorConfigsList *CollatorConfigsListSession) AddNewCfg(config CollatorConfig) (*types.Transaction, error) {
	return _CollatorConfigsList.Contract.AddNewCfg(&_CollatorConfigsList.TransactOpts, config)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x79f78099.
//
// Solidity: function addNewCfg((uint64,uint64) config) returns()
func (_CollatorConfigsList *CollatorConfigsListTransactorSession) AddNewCfg(config CollatorConfig) (*types.Transaction, error) {
	return _CollatorConfigsList.Contract.AddNewCfg(&_CollatorConfigsList.TransactOpts, config)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CollatorConfigsList *CollatorConfigsListTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CollatorConfigsList.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CollatorConfigsList *CollatorConfigsListSession) RenounceOwnership() (*types.Transaction, error) {
	return _CollatorConfigsList.Contract.RenounceOwnership(&_CollatorConfigsList.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CollatorConfigsList *CollatorConfigsListTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _CollatorConfigsList.Contract.RenounceOwnership(&_CollatorConfigsList.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CollatorConfigsList *CollatorConfigsListTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _CollatorConfigsList.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CollatorConfigsList *CollatorConfigsListSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _CollatorConfigsList.Contract.TransferOwnership(&_CollatorConfigsList.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CollatorConfigsList *CollatorConfigsListTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _CollatorConfigsList.Contract.TransferOwnership(&_CollatorConfigsList.TransactOpts, newOwner)
}

// CollatorConfigsListNewConfigIterator is returned from FilterNewConfig and is used to iterate over the raw logs and unpacked data for NewConfig events raised by the CollatorConfigsList contract.
type CollatorConfigsListNewConfigIterator struct {
	Event *CollatorConfigsListNewConfig // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CollatorConfigsListNewConfigIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CollatorConfigsListNewConfig)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CollatorConfigsListNewConfig)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CollatorConfigsListNewConfigIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CollatorConfigsListNewConfigIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CollatorConfigsListNewConfig represents a NewConfig event raised by the CollatorConfigsList contract.
type CollatorConfigsListNewConfig struct {
	ActivationBlockNumber uint64
	CollatorSetIndex      uint64
	CollatorConfigIndex   uint64
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewConfig is a free log retrieval operation binding the contract event 0xf1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d9.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 collatorSetIndex, uint64 collatorConfigIndex)
func (_CollatorConfigsList *CollatorConfigsListFilterer) FilterNewConfig(opts *bind.FilterOpts) (*CollatorConfigsListNewConfigIterator, error) {

	logs, sub, err := _CollatorConfigsList.contract.FilterLogs(opts, "NewConfig")
	if err != nil {
		return nil, err
	}
	return &CollatorConfigsListNewConfigIterator{contract: _CollatorConfigsList.contract, event: "NewConfig", logs: logs, sub: sub}, nil
}

// WatchNewConfig is a free log subscription operation binding the contract event 0xf1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d9.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 collatorSetIndex, uint64 collatorConfigIndex)
func (_CollatorConfigsList *CollatorConfigsListFilterer) WatchNewConfig(opts *bind.WatchOpts, sink chan<- *CollatorConfigsListNewConfig) (event.Subscription, error) {

	logs, sub, err := _CollatorConfigsList.contract.WatchLogs(opts, "NewConfig")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CollatorConfigsListNewConfig)
				if err := _CollatorConfigsList.contract.UnpackLog(event, "NewConfig", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewConfig is a log parse operation binding the contract event 0xf1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d9.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 collatorSetIndex, uint64 collatorConfigIndex)
func (_CollatorConfigsList *CollatorConfigsListFilterer) ParseNewConfig(log types.Log) (*CollatorConfigsListNewConfig, error) {
	event := new(CollatorConfigsListNewConfig)
	if err := _CollatorConfigsList.contract.UnpackLog(event, "NewConfig", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CollatorConfigsListOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the CollatorConfigsList contract.
type CollatorConfigsListOwnershipTransferredIterator struct {
	Event *CollatorConfigsListOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CollatorConfigsListOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CollatorConfigsListOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CollatorConfigsListOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CollatorConfigsListOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CollatorConfigsListOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CollatorConfigsListOwnershipTransferred represents a OwnershipTransferred event raised by the CollatorConfigsList contract.
type CollatorConfigsListOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CollatorConfigsList *CollatorConfigsListFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*CollatorConfigsListOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CollatorConfigsList.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &CollatorConfigsListOwnershipTransferredIterator{contract: _CollatorConfigsList.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CollatorConfigsList *CollatorConfigsListFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *CollatorConfigsListOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CollatorConfigsList.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CollatorConfigsListOwnershipTransferred)
				if err := _CollatorConfigsList.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CollatorConfigsList *CollatorConfigsListFilterer) ParseOwnershipTransferred(log types.Log) (*CollatorConfigsListOwnershipTransferred, error) {
	event := new(CollatorConfigsListOwnershipTransferred)
	if err := _CollatorConfigsList.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContextMetaData contains all meta data concerning the Context contract.
var ContextMetaData = &bind.MetaData{
	ABI: "[]",
}

// ContextABI is the input ABI used to generate the binding from.
// Deprecated: Use ContextMetaData.ABI instead.
var ContextABI = ContextMetaData.ABI

// Context is an auto generated Go binding around an Ethereum contract.
type Context struct {
	ContextCaller     // Read-only binding to the contract
	ContextTransactor // Write-only binding to the contract
	ContextFilterer   // Log filterer for contract events
}

// ContextCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContextCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContextTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContextTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContextFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContextFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContextSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContextSession struct {
	Contract     *Context          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContextCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContextCallerSession struct {
	Contract *ContextCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// ContextTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContextTransactorSession struct {
	Contract     *ContextTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// ContextRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContextRaw struct {
	Contract *Context // Generic contract binding to access the raw methods on
}

// ContextCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContextCallerRaw struct {
	Contract *ContextCaller // Generic read-only contract binding to access the raw methods on
}

// ContextTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContextTransactorRaw struct {
	Contract *ContextTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContext creates a new instance of Context, bound to a specific deployed contract.
func NewContext(address common.Address, backend bind.ContractBackend) (*Context, error) {
	contract, err := bindContext(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Context{ContextCaller: ContextCaller{contract: contract}, ContextTransactor: ContextTransactor{contract: contract}, ContextFilterer: ContextFilterer{contract: contract}}, nil
}

// NewContextCaller creates a new read-only instance of Context, bound to a specific deployed contract.
func NewContextCaller(address common.Address, caller bind.ContractCaller) (*ContextCaller, error) {
	contract, err := bindContext(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContextCaller{contract: contract}, nil
}

// NewContextTransactor creates a new write-only instance of Context, bound to a specific deployed contract.
func NewContextTransactor(address common.Address, transactor bind.ContractTransactor) (*ContextTransactor, error) {
	contract, err := bindContext(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContextTransactor{contract: contract}, nil
}

// NewContextFilterer creates a new log filterer instance of Context, bound to a specific deployed contract.
func NewContextFilterer(address common.Address, filterer bind.ContractFilterer) (*ContextFilterer, error) {
	contract, err := bindContext(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContextFilterer{contract: contract}, nil
}

// bindContext binds a generic wrapper to an already deployed contract.
func bindContext(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContextMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Context *ContextRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Context.Contract.ContextCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Context *ContextRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Context.Contract.ContextTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Context *ContextRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Context.Contract.ContextTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Context *ContextCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Context.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Context *ContextTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Context.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Context *ContextTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Context.Contract.contract.Transact(opts, method, params...)
}

// EonKeyStorageMetaData contains all meta data concerning the EonKeyStorage contract.
var EonKeyStorageMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"blockNumber\",\"type\":\"uint64\"}],\"name\":\"NotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"key\",\"type\":\"bytes\"}],\"name\":\"Inserted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"blockNumber\",\"type\":\"uint64\"}],\"name\":\"get\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"serializedKey\",\"type\":\"bytes\"},{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"}],\"name\":\"insert\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"keys\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nextIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes\",\"name\":\"key\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"num\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801562000010575f80fd5b5033806200003757604051631e4fbdf760e01b81525f600482015260240160405180910390fd5b620000428162000070565b506060620000535f826001620000bf565b50620000686001600160401b03825f620000bf565b5050620003ac565b5f80546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60018054604080516060810182526001600160401b0380881682528581166020830190815292820187815284860186555f9586528251600286027fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6810180549651851668010000000000000000026001600160801b0319909716929094169190911794909417825551919290917fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf7909101906200017d908262000276565b5050506002816001600160401b031610620001d0577f2f64d9497c8c677c995d99bcc930463dca07bfc5906e28140cbfa4222ddf402c858286604051620001c79392919062000342565b60405180910390a15b949350505050565b634e487b7160e01b5f52604160045260245ffd5b600181811c908216806200020157607f821691505b6020821081036200022057634e487b7160e01b5f52602260045260245ffd5b50919050565b601f8211156200027157805f5260205f20601f840160051c810160208510156200024d5750805b601f840160051c820191505b818110156200026e575f815560010162000259565b50505b505050565b81516001600160401b03811115620002925762000292620001d8565b620002aa81620002a38454620001ec565b8462000226565b602080601f831160018114620002e0575f8415620002c85750858301515b5f19600386901b1c1916600185901b1785556200033a565b5f85815260208120601f198616915b828110156200031057888601518255948401946001909101908401620002ef565b50858210156200032e57878501515f19600388901b60f8161c191681555b505060018460011b0185555b505050505050565b5f60018060401b0380861683526020818616602085015260606040850152845191508160608501525f5b828110156200038a578581018201518582016080015281016200036c565b50505f608082850101526080601f19601f830116840101915050949350505050565b610bcb80620003ba5f395ff3fe608060405234801561000f575f80fd5b506004361061007a575f3560e01c8063715018a611610058578063715018a6146100de5780638da5cb5b146100e6578063ada8679814610100578063f2fde38b14610120575f80fd5b80630cb6aaf11461007e5780633f5fafa4146100a95780634e70b1dc146100be575b5f80fd5b61009161008c36600461083e565b610133565b6040516100a093929190610898565b60405180910390f35b6100bc6100b73660046108fb565b6101fd565b005b6100c66104bb565b6040516001600160401b0390911681526020016100a0565b6100bc6104d1565b5f546040516001600160a01b0390911681526020016100a0565b61011361010e3660046109b6565b6104e4565b6040516100a091906109d6565b6100bc61012e3660046109e8565b610677565b60018181548110610142575f80fd5b5f918252602090912060029091020180546001820180546001600160401b038084169550600160401b909304909216929161017c90610a0e565b80601f01602080910402602001604051908101604052809291908181526020018280546101a890610a0e565b80156101f35780601f106101ca576101008083540402835291602001916101f3565b820191905f5260205f20905b8154815290600101906020018083116101d657829003601f168201915b5050505050905083565b6102056106b4565b5f80600190505f6001826001600160401b03168154811061022857610228610a46565b5f91825260209182902060408051606081018252600290930290910180546001600160401b038082168552600160401b90910416938301939093526001830180549293929184019161027990610a0e565b80601f01602080910402602001604051908101604052809291908181526020018280546102a590610a0e565b80156102f05780601f106102c7576101008083540402835291602001916102f0565b820191905f5260205f20905b8154815290600101906020018083116102d357829003601f168201915b50505050508152505090505b5f600182602001516001600160401b03168154811061031d5761031d610a46565b5f91825260209182902060408051606081018252600290930290910180546001600160401b038082168552600160401b90910416938301939093526001830180549293929184019161036e90610a0e565b80601f016020809104026020016040519081016040528092919081815260200182805461039a90610a0e565b80156103e55780601f106103bc576101008083540402835291602001916103e5565b820191905f5260205f20905b8154815290600101906020018083116103c857829003601f168201915b5050505050815250509050846001600160401b0316815f01516001600160401b0316116104a75761041b858784602001516106e0565b6001600160401b0380821660208501526001805492965084929091861690811061044757610447610a46565b5f91825260209182902083516002909202018054928401516001600160401b03908116600160401b026001600160801b03199094169216919091179190911781556040820151600182019061049c9082610aa3565b505050505050505050565b60209091015191506102fc565b5050505050565b6001545f906104cc90600290610b62565b905090565b6104d96106b4565b6104e25f6107ef565b565b60605f600180815481106104fa576104fa610a46565b5f918252602090912060029091020154600160401b90046001600160401b031690505b6001600160401b0381161561064e575f6001826001600160401b03168154811061054957610549610a46565b5f91825260209182902060408051606081018252600290930290910180546001600160401b038082168552600160401b90910416938301939093526001830180549293929184019161059a90610a0e565b80601f01602080910402602001604051908101604052809291908181526020018280546105c690610a0e565b80156106115780601f106105e857610100808354040283529160200191610611565b820191905f5260205f20905b8154815290600101906020018083116105f457829003601f168201915b5050505050815250509050836001600160401b0316815f01516001600160401b03161161064357604001519392505050565b60200151905061051d565b604051636be0ee8760e01b81526001600160401b03841660048201526024015b60405180910390fd5b61067f6106b4565b6001600160a01b0381166106a857604051631e4fbdf760e01b81525f600482015260240161066e565b6106b1816107ef565b50565b5f546001600160a01b031633146104e25760405163118cdaa760e01b815233600482015260240161066e565b60018054604080516060810182526001600160401b0380881682528581166020830190815292820187815284860186555f9586528251600286027fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf68101805496518516600160401b026001600160801b0319909716929094169190911794909417825551919290917fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf7909101906107979082610aa3565b5050506002816001600160401b0316106107e7577f2f64d9497c8c677c995d99bcc930463dca07bfc5906e28140cbfa4222ddf402c8582866040516107de93929190610898565b60405180910390a15b949350505050565b5f80546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b5f6020828403121561084e575f80fd5b5035919050565b5f81518084525f5b818110156108795760208185018101518683018201520161085d565b505f602082860101526020601f19601f83011685010191505092915050565b5f6001600160401b038086168352808516602084015250606060408301526108c36060830184610855565b95945050505050565b634e487b7160e01b5f52604160045260245ffd5b80356001600160401b03811681146108f6575f80fd5b919050565b5f806040838503121561090c575f80fd5b82356001600160401b0380821115610922575f80fd5b818501915085601f830112610935575f80fd5b813581811115610947576109476108cc565b604051601f8201601f19908116603f0116810190838211818310171561096f5761096f6108cc565b81604052828152886020848701011115610987575f80fd5b826020860160208301375f6020848301015280965050505050506109ad602084016108e0565b90509250929050565b5f602082840312156109c6575f80fd5b6109cf826108e0565b9392505050565b602081525f6109cf6020830184610855565b5f602082840312156109f8575f80fd5b81356001600160a01b03811681146109cf575f80fd5b600181811c90821680610a2257607f821691505b602082108103610a4057634e487b7160e01b5f52602260045260245ffd5b50919050565b634e487b7160e01b5f52603260045260245ffd5b601f821115610a9e57805f5260205f20601f840160051c81016020851015610a7f5750805b601f840160051c820191505b818110156104b4575f8155600101610a8b565b505050565b81516001600160401b03811115610abc57610abc6108cc565b610ad081610aca8454610a0e565b84610a5a565b602080601f831160018114610b03575f8415610aec5750858301515b5f19600386901b1c1916600185901b178555610b5a565b5f85815260208120601f198616915b82811015610b3157888601518255948401946001909101908401610b12565b5085821015610b4e57878501515f19600388901b60f8161c191681555b505060018460011b0185555b505050505050565b6001600160401b03828116828216039080821115610b8e57634e487b7160e01b5f52601160045260245ffd5b509291505056fea26469706673582212207edcfddce17f42840c2f77524c164cadb53c4eb4ec1ed057375111c2c0ca7da964736f6c63430008160033",
}

// EonKeyStorageABI is the input ABI used to generate the binding from.
// Deprecated: Use EonKeyStorageMetaData.ABI instead.
var EonKeyStorageABI = EonKeyStorageMetaData.ABI

// EonKeyStorageBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use EonKeyStorageMetaData.Bin instead.
var EonKeyStorageBin = EonKeyStorageMetaData.Bin

// DeployEonKeyStorage deploys a new Ethereum contract, binding an instance of EonKeyStorage to it.
func DeployEonKeyStorage(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *EonKeyStorage, error) {
	parsed, err := EonKeyStorageMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(EonKeyStorageBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &EonKeyStorage{EonKeyStorageCaller: EonKeyStorageCaller{contract: contract}, EonKeyStorageTransactor: EonKeyStorageTransactor{contract: contract}, EonKeyStorageFilterer: EonKeyStorageFilterer{contract: contract}}, nil
}

// EonKeyStorage is an auto generated Go binding around an Ethereum contract.
type EonKeyStorage struct {
	EonKeyStorageCaller     // Read-only binding to the contract
	EonKeyStorageTransactor // Write-only binding to the contract
	EonKeyStorageFilterer   // Log filterer for contract events
}

// EonKeyStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type EonKeyStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EonKeyStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EonKeyStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EonKeyStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EonKeyStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EonKeyStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EonKeyStorageSession struct {
	Contract     *EonKeyStorage    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EonKeyStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EonKeyStorageCallerSession struct {
	Contract *EonKeyStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// EonKeyStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EonKeyStorageTransactorSession struct {
	Contract     *EonKeyStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// EonKeyStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type EonKeyStorageRaw struct {
	Contract *EonKeyStorage // Generic contract binding to access the raw methods on
}

// EonKeyStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EonKeyStorageCallerRaw struct {
	Contract *EonKeyStorageCaller // Generic read-only contract binding to access the raw methods on
}

// EonKeyStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EonKeyStorageTransactorRaw struct {
	Contract *EonKeyStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEonKeyStorage creates a new instance of EonKeyStorage, bound to a specific deployed contract.
func NewEonKeyStorage(address common.Address, backend bind.ContractBackend) (*EonKeyStorage, error) {
	contract, err := bindEonKeyStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EonKeyStorage{EonKeyStorageCaller: EonKeyStorageCaller{contract: contract}, EonKeyStorageTransactor: EonKeyStorageTransactor{contract: contract}, EonKeyStorageFilterer: EonKeyStorageFilterer{contract: contract}}, nil
}

// NewEonKeyStorageCaller creates a new read-only instance of EonKeyStorage, bound to a specific deployed contract.
func NewEonKeyStorageCaller(address common.Address, caller bind.ContractCaller) (*EonKeyStorageCaller, error) {
	contract, err := bindEonKeyStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EonKeyStorageCaller{contract: contract}, nil
}

// NewEonKeyStorageTransactor creates a new write-only instance of EonKeyStorage, bound to a specific deployed contract.
func NewEonKeyStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*EonKeyStorageTransactor, error) {
	contract, err := bindEonKeyStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EonKeyStorageTransactor{contract: contract}, nil
}

// NewEonKeyStorageFilterer creates a new log filterer instance of EonKeyStorage, bound to a specific deployed contract.
func NewEonKeyStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*EonKeyStorageFilterer, error) {
	contract, err := bindEonKeyStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EonKeyStorageFilterer{contract: contract}, nil
}

// bindEonKeyStorage binds a generic wrapper to an already deployed contract.
func bindEonKeyStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := EonKeyStorageMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EonKeyStorage *EonKeyStorageRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _EonKeyStorage.Contract.EonKeyStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EonKeyStorage *EonKeyStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EonKeyStorage.Contract.EonKeyStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EonKeyStorage *EonKeyStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EonKeyStorage.Contract.EonKeyStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EonKeyStorage *EonKeyStorageCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _EonKeyStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EonKeyStorage *EonKeyStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EonKeyStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EonKeyStorage *EonKeyStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EonKeyStorage.Contract.contract.Transact(opts, method, params...)
}

// Get is a free data retrieval call binding the contract method 0xada86798.
//
// Solidity: function get(uint64 blockNumber) view returns(bytes)
func (_EonKeyStorage *EonKeyStorageCaller) Get(opts *bind.CallOpts, blockNumber uint64) ([]byte, error) {
	var out []interface{}
	err := _EonKeyStorage.contract.Call(opts, &out, "get", blockNumber)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// Get is a free data retrieval call binding the contract method 0xada86798.
//
// Solidity: function get(uint64 blockNumber) view returns(bytes)
func (_EonKeyStorage *EonKeyStorageSession) Get(blockNumber uint64) ([]byte, error) {
	return _EonKeyStorage.Contract.Get(&_EonKeyStorage.CallOpts, blockNumber)
}

// Get is a free data retrieval call binding the contract method 0xada86798.
//
// Solidity: function get(uint64 blockNumber) view returns(bytes)
func (_EonKeyStorage *EonKeyStorageCallerSession) Get(blockNumber uint64) ([]byte, error) {
	return _EonKeyStorage.Contract.Get(&_EonKeyStorage.CallOpts, blockNumber)
}

// Keys is a free data retrieval call binding the contract method 0x0cb6aaf1.
//
// Solidity: function keys(uint256 ) view returns(uint64 activationBlockNumber, uint64 nextIndex, bytes key)
func (_EonKeyStorage *EonKeyStorageCaller) Keys(opts *bind.CallOpts, arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	NextIndex             uint64
	Key                   []byte
}, error) {
	var out []interface{}
	err := _EonKeyStorage.contract.Call(opts, &out, "keys", arg0)

	outstruct := new(struct {
		ActivationBlockNumber uint64
		NextIndex             uint64
		Key                   []byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ActivationBlockNumber = *abi.ConvertType(out[0], new(uint64)).(*uint64)
	outstruct.NextIndex = *abi.ConvertType(out[1], new(uint64)).(*uint64)
	outstruct.Key = *abi.ConvertType(out[2], new([]byte)).(*[]byte)

	return *outstruct, err

}

// Keys is a free data retrieval call binding the contract method 0x0cb6aaf1.
//
// Solidity: function keys(uint256 ) view returns(uint64 activationBlockNumber, uint64 nextIndex, bytes key)
func (_EonKeyStorage *EonKeyStorageSession) Keys(arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	NextIndex             uint64
	Key                   []byte
}, error) {
	return _EonKeyStorage.Contract.Keys(&_EonKeyStorage.CallOpts, arg0)
}

// Keys is a free data retrieval call binding the contract method 0x0cb6aaf1.
//
// Solidity: function keys(uint256 ) view returns(uint64 activationBlockNumber, uint64 nextIndex, bytes key)
func (_EonKeyStorage *EonKeyStorageCallerSession) Keys(arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	NextIndex             uint64
	Key                   []byte
}, error) {
	return _EonKeyStorage.Contract.Keys(&_EonKeyStorage.CallOpts, arg0)
}

// Num is a free data retrieval call binding the contract method 0x4e70b1dc.
//
// Solidity: function num() view returns(uint64)
func (_EonKeyStorage *EonKeyStorageCaller) Num(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _EonKeyStorage.contract.Call(opts, &out, "num")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// Num is a free data retrieval call binding the contract method 0x4e70b1dc.
//
// Solidity: function num() view returns(uint64)
func (_EonKeyStorage *EonKeyStorageSession) Num() (uint64, error) {
	return _EonKeyStorage.Contract.Num(&_EonKeyStorage.CallOpts)
}

// Num is a free data retrieval call binding the contract method 0x4e70b1dc.
//
// Solidity: function num() view returns(uint64)
func (_EonKeyStorage *EonKeyStorageCallerSession) Num() (uint64, error) {
	return _EonKeyStorage.Contract.Num(&_EonKeyStorage.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_EonKeyStorage *EonKeyStorageCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _EonKeyStorage.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_EonKeyStorage *EonKeyStorageSession) Owner() (common.Address, error) {
	return _EonKeyStorage.Contract.Owner(&_EonKeyStorage.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_EonKeyStorage *EonKeyStorageCallerSession) Owner() (common.Address, error) {
	return _EonKeyStorage.Contract.Owner(&_EonKeyStorage.CallOpts)
}

// Insert is a paid mutator transaction binding the contract method 0x3f5fafa4.
//
// Solidity: function insert(bytes serializedKey, uint64 activationBlockNumber) returns()
func (_EonKeyStorage *EonKeyStorageTransactor) Insert(opts *bind.TransactOpts, serializedKey []byte, activationBlockNumber uint64) (*types.Transaction, error) {
	return _EonKeyStorage.contract.Transact(opts, "insert", serializedKey, activationBlockNumber)
}

// Insert is a paid mutator transaction binding the contract method 0x3f5fafa4.
//
// Solidity: function insert(bytes serializedKey, uint64 activationBlockNumber) returns()
func (_EonKeyStorage *EonKeyStorageSession) Insert(serializedKey []byte, activationBlockNumber uint64) (*types.Transaction, error) {
	return _EonKeyStorage.Contract.Insert(&_EonKeyStorage.TransactOpts, serializedKey, activationBlockNumber)
}

// Insert is a paid mutator transaction binding the contract method 0x3f5fafa4.
//
// Solidity: function insert(bytes serializedKey, uint64 activationBlockNumber) returns()
func (_EonKeyStorage *EonKeyStorageTransactorSession) Insert(serializedKey []byte, activationBlockNumber uint64) (*types.Transaction, error) {
	return _EonKeyStorage.Contract.Insert(&_EonKeyStorage.TransactOpts, serializedKey, activationBlockNumber)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EonKeyStorage *EonKeyStorageTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EonKeyStorage.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EonKeyStorage *EonKeyStorageSession) RenounceOwnership() (*types.Transaction, error) {
	return _EonKeyStorage.Contract.RenounceOwnership(&_EonKeyStorage.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EonKeyStorage *EonKeyStorageTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _EonKeyStorage.Contract.RenounceOwnership(&_EonKeyStorage.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_EonKeyStorage *EonKeyStorageTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _EonKeyStorage.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_EonKeyStorage *EonKeyStorageSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EonKeyStorage.Contract.TransferOwnership(&_EonKeyStorage.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_EonKeyStorage *EonKeyStorageTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EonKeyStorage.Contract.TransferOwnership(&_EonKeyStorage.TransactOpts, newOwner)
}

// EonKeyStorageInsertedIterator is returned from FilterInserted and is used to iterate over the raw logs and unpacked data for Inserted events raised by the EonKeyStorage contract.
type EonKeyStorageInsertedIterator struct {
	Event *EonKeyStorageInserted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EonKeyStorageInsertedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EonKeyStorageInserted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EonKeyStorageInserted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EonKeyStorageInsertedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EonKeyStorageInsertedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EonKeyStorageInserted represents a Inserted event raised by the EonKeyStorage contract.
type EonKeyStorageInserted struct {
	ActivationBlockNumber uint64
	Index                 uint64
	Key                   []byte
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterInserted is a free log retrieval operation binding the contract event 0x2f64d9497c8c677c995d99bcc930463dca07bfc5906e28140cbfa4222ddf402c.
//
// Solidity: event Inserted(uint64 activationBlockNumber, uint64 index, bytes key)
func (_EonKeyStorage *EonKeyStorageFilterer) FilterInserted(opts *bind.FilterOpts) (*EonKeyStorageInsertedIterator, error) {

	logs, sub, err := _EonKeyStorage.contract.FilterLogs(opts, "Inserted")
	if err != nil {
		return nil, err
	}
	return &EonKeyStorageInsertedIterator{contract: _EonKeyStorage.contract, event: "Inserted", logs: logs, sub: sub}, nil
}

// WatchInserted is a free log subscription operation binding the contract event 0x2f64d9497c8c677c995d99bcc930463dca07bfc5906e28140cbfa4222ddf402c.
//
// Solidity: event Inserted(uint64 activationBlockNumber, uint64 index, bytes key)
func (_EonKeyStorage *EonKeyStorageFilterer) WatchInserted(opts *bind.WatchOpts, sink chan<- *EonKeyStorageInserted) (event.Subscription, error) {

	logs, sub, err := _EonKeyStorage.contract.WatchLogs(opts, "Inserted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EonKeyStorageInserted)
				if err := _EonKeyStorage.contract.UnpackLog(event, "Inserted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInserted is a log parse operation binding the contract event 0x2f64d9497c8c677c995d99bcc930463dca07bfc5906e28140cbfa4222ddf402c.
//
// Solidity: event Inserted(uint64 activationBlockNumber, uint64 index, bytes key)
func (_EonKeyStorage *EonKeyStorageFilterer) ParseInserted(log types.Log) (*EonKeyStorageInserted, error) {
	event := new(EonKeyStorageInserted)
	if err := _EonKeyStorage.contract.UnpackLog(event, "Inserted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EonKeyStorageOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the EonKeyStorage contract.
type EonKeyStorageOwnershipTransferredIterator struct {
	Event *EonKeyStorageOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EonKeyStorageOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EonKeyStorageOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EonKeyStorageOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EonKeyStorageOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EonKeyStorageOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EonKeyStorageOwnershipTransferred represents a OwnershipTransferred event raised by the EonKeyStorage contract.
type EonKeyStorageOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_EonKeyStorage *EonKeyStorageFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*EonKeyStorageOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EonKeyStorage.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &EonKeyStorageOwnershipTransferredIterator{contract: _EonKeyStorage.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_EonKeyStorage *EonKeyStorageFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *EonKeyStorageOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EonKeyStorage.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EonKeyStorageOwnershipTransferred)
				if err := _EonKeyStorage.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_EonKeyStorage *EonKeyStorageFilterer) ParseOwnershipTransferred(log types.Log) (*EonKeyStorageOwnershipTransferred, error) {
	event := new(EonKeyStorageOwnershipTransferred)
	if err := _EonKeyStorage.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeypersConfigsListMetaData contains all meta data concerning the KeypersConfigsList contract.
var KeypersConfigsListMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"_addrsSeq\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"keyperConfigIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"name\":\"NewConfig\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"internalType\":\"structKeypersConfig\",\"name\":\"config\",\"type\":\"tuple\"}],\"name\":\"addNewCfg\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"addrsSeq\",\"outputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"}],\"name\":\"getActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"internalType\":\"structKeypersConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"keypersConfigs\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b50604051610cc7380380610cc783398101604081905261002e91610274565b338061005457604051631e4fbdf760e01b81525f60048201526024015b60405180910390fd5b61005d81610225565b50600280546001600160a01b0319166001600160a01b038316908117909155604051630545a03f60e31b81525f6004820152632a2d01f890602401602060405180830381865afa1580156100b3573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906100d791906102a1565b6001600160401b03161561013e5760405162461bcd60e51b815260206004820152602860248201527f4164647273536571206d757374206861766520656d707479206c697374206174604482015267020696e64657820360c41b606482015260840161004b565b60408051606080820183525f808352602080840182815284860183815260018054808201825590855295517fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf69096018054925191516001600160401b03908116600160801b02600160801b600160c01b031993821668010000000000000000026001600160801b0319909516919098161792909217169490941790935583518181529283018190529282018390528101919091527f97f0f7a37d08d48af6a5f7140aedcc4fa60e92ee1d0607f2d4868c8fc518cc0e9060800160405180910390a1506102c7565b5f80546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b5f60208284031215610284575f80fd5b81516001600160a01b038116811461029a575f80fd5b9392505050565b5f602082840312156102b1575f80fd5b81516001600160401b038116811461029a575f80fd5b6109f3806102d45f395ff3fe608060405234801561000f575f80fd5b506004361061007a575f3560e01c80638da5cb5b116100585780638da5cb5b146100cb578063b5351b0d146100db578063f2fde38b14610120578063fc6d0c7e14610133575f80fd5b806333419af51461007e5780634d89eaaf14610093578063715018a6146100c3575b5f80fd5b61009161008c3660046107fb565b610170565b005b6002546100a6906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b610091610620565b5f546001600160a01b03166100a6565b6100ee6100e9366004610825565b610633565b6040805182516001600160401b03908116825260208085015182169083015292820151909216908201526060016100ba565b61009161012e366004610847565b610707565b61014661014136600461086d565b610744565b604080516001600160401b03948516815292841660208401529216918101919091526060016100ba565b610178610780565b6101886040820160208301610825565b6001600160401b031660025f9054906101000a90046001600160a01b03166001600160a01b03166306661abd6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156101e1573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102059190610884565b6001600160401b0316116102865760405162461bcd60e51b815260206004820152603a60248201527f4e6f20617070656e6465642073657420696e2073657120636f72726573706f6e60448201527f64696e6720746f20636f6e66696727732073657420696e64657800000000000060648201526084015b60405180910390fd5b6102936020820182610825565b6001600160401b031660018080805490506102ae91906108b3565b815481106102be576102be6108cc565b5f918252602090912001546001600160401b031611156103465760405162461bcd60e51b815260206004820152603860248201527f43616e6e6f7420616464206e6577207365742077697468206c6f77657220626c60448201527f6f636b206e756d626572207468616e2070726576696f75730000000000000000606482015260840161027d565b6002545f906001600160a01b0316632a2d01f86103696040850160208601610825565b6040516001600160e01b031960e084901b1681526001600160401b039091166004820152602401602060405180830381865afa1580156103ab573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906103cf9190610884565b9050806001600160401b03165f03610462576103f16060830160408401610825565b6001600160401b03161561045d5760405162461bcd60e51b815260206004820152602d60248201527f5468726573686f6c64206d757374206265207a65726f206966206b657970657260448201526c2073657420697320656d70747960981b606482015260840161027d565b61054e565b60016104746060840160408501610825565b6001600160401b031610156104cb5760405162461bcd60e51b815260206004820152601e60248201527f5468726573686f6c64206d757374206265206174206c65617374206f6e650000604482015260640161027d565b6001600160401b0381166104e56060840160408501610825565b6001600160401b0316111561054e5760405162461bcd60e51b815260206004820152602960248201527f5468726573686f6c64206d757374206e6f7420657863656564206b6579706572604482015268207365742073697a6560b81b606482015260840161027d565b6001805480820182555f9190915282907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf60161058a82826108e0565b507f97f0f7a37d08d48af6a5f7140aedcc4fa60e92ee1d0607f2d4868c8fc518cc0e90506105bb6020840184610825565b6105cb6040850160208601610825565b600180546105d99190610981565b6105e96060870160408801610825565b604080516001600160401b039586168152938516602085015291841683830152909216606082015290519081900360800190a15050565b610628610780565b6106315f6107ac565b565b604080516060810182525f8082526020820181905291810191909152600180545f9161065e916108b3565b90505b826001600160401b03166001828154811061067e5761067e6108cc565b5f918252602090912001546001600160401b0316116106f557600181815481106106aa576106aa6108cc565b5f9182526020918290206040805160608101825291909201546001600160401b038082168352600160401b8204811694830194909452600160801b9004909216908201529392505050565b806106ff816109a8565b915050610661565b61070f610780565b6001600160a01b03811661073857604051631e4fbdf760e01b81525f600482015260240161027d565b610741816107ac565b50565b60018181548110610753575f80fd5b5f918252602090912001546001600160401b038082169250600160401b8204811691600160801b90041683565b5f546001600160a01b031633146106315760405163118cdaa760e01b815233600482015260240161027d565b5f80546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b5f6060828403121561080b575f80fd5b50919050565b6001600160401b0381168114610741575f80fd5b5f60208284031215610835575f80fd5b813561084081610811565b9392505050565b5f60208284031215610857575f80fd5b81356001600160a01b0381168114610840575f80fd5b5f6020828403121561087d575f80fd5b5035919050565b5f60208284031215610894575f80fd5b815161084081610811565b634e487b7160e01b5f52601160045260245ffd5b818103818111156108c6576108c661089f565b92915050565b634e487b7160e01b5f52603260045260245ffd5b81356108eb81610811565b6001600160401b03811690508154816001600160401b03198216178355602084013561091681610811565b6fffffffffffffffff0000000000000000604091821b166fffffffffffffffffffffffffffffffff1983168417811785559085013561095481610811565b6001600160c01b0319929092169092179190911760809190911b67ffffffffffffffff60801b1617905550565b6001600160401b038281168282160390808211156109a1576109a161089f565b5092915050565b5f816109b6576109b661089f565b505f19019056fea2646970667358221220a196eb4af3420aaa13d98cc5e94bdc591c61a98781f7bed9b465fcec75b62a5764736f6c63430008160033",
}

// KeypersConfigsListABI is the input ABI used to generate the binding from.
// Deprecated: Use KeypersConfigsListMetaData.ABI instead.
var KeypersConfigsListABI = KeypersConfigsListMetaData.ABI

// KeypersConfigsListBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use KeypersConfigsListMetaData.Bin instead.
var KeypersConfigsListBin = KeypersConfigsListMetaData.Bin

// DeployKeypersConfigsList deploys a new Ethereum contract, binding an instance of KeypersConfigsList to it.
func DeployKeypersConfigsList(auth *bind.TransactOpts, backend bind.ContractBackend, _addrsSeq common.Address) (common.Address, *types.Transaction, *KeypersConfigsList, error) {
	parsed, err := KeypersConfigsListMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(KeypersConfigsListBin), backend, _addrsSeq)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &KeypersConfigsList{KeypersConfigsListCaller: KeypersConfigsListCaller{contract: contract}, KeypersConfigsListTransactor: KeypersConfigsListTransactor{contract: contract}, KeypersConfigsListFilterer: KeypersConfigsListFilterer{contract: contract}}, nil
}

// KeypersConfigsList is an auto generated Go binding around an Ethereum contract.
type KeypersConfigsList struct {
	KeypersConfigsListCaller     // Read-only binding to the contract
	KeypersConfigsListTransactor // Write-only binding to the contract
	KeypersConfigsListFilterer   // Log filterer for contract events
}

// KeypersConfigsListCaller is an auto generated read-only Go binding around an Ethereum contract.
type KeypersConfigsListCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeypersConfigsListTransactor is an auto generated write-only Go binding around an Ethereum contract.
type KeypersConfigsListTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeypersConfigsListFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type KeypersConfigsListFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeypersConfigsListSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type KeypersConfigsListSession struct {
	Contract     *KeypersConfigsList // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// KeypersConfigsListCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type KeypersConfigsListCallerSession struct {
	Contract *KeypersConfigsListCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// KeypersConfigsListTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type KeypersConfigsListTransactorSession struct {
	Contract     *KeypersConfigsListTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// KeypersConfigsListRaw is an auto generated low-level Go binding around an Ethereum contract.
type KeypersConfigsListRaw struct {
	Contract *KeypersConfigsList // Generic contract binding to access the raw methods on
}

// KeypersConfigsListCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type KeypersConfigsListCallerRaw struct {
	Contract *KeypersConfigsListCaller // Generic read-only contract binding to access the raw methods on
}

// KeypersConfigsListTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type KeypersConfigsListTransactorRaw struct {
	Contract *KeypersConfigsListTransactor // Generic write-only contract binding to access the raw methods on
}

// NewKeypersConfigsList creates a new instance of KeypersConfigsList, bound to a specific deployed contract.
func NewKeypersConfigsList(address common.Address, backend bind.ContractBackend) (*KeypersConfigsList, error) {
	contract, err := bindKeypersConfigsList(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &KeypersConfigsList{KeypersConfigsListCaller: KeypersConfigsListCaller{contract: contract}, KeypersConfigsListTransactor: KeypersConfigsListTransactor{contract: contract}, KeypersConfigsListFilterer: KeypersConfigsListFilterer{contract: contract}}, nil
}

// NewKeypersConfigsListCaller creates a new read-only instance of KeypersConfigsList, bound to a specific deployed contract.
func NewKeypersConfigsListCaller(address common.Address, caller bind.ContractCaller) (*KeypersConfigsListCaller, error) {
	contract, err := bindKeypersConfigsList(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &KeypersConfigsListCaller{contract: contract}, nil
}

// NewKeypersConfigsListTransactor creates a new write-only instance of KeypersConfigsList, bound to a specific deployed contract.
func NewKeypersConfigsListTransactor(address common.Address, transactor bind.ContractTransactor) (*KeypersConfigsListTransactor, error) {
	contract, err := bindKeypersConfigsList(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &KeypersConfigsListTransactor{contract: contract}, nil
}

// NewKeypersConfigsListFilterer creates a new log filterer instance of KeypersConfigsList, bound to a specific deployed contract.
func NewKeypersConfigsListFilterer(address common.Address, filterer bind.ContractFilterer) (*KeypersConfigsListFilterer, error) {
	contract, err := bindKeypersConfigsList(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &KeypersConfigsListFilterer{contract: contract}, nil
}

// bindKeypersConfigsList binds a generic wrapper to an already deployed contract.
func bindKeypersConfigsList(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := KeypersConfigsListMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KeypersConfigsList *KeypersConfigsListRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _KeypersConfigsList.Contract.KeypersConfigsListCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KeypersConfigsList *KeypersConfigsListRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.KeypersConfigsListTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KeypersConfigsList *KeypersConfigsListRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.KeypersConfigsListTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KeypersConfigsList *KeypersConfigsListCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _KeypersConfigsList.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KeypersConfigsList *KeypersConfigsListTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KeypersConfigsList *KeypersConfigsListTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.contract.Transact(opts, method, params...)
}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_KeypersConfigsList *KeypersConfigsListCaller) AddrsSeq(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _KeypersConfigsList.contract.Call(opts, &out, "addrsSeq")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_KeypersConfigsList *KeypersConfigsListSession) AddrsSeq() (common.Address, error) {
	return _KeypersConfigsList.Contract.AddrsSeq(&_KeypersConfigsList.CallOpts)
}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_KeypersConfigsList *KeypersConfigsListCallerSession) AddrsSeq() (common.Address, error) {
	return _KeypersConfigsList.Contract.AddrsSeq(&_KeypersConfigsList.CallOpts)
}

// GetActiveConfig is a free data retrieval call binding the contract method 0xb5351b0d.
//
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64,uint64))
func (_KeypersConfigsList *KeypersConfigsListCaller) GetActiveConfig(opts *bind.CallOpts, activationBlockNumber uint64) (KeypersConfig, error) {
	var out []interface{}
	err := _KeypersConfigsList.contract.Call(opts, &out, "getActiveConfig", activationBlockNumber)

	if err != nil {
		return *new(KeypersConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(KeypersConfig)).(*KeypersConfig)

	return out0, err

}

// GetActiveConfig is a free data retrieval call binding the contract method 0xb5351b0d.
//
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64,uint64))
func (_KeypersConfigsList *KeypersConfigsListSession) GetActiveConfig(activationBlockNumber uint64) (KeypersConfig, error) {
	return _KeypersConfigsList.Contract.GetActiveConfig(&_KeypersConfigsList.CallOpts, activationBlockNumber)
}

// GetActiveConfig is a free data retrieval call binding the contract method 0xb5351b0d.
//
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64,uint64))
func (_KeypersConfigsList *KeypersConfigsListCallerSession) GetActiveConfig(activationBlockNumber uint64) (KeypersConfig, error) {
	return _KeypersConfigsList.Contract.GetActiveConfig(&_KeypersConfigsList.CallOpts, activationBlockNumber)
}

// KeypersConfigs is a free data retrieval call binding the contract method 0xfc6d0c7e.
//
// Solidity: function keypersConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex, uint64 threshold)
func (_KeypersConfigsList *KeypersConfigsListCaller) KeypersConfigs(opts *bind.CallOpts, arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
	Threshold             uint64
}, error) {
	var out []interface{}
	err := _KeypersConfigsList.contract.Call(opts, &out, "keypersConfigs", arg0)

	outstruct := new(struct {
		ActivationBlockNumber uint64
		SetIndex              uint64
		Threshold             uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ActivationBlockNumber = *abi.ConvertType(out[0], new(uint64)).(*uint64)
	outstruct.SetIndex = *abi.ConvertType(out[1], new(uint64)).(*uint64)
	outstruct.Threshold = *abi.ConvertType(out[2], new(uint64)).(*uint64)

	return *outstruct, err

}

// KeypersConfigs is a free data retrieval call binding the contract method 0xfc6d0c7e.
//
// Solidity: function keypersConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex, uint64 threshold)
func (_KeypersConfigsList *KeypersConfigsListSession) KeypersConfigs(arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
	Threshold             uint64
}, error) {
	return _KeypersConfigsList.Contract.KeypersConfigs(&_KeypersConfigsList.CallOpts, arg0)
}

// KeypersConfigs is a free data retrieval call binding the contract method 0xfc6d0c7e.
//
// Solidity: function keypersConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex, uint64 threshold)
func (_KeypersConfigsList *KeypersConfigsListCallerSession) KeypersConfigs(arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
	Threshold             uint64
}, error) {
	return _KeypersConfigsList.Contract.KeypersConfigs(&_KeypersConfigsList.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeypersConfigsList *KeypersConfigsListCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _KeypersConfigsList.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeypersConfigsList *KeypersConfigsListSession) Owner() (common.Address, error) {
	return _KeypersConfigsList.Contract.Owner(&_KeypersConfigsList.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeypersConfigsList *KeypersConfigsListCallerSession) Owner() (common.Address, error) {
	return _KeypersConfigsList.Contract.Owner(&_KeypersConfigsList.CallOpts)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x33419af5.
//
// Solidity: function addNewCfg((uint64,uint64,uint64) config) returns()
func (_KeypersConfigsList *KeypersConfigsListTransactor) AddNewCfg(opts *bind.TransactOpts, config KeypersConfig) (*types.Transaction, error) {
	return _KeypersConfigsList.contract.Transact(opts, "addNewCfg", config)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x33419af5.
//
// Solidity: function addNewCfg((uint64,uint64,uint64) config) returns()
func (_KeypersConfigsList *KeypersConfigsListSession) AddNewCfg(config KeypersConfig) (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.AddNewCfg(&_KeypersConfigsList.TransactOpts, config)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x33419af5.
//
// Solidity: function addNewCfg((uint64,uint64,uint64) config) returns()
func (_KeypersConfigsList *KeypersConfigsListTransactorSession) AddNewCfg(config KeypersConfig) (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.AddNewCfg(&_KeypersConfigsList.TransactOpts, config)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_KeypersConfigsList *KeypersConfigsListTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeypersConfigsList.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_KeypersConfigsList *KeypersConfigsListSession) RenounceOwnership() (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.RenounceOwnership(&_KeypersConfigsList.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_KeypersConfigsList *KeypersConfigsListTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.RenounceOwnership(&_KeypersConfigsList.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_KeypersConfigsList *KeypersConfigsListTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _KeypersConfigsList.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_KeypersConfigsList *KeypersConfigsListSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.TransferOwnership(&_KeypersConfigsList.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_KeypersConfigsList *KeypersConfigsListTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.TransferOwnership(&_KeypersConfigsList.TransactOpts, newOwner)
}

// KeypersConfigsListNewConfigIterator is returned from FilterNewConfig and is used to iterate over the raw logs and unpacked data for NewConfig events raised by the KeypersConfigsList contract.
type KeypersConfigsListNewConfigIterator struct {
	Event *KeypersConfigsListNewConfig // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeypersConfigsListNewConfigIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeypersConfigsListNewConfig)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeypersConfigsListNewConfig)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeypersConfigsListNewConfigIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeypersConfigsListNewConfigIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeypersConfigsListNewConfig represents a NewConfig event raised by the KeypersConfigsList contract.
type KeypersConfigsListNewConfig struct {
	ActivationBlockNumber uint64
	KeyperSetIndex        uint64
	KeyperConfigIndex     uint64
	Threshold             uint64
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewConfig is a free log retrieval operation binding the contract event 0x97f0f7a37d08d48af6a5f7140aedcc4fa60e92ee1d0607f2d4868c8fc518cc0e.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 keyperSetIndex, uint64 keyperConfigIndex, uint64 threshold)
func (_KeypersConfigsList *KeypersConfigsListFilterer) FilterNewConfig(opts *bind.FilterOpts) (*KeypersConfigsListNewConfigIterator, error) {

	logs, sub, err := _KeypersConfigsList.contract.FilterLogs(opts, "NewConfig")
	if err != nil {
		return nil, err
	}
	return &KeypersConfigsListNewConfigIterator{contract: _KeypersConfigsList.contract, event: "NewConfig", logs: logs, sub: sub}, nil
}

// WatchNewConfig is a free log subscription operation binding the contract event 0x97f0f7a37d08d48af6a5f7140aedcc4fa60e92ee1d0607f2d4868c8fc518cc0e.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 keyperSetIndex, uint64 keyperConfigIndex, uint64 threshold)
func (_KeypersConfigsList *KeypersConfigsListFilterer) WatchNewConfig(opts *bind.WatchOpts, sink chan<- *KeypersConfigsListNewConfig) (event.Subscription, error) {

	logs, sub, err := _KeypersConfigsList.contract.WatchLogs(opts, "NewConfig")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeypersConfigsListNewConfig)
				if err := _KeypersConfigsList.contract.UnpackLog(event, "NewConfig", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewConfig is a log parse operation binding the contract event 0x97f0f7a37d08d48af6a5f7140aedcc4fa60e92ee1d0607f2d4868c8fc518cc0e.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 keyperSetIndex, uint64 keyperConfigIndex, uint64 threshold)
func (_KeypersConfigsList *KeypersConfigsListFilterer) ParseNewConfig(log types.Log) (*KeypersConfigsListNewConfig, error) {
	event := new(KeypersConfigsListNewConfig)
	if err := _KeypersConfigsList.contract.UnpackLog(event, "NewConfig", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeypersConfigsListOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the KeypersConfigsList contract.
type KeypersConfigsListOwnershipTransferredIterator struct {
	Event *KeypersConfigsListOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *KeypersConfigsListOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeypersConfigsListOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(KeypersConfigsListOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *KeypersConfigsListOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeypersConfigsListOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeypersConfigsListOwnershipTransferred represents a OwnershipTransferred event raised by the KeypersConfigsList contract.
type KeypersConfigsListOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_KeypersConfigsList *KeypersConfigsListFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*KeypersConfigsListOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _KeypersConfigsList.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &KeypersConfigsListOwnershipTransferredIterator{contract: _KeypersConfigsList.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_KeypersConfigsList *KeypersConfigsListFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *KeypersConfigsListOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _KeypersConfigsList.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeypersConfigsListOwnershipTransferred)
				if err := _KeypersConfigsList.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_KeypersConfigsList *KeypersConfigsListFilterer) ParseOwnershipTransferred(log types.Log) (*KeypersConfigsListOwnershipTransferred, error) {
	event := new(KeypersConfigsListOwnershipTransferred)
	if err := _KeypersConfigsList.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OwnableMetaData contains all meta data concerning the Ownable contract.
var OwnableMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// OwnableABI is the input ABI used to generate the binding from.
// Deprecated: Use OwnableMetaData.ABI instead.
var OwnableABI = OwnableMetaData.ABI

// Ownable is an auto generated Go binding around an Ethereum contract.
type Ownable struct {
	OwnableCaller     // Read-only binding to the contract
	OwnableTransactor // Write-only binding to the contract
	OwnableFilterer   // Log filterer for contract events
}

// OwnableCaller is an auto generated read-only Go binding around an Ethereum contract.
type OwnableCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnableTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OwnableTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnableFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OwnableFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnableSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OwnableSession struct {
	Contract     *Ownable          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OwnableCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OwnableCallerSession struct {
	Contract *OwnableCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// OwnableTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OwnableTransactorSession struct {
	Contract     *OwnableTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// OwnableRaw is an auto generated low-level Go binding around an Ethereum contract.
type OwnableRaw struct {
	Contract *Ownable // Generic contract binding to access the raw methods on
}

// OwnableCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OwnableCallerRaw struct {
	Contract *OwnableCaller // Generic read-only contract binding to access the raw methods on
}

// OwnableTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OwnableTransactorRaw struct {
	Contract *OwnableTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOwnable creates a new instance of Ownable, bound to a specific deployed contract.
func NewOwnable(address common.Address, backend bind.ContractBackend) (*Ownable, error) {
	contract, err := bindOwnable(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Ownable{OwnableCaller: OwnableCaller{contract: contract}, OwnableTransactor: OwnableTransactor{contract: contract}, OwnableFilterer: OwnableFilterer{contract: contract}}, nil
}

// NewOwnableCaller creates a new read-only instance of Ownable, bound to a specific deployed contract.
func NewOwnableCaller(address common.Address, caller bind.ContractCaller) (*OwnableCaller, error) {
	contract, err := bindOwnable(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OwnableCaller{contract: contract}, nil
}

// NewOwnableTransactor creates a new write-only instance of Ownable, bound to a specific deployed contract.
func NewOwnableTransactor(address common.Address, transactor bind.ContractTransactor) (*OwnableTransactor, error) {
	contract, err := bindOwnable(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OwnableTransactor{contract: contract}, nil
}

// NewOwnableFilterer creates a new log filterer instance of Ownable, bound to a specific deployed contract.
func NewOwnableFilterer(address common.Address, filterer bind.ContractFilterer) (*OwnableFilterer, error) {
	contract, err := bindOwnable(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OwnableFilterer{contract: contract}, nil
}

// bindOwnable binds a generic wrapper to an already deployed contract.
func bindOwnable(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OwnableMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ownable *OwnableRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ownable.Contract.OwnableCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ownable *OwnableRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ownable.Contract.OwnableTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ownable *OwnableRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ownable.Contract.OwnableTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ownable *OwnableCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ownable.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ownable *OwnableTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ownable.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ownable *OwnableTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ownable.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ownable *OwnableCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Ownable.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ownable *OwnableSession) Owner() (common.Address, error) {
	return _Ownable.Contract.Owner(&_Ownable.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Ownable *OwnableCallerSession) Owner() (common.Address, error) {
	return _Ownable.Contract.Owner(&_Ownable.CallOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Ownable *OwnableTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ownable.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Ownable *OwnableSession) RenounceOwnership() (*types.Transaction, error) {
	return _Ownable.Contract.RenounceOwnership(&_Ownable.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Ownable *OwnableTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Ownable.Contract.RenounceOwnership(&_Ownable.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ownable *OwnableTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Ownable.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ownable *OwnableSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Ownable.Contract.TransferOwnership(&_Ownable.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Ownable *OwnableTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Ownable.Contract.TransferOwnership(&_Ownable.TransactOpts, newOwner)
}

// OwnableOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Ownable contract.
type OwnableOwnershipTransferredIterator struct {
	Event *OwnableOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OwnableOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OwnableOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OwnableOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OwnableOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OwnableOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OwnableOwnershipTransferred represents a OwnershipTransferred event raised by the Ownable contract.
type OwnableOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Ownable *OwnableFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*OwnableOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Ownable.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &OwnableOwnershipTransferredIterator{contract: _Ownable.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Ownable *OwnableFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *OwnableOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Ownable.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OwnableOwnershipTransferred)
				if err := _Ownable.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Ownable *OwnableFilterer) ParseOwnershipTransferred(log types.Log) (*OwnableOwnershipTransferred, error) {
	event := new(OwnableOwnershipTransferred)
	if err := _Ownable.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ConsoleMetaData contains all meta data concerning the Console contract.
var ConsoleMetaData = &bind.MetaData{
	ABI: "[]",
	Bin: "0x60556032600b8282823980515f1a607314602657634e487b7160e01b5f525f60045260245ffd5b305f52607381538281f3fe730000000000000000000000000000000000000000301460806040525f80fdfea26469706673582212206fbc4cd51981c6aa0b18ba2d9367a71aeb0eb9e3e23d48c1508b378fcd01583e64736f6c63430008160033",
}

// ConsoleABI is the input ABI used to generate the binding from.
// Deprecated: Use ConsoleMetaData.ABI instead.
var ConsoleABI = ConsoleMetaData.ABI

// ConsoleBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ConsoleMetaData.Bin instead.
var ConsoleBin = ConsoleMetaData.Bin

// DeployConsole deploys a new Ethereum contract, binding an instance of Console to it.
func DeployConsole(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Console, error) {
	parsed, err := ConsoleMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ConsoleBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Console{ConsoleCaller: ConsoleCaller{contract: contract}, ConsoleTransactor: ConsoleTransactor{contract: contract}, ConsoleFilterer: ConsoleFilterer{contract: contract}}, nil
}

// Console is an auto generated Go binding around an Ethereum contract.
type Console struct {
	ConsoleCaller     // Read-only binding to the contract
	ConsoleTransactor // Write-only binding to the contract
	ConsoleFilterer   // Log filterer for contract events
}

// ConsoleCaller is an auto generated read-only Go binding around an Ethereum contract.
type ConsoleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ConsoleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ConsoleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ConsoleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ConsoleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ConsoleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ConsoleSession struct {
	Contract     *Console          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ConsoleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ConsoleCallerSession struct {
	Contract *ConsoleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// ConsoleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ConsoleTransactorSession struct {
	Contract     *ConsoleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// ConsoleRaw is an auto generated low-level Go binding around an Ethereum contract.
type ConsoleRaw struct {
	Contract *Console // Generic contract binding to access the raw methods on
}

// ConsoleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ConsoleCallerRaw struct {
	Contract *ConsoleCaller // Generic read-only contract binding to access the raw methods on
}

// ConsoleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ConsoleTransactorRaw struct {
	Contract *ConsoleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewConsole creates a new instance of Console, bound to a specific deployed contract.
func NewConsole(address common.Address, backend bind.ContractBackend) (*Console, error) {
	contract, err := bindConsole(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Console{ConsoleCaller: ConsoleCaller{contract: contract}, ConsoleTransactor: ConsoleTransactor{contract: contract}, ConsoleFilterer: ConsoleFilterer{contract: contract}}, nil
}

// NewConsoleCaller creates a new read-only instance of Console, bound to a specific deployed contract.
func NewConsoleCaller(address common.Address, caller bind.ContractCaller) (*ConsoleCaller, error) {
	contract, err := bindConsole(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ConsoleCaller{contract: contract}, nil
}

// NewConsoleTransactor creates a new write-only instance of Console, bound to a specific deployed contract.
func NewConsoleTransactor(address common.Address, transactor bind.ContractTransactor) (*ConsoleTransactor, error) {
	contract, err := bindConsole(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ConsoleTransactor{contract: contract}, nil
}

// NewConsoleFilterer creates a new log filterer instance of Console, bound to a specific deployed contract.
func NewConsoleFilterer(address common.Address, filterer bind.ContractFilterer) (*ConsoleFilterer, error) {
	contract, err := bindConsole(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ConsoleFilterer{contract: contract}, nil
}

// bindConsole binds a generic wrapper to an already deployed contract.
func bindConsole(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ConsoleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Console *ConsoleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Console.Contract.ConsoleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Console *ConsoleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Console.Contract.ConsoleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Console *ConsoleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Console.Contract.ConsoleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Console *ConsoleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Console.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Console *ConsoleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Console.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Console *ConsoleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Console.Contract.contract.Transact(opts, method, params...)
}

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
)

// AddrsSeqMetaData contains all meta data concerning the AddrsSeq contract.
var AddrsSeqMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"}],\"name\":\"Appended\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"newAddrs\",\"type\":\"address[]\"}],\"name\":\"add\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"append\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"i\",\"type\":\"uint64\"}],\"name\":\"at\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"}],\"name\":\"countNth\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b5061001a33610027565b610022610077565b610150565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080516000602080830182815283850190945292825260018054808201825591528151805192937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909201926100d192849201906100d6565b505050565b82805482825590600052602060002090810192821561012b579160200282015b8281111561012b57825182546001600160a01b0319166001600160a01b039091161782556020909201916001909101906100f6565b5061013792915061013b565b5090565b5b80821115610137576000815560010161013c565b61092b8061015f6000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c80637f353d551161005b5780637f353d55146100fa5780638da5cb5b14610102578063c4c1c94f14610113578063f2fde38b1461012657600080fd5b806306661abd1461008d5780632a2d01f8146100b257806335147092146100c5578063715018a6146100f0575b600080fd5b610095610139565b6040516001600160401b0390911681526020015b60405180910390f35b6100956100c036600461073c565b61014e565b6100d86100d336600461075e565b6101f5565b6040516001600160a01b0390911681526020016100a9565b6100f861033d565b005b6100f8610373565b6000546001600160a01b03166100d8565b6100f8610121366004610791565b610469565b6100f8610134366004610805565b61055c565b6001805460009161014991610844565b905090565b6000610158610139565b6001600160401b0316826001600160401b0316106101c75760405162461bcd60e51b815260206004820152602160248201527f41646472735365712e636f756e744e74683a206e206f7574206f662072616e676044820152606560f81b60648201526084015b60405180910390fd5b6001826001600160401b0316815481106101e3576101e361086c565b60009182526020909120015492915050565b60006101ff610139565b6001600160401b0316836001600160401b03161061025f5760405162461bcd60e51b815260206004820152601b60248201527f41646472735365712e61743a206e206f7574206f662072616e6765000000000060448201526064016101be565b6001836001600160401b03168154811061027b5761027b61086c565b6000918252602090912001546001600160401b038316106102de5760405162461bcd60e51b815260206004820152601b60248201527f41646472735365712e61743a2069206f7574206f662072616e6765000000000060448201526064016101be565b6001836001600160401b0316815481106102fa576102fa61086c565b90600052602060002001600001826001600160401b0316815481106103215761032161086c565b6000918252602090912001546001600160a01b03169392505050565b6000546001600160a01b031633146103675760405162461bcd60e51b81526004016101be90610882565b61037160006105f7565b565b6000546001600160a01b0316331461039d5760405162461bcd60e51b81526004016101be90610882565b6103af60016001600160401b03610844565b6001600160401b0316600180549050106104175760405162461bcd60e51b815260206004820152602360248201527f41646472735365712e617070656e643a20736571206578636565656473206c696044820152621b5a5d60ea1b60648201526084016101be565b600180547f5ff9c98a1faf73c018d22371cb08c08dec1412825b68523a8e7deaa17683a6b99161044691610844565b6040516001600160401b03909116815260200160405180910390a1610371610647565b6000546001600160a01b031633146104935760405162461bcd60e51b81526004016101be90610882565b600180546000916104a3916108b7565b905060005b6001600160401b038116831115610556576001826001600160401b0316815481106104d5576104d561086c565b906000526020600020016000018484836001600160401b03168181106104fd576104fd61086c565b90506020020160208101906105129190610805565b81546001810183556000928352602090922090910180546001600160a01b0319166001600160a01b039092169190911790558061054e816108ce565b9150506104a8565b50505050565b6000546001600160a01b031633146105865760405162461bcd60e51b81526004016101be90610882565b6001600160a01b0381166105eb5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016101be565b6105f4816105f7565b50565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080516000602080830182815283850190945292825260018054808201825591528151805192937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909201926106a192849201906106a6565b505050565b8280548282559060005260206000209081019282156106fb579160200282015b828111156106fb57825182546001600160a01b0319166001600160a01b039091161782556020909201916001909101906106c6565b5061070792915061070b565b5090565b5b80821115610707576000815560010161070c565b80356001600160401b038116811461073757600080fd5b919050565b60006020828403121561074e57600080fd5b61075782610720565b9392505050565b6000806040838503121561077157600080fd5b61077a83610720565b915061078860208401610720565b90509250929050565b600080602083850312156107a457600080fd5b82356001600160401b03808211156107bb57600080fd5b818501915085601f8301126107cf57600080fd5b8135818111156107de57600080fd5b8660208260051b85010111156107f357600080fd5b60209290920196919550909350505050565b60006020828403121561081757600080fd5b81356001600160a01b038116811461075757600080fd5b634e487b7160e01b600052601160045260246000fd5b60006001600160401b03838116908316818110156108645761086461082e565b039392505050565b634e487b7160e01b600052603260045260246000fd5b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b6000828210156108c9576108c961082e565b500390565b60006001600160401b03808316818114156108eb576108eb61082e565b600101939250505056fea26469706673582212203251272da89b497480f4a09582108ad651d8da427ef504ecbe4b334d356af24b64736f6c63430008090033",
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
	parsed, err := abi.JSON(strings.NewReader(AddrsSeqABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
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
	parsed, err := abi.JSON(strings.NewReader(ContextABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
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

// OwnableMetaData contains all meta data concerning the Ownable contract.
var OwnableMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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
	parsed, err := abi.JSON(strings.NewReader(OwnableABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
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

// RegistryMetaData contains all meta data concerning the Registry contract.
var RegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"_addrsSeq\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"i\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"a\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"Registered\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"a\",\"type\":\"address\"}],\"name\":\"get\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"i\",\"type\":\"uint64\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b506040516107be3803806107be83398101604081905261002f91610054565b600080546001600160a01b0319166001600160a01b0392909216919091179055610084565b60006020828403121561006657600080fd5b81516001600160a01b038116811461007d57600080fd5b9392505050565b61072b806100936000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c806354d77fe41461003b578063c2bc2efc14610050575b600080fd5b61004e6100493660046104bf565b610079565b005b61006361005e3660046105a9565b6102dd565b604051610070919061061a565b60405180910390f35b60008054604051631a8a384960e11b815267ffffffffffffffff8087166004830152851660248201526001600160a01b039091169063351470929060440160206040518083038186803b1580156100cf57600080fd5b505afa1580156100e3573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610107919061062d565b90506001600160a01b03811633146101665760405162461bcd60e51b815260206004820152601f60248201527f52656769737472793a2073656e646572206973206e6f7420616c6c6f7765640060448201526064015b60405180910390fd5b33600090815260016020526040902080546101809061064a565b1590506101db5760405162461bcd60e51b815260206004820152602360248201527f52656769737472793a2073656e64657220616c726561647920726567697374656044820152621c995960ea1b606482015260840161015d565b600082511161023a5760405162461bcd60e51b815260206004820152602560248201527f52656769737472793a2063616e6e6f7420726567697374657220656d7074792060448201526476616c756560d81b606482015260840161015d565b336000908152600160209081526040909120835161025a928501906103f3565b5061029a6040518060400160405280601e81526020017f526567697374657265642076616c756520666f722073656e646572202573000081525033610389565b7f2791b5fcdbb8707a39509d0547670c51cfca994cc587718f55192d409fd514ca848433856040516102cf9493929190610685565b60405180910390a150505050565b6001600160a01b03811660009081526001602052604090208054606091906103049061064a565b80601f01602080910402602001604051908101604052809291908181526020018280546103309061064a565b801561037d5780601f106103525761010080835404028352916020019161037d565b820191906000526020600020905b81548152906001019060200180831161036057829003601f168201915b50505050509050919050565b6103ce828260405160240161039f9291906106cb565b60408051601f198184030181529190526020810180516001600160e01b031663319af33360e01b1790526103d2565b5050565b80516a636f6e736f6c652e6c6f67602083016000808483855afa5050505050565b8280546103ff9061064a565b90600052602060002090601f0160209004810192826104215760008555610467565b82601f1061043a57805160ff1916838001178555610467565b82800160010185558215610467579182015b8281111561046757825182559160200191906001019061044c565b50610473929150610477565b5090565b5b808211156104735760008155600101610478565b803567ffffffffffffffff811681146104a457600080fd5b919050565b634e487b7160e01b600052604160045260246000fd5b6000806000606084860312156104d457600080fd5b6104dd8461048c565b92506104eb6020850161048c565b9150604084013567ffffffffffffffff8082111561050857600080fd5b818601915086601f83011261051c57600080fd5b81358181111561052e5761052e6104a9565b604051601f8201601f19908116603f01168101908382118183101715610556576105566104a9565b8160405282815289602084870101111561056f57600080fd5b8260208601602083013760006020848301015280955050505050509250925092565b6001600160a01b03811681146105a657600080fd5b50565b6000602082840312156105bb57600080fd5b81356105c681610591565b9392505050565b6000815180845260005b818110156105f3576020818501810151868301820152016105d7565b81811115610605576000602083870101525b50601f01601f19169290920160200192915050565b6020815260006105c660208301846105cd565b60006020828403121561063f57600080fd5b81516105c681610591565b600181811c9082168061065e57607f821691505b6020821081141561067f57634e487b7160e01b600052602260045260246000fd5b50919050565b67ffffffffffffffff8581168252841660208201526001600160a01b03831660408201526080606082018190526000906106c1908301846105cd565b9695505050505050565b6040815260006106de60408301856105cd565b905060018060a01b0383166020830152939250505056fea2646970667358221220ed269b5f5df69592d93fa80cb85caf5bbaa5b4bc696f5752770a43eeb7cc723964736f6c63430008090033",
}

// RegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use RegistryMetaData.ABI instead.
var RegistryABI = RegistryMetaData.ABI

// RegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use RegistryMetaData.Bin instead.
var RegistryBin = RegistryMetaData.Bin

// DeployRegistry deploys a new Ethereum contract, binding an instance of Registry to it.
func DeployRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, _addrsSeq common.Address) (common.Address, *types.Transaction, *Registry, error) {
	parsed, err := RegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(RegistryBin), backend, _addrsSeq)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Registry{RegistryCaller: RegistryCaller{contract: contract}, RegistryTransactor: RegistryTransactor{contract: contract}, RegistryFilterer: RegistryFilterer{contract: contract}}, nil
}

// Registry is an auto generated Go binding around an Ethereum contract.
type Registry struct {
	RegistryCaller     // Read-only binding to the contract
	RegistryTransactor // Write-only binding to the contract
	RegistryFilterer   // Log filterer for contract events
}

// RegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type RegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RegistrySession struct {
	Contract     *Registry         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RegistryCallerSession struct {
	Contract *RegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// RegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RegistryTransactorSession struct {
	Contract     *RegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// RegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type RegistryRaw struct {
	Contract *Registry // Generic contract binding to access the raw methods on
}

// RegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RegistryCallerRaw struct {
	Contract *RegistryCaller // Generic read-only contract binding to access the raw methods on
}

// RegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RegistryTransactorRaw struct {
	Contract *RegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRegistry creates a new instance of Registry, bound to a specific deployed contract.
func NewRegistry(address common.Address, backend bind.ContractBackend) (*Registry, error) {
	contract, err := bindRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Registry{RegistryCaller: RegistryCaller{contract: contract}, RegistryTransactor: RegistryTransactor{contract: contract}, RegistryFilterer: RegistryFilterer{contract: contract}}, nil
}

// NewRegistryCaller creates a new read-only instance of Registry, bound to a specific deployed contract.
func NewRegistryCaller(address common.Address, caller bind.ContractCaller) (*RegistryCaller, error) {
	contract, err := bindRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RegistryCaller{contract: contract}, nil
}

// NewRegistryTransactor creates a new write-only instance of Registry, bound to a specific deployed contract.
func NewRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*RegistryTransactor, error) {
	contract, err := bindRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RegistryTransactor{contract: contract}, nil
}

// NewRegistryFilterer creates a new log filterer instance of Registry, bound to a specific deployed contract.
func NewRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*RegistryFilterer, error) {
	contract, err := bindRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RegistryFilterer{contract: contract}, nil
}

// bindRegistry binds a generic wrapper to an already deployed contract.
func bindRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RegistryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Registry *RegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Registry.Contract.RegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Registry *RegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Registry.Contract.RegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Registry *RegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Registry.Contract.RegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Registry *RegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Registry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Registry *RegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Registry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Registry *RegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Registry.Contract.contract.Transact(opts, method, params...)
}

// Get is a free data retrieval call binding the contract method 0xc2bc2efc.
//
// Solidity: function get(address a) view returns(bytes)
func (_Registry *RegistryCaller) Get(opts *bind.CallOpts, a common.Address) ([]byte, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "get", a)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// Get is a free data retrieval call binding the contract method 0xc2bc2efc.
//
// Solidity: function get(address a) view returns(bytes)
func (_Registry *RegistrySession) Get(a common.Address) ([]byte, error) {
	return _Registry.Contract.Get(&_Registry.CallOpts, a)
}

// Get is a free data retrieval call binding the contract method 0xc2bc2efc.
//
// Solidity: function get(address a) view returns(bytes)
func (_Registry *RegistryCallerSession) Get(a common.Address) ([]byte, error) {
	return _Registry.Contract.Get(&_Registry.CallOpts, a)
}

// Register is a paid mutator transaction binding the contract method 0x54d77fe4.
//
// Solidity: function register(uint64 n, uint64 i, bytes data) returns()
func (_Registry *RegistryTransactor) Register(opts *bind.TransactOpts, n uint64, i uint64, data []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "register", n, i, data)
}

// Register is a paid mutator transaction binding the contract method 0x54d77fe4.
//
// Solidity: function register(uint64 n, uint64 i, bytes data) returns()
func (_Registry *RegistrySession) Register(n uint64, i uint64, data []byte) (*types.Transaction, error) {
	return _Registry.Contract.Register(&_Registry.TransactOpts, n, i, data)
}

// Register is a paid mutator transaction binding the contract method 0x54d77fe4.
//
// Solidity: function register(uint64 n, uint64 i, bytes data) returns()
func (_Registry *RegistryTransactorSession) Register(n uint64, i uint64, data []byte) (*types.Transaction, error) {
	return _Registry.Contract.Register(&_Registry.TransactOpts, n, i, data)
}

// RegistryRegisteredIterator is returned from FilterRegistered and is used to iterate over the raw logs and unpacked data for Registered events raised by the Registry contract.
type RegistryRegisteredIterator struct {
	Event *RegistryRegistered // Event containing the contract specifics and raw log

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
func (it *RegistryRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryRegistered)
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
		it.Event = new(RegistryRegistered)
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
func (it *RegistryRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryRegistered represents a Registered event raised by the Registry contract.
type RegistryRegistered struct {
	N    uint64
	I    uint64
	A    common.Address
	Data []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterRegistered is a free log retrieval operation binding the contract event 0x2791b5fcdbb8707a39509d0547670c51cfca994cc587718f55192d409fd514ca.
//
// Solidity: event Registered(uint64 n, uint64 i, address a, bytes data)
func (_Registry *RegistryFilterer) FilterRegistered(opts *bind.FilterOpts) (*RegistryRegisteredIterator, error) {

	logs, sub, err := _Registry.contract.FilterLogs(opts, "Registered")
	if err != nil {
		return nil, err
	}
	return &RegistryRegisteredIterator{contract: _Registry.contract, event: "Registered", logs: logs, sub: sub}, nil
}

// WatchRegistered is a free log subscription operation binding the contract event 0x2791b5fcdbb8707a39509d0547670c51cfca994cc587718f55192d409fd514ca.
//
// Solidity: event Registered(uint64 n, uint64 i, address a, bytes data)
func (_Registry *RegistryFilterer) WatchRegistered(opts *bind.WatchOpts, sink chan<- *RegistryRegistered) (event.Subscription, error) {

	logs, sub, err := _Registry.contract.WatchLogs(opts, "Registered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryRegistered)
				if err := _Registry.contract.UnpackLog(event, "Registered", log); err != nil {
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

// ParseRegistered is a log parse operation binding the contract event 0x2791b5fcdbb8707a39509d0547670c51cfca994cc587718f55192d409fd514ca.
//
// Solidity: event Registered(uint64 n, uint64 i, address a, bytes data)
func (_Registry *RegistryFilterer) ParseRegistered(log types.Log) (*RegistryRegistered, error) {
	event := new(RegistryRegistered)
	if err := _Registry.contract.UnpackLog(event, "Registered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ConsoleMetaData contains all meta data concerning the Console contract.
var ConsoleMetaData = &bind.MetaData{
	ABI: "[]",
	Bin: "0x60566037600b82828239805160001a607314602a57634e487b7160e01b600052600060045260246000fd5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea2646970667358221220bb07727afc6a474f8c7e705d6fe48e9dd5d02f16b2af91738cfe76087e4b1fd064736f6c63430008090033",
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
	parsed, err := abi.JSON(strings.NewReader(ConsoleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
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

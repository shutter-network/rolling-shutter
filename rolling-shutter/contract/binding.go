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

// DecryptorsConfig is an auto generated low-level Go binding around an user-defined struct.
type DecryptorsConfig struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}

// KeypersConfig is an auto generated low-level Go binding around an user-defined struct.
type KeypersConfig struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}

// AddrsSeqMetaData contains all meta data concerning the AddrsSeq contract.
var AddrsSeqMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"}],\"name\":\"Appended\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"newAddrs\",\"type\":\"address[]\"}],\"name\":\"add\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"append\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"i\",\"type\":\"uint64\"}],\"name\":\"at\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"}],\"name\":\"countNth\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b5061001a33610027565b610022610077565b610150565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080516000602080830182815283850190945292825260018054808201825591528151805192937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909201926100d192849201906100d6565b505050565b82805482825590600052602060002090810192821561012b579160200282015b8281111561012b57825182546001600160a01b0319166001600160a01b039091161782556020909201916001909101906100f6565b5061013792915061013b565b5090565b5b80821115610137576000815560010161013c565b61092b8061015f6000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c80637f353d551161005b5780637f353d55146100fa5780638da5cb5b14610102578063c4c1c94f14610113578063f2fde38b1461012657600080fd5b806306661abd1461008d5780632a2d01f8146100b257806335147092146100c5578063715018a6146100f0575b600080fd5b610095610139565b6040516001600160401b0390911681526020015b60405180910390f35b6100956100c036600461073c565b61014e565b6100d86100d336600461075e565b6101f5565b6040516001600160a01b0390911681526020016100a9565b6100f861033d565b005b6100f8610373565b6000546001600160a01b03166100d8565b6100f8610121366004610791565b610469565b6100f8610134366004610805565b61055c565b6001805460009161014991610844565b905090565b6000610158610139565b6001600160401b0316826001600160401b0316106101c75760405162461bcd60e51b815260206004820152602160248201527f41646472735365712e636f756e744e74683a206e206f7574206f662072616e676044820152606560f81b60648201526084015b60405180910390fd5b6001826001600160401b0316815481106101e3576101e361086c565b60009182526020909120015492915050565b60006101ff610139565b6001600160401b0316836001600160401b03161061025f5760405162461bcd60e51b815260206004820152601b60248201527f41646472735365712e61743a206e206f7574206f662072616e6765000000000060448201526064016101be565b6001836001600160401b03168154811061027b5761027b61086c565b6000918252602090912001546001600160401b038316106102de5760405162461bcd60e51b815260206004820152601b60248201527f41646472735365712e61743a2069206f7574206f662072616e6765000000000060448201526064016101be565b6001836001600160401b0316815481106102fa576102fa61086c565b90600052602060002001600001826001600160401b0316815481106103215761032161086c565b6000918252602090912001546001600160a01b03169392505050565b6000546001600160a01b031633146103675760405162461bcd60e51b81526004016101be90610882565b61037160006105f7565b565b6000546001600160a01b0316331461039d5760405162461bcd60e51b81526004016101be90610882565b6103af60016001600160401b03610844565b6001600160401b0316600180549050106104175760405162461bcd60e51b815260206004820152602360248201527f41646472735365712e617070656e643a20736571206578636565656473206c696044820152621b5a5d60ea1b60648201526084016101be565b600180547f5ff9c98a1faf73c018d22371cb08c08dec1412825b68523a8e7deaa17683a6b99161044691610844565b6040516001600160401b03909116815260200160405180910390a1610371610647565b6000546001600160a01b031633146104935760405162461bcd60e51b81526004016101be90610882565b600180546000916104a3916108b7565b905060005b6001600160401b038116831115610556576001826001600160401b0316815481106104d5576104d561086c565b906000526020600020016000018484836001600160401b03168181106104fd576104fd61086c565b90506020020160208101906105129190610805565b81546001810183556000928352602090922090910180546001600160a01b0319166001600160a01b039092169190911790558061054e816108ce565b9150506104a8565b50505050565b6000546001600160a01b031633146105865760405162461bcd60e51b81526004016101be90610882565b6001600160a01b0381166105eb5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016101be565b6105f4816105f7565b50565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080516000602080830182815283850190945292825260018054808201825591528151805192937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909201926106a192849201906106a6565b505050565b8280548282559060005260206000209081019282156106fb579160200282015b828111156106fb57825182546001600160a01b0319166001600160a01b039091161782556020909201916001909101906106c6565b5061070792915061070b565b5090565b5b80821115610707576000815560010161070c565b80356001600160401b038116811461073757600080fd5b919050565b60006020828403121561074e57600080fd5b61075782610720565b9392505050565b6000806040838503121561077157600080fd5b61077a83610720565b915061078860208401610720565b90509250929050565b600080602083850312156107a457600080fd5b82356001600160401b03808211156107bb57600080fd5b818501915085601f8301126107cf57600080fd5b8135818111156107de57600080fd5b8660208260051b85010111156107f357600080fd5b60209290920196919550909350505050565b60006020828403121561081757600080fd5b81356001600160a01b038116811461075757600080fd5b634e487b7160e01b600052601160045260246000fd5b60006001600160401b03838116908316818110156108645761086461082e565b039392505050565b634e487b7160e01b600052603260045260246000fd5b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b6000828210156108c9576108c961082e565b500390565b60006001600160401b03808316818114156108eb576108eb61082e565b600101939250505056fea26469706673582212207bd5d3542ff670aee0276b5f66ebed986f8b0b546169d49735fc066a03db9b6f64736f6c63430008090033",
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

// DecryptorsConfigsListMetaData contains all meta data concerning the DecryptorsConfigsList contract.
var DecryptorsConfigsListMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"_addrsSeq\",\"type\":\"address\"},{\"internalType\":\"contractRegistry\",\"name\":\"_BLSKeysRegistry\",\"type\":\"address\"},{\"internalType\":\"contractRegistry\",\"name\":\"_KeySignaturesRegistry\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"}],\"name\":\"NewConfig\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BLSKeysRegistry\",\"outputs\":[{\"internalType\":\"contractRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"KeySignaturesRegistry\",\"outputs\":[{\"internalType\":\"contractRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structDecryptorsConfig\",\"name\":\"config\",\"type\":\"tuple\"}],\"name\":\"addNewCfg\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"addrsSeq\",\"outputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"decryptorsConfigs\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"}],\"name\":\"getActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structDecryptorsConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structDecryptorsConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040523480156200001157600080fd5b5060405162000e8b38038062000e8b8339810160408190526200003491620004e4565b6200003f336200047b565b604051630545a03f60e31b8152600060048201526001600160a01b03841690632a2d01f89060240160206040518083038186803b1580156200008057600080fd5b505afa15801562000095573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190620000bb919062000538565b6001600160401b031615620001285760405162461bcd60e51b815260206004820152602860248201527f4164647273536571206d757374206861766520656d707479206c697374206174604482015267020696e64657820360c41b60648201526084015b60405180910390fd5b826001600160a01b0316826001600160a01b0316634d89eaaf6040518163ffffffff1660e01b815260040160206040518083038186803b1580156200016c57600080fd5b505afa15801562000181573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190620001a791906200056a565b6001600160a01b031614620002165760405162461bcd60e51b815260206004820152602e60248201527f4164647273536571206f66205f424c534b6579735265676973747279206d757360448201526d74206265205f616464727353657160901b60648201526084016200011f565b826001600160a01b0316816001600160a01b0316634d89eaaf6040518163ffffffff1660e01b815260040160206040518083038186803b1580156200025a57600080fd5b505afa1580156200026f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906200029591906200056a565b6001600160a01b031614620003135760405162461bcd60e51b815260206004820152603460248201527f4164647273536571206f66205f4b65795369676e61747572657352656769737460448201527f7279206d757374206265205f616464727353657100000000000000000000000060648201526084016200011f565b816001600160a01b0316816001600160a01b03161415620003895760405162461bcd60e51b815260206004820152602960248201527f5468652074776f20757365642072656769737472696573206d75737420626520604482015268191a5999995c995b9d60ba1b60648201526084016200011f565b600280546001600160a01b038581166001600160a01b0319928316179092556003805485841690831617905560048054928416929091169190911790556040805180820182526000808252602080830182815260018054808201825590845293517fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909401805491516001600160401b0390811668010000000000000000026001600160801b0319909316951694909417179092558251818152918201527ff991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7910160405180910390a15050506200058a565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6001600160a01b0381168114620004e157600080fd5b50565b600080600060608486031215620004fa57600080fd5b83516200050781620004cb565b60208501519093506200051a81620004cb565b60408501519092506200052d81620004cb565b809150509250925092565b6000602082840312156200054b57600080fd5b81516001600160401b03811681146200056357600080fd5b9392505050565b6000602082840312156200057d57600080fd5b81516200056381620004cb565b6108f1806200059a6000396000f3fe608060405234801561001057600080fd5b506004361061009e5760003560e01c806379f780991161006657806379f78099146101365780638da5cb5b14610149578063b5351b0d1461015a578063f2fde38b14610194578063f9c72b53146101a757600080fd5b806304355dfa146100a3578063139ff1c0146100d35780634d89eaaf146100e65780635c8e4c97146100f9578063715018a61461012c575b600080fd5b6004546100b6906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b6003546100b6906001600160a01b031681565b6002546100b6906001600160a01b031681565b61010c610107366004610710565b6101af565b604080516001600160401b039384168152929091166020830152016100ca565b6101346101e4565b005b610134610144366004610729565b610223565b6000546001600160a01b03166100b6565b61016d610168366004610756565b610544565b6040805182516001600160401b0390811682526020938401511692810192909252016100ca565b6101346101a236600461077a565b610603565b61016d61069e565b600181815481106101bf57600080fd5b6000918252602090912001546001600160401b038082169250600160401b9091041682565b6000546001600160a01b031633146102175760405162461bcd60e51b815260040161020e906107a3565b60405180910390fd5b61022160006106c0565b565b6000546001600160a01b0316331461024d5760405162461bcd60e51b815260040161020e906107a3565b61025d6040820160208301610756565b6001600160401b0316600260009054906101000a90046001600160a01b03166001600160a01b03166306661abd6040518163ffffffff1660e01b815260040160206040518083038186803b1580156102b457600080fd5b505afa1580156102c8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906102ec91906107d8565b6001600160401b0316116103685760405162461bcd60e51b815260206004820152603a60248201527f4e6f20617070656e6465642073657420696e2073657120636f72726573706f6e60448201527f64696e6720746f20636f6e66696727732073657420696e646578000000000000606482015260840161020e565b6103756020820182610756565b6001600160401b03166001808080549050610390919061080b565b815481106103a0576103a0610822565b6000918252602090912001546001600160401b031611156104295760405162461bcd60e51b815260206004820152603860248201527f43616e6e6f7420616464206e6577207365742077697468206c6f77657220626c60448201527f6f636b206e756d626572207468616e2070726576696f75730000000000000000606482015260840161020e565b6104366020820182610756565b6001600160401b03164311156104a05760405162461bcd60e51b815260206004820152602960248201527f43616e6e6f7420616464206e6577207365742077697468207061737420626c6f60448201526831b590373ab6b132b960b91b606482015260840161020e565b60018054808201825560009190915281907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6016104dd8282610838565b507ff991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7905061050e6020830183610756565b61051e6040840160208501610756565b604080516001600160401b0393841681529290911660208301520160405180910390a150565b6040805180820190915260008082526020820152600180546000916105689161080b565b90505b826001600160401b03166001828154811061058857610588610822565b6000918252602090912001546001600160401b0316116105f157600181815481106105b5576105b5610822565b6000918252602091829020604080518082019091529101546001600160401b038082168352600160401b90910416918101919091529392505050565b806105fb816108a4565b91505061056b565b6000546001600160a01b0316331461062d5760405162461bcd60e51b815260040161020e906107a3565b6001600160a01b0381166106925760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b606482015260840161020e565b61069b816106c0565b50565b60408051808201909152600080825260208201526106bb43610544565b905090565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60006020828403121561072257600080fd5b5035919050565b60006040828403121561073b57600080fd5b50919050565b6001600160401b038116811461069b57600080fd5b60006020828403121561076857600080fd5b813561077381610741565b9392505050565b60006020828403121561078c57600080fd5b81356001600160a01b038116811461077357600080fd5b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b6000602082840312156107ea57600080fd5b815161077381610741565b634e487b7160e01b600052601160045260246000fd5b60008282101561081d5761081d6107f5565b500390565b634e487b7160e01b600052603260045260246000fd5b813561084381610741565b6001600160401b03811690508154816001600160401b03198216178355602084013561086e81610741565b6fffffffffffffffff00000000000000008160401b16836fffffffffffffffffffffffffffffffff198416171784555050505050565b6000816108b3576108b36107f5565b50600019019056fea264697066735822122040b870ce5bdfaa32b2948f87dd26fc7bf65c2764a8e727fe2a99da41005988a064736f6c63430008090033",
}

// DecryptorsConfigsListABI is the input ABI used to generate the binding from.
// Deprecated: Use DecryptorsConfigsListMetaData.ABI instead.
var DecryptorsConfigsListABI = DecryptorsConfigsListMetaData.ABI

// DecryptorsConfigsListBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use DecryptorsConfigsListMetaData.Bin instead.
var DecryptorsConfigsListBin = DecryptorsConfigsListMetaData.Bin

// DeployDecryptorsConfigsList deploys a new Ethereum contract, binding an instance of DecryptorsConfigsList to it.
func DeployDecryptorsConfigsList(auth *bind.TransactOpts, backend bind.ContractBackend, _addrsSeq common.Address, _BLSKeysRegistry common.Address, _KeySignaturesRegistry common.Address) (common.Address, *types.Transaction, *DecryptorsConfigsList, error) {
	parsed, err := DecryptorsConfigsListMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(DecryptorsConfigsListBin), backend, _addrsSeq, _BLSKeysRegistry, _KeySignaturesRegistry)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &DecryptorsConfigsList{DecryptorsConfigsListCaller: DecryptorsConfigsListCaller{contract: contract}, DecryptorsConfigsListTransactor: DecryptorsConfigsListTransactor{contract: contract}, DecryptorsConfigsListFilterer: DecryptorsConfigsListFilterer{contract: contract}}, nil
}

// DecryptorsConfigsList is an auto generated Go binding around an Ethereum contract.
type DecryptorsConfigsList struct {
	DecryptorsConfigsListCaller     // Read-only binding to the contract
	DecryptorsConfigsListTransactor // Write-only binding to the contract
	DecryptorsConfigsListFilterer   // Log filterer for contract events
}

// DecryptorsConfigsListCaller is an auto generated read-only Go binding around an Ethereum contract.
type DecryptorsConfigsListCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DecryptorsConfigsListTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DecryptorsConfigsListTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DecryptorsConfigsListFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DecryptorsConfigsListFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DecryptorsConfigsListSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DecryptorsConfigsListSession struct {
	Contract     *DecryptorsConfigsList // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// DecryptorsConfigsListCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DecryptorsConfigsListCallerSession struct {
	Contract *DecryptorsConfigsListCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// DecryptorsConfigsListTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DecryptorsConfigsListTransactorSession struct {
	Contract     *DecryptorsConfigsListTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// DecryptorsConfigsListRaw is an auto generated low-level Go binding around an Ethereum contract.
type DecryptorsConfigsListRaw struct {
	Contract *DecryptorsConfigsList // Generic contract binding to access the raw methods on
}

// DecryptorsConfigsListCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DecryptorsConfigsListCallerRaw struct {
	Contract *DecryptorsConfigsListCaller // Generic read-only contract binding to access the raw methods on
}

// DecryptorsConfigsListTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DecryptorsConfigsListTransactorRaw struct {
	Contract *DecryptorsConfigsListTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDecryptorsConfigsList creates a new instance of DecryptorsConfigsList, bound to a specific deployed contract.
func NewDecryptorsConfigsList(address common.Address, backend bind.ContractBackend) (*DecryptorsConfigsList, error) {
	contract, err := bindDecryptorsConfigsList(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DecryptorsConfigsList{DecryptorsConfigsListCaller: DecryptorsConfigsListCaller{contract: contract}, DecryptorsConfigsListTransactor: DecryptorsConfigsListTransactor{contract: contract}, DecryptorsConfigsListFilterer: DecryptorsConfigsListFilterer{contract: contract}}, nil
}

// NewDecryptorsConfigsListCaller creates a new read-only instance of DecryptorsConfigsList, bound to a specific deployed contract.
func NewDecryptorsConfigsListCaller(address common.Address, caller bind.ContractCaller) (*DecryptorsConfigsListCaller, error) {
	contract, err := bindDecryptorsConfigsList(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DecryptorsConfigsListCaller{contract: contract}, nil
}

// NewDecryptorsConfigsListTransactor creates a new write-only instance of DecryptorsConfigsList, bound to a specific deployed contract.
func NewDecryptorsConfigsListTransactor(address common.Address, transactor bind.ContractTransactor) (*DecryptorsConfigsListTransactor, error) {
	contract, err := bindDecryptorsConfigsList(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DecryptorsConfigsListTransactor{contract: contract}, nil
}

// NewDecryptorsConfigsListFilterer creates a new log filterer instance of DecryptorsConfigsList, bound to a specific deployed contract.
func NewDecryptorsConfigsListFilterer(address common.Address, filterer bind.ContractFilterer) (*DecryptorsConfigsListFilterer, error) {
	contract, err := bindDecryptorsConfigsList(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DecryptorsConfigsListFilterer{contract: contract}, nil
}

// bindDecryptorsConfigsList binds a generic wrapper to an already deployed contract.
func bindDecryptorsConfigsList(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DecryptorsConfigsListABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DecryptorsConfigsList *DecryptorsConfigsListRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DecryptorsConfigsList.Contract.DecryptorsConfigsListCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DecryptorsConfigsList *DecryptorsConfigsListRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DecryptorsConfigsList.Contract.DecryptorsConfigsListTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DecryptorsConfigsList *DecryptorsConfigsListRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DecryptorsConfigsList.Contract.DecryptorsConfigsListTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DecryptorsConfigsList *DecryptorsConfigsListCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DecryptorsConfigsList.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DecryptorsConfigsList *DecryptorsConfigsListTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DecryptorsConfigsList.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DecryptorsConfigsList *DecryptorsConfigsListTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DecryptorsConfigsList.Contract.contract.Transact(opts, method, params...)
}

// BLSKeysRegistry is a free data retrieval call binding the contract method 0x139ff1c0.
//
// Solidity: function BLSKeysRegistry() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListCaller) BLSKeysRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DecryptorsConfigsList.contract.Call(opts, &out, "BLSKeysRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BLSKeysRegistry is a free data retrieval call binding the contract method 0x139ff1c0.
//
// Solidity: function BLSKeysRegistry() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListSession) BLSKeysRegistry() (common.Address, error) {
	return _DecryptorsConfigsList.Contract.BLSKeysRegistry(&_DecryptorsConfigsList.CallOpts)
}

// BLSKeysRegistry is a free data retrieval call binding the contract method 0x139ff1c0.
//
// Solidity: function BLSKeysRegistry() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListCallerSession) BLSKeysRegistry() (common.Address, error) {
	return _DecryptorsConfigsList.Contract.BLSKeysRegistry(&_DecryptorsConfigsList.CallOpts)
}

// KeySignaturesRegistry is a free data retrieval call binding the contract method 0x04355dfa.
//
// Solidity: function KeySignaturesRegistry() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListCaller) KeySignaturesRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DecryptorsConfigsList.contract.Call(opts, &out, "KeySignaturesRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// KeySignaturesRegistry is a free data retrieval call binding the contract method 0x04355dfa.
//
// Solidity: function KeySignaturesRegistry() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListSession) KeySignaturesRegistry() (common.Address, error) {
	return _DecryptorsConfigsList.Contract.KeySignaturesRegistry(&_DecryptorsConfigsList.CallOpts)
}

// KeySignaturesRegistry is a free data retrieval call binding the contract method 0x04355dfa.
//
// Solidity: function KeySignaturesRegistry() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListCallerSession) KeySignaturesRegistry() (common.Address, error) {
	return _DecryptorsConfigsList.Contract.KeySignaturesRegistry(&_DecryptorsConfigsList.CallOpts)
}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListCaller) AddrsSeq(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DecryptorsConfigsList.contract.Call(opts, &out, "addrsSeq")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListSession) AddrsSeq() (common.Address, error) {
	return _DecryptorsConfigsList.Contract.AddrsSeq(&_DecryptorsConfigsList.CallOpts)
}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListCallerSession) AddrsSeq() (common.Address, error) {
	return _DecryptorsConfigsList.Contract.AddrsSeq(&_DecryptorsConfigsList.CallOpts)
}

// DecryptorsConfigs is a free data retrieval call binding the contract method 0x5c8e4c97.
//
// Solidity: function decryptorsConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex)
func (_DecryptorsConfigsList *DecryptorsConfigsListCaller) DecryptorsConfigs(opts *bind.CallOpts, arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}, error) {
	var out []interface{}
	err := _DecryptorsConfigsList.contract.Call(opts, &out, "decryptorsConfigs", arg0)

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

// DecryptorsConfigs is a free data retrieval call binding the contract method 0x5c8e4c97.
//
// Solidity: function decryptorsConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex)
func (_DecryptorsConfigsList *DecryptorsConfigsListSession) DecryptorsConfigs(arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}, error) {
	return _DecryptorsConfigsList.Contract.DecryptorsConfigs(&_DecryptorsConfigsList.CallOpts, arg0)
}

// DecryptorsConfigs is a free data retrieval call binding the contract method 0x5c8e4c97.
//
// Solidity: function decryptorsConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex)
func (_DecryptorsConfigsList *DecryptorsConfigsListCallerSession) DecryptorsConfigs(arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}, error) {
	return _DecryptorsConfigsList.Contract.DecryptorsConfigs(&_DecryptorsConfigsList.CallOpts, arg0)
}

// GetActiveConfig is a free data retrieval call binding the contract method 0xb5351b0d.
//
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64))
func (_DecryptorsConfigsList *DecryptorsConfigsListCaller) GetActiveConfig(opts *bind.CallOpts, activationBlockNumber uint64) (DecryptorsConfig, error) {
	var out []interface{}
	err := _DecryptorsConfigsList.contract.Call(opts, &out, "getActiveConfig", activationBlockNumber)

	if err != nil {
		return *new(DecryptorsConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(DecryptorsConfig)).(*DecryptorsConfig)

	return out0, err

}

// GetActiveConfig is a free data retrieval call binding the contract method 0xb5351b0d.
//
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64))
func (_DecryptorsConfigsList *DecryptorsConfigsListSession) GetActiveConfig(activationBlockNumber uint64) (DecryptorsConfig, error) {
	return _DecryptorsConfigsList.Contract.GetActiveConfig(&_DecryptorsConfigsList.CallOpts, activationBlockNumber)
}

// GetActiveConfig is a free data retrieval call binding the contract method 0xb5351b0d.
//
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64))
func (_DecryptorsConfigsList *DecryptorsConfigsListCallerSession) GetActiveConfig(activationBlockNumber uint64) (DecryptorsConfig, error) {
	return _DecryptorsConfigsList.Contract.GetActiveConfig(&_DecryptorsConfigsList.CallOpts, activationBlockNumber)
}

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64))
func (_DecryptorsConfigsList *DecryptorsConfigsListCaller) GetCurrentActiveConfig(opts *bind.CallOpts) (DecryptorsConfig, error) {
	var out []interface{}
	err := _DecryptorsConfigsList.contract.Call(opts, &out, "getCurrentActiveConfig")

	if err != nil {
		return *new(DecryptorsConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(DecryptorsConfig)).(*DecryptorsConfig)

	return out0, err

}

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64))
func (_DecryptorsConfigsList *DecryptorsConfigsListSession) GetCurrentActiveConfig() (DecryptorsConfig, error) {
	return _DecryptorsConfigsList.Contract.GetCurrentActiveConfig(&_DecryptorsConfigsList.CallOpts)
}

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64))
func (_DecryptorsConfigsList *DecryptorsConfigsListCallerSession) GetCurrentActiveConfig() (DecryptorsConfig, error) {
	return _DecryptorsConfigsList.Contract.GetCurrentActiveConfig(&_DecryptorsConfigsList.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DecryptorsConfigsList.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListSession) Owner() (common.Address, error) {
	return _DecryptorsConfigsList.Contract.Owner(&_DecryptorsConfigsList.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_DecryptorsConfigsList *DecryptorsConfigsListCallerSession) Owner() (common.Address, error) {
	return _DecryptorsConfigsList.Contract.Owner(&_DecryptorsConfigsList.CallOpts)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x79f78099.
//
// Solidity: function addNewCfg((uint64,uint64) config) returns()
func (_DecryptorsConfigsList *DecryptorsConfigsListTransactor) AddNewCfg(opts *bind.TransactOpts, config DecryptorsConfig) (*types.Transaction, error) {
	return _DecryptorsConfigsList.contract.Transact(opts, "addNewCfg", config)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x79f78099.
//
// Solidity: function addNewCfg((uint64,uint64) config) returns()
func (_DecryptorsConfigsList *DecryptorsConfigsListSession) AddNewCfg(config DecryptorsConfig) (*types.Transaction, error) {
	return _DecryptorsConfigsList.Contract.AddNewCfg(&_DecryptorsConfigsList.TransactOpts, config)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x79f78099.
//
// Solidity: function addNewCfg((uint64,uint64) config) returns()
func (_DecryptorsConfigsList *DecryptorsConfigsListTransactorSession) AddNewCfg(config DecryptorsConfig) (*types.Transaction, error) {
	return _DecryptorsConfigsList.Contract.AddNewCfg(&_DecryptorsConfigsList.TransactOpts, config)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_DecryptorsConfigsList *DecryptorsConfigsListTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DecryptorsConfigsList.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_DecryptorsConfigsList *DecryptorsConfigsListSession) RenounceOwnership() (*types.Transaction, error) {
	return _DecryptorsConfigsList.Contract.RenounceOwnership(&_DecryptorsConfigsList.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_DecryptorsConfigsList *DecryptorsConfigsListTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _DecryptorsConfigsList.Contract.RenounceOwnership(&_DecryptorsConfigsList.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_DecryptorsConfigsList *DecryptorsConfigsListTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _DecryptorsConfigsList.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_DecryptorsConfigsList *DecryptorsConfigsListSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _DecryptorsConfigsList.Contract.TransferOwnership(&_DecryptorsConfigsList.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_DecryptorsConfigsList *DecryptorsConfigsListTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _DecryptorsConfigsList.Contract.TransferOwnership(&_DecryptorsConfigsList.TransactOpts, newOwner)
}

// DecryptorsConfigsListNewConfigIterator is returned from FilterNewConfig and is used to iterate over the raw logs and unpacked data for NewConfig events raised by the DecryptorsConfigsList contract.
type DecryptorsConfigsListNewConfigIterator struct {
	Event *DecryptorsConfigsListNewConfig // Event containing the contract specifics and raw log

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
func (it *DecryptorsConfigsListNewConfigIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DecryptorsConfigsListNewConfig)
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
		it.Event = new(DecryptorsConfigsListNewConfig)
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
func (it *DecryptorsConfigsListNewConfigIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DecryptorsConfigsListNewConfigIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DecryptorsConfigsListNewConfig represents a NewConfig event raised by the DecryptorsConfigsList contract.
type DecryptorsConfigsListNewConfig struct {
	ActivationBlockNumber uint64
	Index                 uint64
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewConfig is a free log retrieval operation binding the contract event 0xf991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index)
func (_DecryptorsConfigsList *DecryptorsConfigsListFilterer) FilterNewConfig(opts *bind.FilterOpts) (*DecryptorsConfigsListNewConfigIterator, error) {

	logs, sub, err := _DecryptorsConfigsList.contract.FilterLogs(opts, "NewConfig")
	if err != nil {
		return nil, err
	}
	return &DecryptorsConfigsListNewConfigIterator{contract: _DecryptorsConfigsList.contract, event: "NewConfig", logs: logs, sub: sub}, nil
}

// WatchNewConfig is a free log subscription operation binding the contract event 0xf991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index)
func (_DecryptorsConfigsList *DecryptorsConfigsListFilterer) WatchNewConfig(opts *bind.WatchOpts, sink chan<- *DecryptorsConfigsListNewConfig) (event.Subscription, error) {

	logs, sub, err := _DecryptorsConfigsList.contract.WatchLogs(opts, "NewConfig")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DecryptorsConfigsListNewConfig)
				if err := _DecryptorsConfigsList.contract.UnpackLog(event, "NewConfig", log); err != nil {
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

// ParseNewConfig is a log parse operation binding the contract event 0xf991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index)
func (_DecryptorsConfigsList *DecryptorsConfigsListFilterer) ParseNewConfig(log types.Log) (*DecryptorsConfigsListNewConfig, error) {
	event := new(DecryptorsConfigsListNewConfig)
	if err := _DecryptorsConfigsList.contract.UnpackLog(event, "NewConfig", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DecryptorsConfigsListOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the DecryptorsConfigsList contract.
type DecryptorsConfigsListOwnershipTransferredIterator struct {
	Event *DecryptorsConfigsListOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *DecryptorsConfigsListOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DecryptorsConfigsListOwnershipTransferred)
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
		it.Event = new(DecryptorsConfigsListOwnershipTransferred)
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
func (it *DecryptorsConfigsListOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DecryptorsConfigsListOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DecryptorsConfigsListOwnershipTransferred represents a OwnershipTransferred event raised by the DecryptorsConfigsList contract.
type DecryptorsConfigsListOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_DecryptorsConfigsList *DecryptorsConfigsListFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*DecryptorsConfigsListOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _DecryptorsConfigsList.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &DecryptorsConfigsListOwnershipTransferredIterator{contract: _DecryptorsConfigsList.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_DecryptorsConfigsList *DecryptorsConfigsListFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *DecryptorsConfigsListOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _DecryptorsConfigsList.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DecryptorsConfigsListOwnershipTransferred)
				if err := _DecryptorsConfigsList.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_DecryptorsConfigsList *DecryptorsConfigsListFilterer) ParseOwnershipTransferred(log types.Log) (*DecryptorsConfigsListOwnershipTransferred, error) {
	event := new(DecryptorsConfigsListOwnershipTransferred)
	if err := _DecryptorsConfigsList.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeypersConfigsListMetaData contains all meta data concerning the KeypersConfigsList contract.
var KeypersConfigsListMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"_addrsSeq\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"}],\"name\":\"NewConfig\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structKeypersConfig\",\"name\":\"config\",\"type\":\"tuple\"}],\"name\":\"addNewCfg\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"addrsSeq\",\"outputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"}],\"name\":\"getActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structKeypersConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structKeypersConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"keypersConfigs\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50604051610b4d380380610b4d83398101604081905261002f91610230565b610038336101e0565b600280546001600160a01b0319166001600160a01b038316908117909155604051630545a03f60e31b815260006004820152632a2d01f89060240160206040518083038186803b15801561008b57600080fd5b505afa15801561009f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906100c39190610260565b6001600160401b03161561012e5760405162461bcd60e51b815260206004820152602860248201527f4164647273536571206d757374206861766520656d707479206c697374206174604482015267020696e64657820360c41b606482015260840160405180910390fd5b6040805180820182526000808252602080830182815260018054808201825590845293517fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909401805491516001600160401b0390811668010000000000000000026001600160801b0319909316951694909417179092558251818152918201527ff991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7910160405180910390a150610289565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60006020828403121561024257600080fd5b81516001600160a01b038116811461025957600080fd5b9392505050565b60006020828403121561027257600080fd5b81516001600160401b038116811461025957600080fd5b6108b5806102986000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c8063b5351b0d1161005b578063b5351b0d146100eb578063f2fde38b14610125578063f9c72b5314610138578063fc6d0c7e1461014057600080fd5b80634d89eaaf1461008d578063715018a6146100bd57806379f78099146100c75780638da5cb5b146100da575b600080fd5b6002546100a0906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b6100c5610173565b005b6100c56100d53660046106d4565b6101b2565b6000546001600160a01b03166100a0565b6100fe6100f9366004610701565b6104d3565b6040805182516001600160401b0390811682526020938401511692810192909252016100b4565b6100c5610133366004610725565b610592565b6100fe61062d565b61015361014e36600461074e565b61064f565b604080516001600160401b039384168152929091166020830152016100b4565b6000546001600160a01b031633146101a65760405162461bcd60e51b815260040161019d90610767565b60405180910390fd5b6101b06000610684565b565b6000546001600160a01b031633146101dc5760405162461bcd60e51b815260040161019d90610767565b6101ec6040820160208301610701565b6001600160401b0316600260009054906101000a90046001600160a01b03166001600160a01b03166306661abd6040518163ffffffff1660e01b815260040160206040518083038186803b15801561024357600080fd5b505afa158015610257573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061027b919061079c565b6001600160401b0316116102f75760405162461bcd60e51b815260206004820152603a60248201527f4e6f20617070656e6465642073657420696e2073657120636f72726573706f6e60448201527f64696e6720746f20636f6e66696727732073657420696e646578000000000000606482015260840161019d565b6103046020820182610701565b6001600160401b0316600180808054905061031f91906107cf565b8154811061032f5761032f6107e6565b6000918252602090912001546001600160401b031611156103b85760405162461bcd60e51b815260206004820152603860248201527f43616e6e6f7420616464206e6577207365742077697468206c6f77657220626c60448201527f6f636b206e756d626572207468616e2070726576696f75730000000000000000606482015260840161019d565b6103c56020820182610701565b6001600160401b031643111561042f5760405162461bcd60e51b815260206004820152602960248201527f43616e6e6f7420616464206e6577207365742077697468207061737420626c6f60448201526831b590373ab6b132b960b91b606482015260840161019d565b60018054808201825560009190915281907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf60161046c82826107fc565b507ff991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7905061049d6020830183610701565b6104ad6040840160208501610701565b604080516001600160401b0393841681529290911660208301520160405180910390a150565b6040805180820190915260008082526020820152600180546000916104f7916107cf565b90505b826001600160401b031660018281548110610517576105176107e6565b6000918252602090912001546001600160401b0316116105805760018181548110610544576105446107e6565b6000918252602091829020604080518082019091529101546001600160401b038082168352600160401b90910416918101919091529392505050565b8061058a81610868565b9150506104fa565b6000546001600160a01b031633146105bc5760405162461bcd60e51b815260040161019d90610767565b6001600160a01b0381166106215760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b606482015260840161019d565b61062a81610684565b50565b604080518082019091526000808252602082015261064a436104d3565b905090565b6001818154811061065f57600080fd5b6000918252602090912001546001600160401b038082169250600160401b9091041682565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6000604082840312156106e657600080fd5b50919050565b6001600160401b038116811461062a57600080fd5b60006020828403121561071357600080fd5b813561071e816106ec565b9392505050565b60006020828403121561073757600080fd5b81356001600160a01b038116811461071e57600080fd5b60006020828403121561076057600080fd5b5035919050565b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b6000602082840312156107ae57600080fd5b815161071e816106ec565b634e487b7160e01b600052601160045260246000fd5b6000828210156107e1576107e16107b9565b500390565b634e487b7160e01b600052603260045260246000fd5b8135610807816106ec565b6001600160401b03811690508154816001600160401b031982161783556020840135610832816106ec565b6fffffffffffffffff00000000000000008160401b16836fffffffffffffffffffffffffffffffff198416171784555050505050565b600081610877576108776107b9565b50600019019056fea264697066735822122082357d9f3d350e32782f684475d247d5d704440c0b9fbadccef9a135ac13275b64736f6c63430008090033",
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
	parsed, err := abi.JSON(strings.NewReader(KeypersConfigsListABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
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
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64))
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
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64))
func (_KeypersConfigsList *KeypersConfigsListSession) GetActiveConfig(activationBlockNumber uint64) (KeypersConfig, error) {
	return _KeypersConfigsList.Contract.GetActiveConfig(&_KeypersConfigsList.CallOpts, activationBlockNumber)
}

// GetActiveConfig is a free data retrieval call binding the contract method 0xb5351b0d.
//
// Solidity: function getActiveConfig(uint64 activationBlockNumber) view returns((uint64,uint64))
func (_KeypersConfigsList *KeypersConfigsListCallerSession) GetActiveConfig(activationBlockNumber uint64) (KeypersConfig, error) {
	return _KeypersConfigsList.Contract.GetActiveConfig(&_KeypersConfigsList.CallOpts, activationBlockNumber)
}

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64))
func (_KeypersConfigsList *KeypersConfigsListCaller) GetCurrentActiveConfig(opts *bind.CallOpts) (KeypersConfig, error) {
	var out []interface{}
	err := _KeypersConfigsList.contract.Call(opts, &out, "getCurrentActiveConfig")

	if err != nil {
		return *new(KeypersConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(KeypersConfig)).(*KeypersConfig)

	return out0, err

}

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64))
func (_KeypersConfigsList *KeypersConfigsListSession) GetCurrentActiveConfig() (KeypersConfig, error) {
	return _KeypersConfigsList.Contract.GetCurrentActiveConfig(&_KeypersConfigsList.CallOpts)
}

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64))
func (_KeypersConfigsList *KeypersConfigsListCallerSession) GetCurrentActiveConfig() (KeypersConfig, error) {
	return _KeypersConfigsList.Contract.GetCurrentActiveConfig(&_KeypersConfigsList.CallOpts)
}

// KeypersConfigs is a free data retrieval call binding the contract method 0xfc6d0c7e.
//
// Solidity: function keypersConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex)
func (_KeypersConfigsList *KeypersConfigsListCaller) KeypersConfigs(opts *bind.CallOpts, arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}, error) {
	var out []interface{}
	err := _KeypersConfigsList.contract.Call(opts, &out, "keypersConfigs", arg0)

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

// KeypersConfigs is a free data retrieval call binding the contract method 0xfc6d0c7e.
//
// Solidity: function keypersConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex)
func (_KeypersConfigsList *KeypersConfigsListSession) KeypersConfigs(arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
}, error) {
	return _KeypersConfigsList.Contract.KeypersConfigs(&_KeypersConfigsList.CallOpts, arg0)
}

// KeypersConfigs is a free data retrieval call binding the contract method 0xfc6d0c7e.
//
// Solidity: function keypersConfigs(uint256 ) view returns(uint64 activationBlockNumber, uint64 setIndex)
func (_KeypersConfigsList *KeypersConfigsListCallerSession) KeypersConfigs(arg0 *big.Int) (struct {
	ActivationBlockNumber uint64
	SetIndex              uint64
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

// AddNewCfg is a paid mutator transaction binding the contract method 0x79f78099.
//
// Solidity: function addNewCfg((uint64,uint64) config) returns()
func (_KeypersConfigsList *KeypersConfigsListTransactor) AddNewCfg(opts *bind.TransactOpts, config KeypersConfig) (*types.Transaction, error) {
	return _KeypersConfigsList.contract.Transact(opts, "addNewCfg", config)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x79f78099.
//
// Solidity: function addNewCfg((uint64,uint64) config) returns()
func (_KeypersConfigsList *KeypersConfigsListSession) AddNewCfg(config KeypersConfig) (*types.Transaction, error) {
	return _KeypersConfigsList.Contract.AddNewCfg(&_KeypersConfigsList.TransactOpts, config)
}

// AddNewCfg is a paid mutator transaction binding the contract method 0x79f78099.
//
// Solidity: function addNewCfg((uint64,uint64) config) returns()
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
	Index                 uint64
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewConfig is a free log retrieval operation binding the contract event 0xf991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index)
func (_KeypersConfigsList *KeypersConfigsListFilterer) FilterNewConfig(opts *bind.FilterOpts) (*KeypersConfigsListNewConfigIterator, error) {

	logs, sub, err := _KeypersConfigsList.contract.FilterLogs(opts, "NewConfig")
	if err != nil {
		return nil, err
	}
	return &KeypersConfigsListNewConfigIterator{contract: _KeypersConfigsList.contract, event: "NewConfig", logs: logs, sub: sub}, nil
}

// WatchNewConfig is a free log subscription operation binding the contract event 0xf991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index)
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

// ParseNewConfig is a log parse operation binding the contract event 0xf991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index)
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
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"_addrsSeq\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"i\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"a\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"Registered\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"addrsSeq\",\"outputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"a\",\"type\":\"address\"}],\"name\":\"get\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"i\",\"type\":\"uint64\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b506040516107f03803806107f083398101604081905261002f91610054565b600080546001600160a01b0319166001600160a01b0392909216919091179055610084565b60006020828403121561006657600080fd5b81516001600160a01b038116811461007d57600080fd5b9392505050565b61075d806100936000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c80634d89eaaf1461004657806354d77fe414610076578063c2bc2efc1461008b575b600080fd5b600054610059906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b6100896100843660046104f1565b6100ab565b005b61009e6100993660046105db565b61030f565b60405161006d919061064c565b60008054604051631a8a384960e11b815267ffffffffffffffff8087166004830152851660248201526001600160a01b039091169063351470929060440160206040518083038186803b15801561010157600080fd5b505afa158015610115573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610139919061065f565b90506001600160a01b03811633146101985760405162461bcd60e51b815260206004820152601f60248201527f52656769737472793a2073656e646572206973206e6f7420616c6c6f7765640060448201526064015b60405180910390fd5b33600090815260016020526040902080546101b29061067c565b15905061020d5760405162461bcd60e51b815260206004820152602360248201527f52656769737472793a2073656e64657220616c726561647920726567697374656044820152621c995960ea1b606482015260840161018f565b600082511161026c5760405162461bcd60e51b815260206004820152602560248201527f52656769737472793a2063616e6e6f7420726567697374657220656d7074792060448201526476616c756560d81b606482015260840161018f565b336000908152600160209081526040909120835161028c92850190610425565b506102cc6040518060400160405280601e81526020017f526567697374657265642076616c756520666f722073656e6465722025730000815250336103bb565b7f2791b5fcdbb8707a39509d0547670c51cfca994cc587718f55192d409fd514ca8484338560405161030194939291906106b7565b60405180910390a150505050565b6001600160a01b03811660009081526001602052604090208054606091906103369061067c565b80601f01602080910402602001604051908101604052809291908181526020018280546103629061067c565b80156103af5780601f10610384576101008083540402835291602001916103af565b820191906000526020600020905b81548152906001019060200180831161039257829003601f168201915b50505050509050919050565b61040082826040516024016103d19291906106fd565b60408051601f198184030181529190526020810180516001600160e01b031663319af33360e01b179052610404565b5050565b80516a636f6e736f6c652e6c6f67602083016000808483855afa5050505050565b8280546104319061067c565b90600052602060002090601f0160209004810192826104535760008555610499565b82601f1061046c57805160ff1916838001178555610499565b82800160010185558215610499579182015b8281111561049957825182559160200191906001019061047e565b506104a59291506104a9565b5090565b5b808211156104a557600081556001016104aa565b803567ffffffffffffffff811681146104d657600080fd5b919050565b634e487b7160e01b600052604160045260246000fd5b60008060006060848603121561050657600080fd5b61050f846104be565b925061051d602085016104be565b9150604084013567ffffffffffffffff8082111561053a57600080fd5b818601915086601f83011261054e57600080fd5b813581811115610560576105606104db565b604051601f8201601f19908116603f01168101908382118183101715610588576105886104db565b816040528281528960208487010111156105a157600080fd5b8260208601602083013760006020848301015280955050505050509250925092565b6001600160a01b03811681146105d857600080fd5b50565b6000602082840312156105ed57600080fd5b81356105f8816105c3565b9392505050565b6000815180845260005b8181101561062557602081850181015186830182015201610609565b81811115610637576000602083870101525b50601f01601f19169290920160200192915050565b6020815260006105f860208301846105ff565b60006020828403121561067157600080fd5b81516105f8816105c3565b600181811c9082168061069057607f821691505b602082108114156106b157634e487b7160e01b600052602260045260246000fd5b50919050565b67ffffffffffffffff8581168252841660208201526001600160a01b03831660408201526080606082018190526000906106f3908301846105ff565b9695505050505050565b60408152600061071060408301856105ff565b905060018060a01b0383166020830152939250505056fea2646970667358221220b4df45d08d451d230ac2f8252426160ff79577bb22ce99a4f9ed9eed350ae0f464736f6c63430008090033",
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

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_Registry *RegistryCaller) AddrsSeq(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "addrsSeq")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_Registry *RegistrySession) AddrsSeq() (common.Address, error) {
	return _Registry.Contract.AddrsSeq(&_Registry.CallOpts)
}

// AddrsSeq is a free data retrieval call binding the contract method 0x4d89eaaf.
//
// Solidity: function addrsSeq() view returns(address)
func (_Registry *RegistryCallerSession) AddrsSeq() (common.Address, error) {
	return _Registry.Contract.AddrsSeq(&_Registry.CallOpts)
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

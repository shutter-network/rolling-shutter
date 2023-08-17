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
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"i\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"newAddrs\",\"type\":\"address[]\"}],\"name\":\"Added\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"}],\"name\":\"Appended\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"newAddrs\",\"type\":\"address[]\"}],\"name\":\"add\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"append\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"i\",\"type\":\"uint64\"}],\"name\":\"at\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"n\",\"type\":\"uint64\"}],\"name\":\"countNth\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b5061001a33610027565b610022610077565b610150565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080516000602080830182815283850190945292825260018054808201825591528151805192937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909201926100d192849201906100d6565b505050565b82805482825590600052602060002090810192821561012b579160200282015b8281111561012b57825182546001600160a01b0319166001600160a01b039091161782556020909201916001909101906100f6565b5061013792915061013b565b5090565b5b80821115610137576000815560010161013c565b6109a48061015f6000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c80637f353d551161005b5780637f353d55146100fa5780638da5cb5b14610102578063c4c1c94f14610113578063f2fde38b1461012657600080fd5b806306661abd1461008d5780632a2d01f8146100b257806335147092146100c5578063715018a6146100f0575b600080fd5b610095610139565b6040516001600160401b0390911681526020015b60405180910390f35b6100956100c0366004610775565b61014e565b6100d86100d3366004610797565b6101f5565b6040516001600160a01b0390911681526020016100a9565b6100f861033d565b005b6100f8610351565b6000546001600160a01b03166100d8565b6100f86101213660046107ca565b610425565b6100f8610134366004610855565b61055d565b6001805460009161014991610886565b905090565b6000610158610139565b6001600160401b0316826001600160401b0316106101c75760405162461bcd60e51b815260206004820152602160248201527f41646472735365712e636f756e744e74683a206e206f7574206f662072616e676044820152606560f81b60648201526084015b60405180910390fd5b6001826001600160401b0316815481106101e3576101e36108ae565b60009182526020909120015492915050565b60006101ff610139565b6001600160401b0316836001600160401b03161061025f5760405162461bcd60e51b815260206004820152601b60248201527f41646472735365712e61743a206e206f7574206f662072616e6765000000000060448201526064016101be565b6001836001600160401b03168154811061027b5761027b6108ae565b6000918252602090912001546001600160401b038316106102de5760405162461bcd60e51b815260206004820152601b60248201527f41646472735365712e61743a2069206f7574206f662072616e6765000000000060448201526064016101be565b6001836001600160401b0316815481106102fa576102fa6108ae565b90600052602060002001600001826001600160401b031681548110610321576103216108ae565b6000918252602090912001546001600160a01b03169392505050565b6103456105d6565b61034f6000610630565b565b6103596105d6565b61036b60016001600160401b03610886565b6001600160401b0316600180549050106103d35760405162461bcd60e51b815260206004820152602360248201527f41646472735365712e617070656e643a20736571206578636565656473206c696044820152621b5a5d60ea1b60648201526084016101be565b600180547f5ff9c98a1faf73c018d22371cb08c08dec1412825b68523a8e7deaa17683a6b99161040291610886565b6040516001600160401b03909116815260200160405180910390a161034f610680565b61042d6105d6565b6001805460009161043d916108c4565b905060006001826001600160401b03168154811061045d5761045d6108ae565b600091825260208220015491505b6001600160401b038116841115610519576001836001600160401b031681548110610498576104986108ae565b906000526020600020016000018585836001600160401b03168181106104c0576104c06108ae565b90506020020160208101906104d59190610855565b81546001810183556000928352602090922090910180546001600160a01b0319166001600160a01b0390921691909117905580610511816108db565b91505061046b565b507f54a93d30cc356a58fe6fe472b453c3ea842500e17a2e9972af429d866f305fbd8282868660405161054f9493929190610902565b60405180910390a150505050565b6105656105d6565b6001600160a01b0381166105ca5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016101be565b6105d381610630565b50565b6000546001600160a01b0316331461034f5760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e657260448201526064016101be565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080516000602080830182815283850190945292825260018054808201825591528151805192937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909201926106da92849201906106df565b505050565b828054828255906000526020600020908101928215610734579160200282015b8281111561073457825182546001600160a01b0319166001600160a01b039091161782556020909201916001909101906106ff565b50610740929150610744565b5090565b5b808211156107405760008155600101610745565b80356001600160401b038116811461077057600080fd5b919050565b60006020828403121561078757600080fd5b61079082610759565b9392505050565b600080604083850312156107aa57600080fd5b6107b383610759565b91506107c160208401610759565b90509250929050565b600080602083850312156107dd57600080fd5b82356001600160401b03808211156107f457600080fd5b818501915085601f83011261080857600080fd5b81358181111561081757600080fd5b8660208260051b850101111561082c57600080fd5b60209290920196919550909350505050565b80356001600160a01b038116811461077057600080fd5b60006020828403121561086757600080fd5b6107908261083e565b634e487b7160e01b600052601160045260246000fd5b60006001600160401b03838116908316818110156108a6576108a6610870565b039392505050565b634e487b7160e01b600052603260045260246000fd5b6000828210156108d6576108d6610870565b500390565b60006001600160401b03808316818114156108f8576108f8610870565b6001019392505050565b6001600160401b0385811682528416602080830191909152606060408301819052820183905260009084906080840190835b86811015610960576001600160a01b0361094d8561083e565b1683529281019291810191600101610934565b50909897505050505050505056fea26469706673582212209d98fae94cd564134f91b36c5df412f613a3c8cda6984ddefd6d23e5792ec81464736f6c63430008090033",
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
	Bin: "0x608060405234801561001057600080fd5b50610284806100206000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c80631d8c311f14610046578063d09de08a1461005b578063e79993f314610063575b600080fd5b6100596100543660046101e3565b610094565b005b61005961011f565b6000546100779067ffffffffffffffff1681565b60405167ffffffffffffffff909116815260200160405180910390f35b33156100b357604051631448d0ef60e01b815260040160405180910390fd5b6000546040805167ffffffffffffffff928316815291831660208301527f5867f9e83f14fb505a43dd58880b1de7e3b5cddbfa99bb92a15dad48b453410b910160405180910390a16000805467ffffffffffffffff191667ffffffffffffffff92909216919091179055565b331561013e57604051631448d0ef60e01b815260040160405180910390fd5b6000547f5867f9e83f14fb505a43dd58880b1de7e3b5cddbfa99bb92a15dad48b453410b9067ffffffffffffffff16610178816001610214565b6040805167ffffffffffffffff93841681529290911660208301520160405180910390a1600080546001919081906101bb90849067ffffffffffffffff16610214565b92506101000a81548167ffffffffffffffff021916908367ffffffffffffffff160217905550565b6000602082840312156101f557600080fd5b813567ffffffffffffffff8116811461020d57600080fd5b9392505050565b600067ffffffffffffffff80831681851680830382111561024557634e487b7160e01b600052601160045260246000fd5b0194935050505056fea2646970667358221220c2b168f9e118d5fd1df81ee5364a3cc7223131b5250eeab7c5b1dfa2a782466c64736f6c63430008090033",
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
	parsed, err := abi.JSON(strings.NewReader(BatchCounterABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
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
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"_addrsSeq\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"collatorSetIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"collatorConfigIndex\",\"type\":\"uint64\"}],\"name\":\"NewConfig\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structCollatorConfig\",\"name\":\"config\",\"type\":\"tuple\"}],\"name\":\"addNewCfg\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"addrsSeq\",\"outputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"collatorConfigs\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"}],\"name\":\"getActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structCollatorConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50604051610aa4380380610aa483398101604081905261002f9161023b565b610038336101eb565b600280546001600160a01b0319166001600160a01b038316908117909155604051630545a03f60e31b815260006004820152632a2d01f89060240160206040518083038186803b15801561008b57600080fd5b505afa15801561009f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906100c3919061026b565b6001600160401b03161561012e5760405162461bcd60e51b815260206004820152602860248201527f4164647273536571206d757374206861766520656d707479206c697374206174604482015267020696e64657820360c41b606482015260840160405180910390fd5b6040805180820182526000808252602080830182815260018054808201825590845293517fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909401805491516001600160401b0390811668010000000000000000026001600160801b0319909316951694909417179092558251818152918201819052918101919091527ff1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d99060600160405180910390a150610294565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60006020828403121561024d57600080fd5b81516001600160a01b038116811461026457600080fd5b9392505050565b60006020828403121561027d57600080fd5b81516001600160401b038116811461026457600080fd5b610801806102a36000396000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c806379f780991161005b57806379f78099146100ef5780638da5cb5b14610102578063b5351b0d14610113578063f2fde38b1461014d57600080fd5b80634d89eaaf14610082578063715018a6146100b257806377e18fc4146100bc575b600080fd5b600254610095906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b6100ba610160565b005b6100cf6100ca36600461062d565b610174565b604080516001600160401b039384168152929091166020830152016100a9565b6100ba6100fd366004610646565b6101a9565b6000546001600160a01b0316610095565b610126610121366004610673565b61044b565b6040805182516001600160401b0390811682526020938401511692810192909252016100a9565b6100ba61015b366004610697565b61050a565b610168610583565b61017260006105dd565b565b6001818154811061018457600080fd5b6000918252602090912001546001600160401b038082169250600160401b9091041682565b6101b1610583565b6101c16040820160208301610673565b6001600160401b0316600260009054906101000a90046001600160a01b03166001600160a01b03166306661abd6040518163ffffffff1660e01b815260040160206040518083038186803b15801561021857600080fd5b505afa15801561022c573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061025091906106c0565b6001600160401b0316116102d15760405162461bcd60e51b815260206004820152603a60248201527f4e6f20617070656e6465642073657420696e2073657120636f72726573706f6e60448201527f64696e6720746f20636f6e66696727732073657420696e64657800000000000060648201526084015b60405180910390fd5b6102de6020820182610673565b6001600160401b031660018080805490506102f991906106f3565b815481106103095761030961070a565b6000918252602090912001546001600160401b031611156103925760405162461bcd60e51b815260206004820152603860248201527f43616e6e6f7420616464206e6577207365742077697468206c6f77657220626c60448201527f6f636b206e756d626572207468616e2070726576696f7573000000000000000060648201526084016102c8565b60018054808201825560009190915281907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6016103cf8282610720565b507ff1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d990506104006020830183610673565b6104106040840160208501610673565b6001805461041e919061078c565b604080516001600160401b039485168152928416602084015292168183015290519081900360600190a150565b60408051808201909152600080825260208201526001805460009161046f916106f3565b90505b826001600160401b03166001828154811061048f5761048f61070a565b6000918252602090912001546001600160401b0316116104f857600181815481106104bc576104bc61070a565b6000918252602091829020604080518082019091529101546001600160401b038082168352600160401b90910416918101919091529392505050565b80610502816107b4565b915050610472565b610512610583565b6001600160a01b0381166105775760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016102c8565b610580816105dd565b50565b6000546001600160a01b031633146101725760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e657260448201526064016102c8565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60006020828403121561063f57600080fd5b5035919050565b60006040828403121561065857600080fd5b50919050565b6001600160401b038116811461058057600080fd5b60006020828403121561068557600080fd5b81356106908161065e565b9392505050565b6000602082840312156106a957600080fd5b81356001600160a01b038116811461069057600080fd5b6000602082840312156106d257600080fd5b81516106908161065e565b634e487b7160e01b600052601160045260246000fd5b600082821015610705576107056106dd565b500390565b634e487b7160e01b600052603260045260246000fd5b813561072b8161065e565b6001600160401b03811690508154816001600160401b0319821617835560208401356107568161065e565b6fffffffffffffffff00000000000000008160401b16836fffffffffffffffffffffffffffffffff198416171784555050505050565b60006001600160401b03838116908316818110156107ac576107ac6106dd565b039392505050565b6000816107c3576107c36106dd565b50600019019056fea2646970667358221220f9c6d8d9790167817293c88109470375d1135399284fc25c38cac36144de4afa64736f6c63430008090033",
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
	parsed, err := abi.JSON(strings.NewReader(CollatorConfigsListABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
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

// EonKeyStorageMetaData contains all meta data concerning the EonKeyStorage contract.
var EonKeyStorageMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"blockNumber\",\"type\":\"uint64\"}],\"name\":\"NotFound\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"key\",\"type\":\"bytes\"}],\"name\":\"Inserted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"blockNumber\",\"type\":\"uint64\"}],\"name\":\"get\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"serializedKey\",\"type\":\"bytes\"},{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"}],\"name\":\"insert\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"keys\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nextIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes\",\"name\":\"key\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"num\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040523480156200001157600080fd5b506200001d336200004c565b60606200002e60008260016200009c565b50620000446001600160401b038260006200009c565b505062000314565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60018054604080516060810182526001600160401b0380881682528581166020808401918252938301888152858701875560009687528351600287027fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6810180549451861668010000000000000000026001600160801b031990951692909516919091179290921783555180519394929362000162937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf7909301929190910190620001bd565b5050506002816001600160401b031610620001b5577f2f64d9497c8c677c995d99bcc930463dca07bfc5906e28140cbfa4222ddf402c858286604051620001ac9392919062000263565b60405180910390a15b949350505050565b828054620001cb90620002d7565b90600052602060002090601f016020900481019282620001ef57600085556200023a565b82601f106200020a57805160ff19168380011785556200023a565b828001600101855582156200023a579182015b828111156200023a5782518255916020019190600101906200021d565b50620002489291506200024c565b5090565b5b808211156200024857600081556001016200024d565b600060018060401b038086168352602081861681850152606060408501528451915081606085015260005b82811015620002ac578581018201518582016080015281016200028e565b82811115620002bf576000608084870101525b5050601f01601f191691909101608001949350505050565b600181811c90821680620002ec57607f821691505b602082108114156200030e57634e487b7160e01b600052602260045260246000fd5b50919050565b610c1080620003246000396000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c8063715018a61161005b578063715018a6146100e25780638da5cb5b146100ea578063ada8679814610105578063f2fde38b1461012557600080fd5b80630cb6aaf1146100825780633f5fafa4146100ad5780634e70b1dc146100c2575b600080fd5b610095610090366004610967565b610138565b6040516100a4939291906109cd565b60405180910390f35b6100c06100bb366004610a34565b610206565b005b6100ca6104cf565b6040516001600160401b0390911681526020016100a4565b6100c06104e6565b6000546040516001600160a01b0390911681526020016100a4565b610118610113366004610af5565b6104fa565b6040516100a49190610b17565b6100c0610133366004610b2a565b610694565b6001818154811061014857600080fd5b6000918252602090912060029091020180546001820180546001600160401b038084169550600160401b909304909216929161018390610b53565b80601f01602080910402602001604051908101604052809291908181526020018280546101af90610b53565b80156101fc5780601f106101d1576101008083540402835291602001916101fc565b820191906000526020600020905b8154815290600101906020018083116101df57829003601f168201915b5050505050905083565b61020e61070d565b6000806001905060006001826001600160401b03168154811061023357610233610b8e565b600091825260209182902060408051606081018252600290930290910180546001600160401b038082168552600160401b90910416938301939093526001830180549293929184019161028590610b53565b80601f01602080910402602001604051908101604052809291908181526020018280546102b190610b53565b80156102fe5780601f106102d3576101008083540402835291602001916102fe565b820191906000526020600020905b8154815290600101906020018083116102e157829003601f168201915b50505050508152505090505b6000600182602001516001600160401b03168154811061032c5761032c610b8e565b600091825260209182902060408051606081018252600290930290910180546001600160401b038082168552600160401b90910416938301939093526001830180549293929184019161037e90610b53565b80601f01602080910402602001604051908101604052809291908181526020018280546103aa90610b53565b80156103f75780601f106103cc576101008083540402835291602001916103f7565b820191906000526020600020905b8154815290600101906020018083116103da57829003601f168201915b5050505050815250509050846001600160401b031681600001516001600160401b0316116104c25761042e85878460200151610767565b6001600160401b0380821660208501526001805492965084929091861690811061045a5761045a610b8e565b600091825260209182902083516002909202018054848401516001600160401b03908116600160401b026001600160801b03199092169316929092179190911781556040830151805191926104b7926001850192909101906108ce565b505050505050505050565b602090910151915061030a565b6001546000906104e190600290610ba4565b905090565b6104ee61070d565b6104f8600061087e565b565b606060006001808154811061051157610511610b8e565b6000918252602090912060029091020154600160401b90046001600160401b031690505b6001600160401b0381161561066b5760006001826001600160401b03168154811061056257610562610b8e565b600091825260209182902060408051606081018252600290930290910180546001600160401b038082168552600160401b9091041693830193909352600183018054929392918401916105b490610b53565b80601f01602080910402602001604051908101604052809291908181526020018280546105e090610b53565b801561062d5780601f106106025761010080835404028352916020019161062d565b820191906000526020600020905b81548152906001019060200180831161061057829003601f168201915b5050505050815250509050836001600160401b031681600001516001600160401b03161161066057604001519392505050565b602001519050610535565b604051636be0ee8760e01b81526001600160401b03841660048201526024015b60405180910390fd5b61069c61070d565b6001600160a01b0381166107015760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b606482015260840161068b565b61070a8161087e565b50565b6000546001600160a01b031633146104f85760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015260640161068b565b60018054604080516060810182526001600160401b0380881682528581166020808401918252938301888152858701875560009687528351600287027fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf68101805494518616600160401b026001600160801b0319909516929095169190911792909217835551805193949293610826937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf79093019291909101906108ce565b5050506002816001600160401b031610610876577f2f64d9497c8c677c995d99bcc930463dca07bfc5906e28140cbfa4222ddf402c85828660405161086d939291906109cd565b60405180910390a15b949350505050565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b8280546108da90610b53565b90600052602060002090601f0160209004810192826108fc5760008555610942565b82601f1061091557805160ff1916838001178555610942565b82800160010185558215610942579182015b82811115610942578251825591602001919060010190610927565b5061094e929150610952565b5090565b5b8082111561094e5760008155600101610953565b60006020828403121561097957600080fd5b5035919050565b6000815180845260005b818110156109a65760208185018101518683018201520161098a565b818111156109b8576000602083870101525b50601f01601f19169290920160200192915050565b60006001600160401b038086168352808516602084015250606060408301526109f96060830184610980565b95945050505050565b634e487b7160e01b600052604160045260246000fd5b80356001600160401b0381168114610a2f57600080fd5b919050565b60008060408385031215610a4757600080fd5b82356001600160401b0380821115610a5e57600080fd5b818501915085601f830112610a7257600080fd5b813581811115610a8457610a84610a02565b604051601f8201601f19908116603f01168101908382118183101715610aac57610aac610a02565b81604052828152886020848701011115610ac557600080fd5b826020860160208301376000602084830101528096505050505050610aec60208401610a18565b90509250929050565b600060208284031215610b0757600080fd5b610b1082610a18565b9392505050565b602081526000610b106020830184610980565b600060208284031215610b3c57600080fd5b81356001600160a01b0381168114610b1057600080fd5b600181811c90821680610b6757607f821691505b60208210811415610b8857634e487b7160e01b600052602260045260246000fd5b50919050565b634e487b7160e01b600052603260045260246000fd5b60006001600160401b0383811690831681811015610bd257634e487b7160e01b600052601160045260246000fd5b03939250505056fea26469706673582212204530dcb90637cc9cf4d00decc61e42bd8bb518b7908b8b77bb730ff91ac6b8a764736f6c63430008090033",
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
	parsed, err := abi.JSON(strings.NewReader(EonKeyStorageABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
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
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"_addrsSeq\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"keyperConfigIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"name\":\"NewConfig\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"internalType\":\"structKeypersConfig\",\"name\":\"config\",\"type\":\"tuple\"}],\"name\":\"addNewCfg\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"addrsSeq\",\"outputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"}],\"name\":\"getActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"internalType\":\"structKeypersConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"keypersConfigs\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50604051610d6d380380610d6d83398101604081905261002f91610266565b61003833610216565b600280546001600160a01b0319166001600160a01b038316908117909155604051630545a03f60e31b815260006004820152632a2d01f89060240160206040518083038186803b15801561008b57600080fd5b505afa15801561009f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906100c39190610296565b6001600160401b03161561012e5760405162461bcd60e51b815260206004820152602860248201527f4164647273536571206d757374206861766520656d707479206c697374206174604482015267020696e64657820360c41b606482015260840160405180910390fd5b60408051606080820183526000808352602080840182815284860183815260018054808201825590855295517fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf69096018054925191516001600160401b03908116600160801b02600160801b600160c01b031993821668010000000000000000026001600160801b0319909516919098161792909217169490941790935583518181529283018190529282018390528101919091527f97f0f7a37d08d48af6a5f7140aedcc4fa60e92ee1d0607f2d4868c8fc518cc0e9060800160405180910390a1506102bf565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60006020828403121561027857600080fd5b81516001600160a01b038116811461028f57600080fd5b9392505050565b6000602082840312156102a857600080fd5b81516001600160401b038116811461028f57600080fd5b610a9f806102ce6000396000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c80638da5cb5b1161005b5780638da5cb5b146100cf578063b5351b0d146100e0578063f2fde38b14610125578063fc6d0c7e1461013857600080fd5b806333419af5146100825780634d89eaaf14610097578063715018a6146100c7575b600080fd5b610095610090366004610896565b610175565b005b6002546100aa906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b610095610649565b6000546001600160a01b03166100aa565b6100f36100ee3660046108c3565b61065d565b6040805182516001600160401b03908116825260208085015182169083015292820151909216908201526060016100be565b6100956101333660046108e7565b610735565b61014b610146366004610910565b6107ae565b604080516001600160401b03948516815292841660208401529216918101919091526060016100be565b61017d6107ec565b61018d60408201602083016108c3565b6001600160401b0316600260009054906101000a90046001600160a01b03166001600160a01b03166306661abd6040518163ffffffff1660e01b815260040160206040518083038186803b1580156101e457600080fd5b505afa1580156101f8573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061021c9190610929565b6001600160401b03161161029d5760405162461bcd60e51b815260206004820152603a60248201527f4e6f20617070656e6465642073657420696e2073657120636f72726573706f6e60448201527f64696e6720746f20636f6e66696727732073657420696e64657800000000000060648201526084015b60405180910390fd5b6102aa60208201826108c3565b6001600160401b031660018080805490506102c5919061095c565b815481106102d5576102d5610973565b6000918252602090912001546001600160401b0316111561035e5760405162461bcd60e51b815260206004820152603860248201527f43616e6e6f7420616464206e6577207365742077697468206c6f77657220626c60448201527f6f636b206e756d626572207468616e2070726576696f757300000000000000006064820152608401610294565b6002546000906001600160a01b0316632a2d01f861038260408501602086016108c3565b6040516001600160e01b031960e084901b1681526001600160401b03909116600482015260240160206040518083038186803b1580156103c157600080fd5b505afa1580156103d5573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103f99190610929565b90506001600160401b03811661048a5761041960608301604084016108c3565b6001600160401b0316156104855760405162461bcd60e51b815260206004820152602d60248201527f5468726573686f6c64206d757374206265207a65726f206966206b657970657260448201526c2073657420697320656d70747960981b6064820152608401610294565b610576565b600161049c60608401604085016108c3565b6001600160401b031610156104f35760405162461bcd60e51b815260206004820152601e60248201527f5468726573686f6c64206d757374206265206174206c65617374206f6e6500006044820152606401610294565b6001600160401b03811661050d60608401604085016108c3565b6001600160401b031611156105765760405162461bcd60e51b815260206004820152602960248201527f5468726573686f6c64206d757374206e6f7420657863656564206b6579706572604482015268207365742073697a6560b81b6064820152608401610294565b60018054808201825560009190915282907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6016105b38282610989565b507f97f0f7a37d08d48af6a5f7140aedcc4fa60e92ee1d0607f2d4868c8fc518cc0e90506105e460208401846108c3565b6105f460408501602086016108c3565b600180546106029190610a2a565b61061260608701604088016108c3565b604080516001600160401b039586168152938516602085015291841683830152909216606082015290519081900360800190a15050565b6106516107ec565b61065b6000610846565b565b60408051606081018252600080825260208201819052918101919091526001805460009161068a9161095c565b90505b826001600160401b0316600182815481106106aa576106aa610973565b6000918252602090912001546001600160401b03161161072357600181815481106106d7576106d7610973565b60009182526020918290206040805160608101825291909201546001600160401b038082168352600160401b8204811694830194909452600160801b9004909216908201529392505050565b8061072d81610a52565b91505061068d565b61073d6107ec565b6001600160a01b0381166107a25760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b6064820152608401610294565b6107ab81610846565b50565b600181815481106107be57600080fd5b6000918252602090912001546001600160401b038082169250600160401b8204811691600160801b90041683565b6000546001600160a01b0316331461065b5760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e65726044820152606401610294565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6000606082840312156108a857600080fd5b50919050565b6001600160401b03811681146107ab57600080fd5b6000602082840312156108d557600080fd5b81356108e0816108ae565b9392505050565b6000602082840312156108f957600080fd5b81356001600160a01b03811681146108e057600080fd5b60006020828403121561092257600080fd5b5035919050565b60006020828403121561093b57600080fd5b81516108e0816108ae565b634e487b7160e01b600052601160045260246000fd5b60008282101561096e5761096e610946565b500390565b634e487b7160e01b600052603260045260246000fd5b8135610994816108ae565b6001600160401b03811690508154816001600160401b0319821617835560208401356109bf816108ae565b6fffffffffffffffff0000000000000000604091821b166fffffffffffffffffffffffffffffffff198316841781178555908501356109fd816108ae565b6001600160c01b0319929092169092179190911760809190911b67ffffffffffffffff60801b1617905550565b60006001600160401b0383811690831681811015610a4a57610a4a610946565b039392505050565b600081610a6157610a61610946565b50600019019056fea2646970667358221220fbc5c07865998bb3513eee9f53ca9da03fbf349050ccc40239f66b0cbfc4a30064736f6c63430008090033",
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

// ConsoleMetaData contains all meta data concerning the Console contract.
var ConsoleMetaData = &bind.MetaData{
	ABI: "[]",
	Bin: "0x60566037600b82828239805160001a607314602a57634e487b7160e01b600052600060045260246000fd5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea264697066735822122012fa3fc355ce3e13dd7f4420ddf7c18bb0c5216a87dce236e8ee669453b117a564736f6c63430008090033",
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

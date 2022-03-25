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
	Bin: "0x608060405234801561001057600080fd5b5061001a33610027565b610022610077565b610150565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080516000602080830182815283850190945292825260018054808201825591528151805192937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909201926100d192849201906100d6565b505050565b82805482825590600052602060002090810192821561012b579160200282015b8281111561012b57825182546001600160a01b0319166001600160a01b039091161782556020909201916001909101906100f6565b5061013792915061013b565b5090565b5b80821115610137576000815560010161013c565b610a078061015f6000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c80637f353d551161005b5780637f353d55146100fa5780638da5cb5b14610102578063c4c1c94f14610113578063f2fde38b1461012657600080fd5b806306661abd1461008d5780632a2d01f8146100b257806335147092146100c5578063715018a6146100f0575b600080fd5b610095610139565b6040516001600160401b0390911681526020015b60405180910390f35b6100956100c03660046107a3565b61014e565b6100d86100d33660046107c5565b6101f5565b6040516001600160a01b0390911681526020016100a9565b6100f861033d565b005b6100f8610373565b6000546001600160a01b03166100d8565b6100f86101213660046107f8565b610469565b6100f8610134366004610883565b6105c3565b60018054600091610149916108b4565b905090565b6000610158610139565b6001600160401b0316826001600160401b0316106101c75760405162461bcd60e51b815260206004820152602160248201527f41646472735365712e636f756e744e74683a206e206f7574206f662072616e676044820152606560f81b60648201526084015b60405180910390fd5b6001826001600160401b0316815481106101e3576101e36108dc565b60009182526020909120015492915050565b60006101ff610139565b6001600160401b0316836001600160401b03161061025f5760405162461bcd60e51b815260206004820152601b60248201527f41646472735365712e61743a206e206f7574206f662072616e6765000000000060448201526064016101be565b6001836001600160401b03168154811061027b5761027b6108dc565b6000918252602090912001546001600160401b038316106102de5760405162461bcd60e51b815260206004820152601b60248201527f41646472735365712e61743a2069206f7574206f662072616e6765000000000060448201526064016101be565b6001836001600160401b0316815481106102fa576102fa6108dc565b90600052602060002001600001826001600160401b031681548110610321576103216108dc565b6000918252602090912001546001600160a01b03169392505050565b6000546001600160a01b031633146103675760405162461bcd60e51b81526004016101be906108f2565b610371600061065e565b565b6000546001600160a01b0316331461039d5760405162461bcd60e51b81526004016101be906108f2565b6103af60016001600160401b036108b4565b6001600160401b0316600180549050106104175760405162461bcd60e51b815260206004820152602360248201527f41646472735365712e617070656e643a20736571206578636565656473206c696044820152621b5a5d60ea1b60648201526084016101be565b600180547f5ff9c98a1faf73c018d22371cb08c08dec1412825b68523a8e7deaa17683a6b991610446916108b4565b6040516001600160401b03909116815260200160405180910390a16103716106ae565b6000546001600160a01b031633146104935760405162461bcd60e51b81526004016101be906108f2565b600180546000916104a391610927565b905060006001826001600160401b0316815481106104c3576104c36108dc565b600091825260208220015491505b6001600160401b03811684111561057f576001836001600160401b0316815481106104fe576104fe6108dc565b906000526020600020016000018585836001600160401b0316818110610526576105266108dc565b905060200201602081019061053b9190610883565b81546001810183556000928352602090922090910180546001600160a01b0319166001600160a01b03909216919091179055806105778161093e565b9150506104d1565b507f54a93d30cc356a58fe6fe472b453c3ea842500e17a2e9972af429d866f305fbd828286866040516105b59493929190610965565b60405180910390a150505050565b6000546001600160a01b031633146105ed5760405162461bcd60e51b81526004016101be906108f2565b6001600160a01b0381166106525760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016101be565b61065b8161065e565b50565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080516000602080830182815283850190945292825260018054808201825591528151805192937fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf690920192610708928492019061070d565b505050565b828054828255906000526020600020908101928215610762579160200282015b8281111561076257825182546001600160a01b0319166001600160a01b0390911617825560209092019160019091019061072d565b5061076e929150610772565b5090565b5b8082111561076e5760008155600101610773565b80356001600160401b038116811461079e57600080fd5b919050565b6000602082840312156107b557600080fd5b6107be82610787565b9392505050565b600080604083850312156107d857600080fd5b6107e183610787565b91506107ef60208401610787565b90509250929050565b6000806020838503121561080b57600080fd5b82356001600160401b038082111561082257600080fd5b818501915085601f83011261083657600080fd5b81358181111561084557600080fd5b8660208260051b850101111561085a57600080fd5b60209290920196919550909350505050565b80356001600160a01b038116811461079e57600080fd5b60006020828403121561089557600080fd5b6107be8261086c565b634e487b7160e01b600052601160045260246000fd5b60006001600160401b03838116908316818110156108d4576108d461089e565b039392505050565b634e487b7160e01b600052603260045260246000fd5b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b6000828210156109395761093961089e565b500390565b60006001600160401b038083168181141561095b5761095b61089e565b6001019392505050565b6001600160401b0385811682528416602080830191909152606060408301819052820183905260009084906080840190835b868110156109c3576001600160a01b036109b08561086c565b1683529281019291810191600101610997565b50909897505050505050505056fea2646970667358221220f66a508257ab16305dfe8b5c3dc831c0ece428ec0b62f0b3ac04a51175947b5064736f6c63430008090033",
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

// CollatorConfigsListMetaData contains all meta data concerning the CollatorConfigsList contract.
var CollatorConfigsListMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"_addrsSeq\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"}],\"name\":\"NewConfig\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structCollatorConfig\",\"name\":\"config\",\"type\":\"tuple\"}],\"name\":\"addNewCfg\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"addrsSeq\",\"outputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"collatorConfigs\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"}],\"name\":\"getActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structCollatorConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"}],\"internalType\":\"structCollatorConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50604051610b4d380380610b4d83398101604081905261002f91610230565b610038336101e0565b600280546001600160a01b0319166001600160a01b038316908117909155604051630545a03f60e31b815260006004820152632a2d01f89060240160206040518083038186803b15801561008b57600080fd5b505afa15801561009f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906100c39190610260565b6001600160401b03161561012e5760405162461bcd60e51b815260206004820152602860248201527f4164647273536571206d757374206861766520656d707479206c697374206174604482015267020696e64657820360c41b606482015260840160405180910390fd5b6040805180820182526000808252602080830182815260018054808201825590845293517fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6909401805491516001600160401b0390811668010000000000000000026001600160801b0319909316951694909417179092558251818152918201527ff991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7910160405180910390a150610289565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60006020828403121561024257600080fd5b81516001600160a01b038116811461025957600080fd5b9392505050565b60006020828403121561027257600080fd5b81516001600160401b038116811461025957600080fd5b6108b5806102986000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c80638da5cb5b1161005b5780638da5cb5b1461010d578063b5351b0d1461011e578063f2fde38b14610158578063f9c72b531461016b57600080fd5b80634d89eaaf1461008d578063715018a6146100bd57806377e18fc4146100c757806379f78099146100fa575b600080fd5b6002546100a0906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b6100c5610173565b005b6100da6100d53660046106d4565b6101b2565b604080516001600160401b039384168152929091166020830152016100b4565b6100c56101083660046106ed565b6101e7565b6000546001600160a01b03166100a0565b61013161012c36600461071a565b610508565b6040805182516001600160401b0390811682526020938401511692810192909252016100b4565b6100c561016636600461073e565b6105c7565b610131610662565b6000546001600160a01b031633146101a65760405162461bcd60e51b815260040161019d90610767565b60405180910390fd5b6101b06000610684565b565b600181815481106101c257600080fd5b6000918252602090912001546001600160401b038082169250600160401b9091041682565b6000546001600160a01b031633146102115760405162461bcd60e51b815260040161019d90610767565b610221604082016020830161071a565b6001600160401b0316600260009054906101000a90046001600160a01b03166001600160a01b03166306661abd6040518163ffffffff1660e01b815260040160206040518083038186803b15801561027857600080fd5b505afa15801561028c573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906102b0919061079c565b6001600160401b03161161032c5760405162461bcd60e51b815260206004820152603a60248201527f4e6f20617070656e6465642073657420696e2073657120636f72726573706f6e60448201527f64696e6720746f20636f6e66696727732073657420696e646578000000000000606482015260840161019d565b610339602082018261071a565b6001600160401b0316600180808054905061035491906107cf565b81548110610364576103646107e6565b6000918252602090912001546001600160401b031611156103ed5760405162461bcd60e51b815260206004820152603860248201527f43616e6e6f7420616464206e6577207365742077697468206c6f77657220626c60448201527f6f636b206e756d626572207468616e2070726576696f75730000000000000000606482015260840161019d565b6103fa602082018261071a565b6001600160401b03164311156104645760405162461bcd60e51b815260206004820152602960248201527f43616e6e6f7420616464206e6577207365742077697468207061737420626c6f60448201526831b590373ab6b132b960b91b606482015260840161019d565b60018054808201825560009190915281907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6016104a182826107fc565b507ff991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a790506104d2602083018361071a565b6104e2604084016020850161071a565b604080516001600160401b0393841681529290911660208301520160405180910390a150565b60408051808201909152600080825260208201526001805460009161052c916107cf565b90505b826001600160401b03166001828154811061054c5761054c6107e6565b6000918252602090912001546001600160401b0316116105b55760018181548110610579576105796107e6565b6000918252602091829020604080518082019091529101546001600160401b038082168352600160401b90910416918101919091529392505050565b806105bf81610868565b91505061052f565b6000546001600160a01b031633146105f15760405162461bcd60e51b815260040161019d90610767565b6001600160a01b0381166106565760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b606482015260840161019d565b61065f81610684565b50565b604080518082019091526000808252602082015261067f43610508565b905090565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6000602082840312156106e657600080fd5b5035919050565b6000604082840312156106ff57600080fd5b50919050565b6001600160401b038116811461065f57600080fd5b60006020828403121561072c57600080fd5b813561073781610705565b9392505050565b60006020828403121561075057600080fd5b81356001600160a01b038116811461073757600080fd5b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b6000602082840312156107ae57600080fd5b815161073781610705565b634e487b7160e01b600052601160045260246000fd5b6000828210156107e1576107e16107b9565b500390565b634e487b7160e01b600052603260045260246000fd5b813561080781610705565b6001600160401b03811690508154816001600160401b03198216178355602084013561083281610705565b6fffffffffffffffff00000000000000008160401b16836fffffffffffffffffffffffffffffffff198416171784555050505050565b600081610877576108776107b9565b50600019019056fea26469706673582212206b9a4fccb460a65b78c10a453dc64b303bd1d42c860d1baafb3d0c345ac4445664736f6c63430008090033",
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

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64))
func (_CollatorConfigsList *CollatorConfigsListCaller) GetCurrentActiveConfig(opts *bind.CallOpts) (CollatorConfig, error) {
	var out []interface{}
	err := _CollatorConfigsList.contract.Call(opts, &out, "getCurrentActiveConfig")

	if err != nil {
		return *new(CollatorConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(CollatorConfig)).(*CollatorConfig)

	return out0, err

}

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64))
func (_CollatorConfigsList *CollatorConfigsListSession) GetCurrentActiveConfig() (CollatorConfig, error) {
	return _CollatorConfigsList.Contract.GetCurrentActiveConfig(&_CollatorConfigsList.CallOpts)
}

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64))
func (_CollatorConfigsList *CollatorConfigsListCallerSession) GetCurrentActiveConfig() (CollatorConfig, error) {
	return _CollatorConfigsList.Contract.GetCurrentActiveConfig(&_CollatorConfigsList.CallOpts)
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
	Index                 uint64
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewConfig is a free log retrieval operation binding the contract event 0xf991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index)
func (_CollatorConfigsList *CollatorConfigsListFilterer) FilterNewConfig(opts *bind.FilterOpts) (*CollatorConfigsListNewConfigIterator, error) {

	logs, sub, err := _CollatorConfigsList.contract.FilterLogs(opts, "NewConfig")
	if err != nil {
		return nil, err
	}
	return &CollatorConfigsListNewConfigIterator{contract: _CollatorConfigsList.contract, event: "NewConfig", logs: logs, sub: sub}, nil
}

// WatchNewConfig is a free log subscription operation binding the contract event 0xf991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index)
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

// ParseNewConfig is a log parse operation binding the contract event 0xf991c74e88b00b8de409caf790045f133e9a8283d3b989db88e2b2d93612c3a7.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index)
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

// KeypersConfigsListMetaData contains all meta data concerning the KeypersConfigsList contract.
var KeypersConfigsListMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"_addrsSeq\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"name\":\"NewConfig\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"internalType\":\"structKeypersConfig\",\"name\":\"config\",\"type\":\"tuple\"}],\"name\":\"addNewCfg\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"addrsSeq\",\"outputs\":[{\"internalType\":\"contractAddrsSeq\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"}],\"name\":\"getActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"internalType\":\"structKeypersConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentActiveConfig\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"internalType\":\"structKeypersConfig\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"keypersConfigs\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"activationBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"setIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"threshold\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50604051610e21380380610e2183398101604081905261002f9161025f565b6100383361020f565b600280546001600160a01b0319166001600160a01b038316908117909155604051630545a03f60e31b815260006004820152632a2d01f89060240160206040518083038186803b15801561008b57600080fd5b505afa15801561009f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906100c3919061028f565b6001600160401b03161561012e5760405162461bcd60e51b815260206004820152602860248201527f4164647273536571206d757374206861766520656d707479206c697374206174604482015267020696e64657820360c41b606482015260840160405180910390fd5b60408051606080820183526000808352602080840182815284860183815260018054808201825590855295517fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf69096018054925191516001600160401b03908116600160801b02600160801b600160c01b031993821668010000000000000000026001600160801b031990951691909816179290921716949094179093558351818152928301819052928201929092527ff1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d9910160405180910390a1506102b8565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60006020828403121561027157600080fd5b81516001600160a01b038116811461028857600080fd5b9392505050565b6000602082840312156102a157600080fd5b81516001600160401b038116811461028857600080fd5b610b5a806102c76000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c8063b5351b0d1161005b578063b5351b0d146100eb578063f2fde38b14610130578063f9c72b5314610143578063fc6d0c7e1461014b57600080fd5b806333419af51461008d5780634d89eaaf146100a2578063715018a6146100d25780638da5cb5b146100da575b600080fd5b6100a061009b366004610944565b610188565b005b6002546100b5906001600160a01b031681565b6040516001600160a01b0390911681526020015b60405180910390f35b6100a06106e2565b6000546001600160a01b03166100b5565b6100fe6100f9366004610971565b610718565b6040805182516001600160401b03908116825260208085015182169083015292820151909216908201526060016100c9565b6100a061013e366004610995565b6107f0565b6100fe61088b565b61015e6101593660046109be565b6108b6565b604080516001600160401b03948516815292841660208401529216918101919091526060016100c9565b6000546001600160a01b031633146101bb5760405162461bcd60e51b81526004016101b2906109d7565b60405180910390fd5b6101cb6040820160208301610971565b6001600160401b0316600260009054906101000a90046001600160a01b03166001600160a01b03166306661abd6040518163ffffffff1660e01b815260040160206040518083038186803b15801561022257600080fd5b505afa158015610236573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061025a9190610a0c565b6001600160401b0316116102d65760405162461bcd60e51b815260206004820152603a60248201527f4e6f20617070656e6465642073657420696e2073657120636f72726573706f6e60448201527f64696e6720746f20636f6e66696727732073657420696e64657800000000000060648201526084016101b2565b6102e36020820182610971565b6001600160401b031660018080805490506102fe9190610a3f565b8154811061030e5761030e610a56565b6000918252602090912001546001600160401b031611156103975760405162461bcd60e51b815260206004820152603860248201527f43616e6e6f7420616464206e6577207365742077697468206c6f77657220626c60448201527f6f636b206e756d626572207468616e2070726576696f7573000000000000000060648201526084016101b2565b6103a46020820182610971565b6001600160401b031643111561040e5760405162461bcd60e51b815260206004820152602960248201527f43616e6e6f7420616464206e6577207365742077697468207061737420626c6f60448201526831b590373ab6b132b960b91b60648201526084016101b2565b6002546000906001600160a01b0316632a2d01f86104326040850160208601610971565b6040516001600160e01b031960e084901b1681526001600160401b03909116600482015260240160206040518083038186803b15801561047157600080fd5b505afa158015610485573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104a99190610a0c565b90506001600160401b03811661053a576104c96060830160408401610971565b6001600160401b0316156105355760405162461bcd60e51b815260206004820152602d60248201527f5468726573686f6c64206d757374206265207a65726f206966206b657970657260448201526c2073657420697320656d70747960981b60648201526084016101b2565b610626565b600161054c6060840160408501610971565b6001600160401b031610156105a35760405162461bcd60e51b815260206004820152601e60248201527f5468726573686f6c64206d757374206265206174206c65617374206f6e65000060448201526064016101b2565b6001600160401b0381166105bd6060840160408501610971565b6001600160401b031611156106265760405162461bcd60e51b815260206004820152602960248201527f5468726573686f6c64206d757374206e6f7420657863656564206b6579706572604482015268207365742073697a6560b81b60648201526084016101b2565b60018054808201825560009190915282907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6016106638282610a6c565b507ff1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d990506106946020840184610971565b6106a46040850160208601610971565b6106b46060860160408701610971565b604080516001600160401b039485168152928416602084015292168183015290519081900360600190a15050565b6000546001600160a01b0316331461070c5760405162461bcd60e51b81526004016101b2906109d7565b61071660006108f4565b565b60408051606081018252600080825260208201819052918101919091526001805460009161074591610a3f565b90505b826001600160401b03166001828154811061076557610765610a56565b6000918252602090912001546001600160401b0316116107de576001818154811061079257610792610a56565b60009182526020918290206040805160608101825291909201546001600160401b038082168352600160401b8204811694830194909452600160801b9004909216908201529392505050565b806107e881610b0d565b915050610748565b6000546001600160a01b0316331461081a5760405162461bcd60e51b81526004016101b2906109d7565b6001600160a01b03811661087f5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016101b2565b610888816108f4565b50565b60408051606081018252600080825260208201819052918101919091526108b143610718565b905090565b600181815481106108c657600080fd5b6000918252602090912001546001600160401b038082169250600160401b8204811691600160801b90041683565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60006060828403121561095657600080fd5b50919050565b6001600160401b038116811461088857600080fd5b60006020828403121561098357600080fd5b813561098e8161095c565b9392505050565b6000602082840312156109a757600080fd5b81356001600160a01b038116811461098e57600080fd5b6000602082840312156109d057600080fd5b5035919050565b6020808252818101527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604082015260600190565b600060208284031215610a1e57600080fd5b815161098e8161095c565b634e487b7160e01b600052601160045260246000fd5b600082821015610a5157610a51610a29565b500390565b634e487b7160e01b600052603260045260246000fd5b8135610a778161095c565b6001600160401b03811690508154816001600160401b031982161783556020840135610aa28161095c565b6fffffffffffffffff0000000000000000604091821b166fffffffffffffffffffffffffffffffff19831684178117855590850135610ae08161095c565b6001600160c01b0319929092169092179190911760809190911b67ffffffffffffffff60801b1617905550565b600081610b1c57610b1c610a29565b50600019019056fea2646970667358221220d72c61c0b8258d97cba5071d62ce65454cecead39c7f5dade7a9e8b8755502d264736f6c63430008090033",
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

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64,uint64))
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
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64,uint64))
func (_KeypersConfigsList *KeypersConfigsListSession) GetCurrentActiveConfig() (KeypersConfig, error) {
	return _KeypersConfigsList.Contract.GetCurrentActiveConfig(&_KeypersConfigsList.CallOpts)
}

// GetCurrentActiveConfig is a free data retrieval call binding the contract method 0xf9c72b53.
//
// Solidity: function getCurrentActiveConfig() view returns((uint64,uint64,uint64))
func (_KeypersConfigsList *KeypersConfigsListCallerSession) GetCurrentActiveConfig() (KeypersConfig, error) {
	return _KeypersConfigsList.Contract.GetCurrentActiveConfig(&_KeypersConfigsList.CallOpts)
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
	Index                 uint64
	Threshold             uint64
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewConfig is a free log retrieval operation binding the contract event 0xf1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d9.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index, uint64 threshold)
func (_KeypersConfigsList *KeypersConfigsListFilterer) FilterNewConfig(opts *bind.FilterOpts) (*KeypersConfigsListNewConfigIterator, error) {

	logs, sub, err := _KeypersConfigsList.contract.FilterLogs(opts, "NewConfig")
	if err != nil {
		return nil, err
	}
	return &KeypersConfigsListNewConfigIterator{contract: _KeypersConfigsList.contract, event: "NewConfig", logs: logs, sub: sub}, nil
}

// WatchNewConfig is a free log subscription operation binding the contract event 0xf1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d9.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index, uint64 threshold)
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

// ParseNewConfig is a log parse operation binding the contract event 0xf1c5613227525376c83485d5a7995987dcfcd90512b0de33df550d2469fba9d9.
//
// Solidity: event NewConfig(uint64 activationBlockNumber, uint64 index, uint64 threshold)
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

// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package help

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

// OwnableMetaData contains all meta data concerning the Ownable contract.
var OwnableMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"8da5cb5b": "owner()",
		"715018a6": "renounceOwnership()",
		"f2fde38b": "transferOwnership(address)",
	},
}

// OwnableABI is the input ABI used to generate the binding from.
// Deprecated: Use OwnableMetaData.ABI instead.
var OwnableABI = OwnableMetaData.ABI

// Deprecated: Use OwnableMetaData.Sigs instead.
// OwnableFuncSigs maps the 4-byte function signature to its string representation.
var OwnableFuncSigs = OwnableMetaData.Sigs

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

// ShutterRegistryMetaData contains all meta data concerning the ShutterRegistry contract.
var ShutterRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"AlreadyRegistered\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidIdentityPrefix\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TTLTooShort\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TimestampInThePast\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"eon\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"identityPrefix\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes[]\",\"name\":\"triggerDefinition\",\"type\":\"bytes[]\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"ttl\",\"type\":\"uint64\"}],\"name\":\"EventTriggerRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"eon\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"identityPrefix\",\"type\":\"bytes32\"},{\"internalType\":\"bytes[]\",\"name\":\"triggerDefinition\",\"type\":\"bytes[]\"},{\"internalType\":\"uint64\",\"name\":\"ttl\",\"type\":\"uint64\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"identity\",\"type\":\"bytes32\"}],\"name\":\"registrations\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"eon\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ttl\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"triggerDefinitionHash\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"8da5cb5b": "owner()",
		"205b396f": "register(uint64,bytes32,bytes[],uint64)",
		"da7c6a42": "registrations(bytes32)",
		"715018a6": "renounceOwnership()",
		"f2fde38b": "transferOwnership(address)",
	},
	Bin: "0x608060405234801561000f575f5ffd5b50335f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603610081575f6040517f1e4fbdf70000000000000000000000000000000000000000000000000000000081526004016100789190610196565b60405180910390fd5b6100908161009660201b60201c565b506101af565b5f5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050815f5f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61018082610157565b9050919050565b61019081610176565b82525050565b5f6020820190506101a95f830184610187565b92915050565b610c5a806101bc5f395ff3fe608060405234801561000f575f5ffd5b5060043610610055575f3560e01c8063205b396f14610059578063715018a6146100755780638da5cb5b1461007f578063da7c6a421461009d578063f2fde38b146100cf575b5f5ffd5b610073600480360381019061006e9190610864565b6100eb565b005b61007d61036c565b005b61008761037f565b6040516100949190610923565b60405180910390f35b6100b760048036038101906100b2919061093c565b6103a6565b6040516100c693929190610985565b60405180910390f35b6100e960048036038101906100e491906109e4565b6103f2565b005b438167ffffffffffffffff16101561012f576040517f84f8e55900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f5f1b830361016a576040517f63a4021d00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f833360405160200161017e929190610a74565b6040516020818303038152906040528051906020012090505f60015f8381526020019081526020015f2090508267ffffffffffffffff16815f0160089054906101000a900467ffffffffffffffff1667ffffffffffffffff161061020e576040517fb5f2184000000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f815f0160089054906101000a900467ffffffffffffffff1667ffffffffffffffff161461029a5780600101548460405160200161024c9190610bba565b6040516020818303038152906040528051906020012014610299576040517f3a81d6fc00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5b82815f0160086101000a81548167ffffffffffffffff021916908367ffffffffffffffff16021790555085815f015f6101000a81548167ffffffffffffffff021916908367ffffffffffffffff160217905550836040516020016102fe9190610bba565b6040516020818303038152906040528051906020012081600101819055508567ffffffffffffffff167f7aae9d6e9dc5cd3510d70ffff8db7974a922a1c919863a85f19881e61da938308633878760405161035c9493929190610bda565b60405180910390a2505050505050565b610374610476565b61037d5f6104fd565b565b5f5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6001602052805f5260405f205f91509050805f015f9054906101000a900467ffffffffffffffff1690805f0160089054906101000a900467ffffffffffffffff16908060010154905083565b6103fa610476565b5f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff160361046a575f6040517f1e4fbdf70000000000000000000000000000000000000000000000000000000081526004016104619190610923565b60405180910390fd5b610473816104fd565b50565b61047e6105be565b73ffffffffffffffffffffffffffffffffffffffff1661049c61037f565b73ffffffffffffffffffffffffffffffffffffffff16146104fb576104bf6105be565b6040517f118cdaa70000000000000000000000000000000000000000000000000000000081526004016104f29190610923565b60405180910390fd5b565b5f5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050815f5f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b5f33905090565b5f604051905090565b5f5ffd5b5f5ffd5b5f67ffffffffffffffff82169050919050565b6105f2816105d6565b81146105fc575f5ffd5b50565b5f8135905061060d816105e9565b92915050565b5f819050919050565b61062581610613565b811461062f575f5ffd5b50565b5f813590506106408161061c565b92915050565b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6106908261064a565b810181811067ffffffffffffffff821117156106af576106ae61065a565b5b80604052505050565b5f6106c16105c5565b90506106cd8282610687565b919050565b5f67ffffffffffffffff8211156106ec576106eb61065a565b5b602082029050602081019050919050565b5f5ffd5b5f5ffd5b5f67ffffffffffffffff82111561071f5761071e61065a565b5b6107288261064a565b9050602081019050919050565b828183375f83830152505050565b5f61075561075084610705565b6106b8565b90508281526020810184848401111561077157610770610701565b5b61077c848285610735565b509392505050565b5f82601f83011261079857610797610646565b5b81356107a8848260208601610743565b91505092915050565b5f6107c36107be846106d2565b6106b8565b905080838252602082019050602084028301858111156107e6576107e56106fd565b5b835b8181101561082d57803567ffffffffffffffff81111561080b5761080a610646565b5b8086016108188982610784565b855260208501945050506020810190506107e8565b5050509392505050565b5f82601f83011261084b5761084a610646565b5b813561085b8482602086016107b1565b91505092915050565b5f5f5f5f6080858703121561087c5761087b6105ce565b5b5f610889878288016105ff565b945050602061089a87828801610632565b935050604085013567ffffffffffffffff8111156108bb576108ba6105d2565b5b6108c787828801610837565b92505060606108d8878288016105ff565b91505092959194509250565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61090d826108e4565b9050919050565b61091d81610903565b82525050565b5f6020820190506109365f830184610914565b92915050565b5f60208284031215610951576109506105ce565b5b5f61095e84828501610632565b91505092915050565b610970816105d6565b82525050565b61097f81610613565b82525050565b5f6060820190506109985f830186610967565b6109a56020830185610967565b6109b26040830184610976565b949350505050565b6109c381610903565b81146109cd575f5ffd5b50565b5f813590506109de816109ba565b92915050565b5f602082840312156109f9576109f86105ce565b5b5f610a06848285016109d0565b91505092915050565b5f819050919050565b610a29610a2482610613565b610a0f565b82525050565b5f8160601b9050919050565b5f610a4582610a2f565b9050919050565b5f610a5682610a3b565b9050919050565b610a6e610a6982610903565b610a4c565b82525050565b5f610a7f8285610a18565b602082019150610a8f8284610a5d565b6014820191508190509392505050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f610afa82610ac8565b610b048185610ad2565b9350610b14818560208601610ae2565b610b1d8161064a565b840191505092915050565b5f610b338383610af0565b905092915050565b5f602082019050919050565b5f610b5182610a9f565b610b5b8185610aa9565b935083602082028501610b6d85610ab9565b805f5b85811015610ba85784840389528151610b898582610b28565b9450610b9483610b3b565b925060208a01995050600181019050610b70565b50829750879550505050505092915050565b5f6020820190508181035f830152610bd28184610b47565b905092915050565b5f608082019050610bed5f830187610976565b610bfa6020830186610914565b8181036040830152610c0c8185610b47565b9050610c1b6060830184610967565b9594505050505056fea26469706673582212209ad31602fbb4bc228e7ee36774a68a11d1c3d06a60fa340db05278e6658e25bc64736f6c634300081c0033",
}

// ShutterRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ShutterRegistryMetaData.ABI instead.
var ShutterRegistryABI = ShutterRegistryMetaData.ABI

// Deprecated: Use ShutterRegistryMetaData.Sigs instead.
// ShutterRegistryFuncSigs maps the 4-byte function signature to its string representation.
var ShutterRegistryFuncSigs = ShutterRegistryMetaData.Sigs

// ShutterRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ShutterRegistryMetaData.Bin instead.
var ShutterRegistryBin = ShutterRegistryMetaData.Bin

// DeployShutterRegistry deploys a new Ethereum contract, binding an instance of ShutterRegistry to it.
func DeployShutterRegistry(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ShutterRegistry, error) {
	parsed, err := ShutterRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ShutterRegistryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ShutterRegistry{ShutterRegistryCaller: ShutterRegistryCaller{contract: contract}, ShutterRegistryTransactor: ShutterRegistryTransactor{contract: contract}, ShutterRegistryFilterer: ShutterRegistryFilterer{contract: contract}}, nil
}

// ShutterRegistry is an auto generated Go binding around an Ethereum contract.
type ShutterRegistry struct {
	ShutterRegistryCaller     // Read-only binding to the contract
	ShutterRegistryTransactor // Write-only binding to the contract
	ShutterRegistryFilterer   // Log filterer for contract events
}

// ShutterRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ShutterRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ShutterRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ShutterRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ShutterRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ShutterRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ShutterRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ShutterRegistrySession struct {
	Contract     *ShutterRegistry  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ShutterRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ShutterRegistryCallerSession struct {
	Contract *ShutterRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// ShutterRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ShutterRegistryTransactorSession struct {
	Contract     *ShutterRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// ShutterRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ShutterRegistryRaw struct {
	Contract *ShutterRegistry // Generic contract binding to access the raw methods on
}

// ShutterRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ShutterRegistryCallerRaw struct {
	Contract *ShutterRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ShutterRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ShutterRegistryTransactorRaw struct {
	Contract *ShutterRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewShutterRegistry creates a new instance of ShutterRegistry, bound to a specific deployed contract.
func NewShutterRegistry(address common.Address, backend bind.ContractBackend) (*ShutterRegistry, error) {
	contract, err := bindShutterRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ShutterRegistry{ShutterRegistryCaller: ShutterRegistryCaller{contract: contract}, ShutterRegistryTransactor: ShutterRegistryTransactor{contract: contract}, ShutterRegistryFilterer: ShutterRegistryFilterer{contract: contract}}, nil
}

// NewShutterRegistryCaller creates a new read-only instance of ShutterRegistry, bound to a specific deployed contract.
func NewShutterRegistryCaller(address common.Address, caller bind.ContractCaller) (*ShutterRegistryCaller, error) {
	contract, err := bindShutterRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ShutterRegistryCaller{contract: contract}, nil
}

// NewShutterRegistryTransactor creates a new write-only instance of ShutterRegistry, bound to a specific deployed contract.
func NewShutterRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ShutterRegistryTransactor, error) {
	contract, err := bindShutterRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ShutterRegistryTransactor{contract: contract}, nil
}

// NewShutterRegistryFilterer creates a new log filterer instance of ShutterRegistry, bound to a specific deployed contract.
func NewShutterRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ShutterRegistryFilterer, error) {
	contract, err := bindShutterRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ShutterRegistryFilterer{contract: contract}, nil
}

// bindShutterRegistry binds a generic wrapper to an already deployed contract.
func bindShutterRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ShutterRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ShutterRegistry *ShutterRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ShutterRegistry.Contract.ShutterRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ShutterRegistry *ShutterRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ShutterRegistry.Contract.ShutterRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ShutterRegistry *ShutterRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ShutterRegistry.Contract.ShutterRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ShutterRegistry *ShutterRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ShutterRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ShutterRegistry *ShutterRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ShutterRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ShutterRegistry *ShutterRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ShutterRegistry.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ShutterRegistry *ShutterRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ShutterRegistry.contract.Call(opts, &out, "owner")
	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ShutterRegistry *ShutterRegistrySession) Owner() (common.Address, error) {
	return _ShutterRegistry.Contract.Owner(&_ShutterRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ShutterRegistry *ShutterRegistryCallerSession) Owner() (common.Address, error) {
	return _ShutterRegistry.Contract.Owner(&_ShutterRegistry.CallOpts)
}

// Registrations is a free data retrieval call binding the contract method 0xda7c6a42.
//
// Solidity: function registrations(bytes32 identity) view returns(uint64 eon, uint64 ttl, bytes32 triggerDefinitionHash)
func (_ShutterRegistry *ShutterRegistryCaller) Registrations(opts *bind.CallOpts, identity [32]byte) (struct {
	Eon                   uint64
	Ttl                   uint64
	TriggerDefinitionHash [32]byte
}, error,
) {
	var out []interface{}
	err := _ShutterRegistry.contract.Call(opts, &out, "registrations", identity)

	outstruct := new(struct {
		Eon                   uint64
		Ttl                   uint64
		TriggerDefinitionHash [32]byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Eon = *abi.ConvertType(out[0], new(uint64)).(*uint64)
	outstruct.Ttl = *abi.ConvertType(out[1], new(uint64)).(*uint64)
	outstruct.TriggerDefinitionHash = *abi.ConvertType(out[2], new([32]byte)).(*[32]byte)

	return *outstruct, err
}

// Registrations is a free data retrieval call binding the contract method 0xda7c6a42.
//
// Solidity: function registrations(bytes32 identity) view returns(uint64 eon, uint64 ttl, bytes32 triggerDefinitionHash)
func (_ShutterRegistry *ShutterRegistrySession) Registrations(identity [32]byte) (struct {
	Eon                   uint64
	Ttl                   uint64
	TriggerDefinitionHash [32]byte
}, error,
) {
	return _ShutterRegistry.Contract.Registrations(&_ShutterRegistry.CallOpts, identity)
}

// Registrations is a free data retrieval call binding the contract method 0xda7c6a42.
//
// Solidity: function registrations(bytes32 identity) view returns(uint64 eon, uint64 ttl, bytes32 triggerDefinitionHash)
func (_ShutterRegistry *ShutterRegistryCallerSession) Registrations(identity [32]byte) (struct {
	Eon                   uint64
	Ttl                   uint64
	TriggerDefinitionHash [32]byte
}, error,
) {
	return _ShutterRegistry.Contract.Registrations(&_ShutterRegistry.CallOpts, identity)
}

// Register is a paid mutator transaction binding the contract method 0x205b396f.
//
// Solidity: function register(uint64 eon, bytes32 identityPrefix, bytes[] triggerDefinition, uint64 ttl) returns()
func (_ShutterRegistry *ShutterRegistryTransactor) Register(opts *bind.TransactOpts, eon uint64, identityPrefix [32]byte, triggerDefinition [][]byte, ttl uint64) (*types.Transaction, error) {
	return _ShutterRegistry.contract.Transact(opts, "register", eon, identityPrefix, triggerDefinition, ttl)
}

// Register is a paid mutator transaction binding the contract method 0x205b396f.
//
// Solidity: function register(uint64 eon, bytes32 identityPrefix, bytes[] triggerDefinition, uint64 ttl) returns()
func (_ShutterRegistry *ShutterRegistrySession) Register(eon uint64, identityPrefix [32]byte, triggerDefinition [][]byte, ttl uint64) (*types.Transaction, error) {
	return _ShutterRegistry.Contract.Register(&_ShutterRegistry.TransactOpts, eon, identityPrefix, triggerDefinition, ttl)
}

// Register is a paid mutator transaction binding the contract method 0x205b396f.
//
// Solidity: function register(uint64 eon, bytes32 identityPrefix, bytes[] triggerDefinition, uint64 ttl) returns()
func (_ShutterRegistry *ShutterRegistryTransactorSession) Register(eon uint64, identityPrefix [32]byte, triggerDefinition [][]byte, ttl uint64) (*types.Transaction, error) {
	return _ShutterRegistry.Contract.Register(&_ShutterRegistry.TransactOpts, eon, identityPrefix, triggerDefinition, ttl)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ShutterRegistry *ShutterRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ShutterRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ShutterRegistry *ShutterRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _ShutterRegistry.Contract.RenounceOwnership(&_ShutterRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ShutterRegistry *ShutterRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ShutterRegistry.Contract.RenounceOwnership(&_ShutterRegistry.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ShutterRegistry *ShutterRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ShutterRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ShutterRegistry *ShutterRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ShutterRegistry.Contract.TransferOwnership(&_ShutterRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ShutterRegistry *ShutterRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ShutterRegistry.Contract.TransferOwnership(&_ShutterRegistry.TransactOpts, newOwner)
}

// ShutterRegistryEventTriggerRegisteredIterator is returned from FilterEventTriggerRegistered and is used to iterate over the raw logs and unpacked data for EventTriggerRegistered events raised by the ShutterRegistry contract.
type ShutterRegistryEventTriggerRegisteredIterator struct {
	Event *ShutterRegistryEventTriggerRegistered // Event containing the contract specifics and raw log

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
func (it *ShutterRegistryEventTriggerRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ShutterRegistryEventTriggerRegistered)
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
		it.Event = new(ShutterRegistryEventTriggerRegistered)
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
func (it *ShutterRegistryEventTriggerRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ShutterRegistryEventTriggerRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ShutterRegistryEventTriggerRegistered represents a EventTriggerRegistered event raised by the ShutterRegistry contract.
type ShutterRegistryEventTriggerRegistered struct {
	Eon               uint64
	IdentityPrefix    [32]byte
	Sender            common.Address
	TriggerDefinition [][]byte
	Ttl               uint64
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterEventTriggerRegistered is a free log retrieval operation binding the contract event 0x7aae9d6e9dc5cd3510d70ffff8db7974a922a1c919863a85f19881e61da93830.
//
// Solidity: event EventTriggerRegistered(uint64 indexed eon, bytes32 identityPrefix, address sender, bytes[] triggerDefinition, uint64 ttl)
func (_ShutterRegistry *ShutterRegistryFilterer) FilterEventTriggerRegistered(opts *bind.FilterOpts, eon []uint64) (*ShutterRegistryEventTriggerRegisteredIterator, error) {
	var eonRule []interface{}
	for _, eonItem := range eon {
		eonRule = append(eonRule, eonItem)
	}

	logs, sub, err := _ShutterRegistry.contract.FilterLogs(opts, "EventTriggerRegistered", eonRule)
	if err != nil {
		return nil, err
	}
	return &ShutterRegistryEventTriggerRegisteredIterator{contract: _ShutterRegistry.contract, event: "EventTriggerRegistered", logs: logs, sub: sub}, nil
}

// WatchEventTriggerRegistered is a free log subscription operation binding the contract event 0x7aae9d6e9dc5cd3510d70ffff8db7974a922a1c919863a85f19881e61da93830.
//
// Solidity: event EventTriggerRegistered(uint64 indexed eon, bytes32 identityPrefix, address sender, bytes[] triggerDefinition, uint64 ttl)
func (_ShutterRegistry *ShutterRegistryFilterer) WatchEventTriggerRegistered(opts *bind.WatchOpts, sink chan<- *ShutterRegistryEventTriggerRegistered, eon []uint64) (event.Subscription, error) {
	var eonRule []interface{}
	for _, eonItem := range eon {
		eonRule = append(eonRule, eonItem)
	}

	logs, sub, err := _ShutterRegistry.contract.WatchLogs(opts, "EventTriggerRegistered", eonRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ShutterRegistryEventTriggerRegistered)
				if err := _ShutterRegistry.contract.UnpackLog(event, "EventTriggerRegistered", log); err != nil {
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

// ParseEventTriggerRegistered is a log parse operation binding the contract event 0x7aae9d6e9dc5cd3510d70ffff8db7974a922a1c919863a85f19881e61da93830.
//
// Solidity: event EventTriggerRegistered(uint64 indexed eon, bytes32 identityPrefix, address sender, bytes[] triggerDefinition, uint64 ttl)
func (_ShutterRegistry *ShutterRegistryFilterer) ParseEventTriggerRegistered(log types.Log) (*ShutterRegistryEventTriggerRegistered, error) {
	event := new(ShutterRegistryEventTriggerRegistered)
	if err := _ShutterRegistry.contract.UnpackLog(event, "EventTriggerRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ShutterRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ShutterRegistry contract.
type ShutterRegistryOwnershipTransferredIterator struct {
	Event *ShutterRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ShutterRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ShutterRegistryOwnershipTransferred)
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
		it.Event = new(ShutterRegistryOwnershipTransferred)
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
func (it *ShutterRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ShutterRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ShutterRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the ShutterRegistry contract.
type ShutterRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ShutterRegistry *ShutterRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ShutterRegistryOwnershipTransferredIterator, error) {
	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ShutterRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ShutterRegistryOwnershipTransferredIterator{contract: _ShutterRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ShutterRegistry *ShutterRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ShutterRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {
	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ShutterRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ShutterRegistryOwnershipTransferred)
				if err := _ShutterRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_ShutterRegistry *ShutterRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*ShutterRegistryOwnershipTransferred, error) {
	event := new(ShutterRegistryOwnershipTransferred)
	if err := _ShutterRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

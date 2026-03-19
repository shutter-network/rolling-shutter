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

// EmitterMetaData contains all meta data concerning the Emitter contract.
var EmitterMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"five\",\"type\":\"bytes\"}],\"name\":\"Five\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"}],\"name\":\"Four\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"two\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"three\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"five\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"six\",\"type\":\"bytes\"}],\"name\":\"Six\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Two\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"five\",\"type\":\"bytes\"}],\"name\":\"emitFive\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"}],\"name\":\"emitFour\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"two\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"three\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"five\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"six\",\"type\":\"bytes\"}],\"name\":\"emitSix\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"emitTwo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b506108ed806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80636995a2d9146100515780637155880d1461006d5780638cc5e89214610089578063f70770e0146100a5575b600080fd5b61006b60048036038101906100669190610381565b6100c1565b005b61008760048036038101906100829190610434565b610104565b005b6100a3600480360381019061009e9190610461565b610140565b005b6100bf60048036038101906100ba91906105e3565b610180565b005b8284867f2778059b9d45e2cd0df03a27bbe3e688dfc48aa15a729c42f39dcd986ebd446185856040516100f592919061074c565b60405180910390a45050505050565b807fce34f015a0e20f2b0daf980b28ed50729e87b993e4d30cca4c3f4da05acbd0ac600560405161013591906107c8565b60405180910390a250565b8183857fd82c9bd67140e94b50e0a62e800c51428267b0cd733573daaafad26b62c05afb8460405161017291906107e3565b60405180910390a450505050565b8373ffffffffffffffffffffffffffffffffffffffff16856040516101a5919061084c565b6040518091039020877f1c49d7caedaa8e8be0f2ef7b3c285ccaef7eac5a7fe6b427cbe7e7d8ad3070548686866040516101e193929190610872565b60405180910390a4505050505050565b6000604051905090565b600080fd5b600080fd5b6000819050919050565b61021881610205565b811461022357600080fd5b50565b6000813590506102358161020f565b92915050565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b61028e82610245565b810181811067ffffffffffffffff821117156102ad576102ac610256565b5b80604052505050565b60006102c06101f1565b90506102cc8282610285565b919050565b600067ffffffffffffffff8211156102ec576102eb610256565b5b6102f582610245565b9050602081019050919050565b82818337600083830152505050565b600061032461031f846102d1565b6102b6565b9050828152602081018484840111156103405761033f610240565b5b61034b848285610302565b509392505050565b600082601f8301126103685761036761023b565b5b8135610378848260208601610311565b91505092915050565b600080600080600060a0868803121561039d5761039c6101fb565b5b60006103ab88828901610226565b95505060206103bc88828901610226565b94505060406103cd88828901610226565b935050606086013567ffffffffffffffff8111156103ee576103ed610200565b5b6103fa88828901610353565b925050608086013567ffffffffffffffff81111561041b5761041a610200565b5b61042788828901610353565b9150509295509295909350565b60006020828403121561044a576104496101fb565b5b600061045884828501610226565b91505092915050565b6000806000806080858703121561047b5761047a6101fb565b5b600061048987828801610226565b945050602061049a87828801610226565b93505060406104ab87828801610226565b925050606085013567ffffffffffffffff8111156104cc576104cb610200565b5b6104d887828801610353565b91505092959194509250565b600067ffffffffffffffff8211156104ff576104fe610256565b5b61050882610245565b9050602081019050919050565b6000610528610523846104e4565b6102b6565b90508281526020810184848401111561054457610543610240565b5b61054f848285610302565b509392505050565b600082601f83011261056c5761056b61023b565b5b813561057c848260208601610515565b91505092915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006105b082610585565b9050919050565b6105c0816105a5565b81146105cb57600080fd5b50565b6000813590506105dd816105b7565b92915050565b60008060008060008060c08789031215610600576105ff6101fb565b5b600061060e89828a01610226565b965050602087013567ffffffffffffffff81111561062f5761062e610200565b5b61063b89828a01610557565b955050604061064c89828a016105ce565b945050606087013567ffffffffffffffff81111561066d5761066c610200565b5b61067989828a01610353565b935050608061068a89828a01610226565b92505060a087013567ffffffffffffffff8111156106ab576106aa610200565b5b6106b789828a01610353565b9150509295509295509295565b600081519050919050565b600082825260208201905092915050565b60005b838110156106fe5780820151818401526020810190506106e3565b8381111561070d576000848401525b50505050565b600061071e826106c4565b61072881856106cf565b93506107388185602086016106e0565b61074181610245565b840191505092915050565b600060408201905081810360008301526107668185610713565b9050818103602083015261077a8184610713565b90509392505050565b6000819050919050565b6000819050919050565b60006107b26107ad6107a884610783565b61078d565b610205565b9050919050565b6107c281610797565b82525050565b60006020820190506107dd60008301846107b9565b92915050565b600060208201905081810360008301526107fd8184610713565b905092915050565b600081519050919050565b600081905092915050565b600061082682610805565b6108308185610810565b93506108408185602086016106e0565b80840191505092915050565b6000610858828461081b565b915081905092915050565b61086c81610205565b82525050565b6000606082019050818103600083015261088c8186610713565b905061089b6020830185610863565b81810360408301526108ad8184610713565b905094935050505056fea2646970667358221220848f06dbefbfd7023244910f585f43334b07207284d4f6541eb16b4b0073fa8964736f6c63430008090033",
}

// EmitterABI is the input ABI used to generate the binding from.
// Deprecated: Use EmitterMetaData.ABI instead.
var EmitterABI = EmitterMetaData.ABI

// EmitterBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use EmitterMetaData.Bin instead.
var EmitterBin = EmitterMetaData.Bin

// DeployEmitter deploys a new Ethereum contract, binding an instance of Emitter to it.
func DeployEmitter(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Emitter, error) {
	parsed, err := EmitterMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(EmitterBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Emitter{EmitterCaller: EmitterCaller{contract: contract}, EmitterTransactor: EmitterTransactor{contract: contract}, EmitterFilterer: EmitterFilterer{contract: contract}}, nil
}

// Emitter is an auto generated Go binding around an Ethereum contract.
type Emitter struct {
	EmitterCaller     // Read-only binding to the contract
	EmitterTransactor // Write-only binding to the contract
	EmitterFilterer   // Log filterer for contract events
}

// EmitterCaller is an auto generated read-only Go binding around an Ethereum contract.
type EmitterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EmitterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EmitterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EmitterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EmitterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EmitterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EmitterSession struct {
	Contract     *Emitter          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EmitterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EmitterCallerSession struct {
	Contract *EmitterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// EmitterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EmitterTransactorSession struct {
	Contract     *EmitterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// EmitterRaw is an auto generated low-level Go binding around an Ethereum contract.
type EmitterRaw struct {
	Contract *Emitter // Generic contract binding to access the raw methods on
}

// EmitterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EmitterCallerRaw struct {
	Contract *EmitterCaller // Generic read-only contract binding to access the raw methods on
}

// EmitterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EmitterTransactorRaw struct {
	Contract *EmitterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEmitter creates a new instance of Emitter, bound to a specific deployed contract.
func NewEmitter(address common.Address, backend bind.ContractBackend) (*Emitter, error) {
	contract, err := bindEmitter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Emitter{EmitterCaller: EmitterCaller{contract: contract}, EmitterTransactor: EmitterTransactor{contract: contract}, EmitterFilterer: EmitterFilterer{contract: contract}}, nil
}

// NewEmitterCaller creates a new read-only instance of Emitter, bound to a specific deployed contract.
func NewEmitterCaller(address common.Address, caller bind.ContractCaller) (*EmitterCaller, error) {
	contract, err := bindEmitter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EmitterCaller{contract: contract}, nil
}

// NewEmitterTransactor creates a new write-only instance of Emitter, bound to a specific deployed contract.
func NewEmitterTransactor(address common.Address, transactor bind.ContractTransactor) (*EmitterTransactor, error) {
	contract, err := bindEmitter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EmitterTransactor{contract: contract}, nil
}

// NewEmitterFilterer creates a new log filterer instance of Emitter, bound to a specific deployed contract.
func NewEmitterFilterer(address common.Address, filterer bind.ContractFilterer) (*EmitterFilterer, error) {
	contract, err := bindEmitter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EmitterFilterer{contract: contract}, nil
}

// bindEmitter binds a generic wrapper to an already deployed contract.
func bindEmitter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := EmitterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Emitter *EmitterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Emitter.Contract.EmitterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Emitter *EmitterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Emitter.Contract.EmitterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Emitter *EmitterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Emitter.Contract.EmitterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Emitter *EmitterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Emitter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Emitter *EmitterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Emitter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Emitter *EmitterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Emitter.Contract.contract.Transact(opts, method, params...)
}

// EmitFive is a paid mutator transaction binding the contract method 0x6995a2d9.
//
// Solidity: function emitFive(uint256 one, uint256 two, uint256 three, bytes four, bytes five) returns()
func (_Emitter *EmitterTransactor) EmitFive(opts *bind.TransactOpts, one *big.Int, two *big.Int, three *big.Int, four []byte, five []byte) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitFive", one, two, three, four, five)
}

// EmitFive is a paid mutator transaction binding the contract method 0x6995a2d9.
//
// Solidity: function emitFive(uint256 one, uint256 two, uint256 three, bytes four, bytes five) returns()
func (_Emitter *EmitterSession) EmitFive(one *big.Int, two *big.Int, three *big.Int, four []byte, five []byte) (*types.Transaction, error) {
	return _Emitter.Contract.EmitFive(&_Emitter.TransactOpts, one, two, three, four, five)
}

// EmitFive is a paid mutator transaction binding the contract method 0x6995a2d9.
//
// Solidity: function emitFive(uint256 one, uint256 two, uint256 three, bytes four, bytes five) returns()
func (_Emitter *EmitterTransactorSession) EmitFive(one *big.Int, two *big.Int, three *big.Int, four []byte, five []byte) (*types.Transaction, error) {
	return _Emitter.Contract.EmitFive(&_Emitter.TransactOpts, one, two, three, four, five)
}

// EmitFour is a paid mutator transaction binding the contract method 0x8cc5e892.
//
// Solidity: function emitFour(uint256 one, uint256 two, uint256 three, bytes four) returns()
func (_Emitter *EmitterTransactor) EmitFour(opts *bind.TransactOpts, one *big.Int, two *big.Int, three *big.Int, four []byte) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitFour", one, two, three, four)
}

// EmitFour is a paid mutator transaction binding the contract method 0x8cc5e892.
//
// Solidity: function emitFour(uint256 one, uint256 two, uint256 three, bytes four) returns()
func (_Emitter *EmitterSession) EmitFour(one *big.Int, two *big.Int, three *big.Int, four []byte) (*types.Transaction, error) {
	return _Emitter.Contract.EmitFour(&_Emitter.TransactOpts, one, two, three, four)
}

// EmitFour is a paid mutator transaction binding the contract method 0x8cc5e892.
//
// Solidity: function emitFour(uint256 one, uint256 two, uint256 three, bytes four) returns()
func (_Emitter *EmitterTransactorSession) EmitFour(one *big.Int, two *big.Int, three *big.Int, four []byte) (*types.Transaction, error) {
	return _Emitter.Contract.EmitFour(&_Emitter.TransactOpts, one, two, three, four)
}

// EmitSix is a paid mutator transaction binding the contract method 0xf70770e0.
//
// Solidity: function emitSix(uint256 one, string two, address three, bytes four, uint256 five, bytes six) returns()
func (_Emitter *EmitterTransactor) EmitSix(opts *bind.TransactOpts, one *big.Int, two string, three common.Address, four []byte, five *big.Int, six []byte) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitSix", one, two, three, four, five, six)
}

// EmitSix is a paid mutator transaction binding the contract method 0xf70770e0.
//
// Solidity: function emitSix(uint256 one, string two, address three, bytes four, uint256 five, bytes six) returns()
func (_Emitter *EmitterSession) EmitSix(one *big.Int, two string, three common.Address, four []byte, five *big.Int, six []byte) (*types.Transaction, error) {
	return _Emitter.Contract.EmitSix(&_Emitter.TransactOpts, one, two, three, four, five, six)
}

// EmitSix is a paid mutator transaction binding the contract method 0xf70770e0.
//
// Solidity: function emitSix(uint256 one, string two, address three, bytes four, uint256 five, bytes six) returns()
func (_Emitter *EmitterTransactorSession) EmitSix(one *big.Int, two string, three common.Address, four []byte, five *big.Int, six []byte) (*types.Transaction, error) {
	return _Emitter.Contract.EmitSix(&_Emitter.TransactOpts, one, two, three, four, five, six)
}

// EmitTwo is a paid mutator transaction binding the contract method 0x7155880d.
//
// Solidity: function emitTwo(uint256 value) returns()
func (_Emitter *EmitterTransactor) EmitTwo(opts *bind.TransactOpts, value *big.Int) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitTwo", value)
}

// EmitTwo is a paid mutator transaction binding the contract method 0x7155880d.
//
// Solidity: function emitTwo(uint256 value) returns()
func (_Emitter *EmitterSession) EmitTwo(value *big.Int) (*types.Transaction, error) {
	return _Emitter.Contract.EmitTwo(&_Emitter.TransactOpts, value)
}

// EmitTwo is a paid mutator transaction binding the contract method 0x7155880d.
//
// Solidity: function emitTwo(uint256 value) returns()
func (_Emitter *EmitterTransactorSession) EmitTwo(value *big.Int) (*types.Transaction, error) {
	return _Emitter.Contract.EmitTwo(&_Emitter.TransactOpts, value)
}

// EmitterFiveIterator is returned from FilterFive and is used to iterate over the raw logs and unpacked data for Five events raised by the Emitter contract.
type EmitterFiveIterator struct {
	Event *EmitterFive // Event containing the contract specifics and raw log

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
func (it *EmitterFiveIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EmitterFive)
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
		it.Event = new(EmitterFive)
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
func (it *EmitterFiveIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EmitterFiveIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EmitterFive represents a Five event raised by the Emitter contract.
type EmitterFive struct {
	One   *big.Int
	Two   *big.Int
	Three *big.Int
	Four  []byte
	Five  []byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterFive is a free log retrieval operation binding the contract event 0x2778059b9d45e2cd0df03a27bbe3e688dfc48aa15a729c42f39dcd986ebd4461.
//
// Solidity: event Five(uint256 indexed one, uint256 indexed two, uint256 indexed three, bytes four, bytes five)
func (_Emitter *EmitterFilterer) FilterFive(opts *bind.FilterOpts, one []*big.Int, two []*big.Int, three []*big.Int) (*EmitterFiveIterator, error) {
	var oneRule []interface{}
	for _, oneItem := range one {
		oneRule = append(oneRule, oneItem)
	}
	var twoRule []interface{}
	for _, twoItem := range two {
		twoRule = append(twoRule, twoItem)
	}
	var threeRule []interface{}
	for _, threeItem := range three {
		threeRule = append(threeRule, threeItem)
	}

	logs, sub, err := _Emitter.contract.FilterLogs(opts, "Five", oneRule, twoRule, threeRule)
	if err != nil {
		return nil, err
	}
	return &EmitterFiveIterator{contract: _Emitter.contract, event: "Five", logs: logs, sub: sub}, nil
}

// WatchFive is a free log subscription operation binding the contract event 0x2778059b9d45e2cd0df03a27bbe3e688dfc48aa15a729c42f39dcd986ebd4461.
//
// Solidity: event Five(uint256 indexed one, uint256 indexed two, uint256 indexed three, bytes four, bytes five)
func (_Emitter *EmitterFilterer) WatchFive(opts *bind.WatchOpts, sink chan<- *EmitterFive, one []*big.Int, two []*big.Int, three []*big.Int) (event.Subscription, error) {
	var oneRule []interface{}
	for _, oneItem := range one {
		oneRule = append(oneRule, oneItem)
	}
	var twoRule []interface{}
	for _, twoItem := range two {
		twoRule = append(twoRule, twoItem)
	}
	var threeRule []interface{}
	for _, threeItem := range three {
		threeRule = append(threeRule, threeItem)
	}

	logs, sub, err := _Emitter.contract.WatchLogs(opts, "Five", oneRule, twoRule, threeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EmitterFive)
				if err := _Emitter.contract.UnpackLog(event, "Five", log); err != nil {
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

// ParseFive is a log parse operation binding the contract event 0x2778059b9d45e2cd0df03a27bbe3e688dfc48aa15a729c42f39dcd986ebd4461.
//
// Solidity: event Five(uint256 indexed one, uint256 indexed two, uint256 indexed three, bytes four, bytes five)
func (_Emitter *EmitterFilterer) ParseFive(log types.Log) (*EmitterFive, error) {
	event := new(EmitterFive)
	if err := _Emitter.contract.UnpackLog(event, "Five", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EmitterFourIterator is returned from FilterFour and is used to iterate over the raw logs and unpacked data for Four events raised by the Emitter contract.
type EmitterFourIterator struct {
	Event *EmitterFour // Event containing the contract specifics and raw log

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
func (it *EmitterFourIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EmitterFour)
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
		it.Event = new(EmitterFour)
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
func (it *EmitterFourIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EmitterFourIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EmitterFour represents a Four event raised by the Emitter contract.
type EmitterFour struct {
	One   *big.Int
	Two   *big.Int
	Three *big.Int
	Four  []byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterFour is a free log retrieval operation binding the contract event 0xd82c9bd67140e94b50e0a62e800c51428267b0cd733573daaafad26b62c05afb.
//
// Solidity: event Four(uint256 indexed one, uint256 indexed two, uint256 indexed three, bytes four)
func (_Emitter *EmitterFilterer) FilterFour(opts *bind.FilterOpts, one []*big.Int, two []*big.Int, three []*big.Int) (*EmitterFourIterator, error) {
	var oneRule []interface{}
	for _, oneItem := range one {
		oneRule = append(oneRule, oneItem)
	}
	var twoRule []interface{}
	for _, twoItem := range two {
		twoRule = append(twoRule, twoItem)
	}
	var threeRule []interface{}
	for _, threeItem := range three {
		threeRule = append(threeRule, threeItem)
	}

	logs, sub, err := _Emitter.contract.FilterLogs(opts, "Four", oneRule, twoRule, threeRule)
	if err != nil {
		return nil, err
	}
	return &EmitterFourIterator{contract: _Emitter.contract, event: "Four", logs: logs, sub: sub}, nil
}

// WatchFour is a free log subscription operation binding the contract event 0xd82c9bd67140e94b50e0a62e800c51428267b0cd733573daaafad26b62c05afb.
//
// Solidity: event Four(uint256 indexed one, uint256 indexed two, uint256 indexed three, bytes four)
func (_Emitter *EmitterFilterer) WatchFour(opts *bind.WatchOpts, sink chan<- *EmitterFour, one []*big.Int, two []*big.Int, three []*big.Int) (event.Subscription, error) {
	var oneRule []interface{}
	for _, oneItem := range one {
		oneRule = append(oneRule, oneItem)
	}
	var twoRule []interface{}
	for _, twoItem := range two {
		twoRule = append(twoRule, twoItem)
	}
	var threeRule []interface{}
	for _, threeItem := range three {
		threeRule = append(threeRule, threeItem)
	}

	logs, sub, err := _Emitter.contract.WatchLogs(opts, "Four", oneRule, twoRule, threeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EmitterFour)
				if err := _Emitter.contract.UnpackLog(event, "Four", log); err != nil {
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

// ParseFour is a log parse operation binding the contract event 0xd82c9bd67140e94b50e0a62e800c51428267b0cd733573daaafad26b62c05afb.
//
// Solidity: event Four(uint256 indexed one, uint256 indexed two, uint256 indexed three, bytes four)
func (_Emitter *EmitterFilterer) ParseFour(log types.Log) (*EmitterFour, error) {
	event := new(EmitterFour)
	if err := _Emitter.contract.UnpackLog(event, "Four", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EmitterSixIterator is returned from FilterSix and is used to iterate over the raw logs and unpacked data for Six events raised by the Emitter contract.
type EmitterSixIterator struct {
	Event *EmitterSix // Event containing the contract specifics and raw log

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
func (it *EmitterSixIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EmitterSix)
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
		it.Event = new(EmitterSix)
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
func (it *EmitterSixIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EmitterSixIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EmitterSix represents a Six event raised by the Emitter contract.
type EmitterSix struct {
	One   *big.Int
	Two   common.Hash
	Three common.Address
	Four  []byte
	Five  *big.Int
	Six   []byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSix is a free log retrieval operation binding the contract event 0x1c49d7caedaa8e8be0f2ef7b3c285ccaef7eac5a7fe6b427cbe7e7d8ad307054.
//
// Solidity: event Six(uint256 indexed one, string indexed two, address indexed three, bytes four, uint256 five, bytes six)
func (_Emitter *EmitterFilterer) FilterSix(opts *bind.FilterOpts, one []*big.Int, two []string, three []common.Address) (*EmitterSixIterator, error) {
	var oneRule []interface{}
	for _, oneItem := range one {
		oneRule = append(oneRule, oneItem)
	}
	var twoRule []interface{}
	for _, twoItem := range two {
		twoRule = append(twoRule, twoItem)
	}
	var threeRule []interface{}
	for _, threeItem := range three {
		threeRule = append(threeRule, threeItem)
	}

	logs, sub, err := _Emitter.contract.FilterLogs(opts, "Six", oneRule, twoRule, threeRule)
	if err != nil {
		return nil, err
	}
	return &EmitterSixIterator{contract: _Emitter.contract, event: "Six", logs: logs, sub: sub}, nil
}

// WatchSix is a free log subscription operation binding the contract event 0x1c49d7caedaa8e8be0f2ef7b3c285ccaef7eac5a7fe6b427cbe7e7d8ad307054.
//
// Solidity: event Six(uint256 indexed one, string indexed two, address indexed three, bytes four, uint256 five, bytes six)
func (_Emitter *EmitterFilterer) WatchSix(opts *bind.WatchOpts, sink chan<- *EmitterSix, one []*big.Int, two []string, three []common.Address) (event.Subscription, error) {
	var oneRule []interface{}
	for _, oneItem := range one {
		oneRule = append(oneRule, oneItem)
	}
	var twoRule []interface{}
	for _, twoItem := range two {
		twoRule = append(twoRule, twoItem)
	}
	var threeRule []interface{}
	for _, threeItem := range three {
		threeRule = append(threeRule, threeItem)
	}

	logs, sub, err := _Emitter.contract.WatchLogs(opts, "Six", oneRule, twoRule, threeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EmitterSix)
				if err := _Emitter.contract.UnpackLog(event, "Six", log); err != nil {
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

// ParseSix is a log parse operation binding the contract event 0x1c49d7caedaa8e8be0f2ef7b3c285ccaef7eac5a7fe6b427cbe7e7d8ad307054.
//
// Solidity: event Six(uint256 indexed one, string indexed two, address indexed three, bytes four, uint256 five, bytes six)
func (_Emitter *EmitterFilterer) ParseSix(log types.Log) (*EmitterSix, error) {
	event := new(EmitterSix)
	if err := _Emitter.contract.UnpackLog(event, "Six", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EmitterTwoIterator is returned from FilterTwo and is used to iterate over the raw logs and unpacked data for Two events raised by the Emitter contract.
type EmitterTwoIterator struct {
	Event *EmitterTwo // Event containing the contract specifics and raw log

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
func (it *EmitterTwoIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EmitterTwo)
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
		it.Event = new(EmitterTwo)
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
func (it *EmitterTwoIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EmitterTwoIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EmitterTwo represents a Two event raised by the Emitter contract.
type EmitterTwo struct {
	NewValue *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterTwo is a free log retrieval operation binding the contract event 0xce34f015a0e20f2b0daf980b28ed50729e87b993e4d30cca4c3f4da05acbd0ac.
//
// Solidity: event Two(uint256 indexed newValue, uint256 value)
func (_Emitter *EmitterFilterer) FilterTwo(opts *bind.FilterOpts, newValue []*big.Int) (*EmitterTwoIterator, error) {
	var newValueRule []interface{}
	for _, newValueItem := range newValue {
		newValueRule = append(newValueRule, newValueItem)
	}

	logs, sub, err := _Emitter.contract.FilterLogs(opts, "Two", newValueRule)
	if err != nil {
		return nil, err
	}
	return &EmitterTwoIterator{contract: _Emitter.contract, event: "Two", logs: logs, sub: sub}, nil
}

// WatchTwo is a free log subscription operation binding the contract event 0xce34f015a0e20f2b0daf980b28ed50729e87b993e4d30cca4c3f4da05acbd0ac.
//
// Solidity: event Two(uint256 indexed newValue, uint256 value)
func (_Emitter *EmitterFilterer) WatchTwo(opts *bind.WatchOpts, sink chan<- *EmitterTwo, newValue []*big.Int) (event.Subscription, error) {
	var newValueRule []interface{}
	for _, newValueItem := range newValue {
		newValueRule = append(newValueRule, newValueItem)
	}

	logs, sub, err := _Emitter.contract.WatchLogs(opts, "Two", newValueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EmitterTwo)
				if err := _Emitter.contract.UnpackLog(event, "Two", log); err != nil {
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

// ParseTwo is a log parse operation binding the contract event 0xce34f015a0e20f2b0daf980b28ed50729e87b993e4d30cca4c3f4da05acbd0ac.
//
// Solidity: event Two(uint256 indexed newValue, uint256 value)
func (_Emitter *EmitterFilterer) ParseTwo(log types.Log) (*EmitterTwo, error) {
	event := new(EmitterTwo)
	if err := _Emitter.contract.UnpackLog(event, "Two", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

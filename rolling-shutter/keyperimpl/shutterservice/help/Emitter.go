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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"five\",\"type\":\"bytes\"}],\"name\":\"Five\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"}],\"name\":\"Four\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"two\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"}],\"name\":\"SingleIdx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"two\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"three\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"five\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"six\",\"type\":\"bytes\"}],\"name\":\"Six\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Two\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"five\",\"type\":\"bytes\"}],\"name\":\"emitFive\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"}],\"name\":\"emitFour\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"two\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"}],\"name\":\"emitSingleIdx\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"two\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"three\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"five\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"six\",\"type\":\"bytes\"}],\"name\":\"emitSix\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"emitTwo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b506109f2806100206000396000f3fe608060405234801561001057600080fd5b50600436106100575760003560e01c80636995a2d91461005c5780637155880d146100785780638cc5e89214610094578063bb1c31ec146100b0578063f70770e0146100cc575b600080fd5b610076600480360381019061007191906103e7565b6100e8565b005b610092600480360381019061008d919061049a565b61012b565b005b6100ae60048036038101906100a991906104c7565b610167565b005b6100ca60048036038101906100c5919061054a565b6101a7565b005b6100e660048036038101906100e191906106b8565b6101e6565b005b8284867f2778059b9d45e2cd0df03a27bbe3e688dfc48aa15a729c42f39dcd986ebd4461858560405161011c929190610821565b60405180910390a45050505050565b807fce34f015a0e20f2b0daf980b28ed50729e87b993e4d30cca4c3f4da05acbd0ac600560405161015c919061089d565b60405180910390a250565b8183857fd82c9bd67140e94b50e0a62e800c51428267b0cd733573daaafad26b62c05afb8460405161019991906108b8565b60405180910390a450505050565b827f118f15bfd27a002429cfe56dc0757c904b3e8c0535f4f54771d8b184ecdf381483836040516101d99291906108e9565b60405180910390a2505050565b8373ffffffffffffffffffffffffffffffffffffffff168560405161020b9190610960565b6040518091039020877f1c49d7caedaa8e8be0f2ef7b3c285ccaef7eac5a7fe6b427cbe7e7d8ad30705486868660405161024793929190610977565b60405180910390a4505050505050565b6000604051905090565b600080fd5b600080fd5b6000819050919050565b61027e8161026b565b811461028957600080fd5b50565b60008135905061029b81610275565b92915050565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6102f4826102ab565b810181811067ffffffffffffffff82111715610313576103126102bc565b5b80604052505050565b6000610326610257565b905061033282826102eb565b919050565b600067ffffffffffffffff821115610352576103516102bc565b5b61035b826102ab565b9050602081019050919050565b82818337600083830152505050565b600061038a61038584610337565b61031c565b9050828152602081018484840111156103a6576103a56102a6565b5b6103b1848285610368565b509392505050565b600082601f8301126103ce576103cd6102a1565b5b81356103de848260208601610377565b91505092915050565b600080600080600060a0868803121561040357610402610261565b5b60006104118882890161028c565b95505060206104228882890161028c565b94505060406104338882890161028c565b935050606086013567ffffffffffffffff81111561045457610453610266565b5b610460888289016103b9565b925050608086013567ffffffffffffffff81111561048157610480610266565b5b61048d888289016103b9565b9150509295509295909350565b6000602082840312156104b0576104af610261565b5b60006104be8482850161028c565b91505092915050565b600080600080608085870312156104e1576104e0610261565b5b60006104ef8782880161028c565b94505060206105008782880161028c565b93505060406105118782880161028c565b925050606085013567ffffffffffffffff81111561053257610531610266565b5b61053e878288016103b9565b91505092959194509250565b60008060006060848603121561056357610562610261565b5b60006105718682870161028c565b935050602084013567ffffffffffffffff81111561059257610591610266565b5b61059e868287016103b9565b92505060406105af8682870161028c565b9150509250925092565b600067ffffffffffffffff8211156105d4576105d36102bc565b5b6105dd826102ab565b9050602081019050919050565b60006105fd6105f8846105b9565b61031c565b905082815260208101848484011115610619576106186102a6565b5b610624848285610368565b509392505050565b600082601f830112610641576106406102a1565b5b81356106518482602086016105ea565b91505092915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006106858261065a565b9050919050565b6106958161067a565b81146106a057600080fd5b50565b6000813590506106b28161068c565b92915050565b60008060008060008060c087890312156106d5576106d4610261565b5b60006106e389828a0161028c565b965050602087013567ffffffffffffffff81111561070457610703610266565b5b61071089828a0161062c565b955050604061072189828a016106a3565b945050606087013567ffffffffffffffff81111561074257610741610266565b5b61074e89828a016103b9565b935050608061075f89828a0161028c565b92505060a087013567ffffffffffffffff8111156107805761077f610266565b5b61078c89828a016103b9565b9150509295509295509295565b600081519050919050565b600082825260208201905092915050565b60005b838110156107d35780820151818401526020810190506107b8565b838111156107e2576000848401525b50505050565b60006107f382610799565b6107fd81856107a4565b935061080d8185602086016107b5565b610816816102ab565b840191505092915050565b6000604082019050818103600083015261083b81856107e8565b9050818103602083015261084f81846107e8565b90509392505050565b6000819050919050565b6000819050919050565b600061088761088261087d84610858565b610862565b61026b565b9050919050565b6108978161086c565b82525050565b60006020820190506108b2600083018461088e565b92915050565b600060208201905081810360008301526108d281846107e8565b905092915050565b6108e38161026b565b82525050565b6000604082019050818103600083015261090381856107e8565b905061091260208301846108da565b9392505050565b600081519050919050565b600081905092915050565b600061093a82610919565b6109448185610924565b93506109548185602086016107b5565b80840191505092915050565b600061096c828461092f565b915081905092915050565b6000606082019050818103600083015261099181866107e8565b90506109a060208301856108da565b81810360408301526109b281846107e8565b905094935050505056fea2646970667358221220095412f10dd044a1de6fe1692f95175a761ae75806a97261993a8ab82c278d7d64736f6c63430008090033",
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

// EmitSingleIdx is a paid mutator transaction binding the contract method 0xbb1c31ec.
//
// Solidity: function emitSingleIdx(uint256 one, bytes two, uint256 three) returns()
func (_Emitter *EmitterTransactor) EmitSingleIdx(opts *bind.TransactOpts, one *big.Int, two []byte, three *big.Int) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitSingleIdx", one, two, three)
}

// EmitSingleIdx is a paid mutator transaction binding the contract method 0xbb1c31ec.
//
// Solidity: function emitSingleIdx(uint256 one, bytes two, uint256 three) returns()
func (_Emitter *EmitterSession) EmitSingleIdx(one *big.Int, two []byte, three *big.Int) (*types.Transaction, error) {
	return _Emitter.Contract.EmitSingleIdx(&_Emitter.TransactOpts, one, two, three)
}

// EmitSingleIdx is a paid mutator transaction binding the contract method 0xbb1c31ec.
//
// Solidity: function emitSingleIdx(uint256 one, bytes two, uint256 three) returns()
func (_Emitter *EmitterTransactorSession) EmitSingleIdx(one *big.Int, two []byte, three *big.Int) (*types.Transaction, error) {
	return _Emitter.Contract.EmitSingleIdx(&_Emitter.TransactOpts, one, two, three)
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

// EmitterSingleIdxIterator is returned from FilterSingleIdx and is used to iterate over the raw logs and unpacked data for SingleIdx events raised by the Emitter contract.
type EmitterSingleIdxIterator struct {
	Event *EmitterSingleIdx // Event containing the contract specifics and raw log

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
func (it *EmitterSingleIdxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EmitterSingleIdx)
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
		it.Event = new(EmitterSingleIdx)
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
func (it *EmitterSingleIdxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EmitterSingleIdxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EmitterSingleIdx represents a SingleIdx event raised by the Emitter contract.
type EmitterSingleIdx struct {
	One   *big.Int
	Two   []byte
	Three *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSingleIdx is a free log retrieval operation binding the contract event 0x118f15bfd27a002429cfe56dc0757c904b3e8c0535f4f54771d8b184ecdf3814.
//
// Solidity: event SingleIdx(uint256 indexed one, bytes two, uint256 three)
func (_Emitter *EmitterFilterer) FilterSingleIdx(opts *bind.FilterOpts, one []*big.Int) (*EmitterSingleIdxIterator, error) {
	var oneRule []interface{}
	for _, oneItem := range one {
		oneRule = append(oneRule, oneItem)
	}

	logs, sub, err := _Emitter.contract.FilterLogs(opts, "SingleIdx", oneRule)
	if err != nil {
		return nil, err
	}
	return &EmitterSingleIdxIterator{contract: _Emitter.contract, event: "SingleIdx", logs: logs, sub: sub}, nil
}

// WatchSingleIdx is a free log subscription operation binding the contract event 0x118f15bfd27a002429cfe56dc0757c904b3e8c0535f4f54771d8b184ecdf3814.
//
// Solidity: event SingleIdx(uint256 indexed one, bytes two, uint256 three)
func (_Emitter *EmitterFilterer) WatchSingleIdx(opts *bind.WatchOpts, sink chan<- *EmitterSingleIdx, one []*big.Int) (event.Subscription, error) {
	var oneRule []interface{}
	for _, oneItem := range one {
		oneRule = append(oneRule, oneItem)
	}

	logs, sub, err := _Emitter.contract.WatchLogs(opts, "SingleIdx", oneRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EmitterSingleIdx)
				if err := _Emitter.contract.UnpackLog(event, "SingleIdx", log); err != nil {
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

// ParseSingleIdx is a log parse operation binding the contract event 0x118f15bfd27a002429cfe56dc0757c904b3e8c0535f4f54771d8b184ecdf3814.
//
// Solidity: event SingleIdx(uint256 indexed one, bytes two, uint256 three)
func (_Emitter *EmitterFilterer) ParseSingleIdx(log types.Log) (*EmitterSingleIdx, error) {
	event := new(EmitterSingleIdx)
	if err := _Emitter.contract.UnpackLog(event, "SingleIdx", log); err != nil {
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

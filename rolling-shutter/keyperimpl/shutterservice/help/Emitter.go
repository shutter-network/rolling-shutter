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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"blob\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"nums\",\"type\":\"uint256[]\"}],\"name\":\"DynamicArgs\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"five\",\"type\":\"bytes\"}],\"name\":\"Five\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"}],\"name\":\"Four\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"blob\",\"type\":\"bytes\"}],\"name\":\"IndexedDynamic\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"two\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"}],\"name\":\"SingleIdx\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"two\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"three\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"five\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"six\",\"type\":\"bytes\"}],\"name\":\"Six\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"ok\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"bytes4\",\"name\":\"sig\",\"type\":\"bytes4\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"tag\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"StaticArgs\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"TransferLike\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Two\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"blob\",\"type\":\"bytes\"},{\"internalType\":\"uint256[]\",\"name\":\"nums\",\"type\":\"uint256[]\"}],\"name\":\"emitDynamic\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"emitDynamicSample\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"five\",\"type\":\"bytes\"}],\"name\":\"emitFive\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"two\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"}],\"name\":\"emitFour\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"emitIndexedDynamicSample\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"two\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"three\",\"type\":\"uint256\"}],\"name\":\"emitSingleIdx\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"one\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"two\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"three\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"four\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"five\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"six\",\"type\":\"bytes\"}],\"name\":\"emitSix\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"ok\",\"type\":\"bool\"},{\"internalType\":\"bytes4\",\"name\":\"sig\",\"type\":\"bytes4\"},{\"internalType\":\"bytes32\",\"name\":\"tag\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"emitStaticCustom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"emitStaticSample\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"emitTransferLike\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"emitTwo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b506114a4806100206000396000f3fe608060405234801561001057600080fd5b50600436106100a95760003560e01c80639e40f880116100715780639e40f880146101285780639ea1943614610144578063bb1c31ec14610160578063cc2e28a71461017c578063d9c1406914610186578063f70770e014610190576100a9565b80631b97dd50146100ae5780636995a2d9146100b85780637155880d146100d45780638cc5e892146100f05780638d49ed1b1461010c575b600080fd5b6100b66101ac565b005b6100d260048036038101906100cd9190610738565b610202565b005b6100ee60048036038101906100e991906107eb565b610245565b005b61010a60048036038101906101059190610818565b610281565b005b61012660048036038101906101219190610a04565b6102c1565b005b610142600480360381019061013d9190610b09565b610301565b005b61015e60048036038101906101599190610c22565b61036b565b005b61017a60048036038101906101759190610c9d565b6103b1565b005b6101846103f0565b005b61018e6104bc565b005b6101aa60048036038101906101a59190610d0c565b610537565b005b6040516101b890610e44565b60405180910390206040516101cc90610eb0565b60405180910390207f9c8384689cf9db5fbe62d9a5d0d946cf1a932684350567d1018172005d95802360405160405180910390a3565b8284867f2778059b9d45e2cd0df03a27bbe3e688dfc48aa15a729c42f39dcd986ebd44618585604051610236929190610f4d565b60405180910390a45050505050565b807fce34f015a0e20f2b0daf980b28ed50729e87b993e4d30cca4c3f4da05acbd0ac60056040516102769190610fc9565b60405180910390a250565b8183857fd82c9bd67140e94b50e0a62e800c51428267b0cd733573daaafad26b62c05afb846040516102b39190610fe4565b60405180910390a450505050565b7f0b6ba7609aaae6128169edc7928038e334ca6a048e03deb28dba5559db7a52068383836040516102f493929190611119565b60405180910390a1505050565b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f3b183f1e11b0cc01bb9b37e9d7ad406d3c364a994e5cd954fd550ad3a0497e6f8360405161035e9190611174565b60405180910390a3505050565b7f7d73e750f0b0574a99b768e99ca9d451e8042fc0e8760e2a90b3625385573fa885858585856040516103a29594939291906111cb565b60405180910390a15050505050565b827f118f15bfd27a002429cfe56dc0757c904b3e8c0535f4f54771d8b184ecdf381483836040516103e392919061121e565b60405180910390a2505050565b6000600267ffffffffffffffff81111561040d5761040c61060d565b5b60405190808252806020026020018201604052801561043b5781602001602082028036833780820191505090505b5090506001816000815181106104545761045361124e565b5b6020026020010181815250506002816001815181106104765761047561124e565b5b6020026020010181815250507f0b6ba7609aaae6128169edc7928038e334ca6a048e03deb28dba5559db7a5206816040516104b191906112c3565b60405180910390a150565b7f7d73e750f0b0574a99b768e99ca9d451e8042fc0e8760e2a90b3625385573fa8731111111111111111111111111111111111111111600163deadbeef7f7461670000000000000000000000000000000000000000000000000000000000602a60405161052d95949392919061138e565b60405180910390a1565b8373ffffffffffffffffffffffffffffffffffffffff168560405161055c9190611412565b6040518091039020877f1c49d7caedaa8e8be0f2ef7b3c285ccaef7eac5a7fe6b427cbe7e7d8ad30705486868660405161059893929190611429565b60405180910390a4505050505050565b6000604051905090565b600080fd5b600080fd5b6000819050919050565b6105cf816105bc565b81146105da57600080fd5b50565b6000813590506105ec816105c6565b92915050565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b610645826105fc565b810181811067ffffffffffffffff821117156106645761066361060d565b5b80604052505050565b60006106776105a8565b9050610683828261063c565b919050565b600067ffffffffffffffff8211156106a3576106a261060d565b5b6106ac826105fc565b9050602081019050919050565b82818337600083830152505050565b60006106db6106d684610688565b61066d565b9050828152602081018484840111156106f7576106f66105f7565b5b6107028482856106b9565b509392505050565b600082601f83011261071f5761071e6105f2565b5b813561072f8482602086016106c8565b91505092915050565b600080600080600060a08688031215610754576107536105b2565b5b6000610762888289016105dd565b9550506020610773888289016105dd565b9450506040610784888289016105dd565b935050606086013567ffffffffffffffff8111156107a5576107a46105b7565b5b6107b18882890161070a565b925050608086013567ffffffffffffffff8111156107d2576107d16105b7565b5b6107de8882890161070a565b9150509295509295909350565b600060208284031215610801576108006105b2565b5b600061080f848285016105dd565b91505092915050565b60008060008060808587031215610832576108316105b2565b5b6000610840878288016105dd565b9450506020610851878288016105dd565b9350506040610862878288016105dd565b925050606085013567ffffffffffffffff811115610883576108826105b7565b5b61088f8782880161070a565b91505092959194509250565b600067ffffffffffffffff8211156108b6576108b561060d565b5b6108bf826105fc565b9050602081019050919050565b60006108df6108da8461089b565b61066d565b9050828152602081018484840111156108fb576108fa6105f7565b5b6109068482856106b9565b509392505050565b600082601f830112610923576109226105f2565b5b81356109338482602086016108cc565b91505092915050565b600067ffffffffffffffff8211156109575761095661060d565b5b602082029050602081019050919050565b600080fd5b600061098061097b8461093c565b61066d565b905080838252602082019050602084028301858111156109a3576109a2610968565b5b835b818110156109cc57806109b888826105dd565b8452602084019350506020810190506109a5565b5050509392505050565b600082601f8301126109eb576109ea6105f2565b5b81356109fb84826020860161096d565b91505092915050565b600080600060608486031215610a1d57610a1c6105b2565b5b600084013567ffffffffffffffff811115610a3b57610a3a6105b7565b5b610a478682870161090e565b935050602084013567ffffffffffffffff811115610a6857610a676105b7565b5b610a748682870161070a565b925050604084013567ffffffffffffffff811115610a9557610a946105b7565b5b610aa1868287016109d6565b9150509250925092565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610ad682610aab565b9050919050565b610ae681610acb565b8114610af157600080fd5b50565b600081359050610b0381610add565b92915050565b600080600060608486031215610b2257610b216105b2565b5b6000610b3086828701610af4565b9350506020610b4186828701610af4565b9250506040610b52868287016105dd565b9150509250925092565b60008115159050919050565b610b7181610b5c565b8114610b7c57600080fd5b50565b600081359050610b8e81610b68565b92915050565b60007fffffffff0000000000000000000000000000000000000000000000000000000082169050919050565b610bc981610b94565b8114610bd457600080fd5b50565b600081359050610be681610bc0565b92915050565b6000819050919050565b610bff81610bec565b8114610c0a57600080fd5b50565b600081359050610c1c81610bf6565b92915050565b600080600080600060a08688031215610c3e57610c3d6105b2565b5b6000610c4c88828901610af4565b9550506020610c5d88828901610b7f565b9450506040610c6e88828901610bd7565b9350506060610c7f88828901610c0d565b9250506080610c90888289016105dd565b9150509295509295909350565b600080600060608486031215610cb657610cb56105b2565b5b6000610cc4868287016105dd565b935050602084013567ffffffffffffffff811115610ce557610ce46105b7565b5b610cf18682870161070a565b9250506040610d02868287016105dd565b9150509250925092565b60008060008060008060c08789031215610d2957610d286105b2565b5b6000610d3789828a016105dd565b965050602087013567ffffffffffffffff811115610d5857610d576105b7565b5b610d6489828a0161090e565b9550506040610d7589828a01610af4565b945050606087013567ffffffffffffffff811115610d9657610d956105b7565b5b610da289828a0161070a565b9350506080610db389828a016105dd565b92505060a087013567ffffffffffffffff811115610dd457610dd36105b7565b5b610de089828a0161070a565b9150509295509295509295565b600081905092915050565b7fbeef000000000000000000000000000000000000000000000000000000000000600082015250565b6000610e2e600283610ded565b9150610e3982610df8565b600282019050919050565b6000610e4f82610e21565b9150819050919050565b600081905092915050565b7f68656c6c6f000000000000000000000000000000000000000000000000000000600082015250565b6000610e9a600583610e59565b9150610ea582610e64565b600582019050919050565b6000610ebb82610e8d565b9150819050919050565b600081519050919050565b600082825260208201905092915050565b60005b83811015610eff578082015181840152602081019050610ee4565b83811115610f0e576000848401525b50505050565b6000610f1f82610ec5565b610f298185610ed0565b9350610f39818560208601610ee1565b610f42816105fc565b840191505092915050565b60006040820190508181036000830152610f678185610f14565b90508181036020830152610f7b8184610f14565b90509392505050565b6000819050919050565b6000819050919050565b6000610fb3610fae610fa984610f84565b610f8e565b6105bc565b9050919050565b610fc381610f98565b82525050565b6000602082019050610fde6000830184610fba565b92915050565b60006020820190508181036000830152610ffe8184610f14565b905092915050565b600081519050919050565b600082825260208201905092915050565b600061102d82611006565b6110378185611011565b9350611047818560208601610ee1565b611050816105fc565b840191505092915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b611090816105bc565b82525050565b60006110a28383611087565b60208301905092915050565b6000602082019050919050565b60006110c68261105b565b6110d08185611066565b93506110db83611077565b8060005b8381101561110c5781516110f38882611096565b97506110fe836110ae565b9250506001810190506110df565b5085935050505092915050565b600060608201905081810360008301526111338186611022565b905081810360208301526111478185610f14565b9050818103604083015261115b81846110bb565b9050949350505050565b61116e816105bc565b82525050565b60006020820190506111896000830184611165565b92915050565b61119881610acb565b82525050565b6111a781610b5c565b82525050565b6111b681610b94565b82525050565b6111c581610bec565b82525050565b600060a0820190506111e0600083018861118f565b6111ed602083018761119e565b6111fa60408301866111ad565b61120760608301856111bc565b6112146080830184611165565b9695505050505050565b600060408201905081810360008301526112388185610f14565b90506112476020830184611165565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b600061128a600583611011565b915061129582610e64565b602082019050919050565b60006112ad600283610ed0565b91506112b882610df8565b602082019050919050565b600060608201905081810360008301526112dc8161127d565b905081810360208301526112ef816112a0565b9050818103604083015261130381846110bb565b905092915050565b6000819050919050565b60008160e01b9050919050565b600061133d6113386113338461130b565b611315565b610b94565b9050919050565b61134d81611322565b82525050565b6000819050919050565b600061137861137361136e84611353565b610f8e565b6105bc565b9050919050565b6113888161135d565b82525050565b600060a0820190506113a3600083018861118f565b6113b0602083018761119e565b6113bd6040830186611344565b6113ca60608301856111bc565b6113d7608083018461137f565b9695505050505050565b60006113ec82611006565b6113f68185610e59565b9350611406818560208601610ee1565b80840191505092915050565b600061141e82846113e1565b915081905092915050565b600060608201905081810360008301526114438186610f14565b90506114526020830185611165565b81810360408301526114648184610f14565b905094935050505056fea2646970667358221220582cbc7700dc91f443692c8bf652c7ee02d4f4d4110bcc1059e31bcc7bd47b8764736f6c63430008090033",
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

// EmitDynamic is a paid mutator transaction binding the contract method 0x8d49ed1b.
//
// Solidity: function emitDynamic(string note, bytes blob, uint256[] nums) returns()
func (_Emitter *EmitterTransactor) EmitDynamic(opts *bind.TransactOpts, note string, blob []byte, nums []*big.Int) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitDynamic", note, blob, nums)
}

// EmitDynamic is a paid mutator transaction binding the contract method 0x8d49ed1b.
//
// Solidity: function emitDynamic(string note, bytes blob, uint256[] nums) returns()
func (_Emitter *EmitterSession) EmitDynamic(note string, blob []byte, nums []*big.Int) (*types.Transaction, error) {
	return _Emitter.Contract.EmitDynamic(&_Emitter.TransactOpts, note, blob, nums)
}

// EmitDynamic is a paid mutator transaction binding the contract method 0x8d49ed1b.
//
// Solidity: function emitDynamic(string note, bytes blob, uint256[] nums) returns()
func (_Emitter *EmitterTransactorSession) EmitDynamic(note string, blob []byte, nums []*big.Int) (*types.Transaction, error) {
	return _Emitter.Contract.EmitDynamic(&_Emitter.TransactOpts, note, blob, nums)
}

// EmitDynamicSample is a paid mutator transaction binding the contract method 0xcc2e28a7.
//
// Solidity: function emitDynamicSample() returns()
func (_Emitter *EmitterTransactor) EmitDynamicSample(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitDynamicSample")
}

// EmitDynamicSample is a paid mutator transaction binding the contract method 0xcc2e28a7.
//
// Solidity: function emitDynamicSample() returns()
func (_Emitter *EmitterSession) EmitDynamicSample() (*types.Transaction, error) {
	return _Emitter.Contract.EmitDynamicSample(&_Emitter.TransactOpts)
}

// EmitDynamicSample is a paid mutator transaction binding the contract method 0xcc2e28a7.
//
// Solidity: function emitDynamicSample() returns()
func (_Emitter *EmitterTransactorSession) EmitDynamicSample() (*types.Transaction, error) {
	return _Emitter.Contract.EmitDynamicSample(&_Emitter.TransactOpts)
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

// EmitIndexedDynamicSample is a paid mutator transaction binding the contract method 0x1b97dd50.
//
// Solidity: function emitIndexedDynamicSample() returns()
func (_Emitter *EmitterTransactor) EmitIndexedDynamicSample(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitIndexedDynamicSample")
}

// EmitIndexedDynamicSample is a paid mutator transaction binding the contract method 0x1b97dd50.
//
// Solidity: function emitIndexedDynamicSample() returns()
func (_Emitter *EmitterSession) EmitIndexedDynamicSample() (*types.Transaction, error) {
	return _Emitter.Contract.EmitIndexedDynamicSample(&_Emitter.TransactOpts)
}

// EmitIndexedDynamicSample is a paid mutator transaction binding the contract method 0x1b97dd50.
//
// Solidity: function emitIndexedDynamicSample() returns()
func (_Emitter *EmitterTransactorSession) EmitIndexedDynamicSample() (*types.Transaction, error) {
	return _Emitter.Contract.EmitIndexedDynamicSample(&_Emitter.TransactOpts)
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

// EmitStaticCustom is a paid mutator transaction binding the contract method 0x9ea19436.
//
// Solidity: function emitStaticCustom(address user, bool ok, bytes4 sig, bytes32 tag, uint256 amount) returns()
func (_Emitter *EmitterTransactor) EmitStaticCustom(opts *bind.TransactOpts, user common.Address, ok bool, sig [4]byte, tag [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitStaticCustom", user, ok, sig, tag, amount)
}

// EmitStaticCustom is a paid mutator transaction binding the contract method 0x9ea19436.
//
// Solidity: function emitStaticCustom(address user, bool ok, bytes4 sig, bytes32 tag, uint256 amount) returns()
func (_Emitter *EmitterSession) EmitStaticCustom(user common.Address, ok bool, sig [4]byte, tag [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Emitter.Contract.EmitStaticCustom(&_Emitter.TransactOpts, user, ok, sig, tag, amount)
}

// EmitStaticCustom is a paid mutator transaction binding the contract method 0x9ea19436.
//
// Solidity: function emitStaticCustom(address user, bool ok, bytes4 sig, bytes32 tag, uint256 amount) returns()
func (_Emitter *EmitterTransactorSession) EmitStaticCustom(user common.Address, ok bool, sig [4]byte, tag [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Emitter.Contract.EmitStaticCustom(&_Emitter.TransactOpts, user, ok, sig, tag, amount)
}

// EmitStaticSample is a paid mutator transaction binding the contract method 0xd9c14069.
//
// Solidity: function emitStaticSample() returns()
func (_Emitter *EmitterTransactor) EmitStaticSample(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitStaticSample")
}

// EmitStaticSample is a paid mutator transaction binding the contract method 0xd9c14069.
//
// Solidity: function emitStaticSample() returns()
func (_Emitter *EmitterSession) EmitStaticSample() (*types.Transaction, error) {
	return _Emitter.Contract.EmitStaticSample(&_Emitter.TransactOpts)
}

// EmitStaticSample is a paid mutator transaction binding the contract method 0xd9c14069.
//
// Solidity: function emitStaticSample() returns()
func (_Emitter *EmitterTransactorSession) EmitStaticSample() (*types.Transaction, error) {
	return _Emitter.Contract.EmitStaticSample(&_Emitter.TransactOpts)
}

// EmitTransferLike is a paid mutator transaction binding the contract method 0x9e40f880.
//
// Solidity: function emitTransferLike(address from, address to, uint256 value) returns()
func (_Emitter *EmitterTransactor) EmitTransferLike(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _Emitter.contract.Transact(opts, "emitTransferLike", from, to, value)
}

// EmitTransferLike is a paid mutator transaction binding the contract method 0x9e40f880.
//
// Solidity: function emitTransferLike(address from, address to, uint256 value) returns()
func (_Emitter *EmitterSession) EmitTransferLike(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _Emitter.Contract.EmitTransferLike(&_Emitter.TransactOpts, from, to, value)
}

// EmitTransferLike is a paid mutator transaction binding the contract method 0x9e40f880.
//
// Solidity: function emitTransferLike(address from, address to, uint256 value) returns()
func (_Emitter *EmitterTransactorSession) EmitTransferLike(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _Emitter.Contract.EmitTransferLike(&_Emitter.TransactOpts, from, to, value)
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

// EmitterDynamicArgsIterator is returned from FilterDynamicArgs and is used to iterate over the raw logs and unpacked data for DynamicArgs events raised by the Emitter contract.
type EmitterDynamicArgsIterator struct {
	Event *EmitterDynamicArgs // Event containing the contract specifics and raw log

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
func (it *EmitterDynamicArgsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EmitterDynamicArgs)
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
		it.Event = new(EmitterDynamicArgs)
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
func (it *EmitterDynamicArgsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EmitterDynamicArgsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EmitterDynamicArgs represents a DynamicArgs event raised by the Emitter contract.
type EmitterDynamicArgs struct {
	Note string
	Blob []byte
	Nums []*big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDynamicArgs is a free log retrieval operation binding the contract event 0x0b6ba7609aaae6128169edc7928038e334ca6a048e03deb28dba5559db7a5206.
//
// Solidity: event DynamicArgs(string note, bytes blob, uint256[] nums)
func (_Emitter *EmitterFilterer) FilterDynamicArgs(opts *bind.FilterOpts) (*EmitterDynamicArgsIterator, error) {
	logs, sub, err := _Emitter.contract.FilterLogs(opts, "DynamicArgs")
	if err != nil {
		return nil, err
	}
	return &EmitterDynamicArgsIterator{contract: _Emitter.contract, event: "DynamicArgs", logs: logs, sub: sub}, nil
}

// WatchDynamicArgs is a free log subscription operation binding the contract event 0x0b6ba7609aaae6128169edc7928038e334ca6a048e03deb28dba5559db7a5206.
//
// Solidity: event DynamicArgs(string note, bytes blob, uint256[] nums)
func (_Emitter *EmitterFilterer) WatchDynamicArgs(opts *bind.WatchOpts, sink chan<- *EmitterDynamicArgs) (event.Subscription, error) {
	logs, sub, err := _Emitter.contract.WatchLogs(opts, "DynamicArgs")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EmitterDynamicArgs)
				if err := _Emitter.contract.UnpackLog(event, "DynamicArgs", log); err != nil {
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

// ParseDynamicArgs is a log parse operation binding the contract event 0x0b6ba7609aaae6128169edc7928038e334ca6a048e03deb28dba5559db7a5206.
//
// Solidity: event DynamicArgs(string note, bytes blob, uint256[] nums)
func (_Emitter *EmitterFilterer) ParseDynamicArgs(log types.Log) (*EmitterDynamicArgs, error) {
	event := new(EmitterDynamicArgs)
	if err := _Emitter.contract.UnpackLog(event, "DynamicArgs", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
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

// EmitterIndexedDynamicIterator is returned from FilterIndexedDynamic and is used to iterate over the raw logs and unpacked data for IndexedDynamic events raised by the Emitter contract.
type EmitterIndexedDynamicIterator struct {
	Event *EmitterIndexedDynamic // Event containing the contract specifics and raw log

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
func (it *EmitterIndexedDynamicIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EmitterIndexedDynamic)
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
		it.Event = new(EmitterIndexedDynamic)
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
func (it *EmitterIndexedDynamicIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EmitterIndexedDynamicIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EmitterIndexedDynamic represents a IndexedDynamic event raised by the Emitter contract.
type EmitterIndexedDynamic struct {
	Note common.Hash
	Blob common.Hash
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterIndexedDynamic is a free log retrieval operation binding the contract event 0x9c8384689cf9db5fbe62d9a5d0d946cf1a932684350567d1018172005d958023.
//
// Solidity: event IndexedDynamic(string indexed note, bytes indexed blob)
func (_Emitter *EmitterFilterer) FilterIndexedDynamic(opts *bind.FilterOpts, note []string, blob [][]byte) (*EmitterIndexedDynamicIterator, error) {
	var noteRule []interface{}
	for _, noteItem := range note {
		noteRule = append(noteRule, noteItem)
	}
	var blobRule []interface{}
	for _, blobItem := range blob {
		blobRule = append(blobRule, blobItem)
	}

	logs, sub, err := _Emitter.contract.FilterLogs(opts, "IndexedDynamic", noteRule, blobRule)
	if err != nil {
		return nil, err
	}
	return &EmitterIndexedDynamicIterator{contract: _Emitter.contract, event: "IndexedDynamic", logs: logs, sub: sub}, nil
}

// WatchIndexedDynamic is a free log subscription operation binding the contract event 0x9c8384689cf9db5fbe62d9a5d0d946cf1a932684350567d1018172005d958023.
//
// Solidity: event IndexedDynamic(string indexed note, bytes indexed blob)
func (_Emitter *EmitterFilterer) WatchIndexedDynamic(opts *bind.WatchOpts, sink chan<- *EmitterIndexedDynamic, note []string, blob [][]byte) (event.Subscription, error) {
	var noteRule []interface{}
	for _, noteItem := range note {
		noteRule = append(noteRule, noteItem)
	}
	var blobRule []interface{}
	for _, blobItem := range blob {
		blobRule = append(blobRule, blobItem)
	}

	logs, sub, err := _Emitter.contract.WatchLogs(opts, "IndexedDynamic", noteRule, blobRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EmitterIndexedDynamic)
				if err := _Emitter.contract.UnpackLog(event, "IndexedDynamic", log); err != nil {
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

// ParseIndexedDynamic is a log parse operation binding the contract event 0x9c8384689cf9db5fbe62d9a5d0d946cf1a932684350567d1018172005d958023.
//
// Solidity: event IndexedDynamic(string indexed note, bytes indexed blob)
func (_Emitter *EmitterFilterer) ParseIndexedDynamic(log types.Log) (*EmitterIndexedDynamic, error) {
	event := new(EmitterIndexedDynamic)
	if err := _Emitter.contract.UnpackLog(event, "IndexedDynamic", log); err != nil {
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

// EmitterStaticArgsIterator is returned from FilterStaticArgs and is used to iterate over the raw logs and unpacked data for StaticArgs events raised by the Emitter contract.
type EmitterStaticArgsIterator struct {
	Event *EmitterStaticArgs // Event containing the contract specifics and raw log

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
func (it *EmitterStaticArgsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EmitterStaticArgs)
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
		it.Event = new(EmitterStaticArgs)
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
func (it *EmitterStaticArgsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EmitterStaticArgsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EmitterStaticArgs represents a StaticArgs event raised by the Emitter contract.
type EmitterStaticArgs struct {
	User   common.Address
	Ok     bool
	Sig    [4]byte
	Tag    [32]byte
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterStaticArgs is a free log retrieval operation binding the contract event 0x7d73e750f0b0574a99b768e99ca9d451e8042fc0e8760e2a90b3625385573fa8.
//
// Solidity: event StaticArgs(address user, bool ok, bytes4 sig, bytes32 tag, uint256 amount)
func (_Emitter *EmitterFilterer) FilterStaticArgs(opts *bind.FilterOpts) (*EmitterStaticArgsIterator, error) {
	logs, sub, err := _Emitter.contract.FilterLogs(opts, "StaticArgs")
	if err != nil {
		return nil, err
	}
	return &EmitterStaticArgsIterator{contract: _Emitter.contract, event: "StaticArgs", logs: logs, sub: sub}, nil
}

// WatchStaticArgs is a free log subscription operation binding the contract event 0x7d73e750f0b0574a99b768e99ca9d451e8042fc0e8760e2a90b3625385573fa8.
//
// Solidity: event StaticArgs(address user, bool ok, bytes4 sig, bytes32 tag, uint256 amount)
func (_Emitter *EmitterFilterer) WatchStaticArgs(opts *bind.WatchOpts, sink chan<- *EmitterStaticArgs) (event.Subscription, error) {
	logs, sub, err := _Emitter.contract.WatchLogs(opts, "StaticArgs")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EmitterStaticArgs)
				if err := _Emitter.contract.UnpackLog(event, "StaticArgs", log); err != nil {
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

// ParseStaticArgs is a log parse operation binding the contract event 0x7d73e750f0b0574a99b768e99ca9d451e8042fc0e8760e2a90b3625385573fa8.
//
// Solidity: event StaticArgs(address user, bool ok, bytes4 sig, bytes32 tag, uint256 amount)
func (_Emitter *EmitterFilterer) ParseStaticArgs(log types.Log) (*EmitterStaticArgs, error) {
	event := new(EmitterStaticArgs)
	if err := _Emitter.contract.UnpackLog(event, "StaticArgs", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EmitterTransferLikeIterator is returned from FilterTransferLike and is used to iterate over the raw logs and unpacked data for TransferLike events raised by the Emitter contract.
type EmitterTransferLikeIterator struct {
	Event *EmitterTransferLike // Event containing the contract specifics and raw log

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
func (it *EmitterTransferLikeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EmitterTransferLike)
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
		it.Event = new(EmitterTransferLike)
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
func (it *EmitterTransferLikeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EmitterTransferLikeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EmitterTransferLike represents a TransferLike event raised by the Emitter contract.
type EmitterTransferLike struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransferLike is a free log retrieval operation binding the contract event 0x3b183f1e11b0cc01bb9b37e9d7ad406d3c364a994e5cd954fd550ad3a0497e6f.
//
// Solidity: event TransferLike(address indexed from, address indexed to, uint256 value)
func (_Emitter *EmitterFilterer) FilterTransferLike(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*EmitterTransferLikeIterator, error) {
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Emitter.contract.FilterLogs(opts, "TransferLike", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &EmitterTransferLikeIterator{contract: _Emitter.contract, event: "TransferLike", logs: logs, sub: sub}, nil
}

// WatchTransferLike is a free log subscription operation binding the contract event 0x3b183f1e11b0cc01bb9b37e9d7ad406d3c364a994e5cd954fd550ad3a0497e6f.
//
// Solidity: event TransferLike(address indexed from, address indexed to, uint256 value)
func (_Emitter *EmitterFilterer) WatchTransferLike(opts *bind.WatchOpts, sink chan<- *EmitterTransferLike, from []common.Address, to []common.Address) (event.Subscription, error) {
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Emitter.contract.WatchLogs(opts, "TransferLike", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EmitterTransferLike)
				if err := _Emitter.contract.UnpackLog(event, "TransferLike", log); err != nil {
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

// ParseTransferLike is a log parse operation binding the contract event 0x3b183f1e11b0cc01bb9b37e9d7ad406d3c364a994e5cd954fd550ad3a0497e6f.
//
// Solidity: event TransferLike(address indexed from, address indexed to, uint256 value)
func (_Emitter *EmitterFilterer) ParseTransferLike(log types.Log) (*EmitterTransferLike, error) {
	event := new(EmitterTransferLike)
	if err := _Emitter.contract.UnpackLog(event, "TransferLike", log); err != nil {
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

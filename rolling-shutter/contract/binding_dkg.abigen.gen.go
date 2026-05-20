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

// DKGContractMetaData contains all meta data concerning the DKGContract contract.
var DKGContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"phaseLength\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"dkgLeadLength\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"keyperSetManagerAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"keyBroadcastContractAddress\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"DKG_LEAD_LENGTH\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PHASE_LENGTH\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"}],\"name\":\"currentPhase\",\"outputs\":[{\"internalType\":\"enumDKGContract.Phase\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"cycleLength\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"}],\"name\":\"dkgStart\",\"outputs\":[{\"internalType\":\"int256\",\"name\":\"\",\"type\":\"int256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"hasVoted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"keyBroadcastContract\",\"outputs\":[{\"internalType\":\"contractKeyBroadcastContract\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"keyperSetManager\",\"outputs\":[{\"internalType\":\"contractKeyperSetManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"keyperIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64[]\",\"name\":\"accusedIndices\",\"type\":\"uint64[]\"}],\"name\":\"submitAccusation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"keyperIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64[]\",\"name\":\"accuserIndices\",\"type\":\"uint64[]\"},{\"internalType\":\"bytes[]\",\"name\":\"polyEvalData\",\"type\":\"bytes[]\"}],\"name\":\"submitApology\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"keyperIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes\",\"name\":\"commitment\",\"type\":\"bytes\"},{\"internalType\":\"bytes[]\",\"name\":\"polyEvals\",\"type\":\"bytes[]\"}],\"name\":\"submitDealing\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"keyperIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes\",\"name\":\"eonPublicKey\",\"type\":\"bytes\"}],\"name\":\"submitSuccessVote\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"name\":\"succeeded\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"voteCount\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"keyperIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64[]\",\"name\":\"accusedIndices\",\"type\":\"uint64[]\"}],\"name\":\"AccusationSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"keyperIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64[]\",\"name\":\"accuserIndices\",\"type\":\"uint64[]\"},{\"indexed\":false,\"internalType\":\"bytes[]\",\"name\":\"polyEvalData\",\"type\":\"bytes[]\"}],\"name\":\"ApologySubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"eonPublicKey\",\"type\":\"bytes\"}],\"name\":\"DKGSucceeded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"keyperIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"commitment\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes[]\",\"name\":\"polyEvals\",\"type\":\"bytes[]\"}],\"name\":\"DealingSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"retryCounter\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"keyperIndex\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"eonPublicKey\",\"type\":\"bytes\"}],\"name\":\"SuccessVoteSubmitted\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"AlreadySucceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"AlreadyVoted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EmptyAccusation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MismatchedArrays\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotKeyperAtIndex\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WrongPhase\",\"type\":\"error\"}]",
	Bin: "0x0x610100604052348015610010575f5ffd5b50604051611efb380380611efb83398181016040528101906100329190610176565b8367ffffffffffffffff1660808167ffffffffffffffff16815250508267ffffffffffffffff1660a08167ffffffffffffffff16815250508173ffffffffffffffffffffffffffffffffffffffff1660c08173ffffffffffffffffffffffffffffffffffffffff16815250508073ffffffffffffffffffffffffffffffffffffffff1660e08173ffffffffffffffffffffffffffffffffffffffff1681525050505050506101da565b5f5ffd5b5f67ffffffffffffffff82169050919050565b6100fb816100df565b8114610105575f5ffd5b50565b5f81519050610116816100f2565b92915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6101458261011c565b9050919050565b6101558161013b565b811461015f575f5ffd5b50565b5f815190506101708161014c565b92915050565b5f5f5f5f6080858703121561018e5761018d6100db565b5b5f61019b87828801610108565b94505060206101ac87828801610108565b93505060406101bd87828801610162565b92505060606101ce87828801610162565b91505092959194509250565b60805160a05160c05160e051611cbd61023e5f395f81816109380152610b1e01525f81816107d701528181610afa01528181610c780152610e6301525f81816103c30152610d3901525f818161030201528181610ad60152610b430152611cbd5ff3fe608060405234801561000f575f5ffd5b50600436106100e8575f3560e01c8063b209e7171161008a578063eac471a011610064578063eac471a014610248578063eb05c28014610266578063f01275b014610296578063f48f3d14146102b2576100e8565b8063b209e717146101ee578063d4b2cedd1461020c578063df6406311461022a576100e8565b80637642d0de116100c65780637642d0de1461015657806378a1b9db1461018657806382b870d3146101b657806394afc12d146101d2576100e8565b80630a3ec4d6146100ec57806316b89c3f1461011c5780633f9800981461013a575b5f5ffd5b61010660048036038101906101019190611021565b6102e2565b60405161011391906110d2565b60405180910390f35b6101246103c1565b60405161013191906110fa565b60405180910390f35b610154600480360381019061014f91906111c9565b6103e5565b005b610170600480360381019061016b91906112da565b61046b565b60405161017d9190611344565b60405180910390f35b6101a0600480360381019061019b919061135d565b6104a0565b6040516101ad9190611344565b60405180910390f35b6101d060048036038101906101cb9190611388565b6104bc565b005b6101ec60048036038101906101e79190611461565b610a18565b005b6101f6610ad4565b60405161020391906110fa565b60405180910390f35b610214610af8565b6040516102219190611540565b60405180910390f35b610232610b1c565b60405161023f9190611579565b60405180910390f35b610250610b40565b60405161025d91906110fa565b60405180910390f35b610280600480360381019061027b91906115c5565b610b73565b60405161028d91906110fa565b60405180910390f35b6102b060048036038101906102ab9190611615565b610baf565b005b6102cc60048036038101906102c79190611021565b610c74565b6040516102d991906116e4565b60405180910390f35b5f5f6102ee8484610c74565b90505f81436102fd919061172a565b90505f7f000000000000000000000000000000000000000000000000000000000000000067ffffffffffffffff1690505f821215610340575f93505050506103bb565b8082121561035457600193505050506103bb565b806002610361919061176a565b82121561037457600293505050506103bb565b806003610381919061176a565b82121561039457600393505050506103bb565b8060046103a1919061176a565b8212156103b457600493505050506103bb565b5f93505050505b92915050565b7f000000000000000000000000000000000000000000000000000000000000000081565b6103ee87610d8a565b6103fa87876001610df5565b6104048786610e60565b8467ffffffffffffffff168667ffffffffffffffff168867ffffffffffffffff167fa5074728b92791250d48ccdc55266aca7b1cca22f89e1ef2ab91ccd7ae0f337d8787878760405161045a9493929190611992565b60405180910390a450505050505050565b6002602052825f5260405f20602052815f5260405f20602052805f5260405f205f92509250509054906101000a900460ff1681565b5f602052805f5260405f205f915054906101000a900460ff1681565b6104c885856004610df5565b6104d28584610e60565b60025f8667ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f8567ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff1615610599576040517f7c9a1cf900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600160025f8767ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f8667ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff0219169083151502179055505f82826040516106459291906119f9565b604051809103902090505f6001805f8967ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f8867ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f8481526020019081526020015f205f9054906101000a900467ffffffffffffffff166106c69190611a11565b90508060015f8967ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f8867ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f8481526020019081526020015f205f6101000a81548167ffffffffffffffff021916908367ffffffffffffffff1602179055508467ffffffffffffffff168667ffffffffffffffff168867ffffffffffffffff167facfc0aa3f6ba36ac0ffda53e00d2cdace2fd4a9e41e2fe3214032794332c37c38787604051610797929190611a4c565b60405180910390a45f5f8867ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff16610a0f575f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663f90f3bed896040518263ffffffff1660e01b815260040161082e91906110fa565b602060405180830381865afa158015610849573d5f5f3e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061086d9190611a82565b90505f8173ffffffffffffffffffffffffffffffffffffffff1663e75235b86040518163ffffffff1660e01b8152600401602060405180830381865afa1580156108b9573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906108dd9190611ac1565b90508067ffffffffffffffff168367ffffffffffffffff1610610a0c5760015f5f8b67ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f6101000a81548160ff0219169083151502179055507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663daade8e88a88886040518463ffffffff1660e01b815260040161099393929190611aec565b5f604051808303815f87803b1580156109aa575f5ffd5b505af19250505080156109bb575060015b508767ffffffffffffffff168967ffffffffffffffff167fe6c9bd3daa50b9218615605cdf333410d405b7f3b8dbeee5a2373769679aa9f58888604051610a03929190611a4c565b60405180910390a35b50505b50505050505050565b610a2185610d8a565b610a2d85856002610df5565b610a378584610e60565b5f8282905003610a73576040517f2eeffe0b00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b8267ffffffffffffffff168467ffffffffffffffff168667ffffffffffffffff167fb5cc05e740a503c13e1b7ac3ad2d5abf41b91d2e5fca79455d34e515fb70849a8585604051610ac5929190611bd8565b60405180910390a45050505050565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b5f7f00000000000000000000000000000000000000000000000000000000000000006004610b6e9190611bfa565b905090565b6001602052825f5260405f20602052815f5260405f20602052805f5260405f205f92509250509054906101000a900467ffffffffffffffff1681565b610bb887610d8a565b610bc487876003610df5565b610bce8786610e60565b818190508484905014610c0d576040517fa121188700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b8467ffffffffffffffff168667ffffffffffffffff168867ffffffffffffffff167f62729f0820cf3ab60fbd63a51cdeae0554d0ab24f843b86bd0d2591594d6913687878787604051610c639493929190611c36565b60405180910390a450505050505050565b5f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663636df979856040518263ffffffff1660e01b8152600401610ccf91906110fa565b602060405180830381865afa158015610cea573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610d0e9190611ac1565b9050610d18610b40565b67ffffffffffffffff168367ffffffffffffffff16610d37919061176a565b7f000000000000000000000000000000000000000000000000000000000000000067ffffffffffffffff168267ffffffffffffffff16610d77919061172a565b610d819190611c6f565b91505092915050565b5f5f8267ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff1615610df2576040517fe0f5ec9d00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b50565b806004811115610e0857610e0761105f565b5b610e1284846102e2565b6004811115610e2457610e2361105f565b5b14610e5b576040517fe2586bcc00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b505050565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663f90f3bed846040518263ffffffff1660e01b8152600401610eba91906110fa565b602060405180830381865afa158015610ed5573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610ef99190611a82565b90503373ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16632e8e6cad846040518263ffffffff1660e01b8152600401610f4b91906110fa565b602060405180830381865afa158015610f66573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610f8a9190611a82565b73ffffffffffffffffffffffffffffffffffffffff1614610fd7576040517f7bf9580300000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b505050565b5f5ffd5b5f5ffd5b5f67ffffffffffffffff82169050919050565b61100081610fe4565b811461100a575f5ffd5b50565b5f8135905061101b81610ff7565b92915050565b5f5f6040838503121561103757611036610fdc565b5b5f6110448582860161100d565b92505060206110558582860161100d565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffd5b6005811061109d5761109c61105f565b5b50565b5f8190506110ad8261108c565b919050565b5f6110bc826110a0565b9050919050565b6110cc816110b2565b82525050565b5f6020820190506110e55f8301846110c3565b92915050565b6110f481610fe4565b82525050565b5f60208201905061110d5f8301846110eb565b92915050565b5f5ffd5b5f5ffd5b5f5ffd5b5f5f83601f84011261113457611133611113565b5b8235905067ffffffffffffffff81111561115157611150611117565b5b60208301915083600182028301111561116d5761116c61111b565b5b9250929050565b5f5f83601f84011261118957611188611113565b5b8235905067ffffffffffffffff8111156111a6576111a5611117565b5b6020830191508360208202830111156111c2576111c161111b565b5b9250929050565b5f5f5f5f5f5f5f60a0888a0312156111e4576111e3610fdc565b5b5f6111f18a828b0161100d565b97505060206112028a828b0161100d565b96505060406112138a828b0161100d565b955050606088013567ffffffffffffffff81111561123457611233610fe0565b5b6112408a828b0161111f565b9450945050608088013567ffffffffffffffff81111561126357611262610fe0565b5b61126f8a828b01611174565b925092505092959891949750929550565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6112a982611280565b9050919050565b6112b98161129f565b81146112c3575f5ffd5b50565b5f813590506112d4816112b0565b92915050565b5f5f5f606084860312156112f1576112f0610fdc565b5b5f6112fe8682870161100d565b935050602061130f8682870161100d565b9250506040611320868287016112c6565b9150509250925092565b5f8115159050919050565b61133e8161132a565b82525050565b5f6020820190506113575f830184611335565b92915050565b5f6020828403121561137257611371610fdc565b5b5f61137f8482850161100d565b91505092915050565b5f5f5f5f5f608086880312156113a1576113a0610fdc565b5b5f6113ae8882890161100d565b95505060206113bf8882890161100d565b94505060406113d08882890161100d565b935050606086013567ffffffffffffffff8111156113f1576113f0610fe0565b5b6113fd8882890161111f565b92509250509295509295909350565b5f5f83601f84011261142157611420611113565b5b8235905067ffffffffffffffff81111561143e5761143d611117565b5b60208301915083602082028301111561145a5761145961111b565b5b9250929050565b5f5f5f5f5f6080868803121561147a57611479610fdc565b5b5f6114878882890161100d565b95505060206114988882890161100d565b94505060406114a98882890161100d565b935050606086013567ffffffffffffffff8111156114ca576114c9610fe0565b5b6114d68882890161140c565b92509250509295509295909350565b5f819050919050565b5f6115086115036114fe84611280565b6114e5565b611280565b9050919050565b5f611519826114ee565b9050919050565b5f61152a8261150f565b9050919050565b61153a81611520565b82525050565b5f6020820190506115535f830184611531565b92915050565b5f6115638261150f565b9050919050565b61157381611559565b82525050565b5f60208201905061158c5f83018461156a565b92915050565b5f819050919050565b6115a481611592565b81146115ae575f5ffd5b50565b5f813590506115bf8161159b565b92915050565b5f5f5f606084860312156115dc576115db610fdc565b5b5f6115e98682870161100d565b93505060206115fa8682870161100d565b925050604061160b868287016115b1565b9150509250925092565b5f5f5f5f5f5f5f60a0888a0312156116305761162f610fdc565b5b5f61163d8a828b0161100d565b975050602061164e8a828b0161100d565b965050604061165f8a828b0161100d565b955050606088013567ffffffffffffffff8111156116805761167f610fe0565b5b61168c8a828b0161140c565b9450945050608088013567ffffffffffffffff8111156116af576116ae610fe0565b5b6116bb8a828b01611174565b925092505092959891949750929550565b5f819050919050565b6116de816116cc565b82525050565b5f6020820190506116f75f8301846116d5565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611734826116cc565b915061173f836116cc565b925082820390508181125f8412168282135f851215161715611764576117636116fd565b5b92915050565b5f611774826116cc565b915061177f836116cc565b925082820261178d816116cc565b91507f800000000000000000000000000000000000000000000000000000000000000084145f841216156117c4576117c36116fd565b5b82820584148315176117d9576117d86116fd565b5b5092915050565b5f82825260208201905092915050565b828183375f83830152505050565b5f601f19601f8301169050919050565b5f61181983856117e0565b93506118268385846117f0565b61182f836117fe565b840190509392505050565b5f82825260208201905092915050565b5f819050919050565b5f82825260208201905092915050565b5f61186e8385611853565b935061187b8385846117f0565b611884836117fe565b840190509392505050565b5f61189b848484611863565b90509392505050565b5f5ffd5b5f5ffd5b5f5ffd5b5f5f833560016020038436030381126118cc576118cb6118ac565b5b83810192508235915060208301925067ffffffffffffffff8211156118f4576118f36118a4565b5b60018202360383131561190a576119096118a8565b5b509250929050565b5f602082019050919050565b5f611929838561183a565b93508360208402850161193b8461184a565b805f5b8781101561198057848403895261195582846118b0565b61196086828461188f565b955061196b84611912565b935060208b019a50505060018101905061193e565b50829750879450505050509392505050565b5f6040820190508181035f8301526119ab81868861180e565b905081810360208301526119c081848661191e565b905095945050505050565b5f81905092915050565b5f6119e083856119cb565b93506119ed8385846117f0565b82840190509392505050565b5f611a058284866119d5565b91508190509392505050565b5f611a1b82610fe4565b9150611a2683610fe4565b9250828201905067ffffffffffffffff811115611a4657611a456116fd565b5b92915050565b5f6020820190508181035f830152611a6581848661180e565b90509392505050565b5f81519050611a7c816112b0565b92915050565b5f60208284031215611a9757611a96610fdc565b5b5f611aa484828501611a6e565b91505092915050565b5f81519050611abb81610ff7565b92915050565b5f60208284031215611ad657611ad5610fdc565b5b5f611ae384828501611aad565b91505092915050565b5f604082019050611aff5f8301866110eb565b8181036020830152611b1281848661180e565b9050949350505050565b5f82825260208201905092915050565b5f819050919050565b611b3e81610fe4565b82525050565b5f611b4f8383611b35565b60208301905092915050565b5f611b69602084018461100d565b905092915050565b5f602082019050919050565b5f611b888385611b1c565b9350611b9382611b2c565b805f5b85811015611bcb57611ba88284611b5b565b611bb28882611b44565b9750611bbd83611b71565b925050600181019050611b96565b5085925050509392505050565b5f6020820190508181035f830152611bf1818486611b7d565b90509392505050565b5f611c0482610fe4565b9150611c0f83610fe4565b9250828202611c1d81610fe4565b9150808214611c2f57611c2e6116fd565b5b5092915050565b5f6040820190508181035f830152611c4f818688611b7d565b90508181036020830152611c6481848661191e565b905095945050505050565b5f611c79826116cc565b9150611c84836116cc565b92508282019050828112155f8312168382125f841215161715611caa57611ca96116fd565b5b9291505056fea164736f6c634300081c000a",
}

// DKGContractABI is the input ABI used to generate the binding from.
// Deprecated: Use DKGContractMetaData.ABI instead.
var DKGContractABI = DKGContractMetaData.ABI

// DKGContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use DKGContractMetaData.Bin instead.
var DKGContractBin = DKGContractMetaData.Bin

// DeployDKGContract deploys a new Ethereum contract, binding an instance of DKGContract to it.
func DeployDKGContract(auth *bind.TransactOpts, backend bind.ContractBackend, phaseLength uint64, dkgLeadLength uint64, keyperSetManagerAddress common.Address, keyBroadcastContractAddress common.Address) (common.Address, *types.Transaction, *DKGContract, error) {
	parsed, err := DKGContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(DKGContractBin), backend, phaseLength, dkgLeadLength, keyperSetManagerAddress, keyBroadcastContractAddress)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &DKGContract{DKGContractCaller: DKGContractCaller{contract: contract}, DKGContractTransactor: DKGContractTransactor{contract: contract}, DKGContractFilterer: DKGContractFilterer{contract: contract}}, nil
}

// DKGContract is an auto generated Go binding around an Ethereum contract.
type DKGContract struct {
	DKGContractCaller     // Read-only binding to the contract
	DKGContractTransactor // Write-only binding to the contract
	DKGContractFilterer   // Log filterer for contract events
}

// DKGContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type DKGContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DKGContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DKGContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DKGContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DKGContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DKGContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DKGContractSession struct {
	Contract     *DKGContract      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DKGContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DKGContractCallerSession struct {
	Contract *DKGContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// DKGContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DKGContractTransactorSession struct {
	Contract     *DKGContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// DKGContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type DKGContractRaw struct {
	Contract *DKGContract // Generic contract binding to access the raw methods on
}

// DKGContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DKGContractCallerRaw struct {
	Contract *DKGContractCaller // Generic read-only contract binding to access the raw methods on
}

// DKGContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DKGContractTransactorRaw struct {
	Contract *DKGContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDKGContract creates a new instance of DKGContract, bound to a specific deployed contract.
func NewDKGContract(address common.Address, backend bind.ContractBackend) (*DKGContract, error) {
	contract, err := bindDKGContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DKGContract{DKGContractCaller: DKGContractCaller{contract: contract}, DKGContractTransactor: DKGContractTransactor{contract: contract}, DKGContractFilterer: DKGContractFilterer{contract: contract}}, nil
}

// NewDKGContractCaller creates a new read-only instance of DKGContract, bound to a specific deployed contract.
func NewDKGContractCaller(address common.Address, caller bind.ContractCaller) (*DKGContractCaller, error) {
	contract, err := bindDKGContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DKGContractCaller{contract: contract}, nil
}

// NewDKGContractTransactor creates a new write-only instance of DKGContract, bound to a specific deployed contract.
func NewDKGContractTransactor(address common.Address, transactor bind.ContractTransactor) (*DKGContractTransactor, error) {
	contract, err := bindDKGContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DKGContractTransactor{contract: contract}, nil
}

// NewDKGContractFilterer creates a new log filterer instance of DKGContract, bound to a specific deployed contract.
func NewDKGContractFilterer(address common.Address, filterer bind.ContractFilterer) (*DKGContractFilterer, error) {
	contract, err := bindDKGContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DKGContractFilterer{contract: contract}, nil
}

// bindDKGContract binds a generic wrapper to an already deployed contract.
func bindDKGContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := DKGContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DKGContract *DKGContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DKGContract.Contract.DKGContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DKGContract *DKGContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DKGContract.Contract.DKGContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DKGContract *DKGContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DKGContract.Contract.DKGContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DKGContract *DKGContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DKGContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DKGContract *DKGContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DKGContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DKGContract *DKGContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DKGContract.Contract.contract.Transact(opts, method, params...)
}

// DKGLEADLENGTH is a free data retrieval call binding the contract method 0x16b89c3f.
//
// Solidity: function DKG_LEAD_LENGTH() view returns(uint64)
func (_DKGContract *DKGContractCaller) DKGLEADLENGTH(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _DKGContract.contract.Call(opts, &out, "DKG_LEAD_LENGTH")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// DKGLEADLENGTH is a free data retrieval call binding the contract method 0x16b89c3f.
//
// Solidity: function DKG_LEAD_LENGTH() view returns(uint64)
func (_DKGContract *DKGContractSession) DKGLEADLENGTH() (uint64, error) {
	return _DKGContract.Contract.DKGLEADLENGTH(&_DKGContract.CallOpts)
}

// DKGLEADLENGTH is a free data retrieval call binding the contract method 0x16b89c3f.
//
// Solidity: function DKG_LEAD_LENGTH() view returns(uint64)
func (_DKGContract *DKGContractCallerSession) DKGLEADLENGTH() (uint64, error) {
	return _DKGContract.Contract.DKGLEADLENGTH(&_DKGContract.CallOpts)
}

// PHASELENGTH is a free data retrieval call binding the contract method 0xb209e717.
//
// Solidity: function PHASE_LENGTH() view returns(uint64)
func (_DKGContract *DKGContractCaller) PHASELENGTH(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _DKGContract.contract.Call(opts, &out, "PHASE_LENGTH")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// PHASELENGTH is a free data retrieval call binding the contract method 0xb209e717.
//
// Solidity: function PHASE_LENGTH() view returns(uint64)
func (_DKGContract *DKGContractSession) PHASELENGTH() (uint64, error) {
	return _DKGContract.Contract.PHASELENGTH(&_DKGContract.CallOpts)
}

// PHASELENGTH is a free data retrieval call binding the contract method 0xb209e717.
//
// Solidity: function PHASE_LENGTH() view returns(uint64)
func (_DKGContract *DKGContractCallerSession) PHASELENGTH() (uint64, error) {
	return _DKGContract.Contract.PHASELENGTH(&_DKGContract.CallOpts)
}

// CurrentPhase is a free data retrieval call binding the contract method 0x0a3ec4d6.
//
// Solidity: function currentPhase(uint64 keyperSetIndex, uint64 retryCounter) view returns(uint8)
func (_DKGContract *DKGContractCaller) CurrentPhase(opts *bind.CallOpts, keyperSetIndex uint64, retryCounter uint64) (uint8, error) {
	var out []interface{}
	err := _DKGContract.contract.Call(opts, &out, "currentPhase", keyperSetIndex, retryCounter)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// CurrentPhase is a free data retrieval call binding the contract method 0x0a3ec4d6.
//
// Solidity: function currentPhase(uint64 keyperSetIndex, uint64 retryCounter) view returns(uint8)
func (_DKGContract *DKGContractSession) CurrentPhase(keyperSetIndex uint64, retryCounter uint64) (uint8, error) {
	return _DKGContract.Contract.CurrentPhase(&_DKGContract.CallOpts, keyperSetIndex, retryCounter)
}

// CurrentPhase is a free data retrieval call binding the contract method 0x0a3ec4d6.
//
// Solidity: function currentPhase(uint64 keyperSetIndex, uint64 retryCounter) view returns(uint8)
func (_DKGContract *DKGContractCallerSession) CurrentPhase(keyperSetIndex uint64, retryCounter uint64) (uint8, error) {
	return _DKGContract.Contract.CurrentPhase(&_DKGContract.CallOpts, keyperSetIndex, retryCounter)
}

// CycleLength is a free data retrieval call binding the contract method 0xeac471a0.
//
// Solidity: function cycleLength() view returns(uint64)
func (_DKGContract *DKGContractCaller) CycleLength(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _DKGContract.contract.Call(opts, &out, "cycleLength")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// CycleLength is a free data retrieval call binding the contract method 0xeac471a0.
//
// Solidity: function cycleLength() view returns(uint64)
func (_DKGContract *DKGContractSession) CycleLength() (uint64, error) {
	return _DKGContract.Contract.CycleLength(&_DKGContract.CallOpts)
}

// CycleLength is a free data retrieval call binding the contract method 0xeac471a0.
//
// Solidity: function cycleLength() view returns(uint64)
func (_DKGContract *DKGContractCallerSession) CycleLength() (uint64, error) {
	return _DKGContract.Contract.CycleLength(&_DKGContract.CallOpts)
}

// DkgStart is a free data retrieval call binding the contract method 0xf48f3d14.
//
// Solidity: function dkgStart(uint64 keyperSetIndex, uint64 retryCounter) view returns(int256)
func (_DKGContract *DKGContractCaller) DkgStart(opts *bind.CallOpts, keyperSetIndex uint64, retryCounter uint64) (*big.Int, error) {
	var out []interface{}
	err := _DKGContract.contract.Call(opts, &out, "dkgStart", keyperSetIndex, retryCounter)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DkgStart is a free data retrieval call binding the contract method 0xf48f3d14.
//
// Solidity: function dkgStart(uint64 keyperSetIndex, uint64 retryCounter) view returns(int256)
func (_DKGContract *DKGContractSession) DkgStart(keyperSetIndex uint64, retryCounter uint64) (*big.Int, error) {
	return _DKGContract.Contract.DkgStart(&_DKGContract.CallOpts, keyperSetIndex, retryCounter)
}

// DkgStart is a free data retrieval call binding the contract method 0xf48f3d14.
//
// Solidity: function dkgStart(uint64 keyperSetIndex, uint64 retryCounter) view returns(int256)
func (_DKGContract *DKGContractCallerSession) DkgStart(keyperSetIndex uint64, retryCounter uint64) (*big.Int, error) {
	return _DKGContract.Contract.DkgStart(&_DKGContract.CallOpts, keyperSetIndex, retryCounter)
}

// HasVoted is a free data retrieval call binding the contract method 0x7642d0de.
//
// Solidity: function hasVoted(uint64 , uint64 , address ) view returns(bool)
func (_DKGContract *DKGContractCaller) HasVoted(opts *bind.CallOpts, arg0 uint64, arg1 uint64, arg2 common.Address) (bool, error) {
	var out []interface{}
	err := _DKGContract.contract.Call(opts, &out, "hasVoted", arg0, arg1, arg2)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasVoted is a free data retrieval call binding the contract method 0x7642d0de.
//
// Solidity: function hasVoted(uint64 , uint64 , address ) view returns(bool)
func (_DKGContract *DKGContractSession) HasVoted(arg0 uint64, arg1 uint64, arg2 common.Address) (bool, error) {
	return _DKGContract.Contract.HasVoted(&_DKGContract.CallOpts, arg0, arg1, arg2)
}

// HasVoted is a free data retrieval call binding the contract method 0x7642d0de.
//
// Solidity: function hasVoted(uint64 , uint64 , address ) view returns(bool)
func (_DKGContract *DKGContractCallerSession) HasVoted(arg0 uint64, arg1 uint64, arg2 common.Address) (bool, error) {
	return _DKGContract.Contract.HasVoted(&_DKGContract.CallOpts, arg0, arg1, arg2)
}

// KeyBroadcastContract is a free data retrieval call binding the contract method 0xdf640631.
//
// Solidity: function keyBroadcastContract() view returns(address)
func (_DKGContract *DKGContractCaller) KeyBroadcastContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DKGContract.contract.Call(opts, &out, "keyBroadcastContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// KeyBroadcastContract is a free data retrieval call binding the contract method 0xdf640631.
//
// Solidity: function keyBroadcastContract() view returns(address)
func (_DKGContract *DKGContractSession) KeyBroadcastContract() (common.Address, error) {
	return _DKGContract.Contract.KeyBroadcastContract(&_DKGContract.CallOpts)
}

// KeyBroadcastContract is a free data retrieval call binding the contract method 0xdf640631.
//
// Solidity: function keyBroadcastContract() view returns(address)
func (_DKGContract *DKGContractCallerSession) KeyBroadcastContract() (common.Address, error) {
	return _DKGContract.Contract.KeyBroadcastContract(&_DKGContract.CallOpts)
}

// KeyperSetManager is a free data retrieval call binding the contract method 0xd4b2cedd.
//
// Solidity: function keyperSetManager() view returns(address)
func (_DKGContract *DKGContractCaller) KeyperSetManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DKGContract.contract.Call(opts, &out, "keyperSetManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// KeyperSetManager is a free data retrieval call binding the contract method 0xd4b2cedd.
//
// Solidity: function keyperSetManager() view returns(address)
func (_DKGContract *DKGContractSession) KeyperSetManager() (common.Address, error) {
	return _DKGContract.Contract.KeyperSetManager(&_DKGContract.CallOpts)
}

// KeyperSetManager is a free data retrieval call binding the contract method 0xd4b2cedd.
//
// Solidity: function keyperSetManager() view returns(address)
func (_DKGContract *DKGContractCallerSession) KeyperSetManager() (common.Address, error) {
	return _DKGContract.Contract.KeyperSetManager(&_DKGContract.CallOpts)
}

// Succeeded is a free data retrieval call binding the contract method 0x78a1b9db.
//
// Solidity: function succeeded(uint64 ) view returns(bool)
func (_DKGContract *DKGContractCaller) Succeeded(opts *bind.CallOpts, arg0 uint64) (bool, error) {
	var out []interface{}
	err := _DKGContract.contract.Call(opts, &out, "succeeded", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Succeeded is a free data retrieval call binding the contract method 0x78a1b9db.
//
// Solidity: function succeeded(uint64 ) view returns(bool)
func (_DKGContract *DKGContractSession) Succeeded(arg0 uint64) (bool, error) {
	return _DKGContract.Contract.Succeeded(&_DKGContract.CallOpts, arg0)
}

// Succeeded is a free data retrieval call binding the contract method 0x78a1b9db.
//
// Solidity: function succeeded(uint64 ) view returns(bool)
func (_DKGContract *DKGContractCallerSession) Succeeded(arg0 uint64) (bool, error) {
	return _DKGContract.Contract.Succeeded(&_DKGContract.CallOpts, arg0)
}

// VoteCount is a free data retrieval call binding the contract method 0xeb05c280.
//
// Solidity: function voteCount(uint64 , uint64 , bytes32 ) view returns(uint64)
func (_DKGContract *DKGContractCaller) VoteCount(opts *bind.CallOpts, arg0 uint64, arg1 uint64, arg2 [32]byte) (uint64, error) {
	var out []interface{}
	err := _DKGContract.contract.Call(opts, &out, "voteCount", arg0, arg1, arg2)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// VoteCount is a free data retrieval call binding the contract method 0xeb05c280.
//
// Solidity: function voteCount(uint64 , uint64 , bytes32 ) view returns(uint64)
func (_DKGContract *DKGContractSession) VoteCount(arg0 uint64, arg1 uint64, arg2 [32]byte) (uint64, error) {
	return _DKGContract.Contract.VoteCount(&_DKGContract.CallOpts, arg0, arg1, arg2)
}

// VoteCount is a free data retrieval call binding the contract method 0xeb05c280.
//
// Solidity: function voteCount(uint64 , uint64 , bytes32 ) view returns(uint64)
func (_DKGContract *DKGContractCallerSession) VoteCount(arg0 uint64, arg1 uint64, arg2 [32]byte) (uint64, error) {
	return _DKGContract.Contract.VoteCount(&_DKGContract.CallOpts, arg0, arg1, arg2)
}

// SubmitAccusation is a paid mutator transaction binding the contract method 0x94afc12d.
//
// Solidity: function submitAccusation(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, uint64[] accusedIndices) returns()
func (_DKGContract *DKGContractTransactor) SubmitAccusation(opts *bind.TransactOpts, keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, accusedIndices []uint64) (*types.Transaction, error) {
	return _DKGContract.contract.Transact(opts, "submitAccusation", keyperSetIndex, retryCounter, keyperIndex, accusedIndices)
}

// SubmitAccusation is a paid mutator transaction binding the contract method 0x94afc12d.
//
// Solidity: function submitAccusation(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, uint64[] accusedIndices) returns()
func (_DKGContract *DKGContractSession) SubmitAccusation(keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, accusedIndices []uint64) (*types.Transaction, error) {
	return _DKGContract.Contract.SubmitAccusation(&_DKGContract.TransactOpts, keyperSetIndex, retryCounter, keyperIndex, accusedIndices)
}

// SubmitAccusation is a paid mutator transaction binding the contract method 0x94afc12d.
//
// Solidity: function submitAccusation(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, uint64[] accusedIndices) returns()
func (_DKGContract *DKGContractTransactorSession) SubmitAccusation(keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, accusedIndices []uint64) (*types.Transaction, error) {
	return _DKGContract.Contract.SubmitAccusation(&_DKGContract.TransactOpts, keyperSetIndex, retryCounter, keyperIndex, accusedIndices)
}

// SubmitApology is a paid mutator transaction binding the contract method 0xf01275b0.
//
// Solidity: function submitApology(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, uint64[] accuserIndices, bytes[] polyEvalData) returns()
func (_DKGContract *DKGContractTransactor) SubmitApology(opts *bind.TransactOpts, keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, accuserIndices []uint64, polyEvalData [][]byte) (*types.Transaction, error) {
	return _DKGContract.contract.Transact(opts, "submitApology", keyperSetIndex, retryCounter, keyperIndex, accuserIndices, polyEvalData)
}

// SubmitApology is a paid mutator transaction binding the contract method 0xf01275b0.
//
// Solidity: function submitApology(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, uint64[] accuserIndices, bytes[] polyEvalData) returns()
func (_DKGContract *DKGContractSession) SubmitApology(keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, accuserIndices []uint64, polyEvalData [][]byte) (*types.Transaction, error) {
	return _DKGContract.Contract.SubmitApology(&_DKGContract.TransactOpts, keyperSetIndex, retryCounter, keyperIndex, accuserIndices, polyEvalData)
}

// SubmitApology is a paid mutator transaction binding the contract method 0xf01275b0.
//
// Solidity: function submitApology(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, uint64[] accuserIndices, bytes[] polyEvalData) returns()
func (_DKGContract *DKGContractTransactorSession) SubmitApology(keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, accuserIndices []uint64, polyEvalData [][]byte) (*types.Transaction, error) {
	return _DKGContract.Contract.SubmitApology(&_DKGContract.TransactOpts, keyperSetIndex, retryCounter, keyperIndex, accuserIndices, polyEvalData)
}

// SubmitDealing is a paid mutator transaction binding the contract method 0x3f980098.
//
// Solidity: function submitDealing(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, bytes commitment, bytes[] polyEvals) returns()
func (_DKGContract *DKGContractTransactor) SubmitDealing(opts *bind.TransactOpts, keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, commitment []byte, polyEvals [][]byte) (*types.Transaction, error) {
	return _DKGContract.contract.Transact(opts, "submitDealing", keyperSetIndex, retryCounter, keyperIndex, commitment, polyEvals)
}

// SubmitDealing is a paid mutator transaction binding the contract method 0x3f980098.
//
// Solidity: function submitDealing(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, bytes commitment, bytes[] polyEvals) returns()
func (_DKGContract *DKGContractSession) SubmitDealing(keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, commitment []byte, polyEvals [][]byte) (*types.Transaction, error) {
	return _DKGContract.Contract.SubmitDealing(&_DKGContract.TransactOpts, keyperSetIndex, retryCounter, keyperIndex, commitment, polyEvals)
}

// SubmitDealing is a paid mutator transaction binding the contract method 0x3f980098.
//
// Solidity: function submitDealing(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, bytes commitment, bytes[] polyEvals) returns()
func (_DKGContract *DKGContractTransactorSession) SubmitDealing(keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, commitment []byte, polyEvals [][]byte) (*types.Transaction, error) {
	return _DKGContract.Contract.SubmitDealing(&_DKGContract.TransactOpts, keyperSetIndex, retryCounter, keyperIndex, commitment, polyEvals)
}

// SubmitSuccessVote is a paid mutator transaction binding the contract method 0x82b870d3.
//
// Solidity: function submitSuccessVote(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, bytes eonPublicKey) returns()
func (_DKGContract *DKGContractTransactor) SubmitSuccessVote(opts *bind.TransactOpts, keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, eonPublicKey []byte) (*types.Transaction, error) {
	return _DKGContract.contract.Transact(opts, "submitSuccessVote", keyperSetIndex, retryCounter, keyperIndex, eonPublicKey)
}

// SubmitSuccessVote is a paid mutator transaction binding the contract method 0x82b870d3.
//
// Solidity: function submitSuccessVote(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, bytes eonPublicKey) returns()
func (_DKGContract *DKGContractSession) SubmitSuccessVote(keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, eonPublicKey []byte) (*types.Transaction, error) {
	return _DKGContract.Contract.SubmitSuccessVote(&_DKGContract.TransactOpts, keyperSetIndex, retryCounter, keyperIndex, eonPublicKey)
}

// SubmitSuccessVote is a paid mutator transaction binding the contract method 0x82b870d3.
//
// Solidity: function submitSuccessVote(uint64 keyperSetIndex, uint64 retryCounter, uint64 keyperIndex, bytes eonPublicKey) returns()
func (_DKGContract *DKGContractTransactorSession) SubmitSuccessVote(keyperSetIndex uint64, retryCounter uint64, keyperIndex uint64, eonPublicKey []byte) (*types.Transaction, error) {
	return _DKGContract.Contract.SubmitSuccessVote(&_DKGContract.TransactOpts, keyperSetIndex, retryCounter, keyperIndex, eonPublicKey)
}

// DKGContractAccusationSubmittedIterator is returned from FilterAccusationSubmitted and is used to iterate over the raw logs and unpacked data for AccusationSubmitted events raised by the DKGContract contract.
type DKGContractAccusationSubmittedIterator struct {
	Event *DKGContractAccusationSubmitted // Event containing the contract specifics and raw log

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
func (it *DKGContractAccusationSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DKGContractAccusationSubmitted)
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
		it.Event = new(DKGContractAccusationSubmitted)
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
func (it *DKGContractAccusationSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DKGContractAccusationSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DKGContractAccusationSubmitted represents a AccusationSubmitted event raised by the DKGContract contract.
type DKGContractAccusationSubmitted struct {
	KeyperSetIndex uint64
	RetryCounter   uint64
	KeyperIndex    uint64
	AccusedIndices []uint64
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterAccusationSubmitted is a free log retrieval operation binding the contract event 0xb5cc05e740a503c13e1b7ac3ad2d5abf41b91d2e5fca79455d34e515fb70849a.
//
// Solidity: event AccusationSubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, uint64[] accusedIndices)
func (_DKGContract *DKGContractFilterer) FilterAccusationSubmitted(opts *bind.FilterOpts, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (*DKGContractAccusationSubmittedIterator, error) {

	var keyperSetIndexRule []interface{}
	for _, keyperSetIndexItem := range keyperSetIndex {
		keyperSetIndexRule = append(keyperSetIndexRule, keyperSetIndexItem)
	}
	var retryCounterRule []interface{}
	for _, retryCounterItem := range retryCounter {
		retryCounterRule = append(retryCounterRule, retryCounterItem)
	}
	var keyperIndexRule []interface{}
	for _, keyperIndexItem := range keyperIndex {
		keyperIndexRule = append(keyperIndexRule, keyperIndexItem)
	}

	logs, sub, err := _DKGContract.contract.FilterLogs(opts, "AccusationSubmitted", keyperSetIndexRule, retryCounterRule, keyperIndexRule)
	if err != nil {
		return nil, err
	}
	return &DKGContractAccusationSubmittedIterator{contract: _DKGContract.contract, event: "AccusationSubmitted", logs: logs, sub: sub}, nil
}

// WatchAccusationSubmitted is a free log subscription operation binding the contract event 0xb5cc05e740a503c13e1b7ac3ad2d5abf41b91d2e5fca79455d34e515fb70849a.
//
// Solidity: event AccusationSubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, uint64[] accusedIndices)
func (_DKGContract *DKGContractFilterer) WatchAccusationSubmitted(opts *bind.WatchOpts, sink chan<- *DKGContractAccusationSubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (event.Subscription, error) {

	var keyperSetIndexRule []interface{}
	for _, keyperSetIndexItem := range keyperSetIndex {
		keyperSetIndexRule = append(keyperSetIndexRule, keyperSetIndexItem)
	}
	var retryCounterRule []interface{}
	for _, retryCounterItem := range retryCounter {
		retryCounterRule = append(retryCounterRule, retryCounterItem)
	}
	var keyperIndexRule []interface{}
	for _, keyperIndexItem := range keyperIndex {
		keyperIndexRule = append(keyperIndexRule, keyperIndexItem)
	}

	logs, sub, err := _DKGContract.contract.WatchLogs(opts, "AccusationSubmitted", keyperSetIndexRule, retryCounterRule, keyperIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DKGContractAccusationSubmitted)
				if err := _DKGContract.contract.UnpackLog(event, "AccusationSubmitted", log); err != nil {
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

// ParseAccusationSubmitted is a log parse operation binding the contract event 0xb5cc05e740a503c13e1b7ac3ad2d5abf41b91d2e5fca79455d34e515fb70849a.
//
// Solidity: event AccusationSubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, uint64[] accusedIndices)
func (_DKGContract *DKGContractFilterer) ParseAccusationSubmitted(log types.Log) (*DKGContractAccusationSubmitted, error) {
	event := new(DKGContractAccusationSubmitted)
	if err := _DKGContract.contract.UnpackLog(event, "AccusationSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DKGContractApologySubmittedIterator is returned from FilterApologySubmitted and is used to iterate over the raw logs and unpacked data for ApologySubmitted events raised by the DKGContract contract.
type DKGContractApologySubmittedIterator struct {
	Event *DKGContractApologySubmitted // Event containing the contract specifics and raw log

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
func (it *DKGContractApologySubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DKGContractApologySubmitted)
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
		it.Event = new(DKGContractApologySubmitted)
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
func (it *DKGContractApologySubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DKGContractApologySubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DKGContractApologySubmitted represents a ApologySubmitted event raised by the DKGContract contract.
type DKGContractApologySubmitted struct {
	KeyperSetIndex uint64
	RetryCounter   uint64
	KeyperIndex    uint64
	AccuserIndices []uint64
	PolyEvalData   [][]byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterApologySubmitted is a free log retrieval operation binding the contract event 0x62729f0820cf3ab60fbd63a51cdeae0554d0ab24f843b86bd0d2591594d69136.
//
// Solidity: event ApologySubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, uint64[] accuserIndices, bytes[] polyEvalData)
func (_DKGContract *DKGContractFilterer) FilterApologySubmitted(opts *bind.FilterOpts, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (*DKGContractApologySubmittedIterator, error) {

	var keyperSetIndexRule []interface{}
	for _, keyperSetIndexItem := range keyperSetIndex {
		keyperSetIndexRule = append(keyperSetIndexRule, keyperSetIndexItem)
	}
	var retryCounterRule []interface{}
	for _, retryCounterItem := range retryCounter {
		retryCounterRule = append(retryCounterRule, retryCounterItem)
	}
	var keyperIndexRule []interface{}
	for _, keyperIndexItem := range keyperIndex {
		keyperIndexRule = append(keyperIndexRule, keyperIndexItem)
	}

	logs, sub, err := _DKGContract.contract.FilterLogs(opts, "ApologySubmitted", keyperSetIndexRule, retryCounterRule, keyperIndexRule)
	if err != nil {
		return nil, err
	}
	return &DKGContractApologySubmittedIterator{contract: _DKGContract.contract, event: "ApologySubmitted", logs: logs, sub: sub}, nil
}

// WatchApologySubmitted is a free log subscription operation binding the contract event 0x62729f0820cf3ab60fbd63a51cdeae0554d0ab24f843b86bd0d2591594d69136.
//
// Solidity: event ApologySubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, uint64[] accuserIndices, bytes[] polyEvalData)
func (_DKGContract *DKGContractFilterer) WatchApologySubmitted(opts *bind.WatchOpts, sink chan<- *DKGContractApologySubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (event.Subscription, error) {

	var keyperSetIndexRule []interface{}
	for _, keyperSetIndexItem := range keyperSetIndex {
		keyperSetIndexRule = append(keyperSetIndexRule, keyperSetIndexItem)
	}
	var retryCounterRule []interface{}
	for _, retryCounterItem := range retryCounter {
		retryCounterRule = append(retryCounterRule, retryCounterItem)
	}
	var keyperIndexRule []interface{}
	for _, keyperIndexItem := range keyperIndex {
		keyperIndexRule = append(keyperIndexRule, keyperIndexItem)
	}

	logs, sub, err := _DKGContract.contract.WatchLogs(opts, "ApologySubmitted", keyperSetIndexRule, retryCounterRule, keyperIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DKGContractApologySubmitted)
				if err := _DKGContract.contract.UnpackLog(event, "ApologySubmitted", log); err != nil {
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

// ParseApologySubmitted is a log parse operation binding the contract event 0x62729f0820cf3ab60fbd63a51cdeae0554d0ab24f843b86bd0d2591594d69136.
//
// Solidity: event ApologySubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, uint64[] accuserIndices, bytes[] polyEvalData)
func (_DKGContract *DKGContractFilterer) ParseApologySubmitted(log types.Log) (*DKGContractApologySubmitted, error) {
	event := new(DKGContractApologySubmitted)
	if err := _DKGContract.contract.UnpackLog(event, "ApologySubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DKGContractDKGSucceededIterator is returned from FilterDKGSucceeded and is used to iterate over the raw logs and unpacked data for DKGSucceeded events raised by the DKGContract contract.
type DKGContractDKGSucceededIterator struct {
	Event *DKGContractDKGSucceeded // Event containing the contract specifics and raw log

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
func (it *DKGContractDKGSucceededIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DKGContractDKGSucceeded)
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
		it.Event = new(DKGContractDKGSucceeded)
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
func (it *DKGContractDKGSucceededIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DKGContractDKGSucceededIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DKGContractDKGSucceeded represents a DKGSucceeded event raised by the DKGContract contract.
type DKGContractDKGSucceeded struct {
	KeyperSetIndex uint64
	RetryCounter   uint64
	EonPublicKey   []byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterDKGSucceeded is a free log retrieval operation binding the contract event 0xe6c9bd3daa50b9218615605cdf333410d405b7f3b8dbeee5a2373769679aa9f5.
//
// Solidity: event DKGSucceeded(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, bytes eonPublicKey)
func (_DKGContract *DKGContractFilterer) FilterDKGSucceeded(opts *bind.FilterOpts, keyperSetIndex []uint64, retryCounter []uint64) (*DKGContractDKGSucceededIterator, error) {

	var keyperSetIndexRule []interface{}
	for _, keyperSetIndexItem := range keyperSetIndex {
		keyperSetIndexRule = append(keyperSetIndexRule, keyperSetIndexItem)
	}
	var retryCounterRule []interface{}
	for _, retryCounterItem := range retryCounter {
		retryCounterRule = append(retryCounterRule, retryCounterItem)
	}

	logs, sub, err := _DKGContract.contract.FilterLogs(opts, "DKGSucceeded", keyperSetIndexRule, retryCounterRule)
	if err != nil {
		return nil, err
	}
	return &DKGContractDKGSucceededIterator{contract: _DKGContract.contract, event: "DKGSucceeded", logs: logs, sub: sub}, nil
}

// WatchDKGSucceeded is a free log subscription operation binding the contract event 0xe6c9bd3daa50b9218615605cdf333410d405b7f3b8dbeee5a2373769679aa9f5.
//
// Solidity: event DKGSucceeded(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, bytes eonPublicKey)
func (_DKGContract *DKGContractFilterer) WatchDKGSucceeded(opts *bind.WatchOpts, sink chan<- *DKGContractDKGSucceeded, keyperSetIndex []uint64, retryCounter []uint64) (event.Subscription, error) {

	var keyperSetIndexRule []interface{}
	for _, keyperSetIndexItem := range keyperSetIndex {
		keyperSetIndexRule = append(keyperSetIndexRule, keyperSetIndexItem)
	}
	var retryCounterRule []interface{}
	for _, retryCounterItem := range retryCounter {
		retryCounterRule = append(retryCounterRule, retryCounterItem)
	}

	logs, sub, err := _DKGContract.contract.WatchLogs(opts, "DKGSucceeded", keyperSetIndexRule, retryCounterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DKGContractDKGSucceeded)
				if err := _DKGContract.contract.UnpackLog(event, "DKGSucceeded", log); err != nil {
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

// ParseDKGSucceeded is a log parse operation binding the contract event 0xe6c9bd3daa50b9218615605cdf333410d405b7f3b8dbeee5a2373769679aa9f5.
//
// Solidity: event DKGSucceeded(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, bytes eonPublicKey)
func (_DKGContract *DKGContractFilterer) ParseDKGSucceeded(log types.Log) (*DKGContractDKGSucceeded, error) {
	event := new(DKGContractDKGSucceeded)
	if err := _DKGContract.contract.UnpackLog(event, "DKGSucceeded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DKGContractDealingSubmittedIterator is returned from FilterDealingSubmitted and is used to iterate over the raw logs and unpacked data for DealingSubmitted events raised by the DKGContract contract.
type DKGContractDealingSubmittedIterator struct {
	Event *DKGContractDealingSubmitted // Event containing the contract specifics and raw log

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
func (it *DKGContractDealingSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DKGContractDealingSubmitted)
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
		it.Event = new(DKGContractDealingSubmitted)
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
func (it *DKGContractDealingSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DKGContractDealingSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DKGContractDealingSubmitted represents a DealingSubmitted event raised by the DKGContract contract.
type DKGContractDealingSubmitted struct {
	KeyperSetIndex uint64
	RetryCounter   uint64
	KeyperIndex    uint64
	Commitment     []byte
	PolyEvals      [][]byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterDealingSubmitted is a free log retrieval operation binding the contract event 0xa5074728b92791250d48ccdc55266aca7b1cca22f89e1ef2ab91ccd7ae0f337d.
//
// Solidity: event DealingSubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, bytes commitment, bytes[] polyEvals)
func (_DKGContract *DKGContractFilterer) FilterDealingSubmitted(opts *bind.FilterOpts, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (*DKGContractDealingSubmittedIterator, error) {

	var keyperSetIndexRule []interface{}
	for _, keyperSetIndexItem := range keyperSetIndex {
		keyperSetIndexRule = append(keyperSetIndexRule, keyperSetIndexItem)
	}
	var retryCounterRule []interface{}
	for _, retryCounterItem := range retryCounter {
		retryCounterRule = append(retryCounterRule, retryCounterItem)
	}
	var keyperIndexRule []interface{}
	for _, keyperIndexItem := range keyperIndex {
		keyperIndexRule = append(keyperIndexRule, keyperIndexItem)
	}

	logs, sub, err := _DKGContract.contract.FilterLogs(opts, "DealingSubmitted", keyperSetIndexRule, retryCounterRule, keyperIndexRule)
	if err != nil {
		return nil, err
	}
	return &DKGContractDealingSubmittedIterator{contract: _DKGContract.contract, event: "DealingSubmitted", logs: logs, sub: sub}, nil
}

// WatchDealingSubmitted is a free log subscription operation binding the contract event 0xa5074728b92791250d48ccdc55266aca7b1cca22f89e1ef2ab91ccd7ae0f337d.
//
// Solidity: event DealingSubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, bytes commitment, bytes[] polyEvals)
func (_DKGContract *DKGContractFilterer) WatchDealingSubmitted(opts *bind.WatchOpts, sink chan<- *DKGContractDealingSubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (event.Subscription, error) {

	var keyperSetIndexRule []interface{}
	for _, keyperSetIndexItem := range keyperSetIndex {
		keyperSetIndexRule = append(keyperSetIndexRule, keyperSetIndexItem)
	}
	var retryCounterRule []interface{}
	for _, retryCounterItem := range retryCounter {
		retryCounterRule = append(retryCounterRule, retryCounterItem)
	}
	var keyperIndexRule []interface{}
	for _, keyperIndexItem := range keyperIndex {
		keyperIndexRule = append(keyperIndexRule, keyperIndexItem)
	}

	logs, sub, err := _DKGContract.contract.WatchLogs(opts, "DealingSubmitted", keyperSetIndexRule, retryCounterRule, keyperIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DKGContractDealingSubmitted)
				if err := _DKGContract.contract.UnpackLog(event, "DealingSubmitted", log); err != nil {
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

// ParseDealingSubmitted is a log parse operation binding the contract event 0xa5074728b92791250d48ccdc55266aca7b1cca22f89e1ef2ab91ccd7ae0f337d.
//
// Solidity: event DealingSubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, bytes commitment, bytes[] polyEvals)
func (_DKGContract *DKGContractFilterer) ParseDealingSubmitted(log types.Log) (*DKGContractDealingSubmitted, error) {
	event := new(DKGContractDealingSubmitted)
	if err := _DKGContract.contract.UnpackLog(event, "DealingSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DKGContractSuccessVoteSubmittedIterator is returned from FilterSuccessVoteSubmitted and is used to iterate over the raw logs and unpacked data for SuccessVoteSubmitted events raised by the DKGContract contract.
type DKGContractSuccessVoteSubmittedIterator struct {
	Event *DKGContractSuccessVoteSubmitted // Event containing the contract specifics and raw log

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
func (it *DKGContractSuccessVoteSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DKGContractSuccessVoteSubmitted)
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
		it.Event = new(DKGContractSuccessVoteSubmitted)
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
func (it *DKGContractSuccessVoteSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DKGContractSuccessVoteSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DKGContractSuccessVoteSubmitted represents a SuccessVoteSubmitted event raised by the DKGContract contract.
type DKGContractSuccessVoteSubmitted struct {
	KeyperSetIndex uint64
	RetryCounter   uint64
	KeyperIndex    uint64
	EonPublicKey   []byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSuccessVoteSubmitted is a free log retrieval operation binding the contract event 0xacfc0aa3f6ba36ac0ffda53e00d2cdace2fd4a9e41e2fe3214032794332c37c3.
//
// Solidity: event SuccessVoteSubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, bytes eonPublicKey)
func (_DKGContract *DKGContractFilterer) FilterSuccessVoteSubmitted(opts *bind.FilterOpts, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (*DKGContractSuccessVoteSubmittedIterator, error) {

	var keyperSetIndexRule []interface{}
	for _, keyperSetIndexItem := range keyperSetIndex {
		keyperSetIndexRule = append(keyperSetIndexRule, keyperSetIndexItem)
	}
	var retryCounterRule []interface{}
	for _, retryCounterItem := range retryCounter {
		retryCounterRule = append(retryCounterRule, retryCounterItem)
	}
	var keyperIndexRule []interface{}
	for _, keyperIndexItem := range keyperIndex {
		keyperIndexRule = append(keyperIndexRule, keyperIndexItem)
	}

	logs, sub, err := _DKGContract.contract.FilterLogs(opts, "SuccessVoteSubmitted", keyperSetIndexRule, retryCounterRule, keyperIndexRule)
	if err != nil {
		return nil, err
	}
	return &DKGContractSuccessVoteSubmittedIterator{contract: _DKGContract.contract, event: "SuccessVoteSubmitted", logs: logs, sub: sub}, nil
}

// WatchSuccessVoteSubmitted is a free log subscription operation binding the contract event 0xacfc0aa3f6ba36ac0ffda53e00d2cdace2fd4a9e41e2fe3214032794332c37c3.
//
// Solidity: event SuccessVoteSubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, bytes eonPublicKey)
func (_DKGContract *DKGContractFilterer) WatchSuccessVoteSubmitted(opts *bind.WatchOpts, sink chan<- *DKGContractSuccessVoteSubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (event.Subscription, error) {

	var keyperSetIndexRule []interface{}
	for _, keyperSetIndexItem := range keyperSetIndex {
		keyperSetIndexRule = append(keyperSetIndexRule, keyperSetIndexItem)
	}
	var retryCounterRule []interface{}
	for _, retryCounterItem := range retryCounter {
		retryCounterRule = append(retryCounterRule, retryCounterItem)
	}
	var keyperIndexRule []interface{}
	for _, keyperIndexItem := range keyperIndex {
		keyperIndexRule = append(keyperIndexRule, keyperIndexItem)
	}

	logs, sub, err := _DKGContract.contract.WatchLogs(opts, "SuccessVoteSubmitted", keyperSetIndexRule, retryCounterRule, keyperIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DKGContractSuccessVoteSubmitted)
				if err := _DKGContract.contract.UnpackLog(event, "SuccessVoteSubmitted", log); err != nil {
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

// ParseSuccessVoteSubmitted is a log parse operation binding the contract event 0xacfc0aa3f6ba36ac0ffda53e00d2cdace2fd4a9e41e2fe3214032794332c37c3.
//
// Solidity: event SuccessVoteSubmitted(uint64 indexed keyperSetIndex, uint64 indexed retryCounter, uint64 indexed keyperIndex, bytes eonPublicKey)
func (_DKGContract *DKGContractFilterer) ParseSuccessVoteSubmitted(log types.Log) (*DKGContractSuccessVoteSubmitted, error) {
	event := new(DKGContractSuccessVoteSubmitted)
	if err := _DKGContract.contract.UnpackLog(event, "SuccessVoteSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECIESKeyRegistryMetaData contains all meta data concerning the ECIESKeyRegistry contract.
var ECIESKeyRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"keyperSetManagerAddress\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"keyper\",\"type\":\"address\"}],\"name\":\"getKey\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getKeyperAt\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getKeyperCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"keyperSetManager\",\"outputs\":[{\"internalType\":\"contractKeyperSetManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"keyperSetIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"keyperIndex\",\"type\":\"uint64\"},{\"internalType\":\"bytes\",\"name\":\"eciesPublicKey\",\"type\":\"bytes\"}],\"name\":\"registerKey\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"keyper\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"eciesPublicKey\",\"type\":\"bytes\"}],\"name\":\"KeyRegistered\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"KeyperSetNotFinalized\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotAMember\",\"type\":\"error\"}]",
	Bin: "0x0x60a060405234801561000f575f5ffd5b50604051610ec8380380610ec8833981810160405281019061003191906100c9565b8073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff1681525050506100f4565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6100988261006f565b9050919050565b6100a88161008e565b81146100b2575f5ffd5b50565b5f815190506100c38161009f565b92915050565b5f602082840312156100de576100dd61006b565b5b5f6100eb848285016100b5565b91505092915050565b608051610db56101135f395f81816101fb01526102200152610db55ff3fe608060405234801561000f575f5ffd5b5060043610610055575f3560e01c806393790f44146100595780639ce385a214610089578063d4b2cedd146100b9578063df713e19146100d7578063ecad263c146100f3575b5f5ffd5b610073600480360381019061006e9190610674565b610111565b604051610080919061070f565b60405180910390f35b6100a3600480360381019061009e9190610762565b6101de565b6040516100b0919061079c565b60405180910390f35b6100c16101f9565b6040516100ce9190610810565b60405180910390f35b6100f160048036038101906100ec91906108c7565b61021d565b005b6100fb6104ef565b6040516101089190610947565b60405180910390f35b606060025f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20805461015b9061098d565b80601f01602080910402602001604051908101604052809291908181526020018280546101879061098d565b80156101d25780601f106101a9576101008083540402835291602001916101d2565b820191905f5260205f20905b8154815290600101906020018083116101b557829003601f168201915b50505050509050919050565b5f6101f2825f6104fe90919063ffffffff16565b9050919050565b7f000000000000000000000000000000000000000000000000000000000000000081565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663f90f3bed866040518263ffffffff1660e01b815260040161027791906109cc565b602060405180830381865afa158015610292573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102b691906109f9565b90508073ffffffffffffffffffffffffffffffffffffffff16638d4e40836040518163ffffffff1660e01b8152600401602060405180830381865afa158015610301573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906103259190610a59565b61035b576040517feac631aa00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b3373ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16632e8e6cad866040518263ffffffff1660e01b81526004016103ab91906109cc565b602060405180830381865afa1580156103c6573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906103ea91906109f9565b73ffffffffffffffffffffffffffffffffffffffff1614610437576040517f2818e88800000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b61044a335f61051590919063ffffffff16565b50828260025f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f209182610497929190610c52565b503373ffffffffffffffffffffffffffffffffffffffff167f9b368ac41b4759b00feab16152db5d57270846c8d258ce8ab77ea58800b288f284846040516104e0929190610d59565b60405180910390a25050505050565b5f6104f95f610542565b905090565b5f61050b835f0183610555565b5f1c905092915050565b5f61053a835f018373ffffffffffffffffffffffffffffffffffffffff165f1b61057c565b905092915050565b5f61054e825f016105e3565b9050919050565b5f825f01828154811061056b5761056a610d7b565b5b905f5260205f200154905092915050565b5f61058783836105f2565b6105d957825f0182908060018154018082558091505060019003905f5260205f20015f9091909190915055825f0180549050836001015f8481526020019081526020015f2081905550600190506105dd565b5f90505b92915050565b5f815f01805490509050919050565b5f5f836001015f8481526020019081526020015f20541415905092915050565b5f5ffd5b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6106438261061a565b9050919050565b61065381610639565b811461065d575f5ffd5b50565b5f8135905061066e8161064a565b92915050565b5f6020828403121561068957610688610612565b5b5f61069684828501610660565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f6106e18261069f565b6106eb81856106a9565b93506106fb8185602086016106b9565b610704816106c7565b840191505092915050565b5f6020820190508181035f83015261072781846106d7565b905092915050565b5f819050919050565b6107418161072f565b811461074b575f5ffd5b50565b5f8135905061075c81610738565b92915050565b5f6020828403121561077757610776610612565b5b5f6107848482850161074e565b91505092915050565b61079681610639565b82525050565b5f6020820190506107af5f83018461078d565b92915050565b5f819050919050565b5f6107d86107d36107ce8461061a565b6107b5565b61061a565b9050919050565b5f6107e9826107be565b9050919050565b5f6107fa826107df565b9050919050565b61080a816107f0565b82525050565b5f6020820190506108235f830184610801565b92915050565b5f67ffffffffffffffff82169050919050565b61084581610829565b811461084f575f5ffd5b50565b5f813590506108608161083c565b92915050565b5f5ffd5b5f5ffd5b5f5ffd5b5f5f83601f84011261088757610886610866565b5b8235905067ffffffffffffffff8111156108a4576108a361086a565b5b6020830191508360018202830111156108c0576108bf61086e565b5b9250929050565b5f5f5f5f606085870312156108df576108de610612565b5b5f6108ec87828801610852565b94505060206108fd87828801610852565b935050604085013567ffffffffffffffff81111561091e5761091d610616565b5b61092a87828801610872565b925092505092959194509250565b6109418161072f565b82525050565b5f60208201905061095a5f830184610938565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f60028204905060018216806109a457607f821691505b6020821081036109b7576109b6610960565b5b50919050565b6109c681610829565b82525050565b5f6020820190506109df5f8301846109bd565b92915050565b5f815190506109f38161064a565b92915050565b5f60208284031215610a0e57610a0d610612565b5b5f610a1b848285016109e5565b91505092915050565b5f8115159050919050565b610a3881610a24565b8114610a42575f5ffd5b50565b5f81519050610a5381610a2f565b92915050565b5f60208284031215610a6e57610a6d610612565b5b5f610a7b84828501610a45565b91505092915050565b5f82905092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f60088302610b177fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610adc565b610b218683610adc565b95508019841693508086168417925050509392505050565b5f610b53610b4e610b498461072f565b6107b5565b61072f565b9050919050565b5f819050919050565b610b6c83610b39565b610b80610b7882610b5a565b848454610ae8565b825550505050565b5f5f905090565b610b97610b88565b610ba2818484610b63565b505050565b5b81811015610bc557610bba5f82610b8f565b600181019050610ba8565b5050565b601f821115610c0a57610bdb81610abb565b610be484610acd565b81016020851015610bf3578190505b610c07610bff85610acd565b830182610ba7565b50505b505050565b5f82821c905092915050565b5f610c2a5f1984600802610c0f565b1980831691505092915050565b5f610c428383610c1b565b9150826002028217905092915050565b610c5c8383610a84565b67ffffffffffffffff811115610c7557610c74610a8e565b5b610c7f825461098d565b610c8a828285610bc9565b5f601f831160018114610cb7575f8415610ca5578287013590505b610caf8582610c37565b865550610d16565b601f198416610cc586610abb565b5f5b82811015610cec57848901358255600182019150602085019450602081019050610cc7565b86831015610d095784890135610d05601f891682610c1b565b8355505b6001600288020188555050505b50505050505050565b828183375f83830152505050565b5f610d3883856106a9565b9350610d45838584610d1f565b610d4e836106c7565b840190509392505050565b5f6020820190508181035f830152610d72818486610d2d565b90509392505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffdfea164736f6c634300081c000a",
}

// ECIESKeyRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ECIESKeyRegistryMetaData.ABI instead.
var ECIESKeyRegistryABI = ECIESKeyRegistryMetaData.ABI

// ECIESKeyRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ECIESKeyRegistryMetaData.Bin instead.
var ECIESKeyRegistryBin = ECIESKeyRegistryMetaData.Bin

// DeployECIESKeyRegistry deploys a new Ethereum contract, binding an instance of ECIESKeyRegistry to it.
func DeployECIESKeyRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, keyperSetManagerAddress common.Address) (common.Address, *types.Transaction, *ECIESKeyRegistry, error) {
	parsed, err := ECIESKeyRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ECIESKeyRegistryBin), backend, keyperSetManagerAddress)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ECIESKeyRegistry{ECIESKeyRegistryCaller: ECIESKeyRegistryCaller{contract: contract}, ECIESKeyRegistryTransactor: ECIESKeyRegistryTransactor{contract: contract}, ECIESKeyRegistryFilterer: ECIESKeyRegistryFilterer{contract: contract}}, nil
}

// ECIESKeyRegistry is an auto generated Go binding around an Ethereum contract.
type ECIESKeyRegistry struct {
	ECIESKeyRegistryCaller     // Read-only binding to the contract
	ECIESKeyRegistryTransactor // Write-only binding to the contract
	ECIESKeyRegistryFilterer   // Log filterer for contract events
}

// ECIESKeyRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ECIESKeyRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECIESKeyRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ECIESKeyRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECIESKeyRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ECIESKeyRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECIESKeyRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ECIESKeyRegistrySession struct {
	Contract     *ECIESKeyRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ECIESKeyRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ECIESKeyRegistryCallerSession struct {
	Contract *ECIESKeyRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// ECIESKeyRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ECIESKeyRegistryTransactorSession struct {
	Contract     *ECIESKeyRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// ECIESKeyRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ECIESKeyRegistryRaw struct {
	Contract *ECIESKeyRegistry // Generic contract binding to access the raw methods on
}

// ECIESKeyRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ECIESKeyRegistryCallerRaw struct {
	Contract *ECIESKeyRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ECIESKeyRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ECIESKeyRegistryTransactorRaw struct {
	Contract *ECIESKeyRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewECIESKeyRegistry creates a new instance of ECIESKeyRegistry, bound to a specific deployed contract.
func NewECIESKeyRegistry(address common.Address, backend bind.ContractBackend) (*ECIESKeyRegistry, error) {
	contract, err := bindECIESKeyRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ECIESKeyRegistry{ECIESKeyRegistryCaller: ECIESKeyRegistryCaller{contract: contract}, ECIESKeyRegistryTransactor: ECIESKeyRegistryTransactor{contract: contract}, ECIESKeyRegistryFilterer: ECIESKeyRegistryFilterer{contract: contract}}, nil
}

// NewECIESKeyRegistryCaller creates a new read-only instance of ECIESKeyRegistry, bound to a specific deployed contract.
func NewECIESKeyRegistryCaller(address common.Address, caller bind.ContractCaller) (*ECIESKeyRegistryCaller, error) {
	contract, err := bindECIESKeyRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ECIESKeyRegistryCaller{contract: contract}, nil
}

// NewECIESKeyRegistryTransactor creates a new write-only instance of ECIESKeyRegistry, bound to a specific deployed contract.
func NewECIESKeyRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ECIESKeyRegistryTransactor, error) {
	contract, err := bindECIESKeyRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ECIESKeyRegistryTransactor{contract: contract}, nil
}

// NewECIESKeyRegistryFilterer creates a new log filterer instance of ECIESKeyRegistry, bound to a specific deployed contract.
func NewECIESKeyRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ECIESKeyRegistryFilterer, error) {
	contract, err := bindECIESKeyRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ECIESKeyRegistryFilterer{contract: contract}, nil
}

// bindECIESKeyRegistry binds a generic wrapper to an already deployed contract.
func bindECIESKeyRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ECIESKeyRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ECIESKeyRegistry *ECIESKeyRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ECIESKeyRegistry.Contract.ECIESKeyRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ECIESKeyRegistry *ECIESKeyRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECIESKeyRegistry.Contract.ECIESKeyRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ECIESKeyRegistry *ECIESKeyRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ECIESKeyRegistry.Contract.ECIESKeyRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ECIESKeyRegistry *ECIESKeyRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ECIESKeyRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ECIESKeyRegistry *ECIESKeyRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECIESKeyRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ECIESKeyRegistry *ECIESKeyRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ECIESKeyRegistry.Contract.contract.Transact(opts, method, params...)
}

// GetKey is a free data retrieval call binding the contract method 0x93790f44.
//
// Solidity: function getKey(address keyper) view returns(bytes)
func (_ECIESKeyRegistry *ECIESKeyRegistryCaller) GetKey(opts *bind.CallOpts, keyper common.Address) ([]byte, error) {
	var out []interface{}
	err := _ECIESKeyRegistry.contract.Call(opts, &out, "getKey", keyper)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetKey is a free data retrieval call binding the contract method 0x93790f44.
//
// Solidity: function getKey(address keyper) view returns(bytes)
func (_ECIESKeyRegistry *ECIESKeyRegistrySession) GetKey(keyper common.Address) ([]byte, error) {
	return _ECIESKeyRegistry.Contract.GetKey(&_ECIESKeyRegistry.CallOpts, keyper)
}

// GetKey is a free data retrieval call binding the contract method 0x93790f44.
//
// Solidity: function getKey(address keyper) view returns(bytes)
func (_ECIESKeyRegistry *ECIESKeyRegistryCallerSession) GetKey(keyper common.Address) ([]byte, error) {
	return _ECIESKeyRegistry.Contract.GetKey(&_ECIESKeyRegistry.CallOpts, keyper)
}

// GetKeyperAt is a free data retrieval call binding the contract method 0x9ce385a2.
//
// Solidity: function getKeyperAt(uint256 index) view returns(address)
func (_ECIESKeyRegistry *ECIESKeyRegistryCaller) GetKeyperAt(opts *bind.CallOpts, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ECIESKeyRegistry.contract.Call(opts, &out, "getKeyperAt", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetKeyperAt is a free data retrieval call binding the contract method 0x9ce385a2.
//
// Solidity: function getKeyperAt(uint256 index) view returns(address)
func (_ECIESKeyRegistry *ECIESKeyRegistrySession) GetKeyperAt(index *big.Int) (common.Address, error) {
	return _ECIESKeyRegistry.Contract.GetKeyperAt(&_ECIESKeyRegistry.CallOpts, index)
}

// GetKeyperAt is a free data retrieval call binding the contract method 0x9ce385a2.
//
// Solidity: function getKeyperAt(uint256 index) view returns(address)
func (_ECIESKeyRegistry *ECIESKeyRegistryCallerSession) GetKeyperAt(index *big.Int) (common.Address, error) {
	return _ECIESKeyRegistry.Contract.GetKeyperAt(&_ECIESKeyRegistry.CallOpts, index)
}

// GetKeyperCount is a free data retrieval call binding the contract method 0xecad263c.
//
// Solidity: function getKeyperCount() view returns(uint256)
func (_ECIESKeyRegistry *ECIESKeyRegistryCaller) GetKeyperCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ECIESKeyRegistry.contract.Call(opts, &out, "getKeyperCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetKeyperCount is a free data retrieval call binding the contract method 0xecad263c.
//
// Solidity: function getKeyperCount() view returns(uint256)
func (_ECIESKeyRegistry *ECIESKeyRegistrySession) GetKeyperCount() (*big.Int, error) {
	return _ECIESKeyRegistry.Contract.GetKeyperCount(&_ECIESKeyRegistry.CallOpts)
}

// GetKeyperCount is a free data retrieval call binding the contract method 0xecad263c.
//
// Solidity: function getKeyperCount() view returns(uint256)
func (_ECIESKeyRegistry *ECIESKeyRegistryCallerSession) GetKeyperCount() (*big.Int, error) {
	return _ECIESKeyRegistry.Contract.GetKeyperCount(&_ECIESKeyRegistry.CallOpts)
}

// KeyperSetManager is a free data retrieval call binding the contract method 0xd4b2cedd.
//
// Solidity: function keyperSetManager() view returns(address)
func (_ECIESKeyRegistry *ECIESKeyRegistryCaller) KeyperSetManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ECIESKeyRegistry.contract.Call(opts, &out, "keyperSetManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// KeyperSetManager is a free data retrieval call binding the contract method 0xd4b2cedd.
//
// Solidity: function keyperSetManager() view returns(address)
func (_ECIESKeyRegistry *ECIESKeyRegistrySession) KeyperSetManager() (common.Address, error) {
	return _ECIESKeyRegistry.Contract.KeyperSetManager(&_ECIESKeyRegistry.CallOpts)
}

// KeyperSetManager is a free data retrieval call binding the contract method 0xd4b2cedd.
//
// Solidity: function keyperSetManager() view returns(address)
func (_ECIESKeyRegistry *ECIESKeyRegistryCallerSession) KeyperSetManager() (common.Address, error) {
	return _ECIESKeyRegistry.Contract.KeyperSetManager(&_ECIESKeyRegistry.CallOpts)
}

// RegisterKey is a paid mutator transaction binding the contract method 0xdf713e19.
//
// Solidity: function registerKey(uint64 keyperSetIndex, uint64 keyperIndex, bytes eciesPublicKey) returns()
func (_ECIESKeyRegistry *ECIESKeyRegistryTransactor) RegisterKey(opts *bind.TransactOpts, keyperSetIndex uint64, keyperIndex uint64, eciesPublicKey []byte) (*types.Transaction, error) {
	return _ECIESKeyRegistry.contract.Transact(opts, "registerKey", keyperSetIndex, keyperIndex, eciesPublicKey)
}

// RegisterKey is a paid mutator transaction binding the contract method 0xdf713e19.
//
// Solidity: function registerKey(uint64 keyperSetIndex, uint64 keyperIndex, bytes eciesPublicKey) returns()
func (_ECIESKeyRegistry *ECIESKeyRegistrySession) RegisterKey(keyperSetIndex uint64, keyperIndex uint64, eciesPublicKey []byte) (*types.Transaction, error) {
	return _ECIESKeyRegistry.Contract.RegisterKey(&_ECIESKeyRegistry.TransactOpts, keyperSetIndex, keyperIndex, eciesPublicKey)
}

// RegisterKey is a paid mutator transaction binding the contract method 0xdf713e19.
//
// Solidity: function registerKey(uint64 keyperSetIndex, uint64 keyperIndex, bytes eciesPublicKey) returns()
func (_ECIESKeyRegistry *ECIESKeyRegistryTransactorSession) RegisterKey(keyperSetIndex uint64, keyperIndex uint64, eciesPublicKey []byte) (*types.Transaction, error) {
	return _ECIESKeyRegistry.Contract.RegisterKey(&_ECIESKeyRegistry.TransactOpts, keyperSetIndex, keyperIndex, eciesPublicKey)
}

// ECIESKeyRegistryKeyRegisteredIterator is returned from FilterKeyRegistered and is used to iterate over the raw logs and unpacked data for KeyRegistered events raised by the ECIESKeyRegistry contract.
type ECIESKeyRegistryKeyRegisteredIterator struct {
	Event *ECIESKeyRegistryKeyRegistered // Event containing the contract specifics and raw log

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
func (it *ECIESKeyRegistryKeyRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECIESKeyRegistryKeyRegistered)
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
		it.Event = new(ECIESKeyRegistryKeyRegistered)
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
func (it *ECIESKeyRegistryKeyRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECIESKeyRegistryKeyRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECIESKeyRegistryKeyRegistered represents a KeyRegistered event raised by the ECIESKeyRegistry contract.
type ECIESKeyRegistryKeyRegistered struct {
	Keyper         common.Address
	EciesPublicKey []byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterKeyRegistered is a free log retrieval operation binding the contract event 0x9b368ac41b4759b00feab16152db5d57270846c8d258ce8ab77ea58800b288f2.
//
// Solidity: event KeyRegistered(address indexed keyper, bytes eciesPublicKey)
func (_ECIESKeyRegistry *ECIESKeyRegistryFilterer) FilterKeyRegistered(opts *bind.FilterOpts, keyper []common.Address) (*ECIESKeyRegistryKeyRegisteredIterator, error) {

	var keyperRule []interface{}
	for _, keyperItem := range keyper {
		keyperRule = append(keyperRule, keyperItem)
	}

	logs, sub, err := _ECIESKeyRegistry.contract.FilterLogs(opts, "KeyRegistered", keyperRule)
	if err != nil {
		return nil, err
	}
	return &ECIESKeyRegistryKeyRegisteredIterator{contract: _ECIESKeyRegistry.contract, event: "KeyRegistered", logs: logs, sub: sub}, nil
}

// WatchKeyRegistered is a free log subscription operation binding the contract event 0x9b368ac41b4759b00feab16152db5d57270846c8d258ce8ab77ea58800b288f2.
//
// Solidity: event KeyRegistered(address indexed keyper, bytes eciesPublicKey)
func (_ECIESKeyRegistry *ECIESKeyRegistryFilterer) WatchKeyRegistered(opts *bind.WatchOpts, sink chan<- *ECIESKeyRegistryKeyRegistered, keyper []common.Address) (event.Subscription, error) {

	var keyperRule []interface{}
	for _, keyperItem := range keyper {
		keyperRule = append(keyperRule, keyperItem)
	}

	logs, sub, err := _ECIESKeyRegistry.contract.WatchLogs(opts, "KeyRegistered", keyperRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECIESKeyRegistryKeyRegistered)
				if err := _ECIESKeyRegistry.contract.UnpackLog(event, "KeyRegistered", log); err != nil {
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

// ParseKeyRegistered is a log parse operation binding the contract event 0x9b368ac41b4759b00feab16152db5d57270846c8d258ce8ab77ea58800b288f2.
//
// Solidity: event KeyRegistered(address indexed keyper, bytes eciesPublicKey)
func (_ECIESKeyRegistry *ECIESKeyRegistryFilterer) ParseKeyRegistered(log types.Log) (*ECIESKeyRegistryKeyRegistered, error) {
	event := new(ECIESKeyRegistryKeyRegistered)
	if err := _ECIESKeyRegistry.contract.UnpackLog(event, "KeyRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

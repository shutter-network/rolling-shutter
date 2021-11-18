package contract

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (_AddrsSeq *AddrsSeqCaller) GetAddrs(opts *bind.CallOpts, n uint64) ([]common.Address, error) {
	numAddresses, err := _AddrsSeq.CountNth(opts, n)
	if err != nil {
		return nil, err
	}
	addresses := []common.Address{}
	for i := uint64(0); i < numAddresses; i++ {
		address, err := _AddrsSeq.At(opts, n, i)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}
	return addresses, nil
}

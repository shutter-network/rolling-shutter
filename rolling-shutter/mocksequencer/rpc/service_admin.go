package rpc

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
)

type AdminService struct {
	processor *mocksequencer.Processor
}

var _ mocksequencer.RPCService = (*AdminService)(nil)

func (s *AdminService) InjectProcessor(p *mocksequencer.Processor) {
	s.processor = p
}

func (s *AdminService) Name() string {
	return "admin"
}

func (s *AdminService) AddCollator(address string, l1BlockNumber uint64) (int, error) {
	collator, err := stringToAddress(address)
	if err != nil {
		return 0, err
	}
	s.processor.Collators.Set(collator, l1BlockNumber)
	return 1, nil
}

func (s *AdminService) AddEonKey(eonKey string, l1BlockNumber uint64) (int, error) {
	eonKeyBytes, err := hexutil.Decode(eonKey)
	if err != nil {
		err = errors.Wrap(err, "eon key could not be decoded")
		return 0, err
	}
	s.processor.EonKeys.Set(eonKeyBytes, l1BlockNumber)
	return 1, nil
}

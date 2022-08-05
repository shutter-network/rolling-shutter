package mocksequencer

import "github.com/ethereum/go-ethereum/common/hexutil"

type AdminService struct {
	processor *SequencerProcessor
}

var _ RPCService = (*AdminService)(nil)

func (s *AdminService) injectProcessor(p *SequencerProcessor) {
	s.processor = p
}

func (s *AdminService) name() string {
	return "admin"
}

func (s *AdminService) AddCollator(address string, l1BlockNumber uint64) (int, error) {
	collator := stringToAddress(address)
	s.processor.collators[l1BlockNumber] = collator
	return 1, nil
}

func (s *AdminService) AddEonKey(eonKey string, l1BlockNumber uint64) (int, error) {
	bytes, err := hexutil.Decode(eonKey)
	if err != nil {
		// TODO return specific decode error
		return 0, err
	}
	s.processor.eonKeys[l1BlockNumber] = bytes
	return 1, nil
}

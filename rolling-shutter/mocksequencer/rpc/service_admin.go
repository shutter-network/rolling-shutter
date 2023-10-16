package rpc

import (
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer/encoding"
)

type AdminService struct {
	processor Sequencer
}

var _ RPCService = (*AdminService)(nil)

func (s *AdminService) InjectProcessor(p Sequencer) {
	s.processor = p
}

func (s *AdminService) Name() string {
	return "admin"
}

func (s *AdminService) AddCollator(address string, l1BlockNumber uint64) (int, error) {
	var err error
	defer func() {
		log.Info().Err(err).Str("address", address).Uint64("l1-blocknumber", l1BlockNumber).Msg("admin method AddCollator called")
	}()

	collator, err := encoding.StringToAddress(address)
	if err != nil {
		return 0, err
	}
	s.processor.SetCollator(collator, l1BlockNumber)
	return 1, nil
}

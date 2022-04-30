package shutterevents

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

// KeyperIndex returns the index of the keyper identified by the given address.
func (bc *BatchConfig) KeyperIndex(address common.Address) (uint64, bool) {
	for i, k := range bc.Keypers {
		if k == address {
			return uint64(i), true
		}
	}
	return 0, false
}

// IsKeyper checks if the given address is a keyper.
func (bc *BatchConfig) IsKeyper(candidate common.Address) bool {
	_, ok := bc.KeyperIndex(candidate)
	return ok
}

// EnsureValid checks if the BatchConfig is valid and returns an error if it's not valid.
func (bc *BatchConfig) EnsureValid() error {
	if len(bc.Keypers) == 0 {
		return errors.Errorf("no keypers in batch config")
	}
	if bc.Threshold == 0 {
		return errors.Errorf("threshold must not be zero")
	}
	if int(bc.Threshold) > len(bc.Keypers) {
		return errors.Errorf("threshold too high")
	}
	// XXX maybe we should check for duplicate addresses
	return nil
}

// BatchConfigFromMessage extracts the batch config received in a message. Started and
// ValidatorsUpdated which are not present on the message are set to false.
func BatchConfigFromMessage(m *shmsg.BatchConfig) (BatchConfig, error) {
	var keypers []common.Address
	for _, b := range m.Keypers {
		if len(b) != common.AddressLength {
			return BatchConfig{}, errors.Errorf("keyper address has invalid length")
		}
		keypers = append(keypers, common.BytesToAddress(b))
	}

	if err := medley.EnsureUniqueAddresses(keypers); err != nil {
		return BatchConfig{}, err
	}

	bc := BatchConfig{
		ActivationBlockNumber: m.ActivationBlockNumber,
		Keypers:               keypers,
		Threshold:             m.Threshold,
		KeyperConfigIndex:     m.KeyperConfigIndex,
		Started:               false,
		ValidatorsUpdated:     false,
	}
	return bc, nil
}

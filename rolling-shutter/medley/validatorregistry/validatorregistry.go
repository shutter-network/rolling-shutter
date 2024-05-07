package validatorregistry

import (
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type RegistrationMessage struct {
	Version                  uint8
	ChainID                  uint64
	ValidatorRegistryAddress common.Address
	ValidatorIndex           uint64
	Nonce                    uint64
	IsRegistration           bool
}

func (m *RegistrationMessage) Marshal() []byte {
	b := make([]byte, 0)
	b = append(b, m.Version)
	b = binary.BigEndian.AppendUint64(b, m.ChainID)
	b = append(b, m.ValidatorRegistryAddress.Bytes()...)
	b = binary.BigEndian.AppendUint64(b, m.ValidatorIndex)
	b = binary.BigEndian.AppendUint64(b, m.Nonce)
	if m.IsRegistration {
		b = append(b, 1)
	} else {
		b = append(b, 0)
	}
	return b
}

func (m *RegistrationMessage) Unmarshal(b []byte) error {
	expectedLength := 1 + 8 + 20 + 8 + 8 + 1
	if len(b) != expectedLength {
		return fmt.Errorf("invalid registration message length %d, expected %d", len(b), expectedLength)
	}

	m.Version = b[0]
	m.ChainID = binary.BigEndian.Uint64(b[1:9])
	m.ValidatorRegistryAddress = common.BytesToAddress(b[9:29])
	m.ValidatorIndex = binary.BigEndian.Uint64(b[29:37])
	m.Nonce = binary.BigEndian.Uint64(b[37:45])
	switch b[45] {
	case 0:
		m.IsRegistration = false
	case 1:
		m.IsRegistration = true
	default:
		return fmt.Errorf("invalid registration message type byte %d", b[45])
	}
	return nil
}

package gnosisaccessnode

import (
	"sync"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
)

type Storage struct {
	mu         sync.Mutex
	eonKeys    map[uint64]*shcrypto.EonPublicKey
	keyperSets map[uint64]*obskeyperdatabase.KeyperSet
}

func NewStorage() *Storage {
	return &Storage{
		mu:         sync.Mutex{},
		eonKeys:    make(map[uint64]*shcrypto.EonPublicKey),
		keyperSets: make(map[uint64]*obskeyperdatabase.KeyperSet),
	}
}

func (s *Storage) AddEonKey(keyperConfigIndex uint64, key *shcrypto.EonPublicKey) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.eonKeys[keyperConfigIndex] = key
}

func (s *Storage) GetEonKey(keyperConfigIndex uint64) (*shcrypto.EonPublicKey, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.eonKeys[keyperConfigIndex]
	return v, ok
}

func (s *Storage) AddKeyperSet(keyperConfigIndex uint64, keyperSet *obskeyperdatabase.KeyperSet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keyperSets[keyperConfigIndex] = keyperSet
}

func (s *Storage) GetKeyperSet(keyperConfigIndex uint64) (*obskeyperdatabase.KeyperSet, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.keyperSets[keyperConfigIndex]
	return v, ok
}

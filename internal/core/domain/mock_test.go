package domain_test

import (
	"strings"
	"sync"

	"github.com/equitas-foundation/bamp-ocean/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// MnemonicStore
type inMemoryMnemonicStore struct {
	mnemonic []string
	lock     *sync.RWMutex
}

func newInMemoryMnemonicStore() domain.IMnemonicStore {
	return &inMemoryMnemonicStore{
		lock: &sync.RWMutex{},
	}
}

func (s *inMemoryMnemonicStore) Set(mnemonic string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.mnemonic = strings.Split(mnemonic, " ")
}

func (s *inMemoryMnemonicStore) Unset() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.mnemonic = nil
}

func (s *inMemoryMnemonicStore) IsSet() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.mnemonic) > 0
}

func (s *inMemoryMnemonicStore) Get() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.mnemonic
}

// MnemonicCypher
type mockMnemonicCypher struct {
	mock.Mock
}

func (m *mockMnemonicCypher) Encrypt(
	mnemonic, password []byte,
) ([]byte, error) {
	args := m.Called(mnemonic, password)

	var res []byte
	if a := args.Get(0); a != nil {
		res = a.([]byte)
	}
	return res, args.Error(1)
}

func (m *mockMnemonicCypher) Decrypt(
	encryptedMnemonic, password []byte,
) ([]byte, error) {
	args := m.Called(encryptedMnemonic, password)

	var res []byte
	if a := args.Get(0); a != nil {
		res = a.([]byte)
	}
	return res, args.Error(1)
}

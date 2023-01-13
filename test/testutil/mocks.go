package testutil

import (
	"github.com/stretchr/testify/mock"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"strings"
	"sync"
)

// inMemoryMnemonicStore
type inMemoryMnemonicStore struct {
	mnemonic []string
	lock     *sync.RWMutex
}

func NewInMemoryMnemonicStore() domain.IMnemonicStore {
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
type MockMnemonicCypher struct {
	mock.Mock
}

func (m *MockMnemonicCypher) Encrypt(
	mnemonic, password []byte,
) ([]byte, error) {
	args := m.Called(mnemonic, password)

	var res []byte
	if a := args.Get(0); a != nil {
		res = a.([]byte)
	}
	return res, args.Error(1)
}

func (m *MockMnemonicCypher) Decrypt(
	encryptedMnemonic, password []byte,
) ([]byte, error) {
	args := m.Called(encryptedMnemonic, password)

	var res []byte
	if a := args.Get(0); a != nil {
		res = a.([]byte)
	}
	return res, args.Error(1)
}

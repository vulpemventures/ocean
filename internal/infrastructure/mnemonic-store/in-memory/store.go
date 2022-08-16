package mnemonic_store

import (
	"strings"

	"github.com/vulpemventures/ocean/internal/config"
)

const (
	mnemonicKey = "MNEMONIC"
)

type MnemonicInMemoryStore struct{}

func NewInMemoryMnemonicStore() *MnemonicInMemoryStore {
	return &MnemonicInMemoryStore{}
}

func (s *MnemonicInMemoryStore) Set(mnemonic string) {
	config.Set(mnemonicKey, mnemonic)
}

func (s *MnemonicInMemoryStore) Unset() {
	config.Unset(mnemonicKey)
}

func (s *MnemonicInMemoryStore) IsSet() bool {
	return len(config.GetString(mnemonicKey)) > 0
}

func (s *MnemonicInMemoryStore) Get() []string {
	mnemonic := config.GetString(mnemonicKey)
	return strings.Split(mnemonic, " ")
}

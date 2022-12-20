package db_test

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
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

func randomUtxosForAccount(
	accountName string,
) ([]*domain.Utxo, []domain.UtxoKey, map[string]*domain.Balance) {
	num := randomIntInRange(1, 5)
	utxos := make([]*domain.Utxo, 0, num)
	keys := make([]domain.UtxoKey, 0, num)
	balanceByAsset := make(map[string]*domain.Balance)
	for i := 0; i < num; i++ {
		key := randomKey()
		utxo := &domain.Utxo{
			UtxoKey:         key,
			Value:           randomValue(),
			Asset:           randomHex(32),
			ValueCommitment: randomValueCommitment(),
			AssetCommitment: randomAssetCommitment(),
			ValueBlinder:    randomBytes(32),
			AssetBlinder:    randomBytes(32),
			Script:          randomScript(),
			Nonce:           randomBytes(33),
			AccountName:     accountName,
		}

		if _, ok := balanceByAsset[utxo.Asset]; !ok {
			balanceByAsset[utxo.Asset] = &domain.Balance{}
		}
		balanceByAsset[utxo.Asset].Unconfirmed += utxo.Value
		keys = append(keys, key)
		utxos = append(utxos, utxo)
	}
	return utxos, keys, balanceByAsset
}

func randomTx(accountName string) *domain.Transaction {
	return &domain.Transaction{
		TxID:  randomHex(32),
		TxHex: randomHex(100),
		Accounts: map[string]struct{}{
			accountName: {},
		},
	}
}

func randomKey() domain.UtxoKey {
	return domain.UtxoKey{
		TxID: randomHex(32),
		VOut: randomVout(),
	}
}

func randomScript() []byte {
	return append([]byte{0, 20}, randomBytes(20)...)
}

func randomValueCommitment() []byte {
	return append([]byte{9}, randomBytes(32)...)
}

func randomAssetCommitment() []byte {
	return append([]byte{10}, randomBytes(32)...)
}

func randomHex(len int) string {
	return hex.EncodeToString(randomBytes(len))
}

func randomVout() uint32 {
	return uint32(randomIntInRange(0, 15))
}

func randomValue() uint64 {
	return uint64(randomIntInRange(1000000, 10000000000))
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	rand.Read(b)
	return b
}

func randomIntInRange(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(int(n.Int64())) + min
}

func b2h(buf []byte) string {
	return hex.EncodeToString(buf)
}

func h2b(str string) []byte {
	buf, _ := hex.DecodeString(str)
	return buf
}

package application_test

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
	"sync"

	"github.com/stretchr/testify/mock"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/ocean/internal/core/application"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

// ports.BlockchainScanner
type mockBcScanner struct {
	mock.Mock
	chTxs   chan *domain.Transaction
	chUtxos chan []*domain.Utxo
}

func newMockedBcScanner() *mockBcScanner {
	return &mockBcScanner{
		chTxs:   make(chan *domain.Transaction),
		chUtxos: make(chan []*domain.Utxo),
	}
}

func (m *mockBcScanner) Start() {}
func (m *mockBcScanner) Stop()  {}
func (m *mockBcScanner) WatchForAccount(
	accountName string, staringBlock uint32, addrInfo []domain.AddressInfo,
) {
	addresses := application.AddressesInfo(addrInfo).Addresses()
	if len(addresses) > 0 {
		utxos := randomUtxos(accountName, addresses)
		m.chUtxos <- utxos

		for _, u := range utxos {
			tx := randomTx(accountName, u.TxID)
			m.chTxs <- tx
		}
	}
}
func (m *mockBcScanner) WatchForUtxos(
	accountName string, utxos []domain.UtxoInfo,
) {
	if len(utxos) > 0 {
		list := make([]*domain.Utxo, 0, len(utxos))
		for _, u := range utxos {
			list = append(list, &domain.Utxo{
				UtxoKey:            u.Key(),
				Value:              u.Value,
				Asset:              u.Asset,
				Script:             u.Script,
				AssetBlinder:       u.AssetBlinder,
				ValueBlinder:       u.ValueBlinder,
				SpentStatus:        u.SpentStatus,
				ConfirmedStatus:    u.ConfirmedStatus,
				FkAccountNamespace: u.FkAccountNamespace,
			})
		}
		m.chUtxos <- list

		for _, u := range utxos {
			tx := randomTx(accountName, u.TxID)
			m.chTxs <- tx
		}
	}
}

func (m *mockBcScanner) StopWatchForAccount(accountName string) {
	close(m.chTxs)
	close(m.chUtxos)
}

func (m *mockBcScanner) GetUtxoChannel(accountName string) chan []*domain.Utxo {
	return m.chUtxos
}

func (m *mockBcScanner) GetTxChannel(accountName string) chan *domain.Transaction {
	return m.chTxs
}

func (m *mockBcScanner) GetLatestBlock() ([]byte, uint32, error) {
	args := m.Called()
	var res []byte
	if a := args.Get(0); a != nil {
		res = a.([]byte)
	}
	var res1 uint32
	if a := args.Get(1); a != nil {
		res1 = a.(uint32)
	}
	return res, res1, args.Error(2)
}

func (m *mockBcScanner) GetBlockHash(height uint32) ([]byte, error) {
	args := m.Called(height)
	var res []byte
	if a := args.Get(0); a != nil {
		res = a.([]byte)
	}
	return res, args.Error(1)
}

func (m *mockBcScanner) GetBlockHeight(hash []byte) (uint32, error) {
	args := m.Called(hash)
	var res uint32
	if a := args.Get(0); a != nil {
		res = a.(uint32)
	}
	return res, args.Error(1)
}

func (m *mockBcScanner) GetUtxos(utxos []domain.Utxo) ([]domain.Utxo, error) {
	args := m.Called(utxos)
	var res []domain.Utxo
	if a := args.Get(0); a != nil {
		res = a.([]domain.Utxo)
	}
	return res, args.Error(1)
}

func (m *mockBcScanner) BroadcastTransaction(txHex string) (string, error) {
	args := m.Called(txHex)
	var res string
	if a := args.Get(0); a != nil {
		res = a.(string)
	}
	return res, args.Error(1)
}

// domain.MnemonicStore
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

// domain.MnemonicCypher
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

func randomUtxos(accountName string, addresses []string) []*domain.Utxo {
	utxos := make([]*domain.Utxo, 0, len(addresses))
	for _, addr := range addresses {
		utxos = append(utxos, randomUtxo(accountName, addr))
	}
	return utxos
}

func randomUtxo(accountNamespace, addr string) *domain.Utxo {
	script, _ := address.ToOutputScript(addr)
	nonce := append([]byte{3}, randomBytes(32)...)
	return &domain.Utxo{
		UtxoKey: domain.UtxoKey{
			TxID: randomHex(32),
			VOut: randomVout(),
		},
		Value:              randomValue(),
		Asset:              randomHex(32),
		ValueCommitment:    randomValueCommitment(),
		AssetCommitment:    randomAssetCommitment(),
		ValueBlinder:       randomBytes(32),
		AssetBlinder:       randomBytes(32),
		Script:             script,
		Nonce:              nonce,
		FkAccountNamespace: accountNamespace,
		ConfirmedStatus: domain.UtxoStatus{
			BlockHeight: uint64(randomIntInRange(1, 10000)),
		},
	}
}

func randomTx(txid, accountName string) *domain.Transaction {
	return &domain.Transaction{
		TxHex:       randomHex(200),
		TxID:        txid,
		BlockHash:   randomHex(32),
		BlockHeight: uint64(randomIntInRange(1, 10000)),
		Accounts: map[string]struct{}{
			accountName: {},
		},
	}
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

func h2b(str string) []byte {
	buf, _ := hex.DecodeString(str)
	return buf
}

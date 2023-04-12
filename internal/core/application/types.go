package application

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	ss_selector "github.com/vulpemventures/ocean/internal/infrastructure/coin-selector/smallest-subset"
	wallet "github.com/vulpemventures/ocean/pkg/wallet"
)

const (
	CoinSelectionStrategySmallestSubset = iota
)

var (
	coinSelectorByType = map[int]CoinSelectorFactory{
		CoinSelectionStrategySmallestSubset: ss_selector.NewSmallestSubsetCoinSelector,
	}

	DefaultCoinSelector = ss_selector.NewSmallestSubsetCoinSelector()
	MinMillisatsPerByte = uint64(100)
)

type WalletStatus struct {
	IsInitialized bool
	IsUnlocked    bool
	IsSynced      bool
}

type WalletInfo struct {
	Network             string
	NativeAsset         string
	RootPath            string
	BirthdayBlockHash   string
	BirthdayBlockHeight uint32
	Accounts            []AccountInfo
	BuildInfo           BuildInfo
}

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

type UtxoInfo struct {
	Spendable Utxos
	Locked    Utxos
}

type TransactionInfo domain.Transaction

type BlockInfo struct {
	Hash      []byte
	Height    uint32
	Timestamp int64
}

type AccountInfo struct {
	domain.AccountInfo
}

type AddressesInfo []domain.AddressInfo

func (info AddressesInfo) Addresses() []string {
	addresses := make([]string, 0, len(info))
	for _, in := range info {
		addresses = append(addresses, in.Address)
	}
	return addresses
}

type BalanceInfo map[string]*domain.Balance

type Utxos []*domain.Utxo

func (u Utxos) Keys() []domain.UtxoKey {
	keys := make([]domain.UtxoKey, 0, len(u))
	for _, utxo := range u {
		keys = append(keys, utxo.Key())
	}
	return keys
}

func (u Utxos) Info() []domain.UtxoInfo {
	info := make([]domain.UtxoInfo, 0, len(u))
	for _, utxo := range u {
		info = append(info, utxo.Info())
	}
	return info
}

type UtxosInfo []domain.UtxoInfo

func (u UtxosInfo) Keys() []domain.UtxoKey {
	keys := make([]domain.UtxoKey, 0, len(u))
	for _, utxo := range u {
		keys = append(keys, utxo.Key())
	}
	return keys
}

type UtxoKeys []domain.UtxoKey

func (u UtxoKeys) String() string {
	str := make([]string, 0, len(u))
	for _, key := range u {
		str = append(str, key.String())
	}
	return strings.Join(str, ", ")
}

type Input struct {
	TxID          string
	VOut          uint32
	Script        string
	ScriptSigSize int
	WitnessSize   int
}

func (i Input) toUtxoKey() domain.UtxoKey {
	return domain.UtxoKey{
		TxID: i.TxID,
		VOut: i.VOut,
	}
}

func (i Input) toUtxo() domain.Utxo {
	buf, _ := hex.DecodeString(i.Script)
	return domain.Utxo{
		UtxoKey: i.toUtxoKey(),
		Script:  buf,
	}
}

type UnblindedInput struct {
	Index         uint32
	Amount        uint64
	Asset         string
	AmountBlinder string
	AssetBlinder  string
}

type Output struct {
	Asset        string
	Amount       uint64
	Script       []byte
	BlindingKey  []byte
	BlinderIndex uint32
}

func (o Output) Validate() error {
	if err := validateAsset(o.Asset); err != nil {
		return err
	}
	if err := validateScript(o.Script); err != nil {
		return err
	}
	if err := validateBlindingKey(o.BlindingKey); err != nil {
		return err
	}
	return nil
}

type Inputs []Input

type Outputs []Output

type CoinSelectorFactory func() ports.CoinSelector

func (o Outputs) totalAmountByAsset() map[string]uint64 {
	totAmount := make(map[string]uint64)
	for _, out := range o {
		totAmount[out.Asset] += out.Amount
	}
	return totAmount
}

func (o Outputs) toWalletOutputs() []wallet.Output {
	outs := make([]wallet.Output, 0, len(o))
	for _, out := range o {
		outs = append(outs, wallet.Output(out))
	}
	return outs
}

type transactionQueue struct {
	lock                *sync.RWMutex
	transactions        []*domain.Transaction
	indexedTransactions map[string]struct{}
}

func newTransactionQueue() *transactionQueue {
	return &transactionQueue{
		&sync.RWMutex{}, make([]*domain.Transaction, 0), make(map[string]struct{}),
	}
}

func (q *transactionQueue) len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return len(q.transactions)
}

func (q *transactionQueue) pushBack(newTx *domain.Transaction) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if _, ok := q.indexedTransactions[newTx.TxID]; !ok {
		q.transactions = append(q.transactions, newTx)
		q.indexedTransactions[newTx.TxID] = struct{}{}
		return
	}

	for i, tx := range q.transactions {
		if tx.TxID == newTx.TxID {
			for _, account := range newTx.GetAccounts() {
				q.transactions[i].AddAccount(account)
				q.transactions[i].BlockHash = newTx.BlockHash
				q.transactions[i].BlockHeight = newTx.BlockHeight
			}
		}
	}
}

func (q *transactionQueue) pop() []*domain.Transaction {
	q.lock.RLock()
	defer q.lock.RUnlock()

	txs := make([]*domain.Transaction, len(q.transactions))
	copy(txs, q.transactions)

	q.transactions = make([]*domain.Transaction, 0)
	q.indexedTransactions = make(map[string]struct{})
	return txs
}

func validateAsset(asset string) error {
	if asset == "" {
		return fmt.Errorf("missing asset")
	}
	buf, err := hex.DecodeString(asset)
	if err != nil {
		return fmt.Errorf("asset is not in hex format")
	}
	if len(buf) != 32 {
		return fmt.Errorf("invalid asset length")
	}
	return nil
}

func validateScript(script []byte) error {
	if len(script) == 0 {
		return nil
	}
	_, err := address.ParseScript(script)
	if err != nil {
		return err
	}
	return nil
}

func validateBlindingKey(key []byte) error {
	if len(key) == 0 {
		return nil
	}
	_, err := btcec.ParsePubKey(key)
	if err != nil {
		return err
	}
	return nil
}

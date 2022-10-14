package application

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	ss_selector "github.com/vulpemventures/ocean/internal/infrastructure/coin-selector/smallest-subset"
	wallet "github.com/vulpemventures/ocean/pkg/single-key-wallet"
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
	MasterBlindingKey   string
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

type AccountInfo domain.AccountInfo

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

type Input domain.UtxoKey

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

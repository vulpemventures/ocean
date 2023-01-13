package testutil

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/application"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"testing"
)

var (
	explorerURL = "http://127.0.0.1:3001"
)

func Faucet(address string) (string, error) {
	url := fmt.Sprintf("%s/faucet", explorerURL)
	payload := map[string]string{"address": address}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if res := string(data); len(res) <= 0 || strings.Contains(res, "sendtoaddress") {
		return "", fmt.Errorf("cannot fund address with faucet: %s", res)
	}

	respBody := map[string]string{}
	if err := json.Unmarshal(data, &respBody); err != nil {
		return "", err
	}
	return respBody["txId"], nil
}

var (
	HandlerFactory = func(t *testing.T, repoType string) ports.WalletEventHandler {
		return func(event domain.WalletEvent) {
			addresses := application.AddressesInfo(event.AccountAddresses).Addresses()
			t.Logf(
				"received event from %s repo: {EventType: %s, AccountName: %s, AccountAddresses: %v}\n",
				repoType, event.EventType, event.AccountNamespace, addresses,
			)
		}
	}
)

func PrepareTestCaseData(
	t *testing.T,
	ctx context.Context,
	walletRepo domain.WalletRepository,
	mnemonic []string,
	password, newPassword, rootPath, regtest string,
	birthdayBlock uint32,
) ([]*domain.Utxo, []domain.UtxoKey, map[string]*domain.Balance, string) {
	w, _ := domain.NewWallet(
		mnemonic, password, rootPath, regtest, birthdayBlock, nil,
	)
	err := walletRepo.CreateWallet(ctx, w)
	require.NoError(t, err)
	err = w.Unlock(password)
	require.NoError(t, err)

	account, err := walletRepo.CreateAccount(ctx, "84", "myAccount", 0)
	require.NoError(t, err)
	namespace := account.Key.Namespace

	newUtxos, utxoKeys, balanceByAsset := RandomUtxosForAccount(namespace)

	return newUtxos, utxoKeys, balanceByAsset, namespace
}

func RandomUtxosForAccount(
	namespace string,
) ([]*domain.Utxo, []domain.UtxoKey, map[string]*domain.Balance) {
	num := RandomIntInRange(1, 5)
	utxos := make([]*domain.Utxo, 0, num)
	keys := make([]domain.UtxoKey, 0, num)
	balanceByAsset := make(map[string]*domain.Balance)
	for i := 0; i < num; i++ {
		key := RandomKey()
		utxo := &domain.Utxo{
			UtxoKey:            key,
			Value:              RandomValue(),
			Asset:              RandomHex(32),
			ValueCommitment:    RandomValueCommitment(),
			AssetCommitment:    RandomAssetCommitment(),
			ValueBlinder:       RandomBytes(32),
			AssetBlinder:       RandomBytes(32),
			Script:             RandomScript(),
			Nonce:              RandomBytes(33),
			FkAccountNamespace: namespace,
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

func RandomTx(accountNamespace string) *domain.Transaction {
	return &domain.Transaction{
		TxID:  RandomHex(32),
		TxHex: RandomHex(100),
		Accounts: map[string]struct{}{
			accountNamespace: {},
		},
	}
}

func RandomKey() domain.UtxoKey {
	return domain.UtxoKey{
		TxID: RandomHex(32),
		VOut: RandomVout(),
	}
}

func RandomScript() []byte {
	return append([]byte{0, 20}, RandomBytes(20)...)
}

func RandomValueCommitment() []byte {
	return append([]byte{9}, RandomBytes(32)...)
}

func RandomAssetCommitment() []byte {
	return append([]byte{10}, RandomBytes(32)...)
}

func RandomHex(len int) string {
	return hex.EncodeToString(RandomBytes(len))
}

func RandomVout() uint32 {
	return uint32(RandomIntInRange(0, 15))
}

func RandomValue() uint64 {
	return uint64(RandomIntInRange(1000000, 10000000000))
}

func RandomBytes(len int) []byte {
	b := make([]byte, len)
	rand.Read(b)
	return b
}

func RandomIntInRange(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(int(n.Int64())) + min
}

func B2h(buf []byte) string {
	return hex.EncodeToString(buf)
}

func H2b(str string) []byte {
	buf, _ := hex.DecodeString(str)
	return buf
}

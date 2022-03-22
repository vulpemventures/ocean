package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/application"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	dbbadger "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/badger"
)

var (
	testAddresses = []string{
		"el1qqfttsemg4sapwrfmmccyztj4wa8gpn5yfetkda4z5uy5e2jysgrszmj0xa8tzftde78kvtl26dtxw6q6gcuawte5xeyvkunws",
		"el1qqfdqyz747wqdtwvf39243a54dkmktexvk6j4ra4h8jkjsp325k54ec35duzlafwpch0h3pp8qt6yhruuwqs9sxf8ukvvuzcxj",
		"el1qqtlphq32x4zpknyfd3hc64cvxymes8stjr7ecxhqjgaxtp9xu9xy0j5d7su2jlasfzv3kg4gnwkyysyk2qy6wumht9qk05r5e",
	}
)

func TestNotificationService(t *testing.T) {
	testGetUtxoChannel(t)

	testGetTxChannel(t)
}

func testGetUtxoChannel(t *testing.T) {
	repoManager, err := newRepoManagerForNotificationService()
	require.NoError(t, err)
	require.NotNil(t, repoManager)

	svc := application.NewNotificationService(repoManager)

	chEvents, err := svc.GetUtxoChannel(ctx)
	require.NoError(t, err)
	require.NotNil(t, chEvents)

	go listenToUtxoEvents(t, chEvents)

	utxos := randomUtxos(accountName, testAddresses)
	repoManager.UtxoRepository().AddUtxos(ctx, utxos)

	time.Sleep(time.Second)

	keys := application.Utxos(utxos).Keys()
	repoManager.UtxoRepository().LockUtxos(ctx, keys, time.Now().Unix())

	time.Sleep(time.Second)

	repoManager.UtxoRepository().UnlockUtxos(ctx, keys)

	time.Sleep(time.Second)

	repoManager.UtxoRepository().SpendUtxos(ctx, keys)

	time.Sleep(time.Second)

	repoManager.UtxoRepository().DeleteUtxosForAccount(ctx, accountName)
}

func testGetTxChannel(t *testing.T) {
	repoManager, err := newRepoManagerForNotificationService()
	require.NoError(t, err)
	require.NotNil(t, repoManager)

	svc := application.NewNotificationService(repoManager)

	chEvents, err := svc.GetTxChannel(ctx)
	require.NoError(t, err)
	require.NotNil(t, chEvents)

	go listenToTxEvents(t, chEvents)

	txid := randomHex(32)
	tx := randomTx(txid, accountName)
	tx.BlockHash = ""
	tx.BlockHeight = 0
	repoManager.TransactionRepository().AddTransaction(ctx, tx)

	time.Sleep(time.Second)

	blockHash := randomHex(32)
	blockHeight := uint32(randomIntInRange(1, 300))
	repoManager.TransactionRepository().ConfirmTransaction(
		ctx, txid, blockHash, blockHeight,
	)

	repoManager.TransactionRepository().UpdateTransaction(
		ctx, txid, func(t *domain.Transaction) (*domain.Transaction, error) {
			t.AddAccount("test2")
			return t, nil
		},
	)
}

func listenToUtxoEvents(t *testing.T, chEvents chan domain.UtxoEvent) {
	for event := range chEvents {
		t.Logf("received event: %+v\n", event)
	}
}

func listenToTxEvents(t *testing.T, chEvents chan domain.TransactionEvent) {
	for event := range chEvents {
		t.Logf(
			"received event: {EventType: %s, Transaction: {TxID: %s, Accounts: %v, Confirmed: %t}}\n",
			event.EventType, event.Transaction.TxID, event.Transaction.GetAccounts(), event.Transaction.IsConfirmed(),
		)
	}
}

func newRepoManagerForNotificationService() (ports.RepoManager, error) {
	rm, err := dbbadger.NewRepoManager("", nil)
	if err != nil {
		return nil, err
	}

	wallet, err := domain.NewWallet(
		mnemonic, password, rootPath, regtest.Name, birthdayBlock, nil,
	)
	if err != nil {
		return nil, err
	}

	if err := rm.WalletRepository().CreateWallet(ctx, wallet); err != nil {
		return nil, err
	}

	if err := rm.WalletRepository().UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			w.Unlock(password)
			w.CreateAccount(accountName, 0)
			return w, nil
		},
	); err != nil {
		return nil, err
	}

	return rm, nil
}

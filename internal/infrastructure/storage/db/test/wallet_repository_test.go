package db_test

import (
	"context"
	"fmt"
	postgresdb "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/ocean/internal/core/application"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	dbbadger "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/badger"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/inmemory"
)

var (
	mnemonic = []string{
		"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
		"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
		"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
	}
	encryptedMnemonic     = "8f29524ee5995c838ca6f28c7ded7da6dc51de804fd2703775989e65ddc1bb3b60122bf0f430bb3b7a267449aaeee103375737d679bfdabf172c3842048925e6f8952e214f6b900435d24cff938be78ad3bb303d305702fbf168534a45a57ac98ca940d4c3319f14d0c97a20b5bcb456d72857d48d0b4f0e0dcf71d1965b6a42aca8d84fcb66aadeabc812a9994cf66e7a75f8718a031418468f023c560312a02f46ec8e65d5dd65c968ddb93e10950e96c8e730ce7a74d33c6ddad9e12f45e534879f1605eb07fe90432f6592f7996091bbb3e3b2"
	password              = "password"
	newPassword           = "newPassword"
	rootPath              = "m/84'/1'"
	regtest               = network.Regtest.Name
	birthdayBlock         = uint32(1)
	ctx                   = context.Background()
	errSomethingWentWrong = fmt.Errorf("something went wrong")
)

func TestMain(m *testing.M) {
	mockedMnemonicCypher := &mockMnemonicCypher{}
	mockedMnemonicCypher.On("Encrypt", mock.Anything, mock.Anything).Return(h2b(encryptedMnemonic), nil)
	mockedMnemonicCypher.On("Decrypt", mock.Anything, []byte(password)).Return([]byte(strings.Join(mnemonic, " ")), nil)
	mockedMnemonicCypher.On("Decrypt", mock.Anything, []byte(newPassword)).Return([]byte(strings.Join(mnemonic, " ")), nil)
	mockedMnemonicCypher.On("Decrypt", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("invalid password"))
	domain.MnemonicCypher = mockedMnemonicCypher

	os.Exit(m.Run())
}

func TestWalletRepository(t *testing.T) {
	repositories, err := newWalletRepositories(
		func(repoType string) ports.WalletEventHandler {
			return func(event domain.WalletEvent) {
				addresses := application.AddressesInfo(event.AccountAddresses).Addresses()
				t.Logf(
					"received event from %s repo: {EventType: %s, AccountName: %s, AccountAddresses: %v}\n",
					repoType, event.EventType, event.AccountName, addresses,
				)
			}
		},
	)
	require.NoError(t, err)

	for name, repo := range repositories {
		t.Run(name, func(t *testing.T) {
			domain.MnemonicStore = newInMemoryMnemonicStore()
			testWalletRepository(t, repo)
		})
	}
}

func testWalletRepository(t *testing.T, repo domain.WalletRepository) {
	testManageWallet(t, repo)

	testManageWalletAccount(t, repo)
}

func testManageWallet(t *testing.T, repo domain.WalletRepository) {
	t.Run("create_wallet", func(t *testing.T) {
		wallet, err := repo.GetWallet(ctx)
		require.Error(t, err)
		require.Nil(t, wallet)

		w, _ := domain.NewWallet(
			mnemonic, password, rootPath, regtest, birthdayBlock, nil,
		)
		err = repo.CreateWallet(ctx, w)
		require.NoError(t, err)

		err = repo.CreateWallet(ctx, w)
		require.Error(t, err)

		wallet, err = repo.GetWallet(ctx)
		require.NoError(t, err)
		require.NotNil(t, wallet)
		require.Exactly(t, *w, *wallet)
		require.True(t, wallet.IsInitialized())
		require.True(t, wallet.IsLocked())
	})

	t.Run("update_unlock_wallet", func(t *testing.T) {
		err := repo.UpdateWallet(
			ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
				if err := w.ChangePassword(password, newPassword); err != nil {
					return nil, err
				}
				return w, nil
			},
		)
		require.NoError(t, err)

		wallet, err := repo.GetWallet(ctx)
		require.NoError(t, err)
		require.NotNil(t, wallet)
		require.True(t, wallet.IsLocked())

		err = repo.UpdateWallet(
			ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
				return nil, errSomethingWentWrong
			},
		)
		require.EqualError(t, errSomethingWentWrong, err.Error())

		err = repo.UnlockWallet(ctx, password)
		require.Error(t, err)

		wallet, err = repo.GetWallet(ctx)
		require.NoError(t, err)
		require.NotNil(t, wallet)
		require.True(t, wallet.IsLocked())

		err = repo.UnlockWallet(ctx, newPassword)
		require.NoError(t, err)

		wallet, err = repo.GetWallet(ctx)
		require.NoError(t, err)
		require.NotNil(t, wallet)
		require.False(t, wallet.IsLocked())
	})
}

func testManageWalletAccount(t *testing.T, repo domain.WalletRepository) {
	t.Run("create_wallet_account", func(t *testing.T) {
		err := repo.DeleteAccount(ctx, accountName)
		require.Error(t, err)

		account, err := repo.CreateAccount(ctx, accountName, 0)
		require.NoError(t, err)
		require.NotNil(t, account)

		account, err = repo.CreateAccount(ctx, account.Key.Name, 0)
		require.Error(t, err)
		require.Nil(t, account)
	})

	t.Run("derive_wallet_account_addresses", func(t *testing.T) {
		addrInfo, err := repo.DeriveNextExternalAddressesForAccount(ctx, accountName, 2)
		require.NoError(t, err)
		require.Len(t, addrInfo, 2)

		addrInfo, err = repo.DeriveNextInternalAddressesForAccount(ctx, accountName, 3)
		require.NoError(t, err)
		require.Len(t, addrInfo, 3)
	})

	t.Run("delete_wallet_account", func(t *testing.T) {
		err := repo.DeleteAccount(ctx, accountName)
		require.NoError(t, err)

		wallet, err := repo.GetWallet(ctx)
		require.NoError(t, err)
		require.NotNil(t, wallet)

		account, err := wallet.GetAccount(accountName)
		require.Error(t, err)
		require.Nil(t, account)
	})
}

func newWalletRepositories(
	handlerFactory func(repoType string) ports.WalletEventHandler,
) (map[string]domain.WalletRepository, error) {
	inmemoryRepoManager := inmemory.NewRepoManager()
	badgerRepoManager, err := dbbadger.NewRepoManager("", nil)
	if err != nil {
		return nil, err
	}
	handlers := []ports.WalletEventHandler{
		handlerFactory("badger"), handlerFactory("inmemory"),
	}

	pgRepoManager, err := postgresdb.NewRepoManager(postgresdb.DbConfig{
		DbUser:             "root",
		DbPassword:         "secret",
		DbHost:             "127.0.0.1",
		DbPort:             5432,
		DbName:             "oceand-db-test",
		MigrationSourceURL: "file://../postgres/migration",
	})
	if err != nil {
		return nil, err
	}

	repoManagers := []ports.RepoManager{badgerRepoManager, inmemoryRepoManager, pgRepoManager}

	for i, handler := range handlers {
		repoManager := repoManagers[i]
		repoManager.RegisterHandlerForWalletEvent(domain.WalletCreated, handler)
		repoManager.RegisterHandlerForWalletEvent(domain.WalletUnlocked, handler)
		repoManager.RegisterHandlerForWalletEvent(domain.WalletAccountCreated, handler)
		repoManager.RegisterHandlerForWalletEvent(domain.WalletAccountAddressesDerived, handler)
		repoManager.RegisterHandlerForWalletEvent(domain.WalletAccountDeleted, handler)
	}
	return map[string]domain.WalletRepository{
		"inmemory": inmemoryRepoManager.WalletRepository(),
		"badger":   badgerRepoManager.WalletRepository(),
		"postgres": pgRepoManager.WalletRepository(),
	}, nil
}

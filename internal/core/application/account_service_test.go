package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/application"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	dbbadger "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/badger"
)

func TestAccountService(t *testing.T) {
	mockedBcScanner := newMockedBcScanner()
	repoManager, err := newRepoManagerForAccountService()
	require.NoError(t, err)
	require.NotNil(t, repoManager)

	svc := application.NewAccountService(repoManager, mockedBcScanner)

	addresses, err := svc.DeriveAddressForAccount(ctx, accountName, 0)
	require.Error(t, err)
	require.Nil(t, addresses)

	accountInfo, err := svc.CreateAccount(ctx, accountName)
	require.NoError(t, err)
	require.NotNil(t, accountInfo)
	require.Equal(t, accountName, accountInfo.Key.Name)
	require.NotEmpty(t, accountInfo.DerivationPath)
	require.NotEmpty(t, accountInfo.Xpub)

	addresses, err = svc.ListAddressesForAccount(ctx, accountName)
	require.NoError(t, err)
	require.Empty(t, addresses)

	addresses, err = svc.DeriveAddressForAccount(ctx, accountName, 2)
	require.NoError(t, err)
	require.Len(t, addresses, 2)

	changeAddresses, err := svc.DeriveChangeAddressForAccount(ctx, accountName, 0)
	require.NoError(t, err)
	require.Len(t, changeAddresses, 1)

	addresses, err = svc.ListAddressesForAccount(ctx, accountName)
	require.NoError(t, err)
	require.Len(t, addresses, 2)

	utxos, err := svc.ListUtxosForAccount(ctx, accountName)
	require.NoError(t, err)
	require.NotNil(t, utxos)
	require.NotEmpty(t, utxos.Spendable)
	require.Empty(t, utxos.Locked)

	balance, err := svc.GetBalanceForAccount(ctx, accountName)
	require.NoError(t, err)
	require.NotNil(t, balance)

	// Cannot delete an account with non-zero balance.
	err = svc.DeleteAccount(ctx, accountName)
	require.Error(t, err)

	// Simulate withdrawing all funds by spending every spendable utxo coming
	// from ListUtxosForAccount.
	_, err = repoManager.UtxoRepository().SpendUtxos(ctx, utxos.Spendable.Keys())
	require.NoError(t, err)

	// Now deleting the account should work without errors.
	err = svc.DeleteAccount(ctx, accountName)
	require.NoError(t, err)
}

func newRepoManagerForAccountService() (ports.RepoManager, error) {
	rm, err := dbbadger.NewRepoManager("", nil)
	if err != nil {
		return nil, err
	}

	wallet, err := domain.NewWallet(mnemonic, password, rootPath, regtest.Name, nil)
	if err != nil {
		return nil, err
	}

	if err := rm.WalletRepository().CreateWallet(ctx, wallet); err != nil {
		return nil, err
	}

	if err := rm.WalletRepository().UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			w.Unlock(password)
			return w, nil
		},
	); err != nil {
		return nil, err
	}

	return rm, nil
}

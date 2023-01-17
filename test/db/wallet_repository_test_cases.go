package dbtest

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"testing"
	"time"
)

var (
	errSomethingWentWrong = fmt.Errorf("something went wrong")
)

func TestWalletRepository(
	t *testing.T,
	ctx context.Context,
	repo domain.WalletRepository,
	mnemonic []string,
	password, newPassword, rootPath, regtest string,
	birthdayBlock uint32,
) {
	testCreateWallet(
		t,
		ctx,
		repo,
		mnemonic,
		password,
		rootPath,
		regtest,
		birthdayBlock,
	)

	testUpdateUnlockWallet(
		t,
		ctx,
		repo,
		password,
		newPassword,
	)

	createWalletAccount(
		t,
		ctx,
		repo,
	)

	deriveWalletAccountAddresses(
		t,
		ctx,
		repo,
	)

	time.Sleep(1 * time.Second) //wait for events
}

func testCreateWallet(
	t *testing.T,
	ctx context.Context,
	repo domain.WalletRepository,
	mnemonic []string,
	password, rootPath, regtest string,
	birthdayBlock uint32,
) {
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
	require.Equal(t, wallet.NextAccountIndex, uint32(0))
}

func testUpdateUnlockWallet(
	t *testing.T,
	ctx context.Context,
	repo domain.WalletRepository,
	password, newPassword string,
) {
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
}

func createWalletAccount(
	t *testing.T,
	ctx context.Context,
	repo domain.WalletRepository,
) {
	err := repo.DeleteAccount(ctx, "dummy")
	require.Error(t, err)

	account, err := repo.CreateAccount(ctx, "myAccount", 0)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, account.Namespace, "bip84-account0")
	require.Equal(t, account.Label, "myAccount")

	wallet, err := repo.GetWallet(ctx)
	require.NoError(t, err)
	require.Equal(
		t,
		wallet.AccountsByNamespace["bip84-account0"].Info.Namespace,
		wallet.AccountsNamespaceByLabel["myAccount"],
	)
	require.Equal(t, wallet.NextAccountIndex, uint32(1))

	account, err = repo.CreateAccount(ctx, "myAccount1", 0)
	require.NoError(t, err)
	require.Equal(t, account.Namespace, "bip84-account1")
	require.Equal(t, account.Label, "myAccount1")

	wallet, err = repo.GetWallet(ctx)
	require.NoError(t, err)
	require.Equal(t,
		wallet.AccountsByNamespace["bip84-account1"].Info.Namespace,
		wallet.AccountsNamespaceByLabel["myAccount1"],
	)
}

func deriveWalletAccountAddresses(
	t *testing.T,
	ctx context.Context,
	repo domain.WalletRepository,
) {
	addrInfo, err := repo.DeriveNextExternalAddressesForAccount(
		ctx,
		"bip84-account0",
		2,
	)
	require.NoError(t, err)
	require.Len(t, addrInfo, 2)

	addrInfo, err = repo.DeriveNextInternalAddressesForAccount(
		ctx,
		"bip84-account0",
		3,
	)
	require.NoError(t, err)
	require.Len(t, addrInfo, 3)
}

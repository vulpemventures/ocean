package dbtest

import (
	"context"
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/test/testutil"
	"testing"
	"time"
)

func TestUtxoRepository(
	t *testing.T,
	ctx context.Context,
	walletRepo domain.WalletRepository,
	utxoRepo domain.UtxoRepository,
	mnemonic []string,
	password, newPassword, rootPath, regtest string,
	birthdayBlock uint32,
) {
	newUtxos, utxoKeys, balanceByAsset, namespace := testutil.PrepareTestCaseData(
		t,
		ctx,
		walletRepo,
		mnemonic,
		password, newPassword, rootPath, regtest,
		birthdayBlock,
	)

	testAddAndGetUtxos(t, ctx, utxoRepo, newUtxos, utxoKeys, namespace)

	testGetBalanceForAccount(t, ctx, utxoRepo, balanceByAsset, namespace)

	testConfirmUtxos(t, ctx, utxoRepo, newUtxos, utxoKeys, balanceByAsset, namespace)

	testLockUtxos(t, ctx, utxoRepo, newUtxos, utxoKeys, balanceByAsset, namespace)

	testUnlockUtxos(t, ctx, utxoRepo, newUtxos, utxoKeys, balanceByAsset, namespace)

	testSpendUtxos(t, ctx, utxoRepo, newUtxos, utxoKeys, namespace)

	time.Sleep(1 * time.Second) //wait for events
}

func testAddAndGetUtxos(
	t *testing.T,
	ctx context.Context,
	utxoRepo domain.UtxoRepository,
	newUtxos []*domain.Utxo,
	utxoKeys []domain.UtxoKey,
	namespace string,
) {
	count, err := utxoRepo.AddUtxos(ctx, newUtxos)
	require.NoError(t, err)
	require.Equal(t, len(newUtxos), count)

	count, err = utxoRepo.AddUtxos(ctx, newUtxos)
	require.NoError(t, err)
	require.Zero(t, count)

	//get utxos
	utxos, err := utxoRepo.GetAllUtxos(ctx)
	require.NoError(t, err)
	require.Len(t, utxos, len(newUtxos))

	utxos, err = utxoRepo.GetAllUtxosForAccount(ctx, namespace)
	require.NoError(t, err)
	require.Len(t, utxos, len(newUtxos))

	utxos, err = utxoRepo.GetAllUtxosForAccount(ctx, "wrongNamespace")
	require.NoError(t, err)
	require.Empty(t, utxos)

	utxos, err = utxoRepo.GetSpendableUtxos(ctx)
	require.NoError(t, err)
	require.Empty(t, utxos)

	utxos, err = utxoRepo.GetSpendableUtxosForAccount(ctx, namespace)
	require.NoError(t, err)
	require.Empty(t, utxos)

	utxos, err = utxoRepo.GetLockedUtxosForAccount(ctx, namespace)
	require.NoError(t, err)
	require.Empty(t, utxos)

	utxos, err = utxoRepo.GetUtxosByKey(ctx, utxoKeys)
	require.NoError(t, err)
	require.Len(t, utxos, len(newUtxos))

	otherKeys := []domain.UtxoKey{testutil.RandomKey()}
	utxos, err = utxoRepo.GetUtxosByKey(ctx, otherKeys)
	require.NoError(t, err)
	require.Empty(t, utxos)

	allKeys := append(utxoKeys, otherKeys...)
	utxos, err = utxoRepo.GetUtxosByKey(ctx, allKeys)
	require.NoError(t, err)
	require.Len(t, utxos, len(newUtxos))
}

func testGetBalanceForAccount(
	t *testing.T,
	ctx context.Context,
	repo domain.UtxoRepository,
	balanceByAsset map[string]*domain.Balance,
	namespace string,
) {
	utxoBalance, err := repo.GetBalanceForAccount(ctx, namespace)
	require.NoError(t, err)
	require.NotNil(t, utxoBalance)
	for asset, balance := range utxoBalance {
		require.Exactly(t, *balanceByAsset[asset], *balance)
	}

	utxoBalance, err = repo.GetBalanceForAccount(ctx, "wrongNamespace")
	require.NoError(t, err)
	require.Empty(t, utxoBalance)
}

func testConfirmUtxos(
	t *testing.T,
	ctx context.Context,
	utxoRepo domain.UtxoRepository,
	newUtxos []*domain.Utxo,
	utxoKeys []domain.UtxoKey,
	balanceByAsset map[string]*domain.Balance,
	namespace string,
) {
	status := domain.UtxoStatus{"", 1, 0, ""}
	count, err := utxoRepo.ConfirmUtxos(ctx, utxoKeys, status)
	require.NoError(t, err)
	require.Equal(t, len(newUtxos), count)

	count, err = utxoRepo.ConfirmUtxos(ctx, utxoKeys, status)
	require.NoError(t, err)
	require.Zero(t, count)

	utxos, err := utxoRepo.GetSpendableUtxos(ctx)
	require.NoError(t, err)
	require.Len(t, utxos, len(newUtxos))

	utxos, err = utxoRepo.GetSpendableUtxosForAccount(ctx, namespace)
	require.NoError(t, err)
	require.Len(t, utxos, len(newUtxos))

	utxoBalance, err := utxoRepo.GetBalanceForAccount(ctx, namespace)
	require.NoError(t, err)
	require.NotNil(t, utxoBalance)
	for asset, balance := range utxoBalance {
		prevBalance := balanceByAsset[asset]
		require.Equal(t, prevBalance.Unconfirmed, balance.Confirmed)
		require.Equal(t, prevBalance.Confirmed, balance.Unconfirmed)
		require.Equal(t, prevBalance.Locked, balance.Locked)
	}
}

func testLockUtxos(
	t *testing.T,
	ctx context.Context,
	utxoRepo domain.UtxoRepository,
	newUtxos []*domain.Utxo,
	utxoKeys []domain.UtxoKey,
	balanceByAsset map[string]*domain.Balance,
	namespace string,
) {
	count, err := utxoRepo.LockUtxos(ctx, utxoKeys, time.Now().Unix(), 0)
	require.NoError(t, err)
	require.Equal(t, len(newUtxos), count)

	count, err = utxoRepo.LockUtxos(ctx, utxoKeys, time.Now().Unix(), 0)
	require.NoError(t, err)
	require.Zero(t, count)

	utxos, err := utxoRepo.GetLockedUtxosForAccount(ctx, namespace)
	require.NoError(t, err)
	require.Len(t, utxos, len(newUtxos))

	utxos, err = utxoRepo.GetSpendableUtxos(ctx)
	require.NoError(t, err)
	require.Empty(t, utxos)

	utxos, err = utxoRepo.GetSpendableUtxosForAccount(ctx, namespace)
	require.NoError(t, err)
	require.Empty(t, utxos)

	utxoBalance, err := utxoRepo.GetBalanceForAccount(ctx, namespace)
	require.NoError(t, err)
	require.NotNil(t, utxoBalance)
	for asset, balance := range utxoBalance {
		prevBalance := balanceByAsset[asset]
		require.Zero(t, balance.Confirmed)
		require.Zero(t, balance.Unconfirmed)
		require.Equal(t, prevBalance.Unconfirmed, balance.Locked)
	}
}

func testUnlockUtxos(
	t *testing.T,
	ctx context.Context,
	utxoRepo domain.UtxoRepository,
	newUtxos []*domain.Utxo,
	utxoKeys []domain.UtxoKey,
	balanceByAsset map[string]*domain.Balance,
	namespace string,
) {
	count, err := utxoRepo.UnlockUtxos(ctx, utxoKeys)
	require.NoError(t, err)
	require.Equal(t, len(newUtxos), count)

	count, err = utxoRepo.UnlockUtxos(ctx, utxoKeys)
	require.NoError(t, err)
	require.Zero(t, count)

	utxos, err := utxoRepo.GetLockedUtxosForAccount(ctx, namespace)
	require.NoError(t, err)
	require.Empty(t, utxos)

	utxos, err = utxoRepo.GetSpendableUtxos(ctx)
	require.NoError(t, err)
	require.Len(t, utxos, len(newUtxos))

	utxos, err = utxoRepo.GetSpendableUtxosForAccount(ctx, namespace)
	require.NoError(t, err)
	require.Len(t, utxos, len(newUtxos))

	utxoBalance, err := utxoRepo.GetBalanceForAccount(ctx, namespace)
	require.NoError(t, err)
	require.NotNil(t, utxoBalance)
	for asset, balance := range utxoBalance {
		prevBalance := balanceByAsset[asset]
		require.Equal(t, prevBalance.Unconfirmed, balance.Confirmed)
		require.Equal(t, prevBalance.Confirmed, balance.Unconfirmed)
		require.Equal(t, prevBalance.Locked, balance.Locked)
	}
}

func testSpendUtxos(
	t *testing.T,
	ctx context.Context,
	utxoRepo domain.UtxoRepository,
	newUtxos []*domain.Utxo,
	utxoKeys []domain.UtxoKey,
	namespace string,
) {
	status := domain.UtxoStatus{
		Txid:        hex.EncodeToString(make([]byte, 32)),
		BlockHeight: 1,
	}
	count, err := utxoRepo.SpendUtxos(ctx, utxoKeys, status)
	require.NoError(t, err)
	require.Equal(t, len(newUtxos), count)

	count, err = utxoRepo.SpendUtxos(ctx, utxoKeys, status)
	require.NoError(t, err)
	require.Zero(t, count)

	utxos, err := utxoRepo.GetSpendableUtxos(ctx)
	require.NoError(t, err)
	require.Empty(t, utxos)

	utxos, err = utxoRepo.GetSpendableUtxosForAccount(ctx, namespace)
	require.NoError(t, err)
	require.Empty(t, utxos)

	utxoBalance, err := utxoRepo.GetBalanceForAccount(ctx, namespace)
	require.NoError(t, err)
	require.Empty(t, utxoBalance)
}

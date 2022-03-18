package db_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	dbbadger "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/badger"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/inmemory"
)

var (
	accountName      = "test1"
	wrongAccountName = "test2"
	newUtxos         []*domain.Utxo
	utxoKeys         []domain.UtxoKey
	balanceByAsset   map[string]*domain.Balance
)

func TestUtxoRepository(t *testing.T) {
	repositories, err := newUtxoRepositories(
		func(repoType string) ports.UtxoEventHandler {
			return func(event domain.UtxoEvent) {
				t.Logf("received event from %s repo: %+v\n", repoType, event)
			}
		},
	)
	require.NoError(t, err)

	for name, repo := range repositories {
		t.Run(name, func(t *testing.T) {
			testUtxoRepository(t, repo)
		})
	}
}

func testUtxoRepository(t *testing.T, repo domain.UtxoRepository) {
	newUtxos, utxoKeys, balanceByAsset = randomUtxosForAccount(accountName)
	testAddUtxos(t, repo)

	testGetUtxos(t, repo)

	testGetBalanceForAccount(t, repo)

	testConfirmUtxos(t, repo)

	testLockUtxos(t, repo)

	testUnlockUtxos(t, repo)

	testSpendUtxos(t, repo)
}

func testAddUtxos(t *testing.T, repo domain.UtxoRepository) {
	t.Run("add_utxos", func(t *testing.T) {
		count, err := repo.AddUtxos(ctx, newUtxos)
		require.NoError(t, err)
		require.Equal(t, len(newUtxos), count)

		count, err = repo.AddUtxos(ctx, newUtxos)
		require.NoError(t, err)
		require.Zero(t, count)
	})
}

func testGetUtxos(t *testing.T, repo domain.UtxoRepository) {
	t.Run("get_utxos", func(t *testing.T) {
		utxos := repo.GetAllUtxos(ctx)
		require.Len(t, utxos, len(newUtxos))

		utxos, err := repo.GetAllUtxosForAccount(ctx, accountName)
		require.NoError(t, err)
		require.Len(t, utxos, len(newUtxos))

		utxos, err = repo.GetAllUtxosForAccount(ctx, wrongAccountName)
		require.NoError(t, err)
		require.Empty(t, utxos)

		utxos, err = repo.GetSpendableUtxos(ctx)
		require.NoError(t, err)
		require.Empty(t, utxos)

		utxos, err = repo.GetSpendableUtxosForAccount(ctx, accountName)
		require.NoError(t, err)
		require.Empty(t, utxos)

		utxos, err = repo.GetLockedUtxosForAccount(ctx, accountName)
		require.NoError(t, err)
		require.Empty(t, utxos)

		utxos, err = repo.GetUtxosByKey(ctx, utxoKeys)
		require.NoError(t, err)
		require.Len(t, utxos, len(newUtxos))

		otherKeys := []domain.UtxoKey{randomKey()}
		utxos, err = repo.GetUtxosByKey(ctx, otherKeys)
		require.Error(t, err)
		require.Empty(t, utxos)

		allKeys := append(utxoKeys, otherKeys...)
		utxos, err = repo.GetUtxosByKey(ctx, allKeys)
		require.NoError(t, err)
		require.Len(t, utxos, len(newUtxos))
	})
}

func testGetBalanceForAccount(t *testing.T, repo domain.UtxoRepository) {
	t.Run("get_balance_for_account", func(t *testing.T) {
		utxoBalance, err := repo.GetBalanceForAccount(ctx, accountName)
		require.NoError(t, err)
		require.NotNil(t, utxoBalance)
		for asset, balance := range utxoBalance {
			require.Exactly(t, *balanceByAsset[asset], *balance)
		}

		utxoBalance, err = repo.GetBalanceForAccount(ctx, wrongAccountName)
		require.NoError(t, err)
		require.Empty(t, utxoBalance)
	})
}

func testConfirmUtxos(t *testing.T, repo domain.UtxoRepository) {
	t.Run("confirm_utxos", func(t *testing.T) {
		count, err := repo.ConfirmUtxos(ctx, utxoKeys)
		require.NoError(t, err)
		require.Equal(t, len(newUtxos), count)

		count, err = repo.ConfirmUtxos(ctx, utxoKeys)
		require.NoError(t, err)
		require.Zero(t, count)

		utxos, err := repo.GetSpendableUtxos(ctx)
		require.NoError(t, err)
		require.Len(t, utxos, len(newUtxos))

		utxos, err = repo.GetSpendableUtxosForAccount(ctx, accountName)
		require.NoError(t, err)
		require.Len(t, utxos, len(newUtxos))

		utxoBalance, err := repo.GetBalanceForAccount(ctx, accountName)
		require.NoError(t, err)
		require.NotNil(t, utxoBalance)
		for asset, balance := range utxoBalance {
			prevBalance := balanceByAsset[asset]
			require.Equal(t, prevBalance.Unconfirmed, balance.Confirmed)
			require.Equal(t, prevBalance.Confirmed, balance.Unconfirmed)
			require.Equal(t, prevBalance.Locked, balance.Locked)
		}
	})
}

func testLockUtxos(t *testing.T, repo domain.UtxoRepository) {
	t.Run("lock_utxos", func(t *testing.T) {
		count, err := repo.LockUtxos(ctx, utxoKeys, time.Now().Unix())
		require.NoError(t, err)
		require.Equal(t, len(newUtxos), count)

		count, err = repo.LockUtxos(ctx, utxoKeys, time.Now().Unix())
		require.NoError(t, err)
		require.Zero(t, count)

		utxos, err := repo.GetLockedUtxosForAccount(ctx, accountName)
		require.NoError(t, err)
		require.Len(t, utxos, len(newUtxos))

		utxos, err = repo.GetSpendableUtxos(ctx)
		require.NoError(t, err)
		require.Empty(t, utxos)

		utxos, err = repo.GetSpendableUtxosForAccount(ctx, accountName)
		require.NoError(t, err)
		require.Empty(t, utxos)

		utxoBalance, err := repo.GetBalanceForAccount(ctx, accountName)
		require.NoError(t, err)
		require.NotNil(t, utxoBalance)
		for asset, balance := range utxoBalance {
			prevBalance := balanceByAsset[asset]
			require.Zero(t, balance.Confirmed)
			require.Zero(t, balance.Unconfirmed)
			require.Equal(t, prevBalance.Unconfirmed, balance.Locked)
		}
	})
}

func testUnlockUtxos(t *testing.T, repo domain.UtxoRepository) {
	t.Run("unlock_utxos", func(t *testing.T) {
		count, err := repo.UnlockUtxos(ctx, utxoKeys)
		require.NoError(t, err)
		require.Equal(t, len(newUtxos), count)

		count, err = repo.UnlockUtxos(ctx, utxoKeys)
		require.NoError(t, err)
		require.Zero(t, count)

		utxos, err := repo.GetLockedUtxosForAccount(ctx, accountName)
		require.NoError(t, err)
		require.Empty(t, utxos)

		utxos, err = repo.GetSpendableUtxos(ctx)
		require.NoError(t, err)
		require.Len(t, utxos, len(newUtxos))

		utxos, err = repo.GetSpendableUtxosForAccount(ctx, accountName)
		require.NoError(t, err)
		require.Len(t, utxos, len(newUtxos))

		utxoBalance, err := repo.GetBalanceForAccount(ctx, accountName)
		require.NoError(t, err)
		require.NotNil(t, utxoBalance)
		for asset, balance := range utxoBalance {
			prevBalance := balanceByAsset[asset]
			require.Equal(t, prevBalance.Unconfirmed, balance.Confirmed)
			require.Equal(t, prevBalance.Confirmed, balance.Unconfirmed)
			require.Equal(t, prevBalance.Locked, balance.Locked)
		}
	})
}

func testSpendUtxos(t *testing.T, repo domain.UtxoRepository) {
	t.Run("spend_utxos", func(t *testing.T) {
		count, err := repo.SpendUtxos(ctx, utxoKeys)
		require.NoError(t, err)
		require.Equal(t, len(newUtxos), count)

		count, err = repo.SpendUtxos(ctx, utxoKeys)
		require.NoError(t, err)
		require.Zero(t, count)

		utxos, err := repo.GetSpendableUtxos(ctx)
		require.NoError(t, err)
		require.Empty(t, utxos)

		utxos, err = repo.GetSpendableUtxosForAccount(ctx, accountName)
		require.NoError(t, err)
		require.Empty(t, utxos)

		utxoBalance, err := repo.GetBalanceForAccount(ctx, accountName)
		require.NoError(t, err)
		require.Empty(t, utxoBalance)
	})
}

func newUtxoRepositories(handlerFactory func(repoType string) ports.UtxoEventHandler) (map[string]domain.UtxoRepository, error) {
	inmemoryRepoManager := inmemory.NewRepoManager()
	badgerRepoManager, err := dbbadger.NewRepoManager("", nil)
	if err != nil {
		return nil, err
	}
	handlers := []ports.UtxoEventHandler{
		handlerFactory("badger"), handlerFactory("inmemory"),
	}
	repoManagers := []ports.RepoManager{badgerRepoManager, inmemoryRepoManager}

	for i, handler := range handlers {
		repoManager := repoManagers[i]
		repoManager.RegisterHandlerForUtxoEvent(domain.UtxoAdded, handler)
		repoManager.RegisterHandlerForUtxoEvent(domain.UtxoConfirmed, handler)
		repoManager.RegisterHandlerForUtxoEvent(domain.UtxoLocked, handler)
		repoManager.RegisterHandlerForUtxoEvent(domain.UtxoUnlocked, handler)
		repoManager.RegisterHandlerForUtxoEvent(domain.UtxoSpent, handler)
	}
	return map[string]domain.UtxoRepository{
		"inmemory": inmemoryRepoManager.UtxoRepository(),
		"badger":   badgerRepoManager.UtxoRepository(),
	}, nil
}

package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	dbbadger "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/badger"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/inmemory"
)

func TestTransactionRepository(t *testing.T) {
	repositories, err := newTransactionRepositories(
		func(repoType string) ports.TxEventHandler {
			return func(event domain.TransactionEvent) {
				t.Logf(
					"received event from %s repo: {EventType: %s, Transaction: "+
						"{TxID: %s, Accounts: %v}}\n", repoType, event.EventType,
					event.Transaction.TxID, event.Transaction.GetAccounts(),
				)
			}
		},
	)
	require.NoError(t, err)

	for name, repo := range repositories {
		t.Run(name, func(t *testing.T) {
			testTransactionRepository(t, repo)
		})
	}
}

func testTransactionRepository(t *testing.T, repo domain.TransactionRepository) {
	accountName := "test1"
	newTx := randomTx(accountName)
	txid := newTx.TxID
	wrongTxid := randomHex(32)

	t.Run("add_transaction", func(t *testing.T) {
		done, err := repo.AddTransaction(ctx, newTx)
		require.NoError(t, err)
		require.True(t, done)

		done, err = repo.AddTransaction(ctx, newTx)
		require.NoError(t, err)
		require.False(t, done)
	})

	t.Run("get_transaction", func(t *testing.T) {
		tx, err := repo.GetTransaction(ctx, txid)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Exactly(t, *newTx, *tx)

		tx, err = repo.GetTransaction(ctx, wrongTxid)
		require.Error(t, err)
		require.Nil(t, tx)
	})

	t.Run("confirm_transaction", func(t *testing.T) {
		blockHash := randomHex(32)
		blockHeight := uint64(randomIntInRange(100, 1000))

		done, err := repo.ConfirmTransaction(ctx, txid, blockHash, blockHeight)
		require.NoError(t, err)
		require.True(t, done)

		done, err = repo.ConfirmTransaction(ctx, txid, blockHash, blockHeight)
		require.NoError(t, err)
		require.False(t, done)

		tx, err := repo.GetTransaction(ctx, txid)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.True(t, tx.IsConfirmed())
	})

	t.Run("update_transaction", func(t *testing.T) {
		tx, err := repo.GetTransaction(ctx, txid)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Len(t, tx.GetAccounts(), 1)

		err = repo.UpdateTransaction(
			ctx, txid, func(tx *domain.Transaction) (*domain.Transaction, error) {
				return tx, nil
			},
		)
		require.NoError(t, err)

		err = repo.UpdateTransaction(
			ctx, txid, func(tx *domain.Transaction) (*domain.Transaction, error) {
				return nil, errSomethingWentWrong
			},
		)
		require.EqualError(t, errSomethingWentWrong, err.Error())
	})
}

func newTransactionRepositories(
	handlerFactory func(repoType string) ports.TxEventHandler,
) (map[string]domain.TransactionRepository, error) {
	inmemoryRepoManager := inmemory.NewRepoManager()
	badgerRepoManager, err := dbbadger.NewRepoManager("", nil)
	if err != nil {
		return nil, err
	}

	handlers := []ports.TxEventHandler{
		handlerFactory("badger"), handlerFactory("inmemory"),
	}
	repoManagers := []ports.RepoManager{badgerRepoManager, inmemoryRepoManager}

	for i, handler := range handlers {
		repoManager := repoManagers[i]
		repoManager.RegisterHandlerForTxEvent(domain.TransactionAdded, handler)
		repoManager.RegisterHandlerForTxEvent(domain.TransactionUnconfirmed, handler)
		repoManager.RegisterHandlerForTxEvent(domain.TransactionConfirmed, handler)
	}

	return map[string]domain.TransactionRepository{
		"inmemory": inmemoryRepoManager.TransactionRepository(),
		"badger":   badgerRepoManager.TransactionRepository(),
	}, nil
}

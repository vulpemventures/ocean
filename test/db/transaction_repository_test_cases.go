package dbtest

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/test/testutil"
	"testing"
	"time"
)

func TestTransactionRepository(
	t *testing.T,
	ctx context.Context,
	walletRepo domain.WalletRepository,
	transactionRepo domain.TransactionRepository,
	mnemonic []string,
	password, newPassword, rootPath, regtest string,
	birthdayBlock uint32,
) {
	_, _, _, namespace := testutil.PrepareTestCaseData(
		t,
		ctx,
		walletRepo,
		mnemonic,
		password, newPassword, rootPath, regtest,
		birthdayBlock,
	)

	testTransactionRepository(t, ctx, transactionRepo, namespace)

	time.Sleep(1 * time.Second) //wait for events
}

func testTransactionRepository(
	t *testing.T,
	ctx context.Context,
	repo domain.TransactionRepository,
	namespace string,
) {
	newTx := testutil.RandomTx(namespace)
	txid := newTx.TxID
	wrongTxid := testutil.RandomHex(32)

	done, err := repo.AddTransaction(ctx, newTx)
	require.NoError(t, err)
	require.True(t, done)

	done, err = repo.AddTransaction(ctx, newTx)
	require.NoError(t, err)
	require.False(t, done)

	tx, err := repo.GetTransaction(ctx, txid)
	require.NoError(t, err)
	require.NotNil(t, tx)
	require.Exactly(t, *newTx, *tx)
	require.Equal(t, tx.GetAccounts()[0], namespace)

	tx, err = repo.GetTransaction(ctx, wrongTxid)
	require.Error(t, err)
	require.Nil(t, tx)

	blockHash := testutil.RandomHex(32)
	blockHeight := uint64(testutil.RandomIntInRange(100, 1000))

	done, err = repo.ConfirmTransaction(ctx, txid, blockHash, blockHeight)
	require.NoError(t, err)
	require.True(t, done)

	done, err = repo.ConfirmTransaction(ctx, txid, blockHash, blockHeight)
	require.NoError(t, err)
	require.False(t, done)

	tx, err = repo.GetTransaction(ctx, txid)
	require.NoError(t, err)
	require.NotNil(t, tx)
	require.True(t, tx.IsConfirmed())

	tx, err = repo.GetTransaction(ctx, txid)
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
}

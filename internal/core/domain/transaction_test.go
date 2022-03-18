package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

func TestConfirmTransaction(t *testing.T) {
	tx := &domain.Transaction{}
	require.False(t, tx.IsConfirmed())

	tx.Confirm("fa84eb6806daf1b3c495ed30554d80573a39335b2993b66b3cc1afaa53816e47", 1728312)
	require.True(t, tx.IsConfirmed())
}

func TestAddAccounts(t *testing.T) {
	tx := &domain.Transaction{}
	accounts := tx.GetAccounts()
	require.Empty(t, accounts)

	tx.AddAccount("test1")
	accounts = tx.GetAccounts()
	require.Len(t, accounts, 1)

	tx.AddAccount("test1")
	accounts = tx.GetAccounts()
	require.Len(t, accounts, 1)

	tx.AddAccount("test2")
	accounts = tx.GetAccounts()
	require.Len(t, accounts, 2)
}

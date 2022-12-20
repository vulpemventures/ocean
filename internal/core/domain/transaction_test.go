package domain_test

import (
	"testing"

	"github.com/equitas-foundation/bamp-ocean/internal/core/domain"
	"github.com/stretchr/testify/require"
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

func TestHasAccounts(t *testing.T) {
	tests := []struct {
		t1       *domain.Transaction
		t2       *domain.Transaction
		expected bool
	}{
		{
			&domain.Transaction{Accounts: map[string]struct{}{"test": {}}},
			&domain.Transaction{Accounts: map[string]struct{}{"test": {}}},
			true,
		},
		{
			&domain.Transaction{Accounts: map[string]struct{}{"test": {}}},
			&domain.Transaction{Accounts: map[string]struct{}{"foo": {}}},
			false,
		},
		{
			&domain.Transaction{Accounts: map[string]struct{}{"test": {}}},
			&domain.Transaction{Accounts: map[string]struct{}{"test": {}, "foo": {}}},
			false,
		},
		{
			&domain.Transaction{Accounts: map[string]struct{}{"test": {}, "foo": {}}},
			&domain.Transaction{Accounts: map[string]struct{}{"test": {}}},
			true,
		},
	}

	for _, tt := range tests {
		res := tt.t1.HasAccounts(tt.t2)
		require.Equal(t, tt.expected, res)
	}
}

package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

func TestSpendUtxo(t *testing.T) {
	t.Parallel()

	u := domain.Utxo{}
	require.False(t, u.IsSpent())

	u.Spend()
	require.True(t, u.IsSpent())
}

func TestConfirmUtxo(t *testing.T) {
	t.Parallel()

	u := domain.Utxo{}
	require.False(t, u.IsConfirmed())

	u.Confirm()
	require.True(t, u.IsConfirmed())
}

func TestLockUnlockUtxo(t *testing.T) {
	t.Parallel()

	u := domain.Utxo{}
	require.False(t, u.IsLocked())

	u.Lock(time.Now().Unix())
	require.True(t, u.IsLocked())

	u.Unlock()
	require.False(t, u.IsLocked())
}

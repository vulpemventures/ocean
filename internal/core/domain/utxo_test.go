package domain_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

func TestSpendUtxo(t *testing.T) {
	t.Parallel()

	u := domain.Utxo{}
	require.False(t, u.IsSpent())

	err := u.Spend(domain.UtxoStatus{hex.EncodeToString(make([]byte, 32)), 1, 0, ""})
	require.NoError(t, err)
	require.True(t, u.IsSpent())
}

func TestConfirmUtxo(t *testing.T) {
	t.Parallel()

	u := domain.Utxo{}
	require.False(t, u.IsConfirmed())

	u.Confirm(domain.UtxoStatus{"", 1, 0, ""})
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

package domain

import (
	"context"
)

const (
	UtxoAdded UtxoEventType = iota
	UtxoConfirmed
	UtxoLocked
	UtxoUnlocked
	UtxoSpent
)

var (
	utxoTypeString = map[UtxoEventType]string{
		UtxoAdded:     "UtxoAdded",
		UtxoConfirmed: "UtxoConfirmed",
		UtxoLocked:    "UtxoLocked",
		UtxoUnlocked:  "UtxoUnlocked",
		UtxoSpent:     "UtxoSpent",
	}
)

type UtxoEventType int

func (t UtxoEventType) String() string {
	return utxoTypeString[t]
}

// UtxoEvent holds info about an event occured within the repository.
type UtxoEvent struct {
	EventType UtxoEventType
	Utxos     []UtxoInfo
}

// UtxoRepository is the abstraction for any kind of database intended to
// persist Utxos.
type UtxoRepository interface {
	// AddUtxos adds the provided utxos to the repository by preventing
	// duplicates.
	// Generates a UtxoAdded event if successfull.
	AddUtxos(ctx context.Context, utxos []*Utxo) (int, error)
	// GetUtxosByKey returns the utxos identified by the given keys.
	GetUtxosByKey(ctx context.Context, utxoKeys []UtxoKey) ([]*Utxo, error)
	// GetAllUtxos returns the entire UTXO set, included those locked or
	// already spent.
	GetAllUtxos(ctx context.Context) []*Utxo
	// GetSpendableUtxos returns all unlocked utxo UTXOs.
	GetSpendableUtxos(ctx context.Context) ([]*Utxo, error)
	// GetAllUtxosForAccount returns the list of all utxos for the given
	// account.
	GetAllUtxosForAccount(ctx context.Context, account string) ([]*Utxo, error)
	// GetSpendableUtxosForAccount returns the list of spendable utxos for the
	// given account. The list incldues only confirmed and unlocked utxos.
	GetSpendableUtxosForAccount(ctx context.Context, account string) ([]*Utxo, error)
	// GetLockedUtxosForAccount returns the list of all currently locked utxos
	// for the given account.
	GetLockedUtxosForAccount(ctx context.Context, account string) ([]*Utxo, error)
	// GetBalanceForAccount returns the confirmed, unconfirmed and locked
	// balances per each asset for the given account.
	GetBalanceForAccount(ctx context.Context, account string) (map[string]*Balance, error)
	// SpendUtxos updates the status of the given list of utxos to "spent".
	// Generates a UtxoSpent event if successfull.
	SpendUtxos(ctx context.Context, utxoKeys []UtxoKey, status UtxoStatus) (int, error)
	// ConfirmUtxos updates the status of the given list of utxos to "confirmed".
	// Generates a UtxoConfirmed event if successfull.
	ConfirmUtxos(ctx context.Context, utxoKeys []UtxoKey, status UtxoStatus) (int, error)
	// LockUtxos updates the status of the given list of utxos to "locked".
	// Generates a UtxoLocked event if successfull.
	LockUtxos(ctx context.Context, utxoKeys []UtxoKey, timestamp int64) (int, error)
	// UnlockUtxos updates the status of the given list of utxos to "unlocked".
	// Generates a UtxoUnlocked event if successfull.
	UnlockUtxos(ctx context.Context, utxoKeys []UtxoKey) (int, error)
	// DeleteUtxosForAccount deletes every utxo associated to the given account
	// from the repository.
	DeleteUtxosForAccount(ctx context.Context, accountName string) error
	// GetEventChannel returns the channel of UtxoEvents.
	GetEventChannel() chan UtxoEvent
}

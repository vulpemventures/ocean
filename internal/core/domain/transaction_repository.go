package domain

import "context"

const (
	TransactionAdded TransactionEventType = iota
	TransactionUnconfirmed
	TransactionConfirmed
)

var (
	txTypeString = map[TransactionEventType]string{
		TransactionAdded:       "TransactionAdded",
		TransactionUnconfirmed: "TransactionUnconfirmed",
		TransactionConfirmed:   "TransactionConfirmed",
	}
)

type TransactionEventType int

func (t TransactionEventType) String() string {
	return txTypeString[t]
}

// TransactionEvent holds info about an event occured within the repository.
type TransactionEvent struct {
	EventType   TransactionEventType
	Transaction *Transaction
}

// TransactionRepository is the abstraction for any kind of database intended
// to persist Transactions.
type TransactionRepository interface {
	// AddTransaction adds the provided transaction to the repository by
	// preventing duplicates.
	// Generates a TransactionAdded event if successful.
	AddTransaction(ctx context.Context, tx *Transaction) (bool, error)
	// ConfirmTransaction adds the given blockhash and block height to the
	// Transaction identified by the given txid.
	// Generates a TransactionConfirmed event if successful.
	ConfirmTransaction(
		ctx context.Context, txid, blockHash string, blockheight uint64,
	) (bool, error)
	// GetTransaction returns the Transaction identified by the given txid.
	GetTransaction(ctx context.Context, txid string) (*Transaction, error)
	// UpdateTransaction allows to commit multiple changes to the same
	// Transaction in a transactional way.
	UpdateTransaction(
		ctx context.Context, txid string,
		updateFn func(tx *Transaction) (*Transaction, error),
	) error
	// GetEventChannel retunrs the channel of TransactionEvents.
	GetEventChannel() chan TransactionEvent
}

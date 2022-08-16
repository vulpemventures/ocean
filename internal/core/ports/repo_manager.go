package ports

import (
	"github.com/vulpemventures/ocean/internal/core/domain"
)

type WalletEventHandler func(event domain.WalletEvent)
type UtxoEventHandler func(event domain.UtxoEvent)
type TxEventHandler func(event domain.TransactionEvent)

// RepoManager is the abstraction for any kind of service intended to manage
// domain repositories implementations of the same concrete type.
type RepoManager interface {
	// WalletRepository returns the concrete implentation as domain interface.
	WalletRepository() domain.WalletRepository
	// UtxoRepository returns the concrete implentation as domain interface.
	UtxoRepository() domain.UtxoRepository
	// TransactionRepository returns the concrete implentation as domain interface.
	TransactionRepository() domain.TransactionRepository

	// RegisterHandlerForWalletEvent registers an handler function, executed
	// whenever the given event type occurs.
	RegisterHandlerForWalletEvent(
		eventType domain.WalletEventType, handler WalletEventHandler,
	)
	// RegisterHandlerForUtxoEvent registers an handler function, executed
	// whenever the given event type occurs.
	RegisterHandlerForUtxoEvent(
		eventType domain.UtxoEventType, handler UtxoEventHandler,
	)
	// RegisterHandlerForTxEvent registers an handler function, executed
	// whenever the given event type occurs.
	RegisterHandlerForTxEvent(
		eventType domain.TransactionEventType, handler TxEventHandler,
	)

	// Close closes the connection with all concrete repositories
	// implementations.
	Close()
}

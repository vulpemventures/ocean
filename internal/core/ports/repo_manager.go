package ports

import (
	"github.com/vulpemventures/ocean/internal/core/domain"
)

type WalletEventHandler func(event domain.WalletEvent)
type UtxoEventHandler func(event domain.UtxoEvent)
type TxEventHandler func(event domain.TransactionEvent)
type ScriptEventHandler func(event domain.ExternalScriptEvent)

// RepoManager is the abstraction for any kind of service intended to manage
// domain repositories implementations of the same concrete type.
type RepoManager interface {
	// WalletRepository returns the wallet repository.
	WalletRepository() domain.WalletRepository
	// UtxoRepository returns the utxo repository.
	UtxoRepository() domain.UtxoRepository
	// TransactionRepository returns the tx repository.
	TransactionRepository() domain.TransactionRepository
	// ExternalScriptRepository returns the external scripts repository.
	ExternalScriptRepository() domain.ExternalScriptRepository

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
	// RegisterHandlerForExternalScriptEvent registers an handler function,
	// executed whenever the given event type occurs.
	RegisterHandlerForExternalScriptEvent(
		eventType domain.ExternalScriptEventType, handler ScriptEventHandler,
	)

	// Reset brings all the repos to their initial state by deleting any persisted data.
	Reset()

	// Close closes the connection with all concrete repositories
	// implementations.
	Close()
}

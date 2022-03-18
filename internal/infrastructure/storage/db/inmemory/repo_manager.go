package inmemory

import (
	"sync"

	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
)

type repoManager struct {
	utxoRepository   *utxoRepository
	walletRepository *walletRepository
	txRepository     *txRepository

	walletEventHandlers map[domain.WalletEventType]ports.WalletEventHandler
	utxoEventHandlers   map[domain.UtxoEventType]ports.UtxoEventHandler
	txEventHandlers     map[domain.TransactionEventType]ports.TxEventHandler
	lock                *sync.Mutex
}

func NewRepoManager() ports.RepoManager {
	utxoRepo := newUtxoRepository()
	walletRepo := newWalletRepository()
	txRepo := newTransactionRepository()

	rm := &repoManager{
		utxoRepository:      utxoRepo,
		walletRepository:    walletRepo,
		txRepository:        txRepo,
		lock:                &sync.Mutex{},
		walletEventHandlers: make(map[domain.WalletEventType]ports.WalletEventHandler),
		utxoEventHandlers:   make(map[domain.UtxoEventType]ports.UtxoEventHandler),
		txEventHandlers:     make(map[domain.TransactionEventType]ports.TxEventHandler),
	}

	go rm.listenToWalletEvents()
	go rm.listenToUtxoEvents()
	go rm.listenToTxEvents()

	return rm
}

func (rm *repoManager) UtxoRepository() domain.UtxoRepository {
	return rm.utxoRepository
}

func (rm *repoManager) WalletRepository() domain.WalletRepository {
	return rm.walletRepository
}

func (rm *repoManager) TransactionRepository() domain.TransactionRepository {
	return rm.txRepository
}

func (rm *repoManager) RegisterHandlerForWalletEvent(
	eventType domain.WalletEventType, handler ports.WalletEventHandler,
) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	if _, ok := rm.walletEventHandlers[eventType]; ok {
		return
	}
	rm.walletEventHandlers[eventType] = handler
}

func (rm *repoManager) RegisterHandlerForUtxoEvent(
	eventType domain.UtxoEventType, handler ports.UtxoEventHandler,
) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	if _, ok := rm.utxoEventHandlers[eventType]; ok {
		return
	}
	rm.utxoEventHandlers[eventType] = handler
}

func (rm *repoManager) RegisterHandlerForTxEvent(
	eventType domain.TransactionEventType, handler ports.TxEventHandler,
) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	if _, ok := rm.txEventHandlers[eventType]; ok {
		return
	}
	rm.txEventHandlers[eventType] = handler
}

func (rm *repoManager) listenToWalletEvents() {
	for event := range rm.walletRepository.chEvents {
		if handler, ok := rm.walletEventHandlers[event.EventType]; ok {
			handler(event)
		}
	}
}

func (rm *repoManager) listenToUtxoEvents() {
	for event := range rm.utxoRepository.chEvents {
		if handler, ok := rm.utxoEventHandlers[event.EventType]; ok {
			handler(event)
		}
	}
}

func (rm *repoManager) listenToTxEvents() {
	for event := range rm.txRepository.chEvents {
		if handler, ok := rm.txEventHandlers[event.EventType]; ok {
			handler(event)
		}
	}
}

func (rm *repoManager) Close() {
	rm.walletRepository.close()
	rm.utxoRepository.close()
	rm.txRepository.close()
}

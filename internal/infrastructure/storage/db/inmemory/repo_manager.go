package inmemory

import (
	"sync"
	"time"

	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
)

type repoManager struct {
	utxoRepository   *utxoRepository
	walletRepository *walletRepository
	txRepository     *txRepository

	walletEventHandlers *handlerMap
	utxoEventHandlers   *handlerMap
	txEventHandlers     *handlerMap
}

func NewRepoManager() ports.RepoManager {
	utxoRepo := newUtxoRepository()
	walletRepo := newWalletRepository()
	txRepo := newTransactionRepository()

	rm := &repoManager{
		utxoRepository:      utxoRepo,
		walletRepository:    walletRepo,
		txRepository:        txRepo,
		walletEventHandlers: newHandlerMap(),
		utxoEventHandlers:   newHandlerMap(),
		txEventHandlers:     newHandlerMap(),
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
	rm.walletEventHandlers.set(int(eventType), handler)
}

func (rm *repoManager) RegisterHandlerForUtxoEvent(
	eventType domain.UtxoEventType, handler ports.UtxoEventHandler,
) {
	rm.utxoEventHandlers.set(int(eventType), handler)
}

func (rm *repoManager) RegisterHandlerForTxEvent(
	eventType domain.TransactionEventType, handler ports.TxEventHandler,
) {
	rm.txEventHandlers.set(int(eventType), handler)
}

func (rm *repoManager) listenToWalletEvents() {
	for event := range rm.walletRepository.chEvents {
		time.Sleep(time.Millisecond)

		if handlers, ok := rm.walletEventHandlers.get(int(event.EventType)); ok {
			for i := range handlers {
				handler := handlers[i]
				go handler.(ports.WalletEventHandler)(event)
			}
		}
	}
}

func (rm *repoManager) listenToUtxoEvents() {
	for event := range rm.utxoRepository.chEvents {
		time.Sleep(time.Millisecond)

		if handlers, ok := rm.utxoEventHandlers.get(int(event.EventType)); ok {
			for i := range handlers {
				handler := handlers[i]
				go handler.(ports.UtxoEventHandler)(event)
			}
		}
	}
}

func (rm *repoManager) listenToTxEvents() {
	for event := range rm.txRepository.chEvents {
		time.Sleep(time.Millisecond)

		if handlers, ok := rm.txEventHandlers.get(int(event.EventType)); ok {
			for i := range handlers {
				handler := handlers[i]
				go handler.(ports.TxEventHandler)(event)
			}
		}
	}
}

func (rm *repoManager) Close() {
	rm.walletRepository.close()
	rm.utxoRepository.close()
	rm.txRepository.close()
}

// handlerMap is a util type to prevent race conditions when registering
// or retrieving handlers for events.
type handlerMap struct {
	handlersByEventType map[int][]interface{}
	lock                *sync.RWMutex
}

func newHandlerMap() *handlerMap {
	return &handlerMap{
		handlersByEventType: make(map[int][]interface{}),
		lock:                &sync.RWMutex{},
	}
}

func (m *handlerMap) set(key int, val interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.handlersByEventType[key] = append(m.handlersByEventType[key], val)
}

func (m *handlerMap) get(key int) ([]interface{}, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	val, ok := m.handlersByEventType[key]
	return val, ok
}

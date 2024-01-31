package inmemory

import (
	"context"
	"fmt"
	"sync"

	"github.com/vulpemventures/ocean/internal/core/domain"
)

type txInmemoryStore struct {
	txs  map[string]*domain.Transaction
	lock *sync.RWMutex
}

type txRepository struct {
	store            *txInmemoryStore
	chEvents         chan domain.TransactionEvent
	externalChEvents chan domain.TransactionEvent
	chLock           *sync.Mutex
}

func NewTransactionRepository() domain.TransactionRepository {
	return newTransactionRepository()
}

func newTransactionRepository() *txRepository {
	return &txRepository{
		store: &txInmemoryStore{
			txs:  make(map[string]*domain.Transaction),
			lock: &sync.RWMutex{},
		},
		chEvents:         make(chan domain.TransactionEvent),
		externalChEvents: make(chan domain.TransactionEvent),
		chLock:           &sync.Mutex{},
	}
}

func (r *txRepository) AddTransaction(
	ctx context.Context, tx *domain.Transaction,
) (bool, error) {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	return r.addTx(ctx, tx)
}

func (r *txRepository) ConfirmTransaction(
	ctx context.Context,
	txid, blockHash string, blockheight uint64, blocktime int64,
) (bool, error) {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	return r.confirmTx(ctx, txid, blockHash, blockheight, blocktime)
}

func (r *txRepository) GetTransaction(
	ctx context.Context, txid string,
) (*domain.Transaction, error) {
	r.store.lock.RLock()
	defer r.store.lock.RUnlock()

	return r.getTx(ctx, txid)
}

func (r *txRepository) UpdateTransaction(
	ctx context.Context, txid string,
	updateFn func(tx *domain.Transaction) (*domain.Transaction, error),
) error {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	tx, err := r.getTx(ctx, txid)
	if err != nil {
		return err
	}

	updatedTx, err := updateFn(tx)
	if err != nil {
		return err
	}

	r.store.txs[txid] = updatedTx
	return nil
}

func (r *txRepository) GetEventChannel() chan domain.TransactionEvent {
	return r.externalChEvents
}

func (r *txRepository) addTx(
	_ context.Context, tx *domain.Transaction,
) (bool, error) {
	if _, ok := r.store.txs[tx.TxID]; ok {
		return false, nil
	}

	r.store.txs[tx.TxID] = tx

	go r.publishEvent(domain.TransactionEvent{
		EventType:   domain.TransactionAdded,
		Transaction: tx,
	})

	return true, nil
}

func (r *txRepository) confirmTx(
	ctx context.Context,
	txid string, blockHash string, blockHeight uint64, blocktime int64,
) (bool, error) {
	tx, err := r.getTx(ctx, txid)
	if err != nil {
		return false, nil
	}

	if tx.IsConfirmed() {
		return false, nil
	}

	tx.Confirm(blockHash, blockHeight, blocktime)

	r.store.txs[txid] = tx

	go r.publishEvent(domain.TransactionEvent{
		EventType:   domain.TransactionConfirmed,
		Transaction: tx,
	})

	return true, nil
}

func (r *txRepository) getTx(
	_ context.Context, txid string,
) (*domain.Transaction, error) {
	tx, ok := r.store.txs[txid]
	if !ok {
		return nil, fmt.Errorf("transaction not found")
	}
	return tx, nil
}

func (r *txRepository) publishEvent(event domain.TransactionEvent) {
	r.chLock.Lock()
	defer r.chLock.Unlock()

	r.chEvents <- event
	// send over channel without blocking in case nobody is listening.
	select {
	case r.externalChEvents <- event:
	default:
	}
}

func (r *txRepository) reset() {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	r.store.txs = make(map[string]*domain.Transaction)
}

func (r *txRepository) close() {
	close(r.chEvents)
	close(r.externalChEvents)
}

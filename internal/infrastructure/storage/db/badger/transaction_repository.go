package dbbadger

import (
	"context"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
	"github.com/timshannon/badgerhold/v4"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

type transactionRepository struct {
	store            *badgerhold.Store
	chEvents         chan domain.TransactionEvent
	externalChEvents chan domain.TransactionEvent
	lock             *sync.Mutex

	log func(format string, a ...interface{})
}

func NewTransactionRepository(
	store *badgerhold.Store,
) domain.TransactionRepository {
	return newTransactionRepository(store)
}

func newTransactionRepository(
	store *badgerhold.Store,
) *transactionRepository {
	chEvents := make(chan domain.TransactionEvent)
	extrernalChEvents := make(chan domain.TransactionEvent)
	lock := &sync.Mutex{}
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("transaction repository: %s", format)
		log.Debugf(format, a...)
	}
	return &transactionRepository{
		store, chEvents, extrernalChEvents, lock, logFn,
	}
}

func (r *transactionRepository) AddTransaction(
	ctx context.Context, tx *domain.Transaction,
) (bool, error) {
	done, err := r.insertTx(ctx, tx)
	if done {
		go r.publishEvent(domain.TransactionEvent{
			EventType:   domain.TransactionAdded,
			Transaction: tx,
		})
	}
	return done, err
}

func (r *transactionRepository) ConfirmTransaction(
	ctx context.Context, txid, blockHash string, blockheight uint64,
) (bool, error) {
	tx, err := r.getTx(ctx, txid)
	if err != nil {
		return false, err
	}

	if tx.IsConfirmed() {
		return false, nil
	}

	tx.Confirm(blockHash, blockheight)

	if err := r.updateTx(ctx, *tx); err != nil {
		return false, err
	}

	go r.publishEvent(domain.TransactionEvent{
		EventType:   domain.TransactionConfirmed,
		Transaction: tx,
	})

	return true, nil
}

func (r *transactionRepository) GetTransaction(
	ctx context.Context, txid string,
) (*domain.Transaction, error) {
	return r.getTx(ctx, txid)
}

func (r *transactionRepository) UpdateTransaction(
	ctx context.Context, txid string,
	updateFn func(*domain.Transaction) (*domain.Transaction, error),
) error {
	tx, err := r.getTx(ctx, txid)
	if err != nil {
		return err
	}

	updatedTx, err := updateFn(tx)
	if err != nil {
		return err
	}

	return r.updateTx(ctx, *updatedTx)
}

func (r *transactionRepository) GetEventChannel() chan domain.TransactionEvent {
	return r.externalChEvents
}

func (r *transactionRepository) insertTx(
	ctx context.Context, tx *domain.Transaction,
) (bool, error) {
	var err error
	if ctx.Value("tx") != nil {
		t := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxInsert(t, tx.TxID, tx)
	} else {
		err = r.store.Insert(tx.TxID, tx)
	}

	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *transactionRepository) getTx(
	ctx context.Context, txid string,
) (*domain.Transaction, error) {
	var err error
	var tx domain.Transaction

	if ctx.Value("tx") != nil {
		t := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxGet(t, txid, &tx)
	} else {
		err = r.store.Get(txid, &tx)
	}

	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, err
	}

	return &tx, nil
}

func (r *transactionRepository) updateTx(
	ctx context.Context, tx domain.Transaction,
) error {
	if ctx.Value("tx") != nil {
		t := ctx.Value("tx").(*badger.Txn)
		return r.store.TxUpdate(t, tx.TxID, tx)
	}
	return r.store.Update(tx.TxID, tx)
}

func (r *transactionRepository) publishEvent(event domain.TransactionEvent) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.log("publish event %s", event.EventType)
	r.chEvents <- event

	// send over channel without blocking in case nobody is listening.
	select {
	case r.externalChEvents <- event:
	default:
	}
}

func (r *transactionRepository) close() {
	r.store.Close()
	close(r.chEvents)
	close(r.externalChEvents)
}

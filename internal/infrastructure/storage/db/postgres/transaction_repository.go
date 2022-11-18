package postgresdb

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres/sqlc/queries"
	"sync"
)

var (
	ErrTxNotFound = errors.New("transaction not found")
)

type txRepositoryPg struct {
	pgxPool          *pgxpool.Pool
	querier          *queries.Queries
	chLock           *sync.Mutex
	chEvents         chan domain.TransactionEvent
	externalChEvents chan domain.TransactionEvent
}

func NewTxRepositoryPgImpl(pgxPool *pgxpool.Pool) domain.TransactionRepository {
	return &txRepositoryPg{
		pgxPool:          pgxPool,
		querier:          queries.New(pgxPool),
		chLock:           &sync.Mutex{},
		chEvents:         make(chan domain.TransactionEvent),
		externalChEvents: make(chan domain.TransactionEvent),
	}
}

func (t *txRepositoryPg) AddTransaction(
	ctx context.Context,
	trx *domain.Transaction,
) (bool, error) {
	conn, err := t.pgxPool.Acquire(ctx)
	if err != nil {
		return false, err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	querierWithTx := t.querier.WithTx(tx)

	txPg, err := querierWithTx.InsertTransaction(ctx, queries.InsertTransactionParams{
		TxID:        trx.TxID,
		TxHex:       trx.TxHex,
		BlockHash:   trx.BlockHash,
		BlockHeight: int32(trx.BlockHeight),
	})
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); pqErr != nil && ok && pqErr.Code == uniqueViolation {
			return false, nil
		} else {
			return false, err
		}
	}

	for k := range trx.Accounts {
		if _, err := querierWithTx.InsertTransactionInputAccount(ctx, queries.InsertTransactionInputAccountParams{
			AccountName: k,
			FkTxID:      txPg.TxID,
		}); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	go t.publishEvent(domain.TransactionEvent{
		EventType:   domain.TransactionAdded,
		Transaction: trx,
	})

	return true, nil
}

func (t *txRepositoryPg) ConfirmTransaction(
	ctx context.Context,
	txid string,
	blockHash string,
	blockHeight uint64,
) (bool, error) {
	tx, err := t.getTx(ctx, txid)
	if err != nil {
		return false, err
	}

	if tx.IsConfirmed() {
		return false, nil
	}

	tx.Confirm(blockHash, blockHeight)

	if err := t.updateTx(ctx, t.querier, *tx); err != nil {
		return false, err
	}

	go t.publishEvent(domain.TransactionEvent{
		EventType:   domain.TransactionConfirmed,
		Transaction: tx,
	})

	return true, nil
}

func (t *txRepositoryPg) GetTransaction(
	ctx context.Context,
	txid string,
) (*domain.Transaction, error) {
	return t.getTx(ctx, txid)
}

func (t *txRepositoryPg) UpdateTransaction(
	ctx context.Context,
	txid string,
	updateFn func(tx *domain.Transaction) (*domain.Transaction, error),
) error {
	tx, err := t.getTx(ctx, txid)
	if err != nil {
		return err
	}

	updatedTx, err := updateFn(tx)
	if err != nil {
		return err
	}

	return t.updateTx(ctx, t.querier, *updatedTx)
}

func (t *txRepositoryPg) GetEventChannel() chan domain.TransactionEvent {
	return t.externalChEvents
}

func (t *txRepositoryPg) publishEvent(event domain.TransactionEvent) {
	t.chLock.Lock()
	defer t.chLock.Unlock()

	t.chEvents <- event
	// send over channel without blocking in case nobody is listening.
	select {
	case t.externalChEvents <- event:
	default:
	}
}

func (t *txRepositoryPg) close() {
	close(t.chEvents)
	close(t.externalChEvents)
}

func (t *txRepositoryPg) updateTx(
	ctx context.Context,
	querier *queries.Queries,
	trx domain.Transaction,
) error {
	if _, err := querier.UpdateTransaction(ctx, queries.UpdateTransactionParams{
		TxHex:       trx.TxHex,
		BlockHash:   trx.BlockHash,
		BlockHeight: int32(trx.BlockHeight),
		TxID:        trx.TxID,
	}); err != nil {
		return err
	}

	if err := querier.DeleteTransactionInputAccounts(ctx, trx.TxID); err != nil {
		return err
	}

	for k := range trx.Accounts {
		if _, err := querier.InsertTransactionInputAccount(ctx, queries.InsertTransactionInputAccountParams{
			AccountName: k,
			FkTxID:      trx.TxID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (t *txRepositoryPg) getTx(
	ctx context.Context, txid string,
) (*domain.Transaction, error) {
	tx, err := t.querier.GetTransaction(ctx, txid)
	if err != nil {
		return nil, err
	}

	if len(tx) == 0 {
		return nil, ErrTxNotFound
	}

	accounts := make(map[string]struct{})
	for _, v := range tx {
		if v.AccountName.Valid {
			accounts[v.AccountName.String] = struct{}{}
		}
	}

	return &domain.Transaction{
		TxID:        tx[0].TxID,
		TxHex:       tx[0].TxHex,
		BlockHash:   tx[0].BlockHash,
		BlockHeight: uint64(tx[0].BlockHeight),
		Accounts:    accounts,
	}, nil
}

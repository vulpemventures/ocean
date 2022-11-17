package postgresdb

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres/sqlc/queries"
	"sync"
)

type txRepositoryPg struct {
	querier          *queries.Queries
	chLock           *sync.Mutex
	chEvents         chan domain.TransactionEvent
	externalChEvents chan domain.TransactionEvent
}

func NewTxRepositoryPgImpl(pgxPool *pgxpool.Pool) domain.TransactionRepository {
	return &txRepositoryPg{
		querier:          queries.New(pgxPool),
		chLock:           &sync.Mutex{},
		chEvents:         make(chan domain.TransactionEvent),
		externalChEvents: make(chan domain.TransactionEvent),
	}
}

func (t *txRepositoryPg) AddTransaction(
	ctx context.Context,
	tx *domain.Transaction,
) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t *txRepositoryPg) ConfirmTransaction(
	ctx context.Context,
	txid string,
	blockHash string,
	blockHeight uint64,
) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t *txRepositoryPg) GetTransaction(
	ctx context.Context,
	txid string,
) (*domain.Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (t *txRepositoryPg) UpdateTransaction(
	ctx context.Context,
	txid string,
	updateFn func(tx *domain.Transaction) (*domain.Transaction, error),
) error {
	//TODO implement me
	panic("implement me")
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

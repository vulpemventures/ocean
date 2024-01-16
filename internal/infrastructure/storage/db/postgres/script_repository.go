package postgresdb

import (
	"context"
	"sync"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres/sqlc/queries"
)

type scriptRepositoryPg struct {
	pgxPool  *pgxpool.Pool
	querier  *queries.Queries
	chLock   *sync.Mutex
	chEvents chan domain.ExternalScriptEvent
}

func NewExternalScriptRepositoryPgImpl(pgxPool *pgxpool.Pool) domain.ExternalScriptRepository {
	return newExternalScriptRepositoryPgImpl(pgxPool)
}

func newExternalScriptRepositoryPgImpl(pgxPool *pgxpool.Pool) *scriptRepositoryPg {
	return &scriptRepositoryPg{
		pgxPool:  pgxPool,
		querier:  queries.New(pgxPool),
		chLock:   &sync.Mutex{},
		chEvents: make(chan domain.ExternalScriptEvent),
	}
}

func (r *scriptRepositoryPg) AddScript(
	ctx context.Context, info domain.AddressInfo,
) (bool, error) {
	conn, err := r.pgxPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	querierWithTx := r.querier.WithTx(tx)
	if err := querierWithTx.InsertScript(ctx, queries.InsertScriptParams{
		Account:     info.Account,
		Script:      info.Script,
		BlindingKey: info.BlindingKey,
	}); err != nil {
		if pqErr, ok := err.(*pgconn.PgError); pqErr != nil && ok && pqErr.Code == uniqueViolation {
			return false, nil
		} else {
			return false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	go r.publishEvent(domain.ExternalScriptEvent{
		EventType: domain.ExternalScriptAdded,
		Info:      info,
	})

	return true, nil
}

func (r *scriptRepositoryPg) GetAllScripts(
	ctx context.Context,
) ([]domain.AddressInfo, error) {
	return r.getScripts(ctx)
}

func (r *scriptRepositoryPg) DeleteScript(
	ctx context.Context, scriptHash string,
) (bool, error) {
	conn, err := r.pgxPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	querierWithTx := r.querier.WithTx(tx)
	if err := querierWithTx.DeleteScript(ctx, scriptHash); err != nil {
		if pqErr, ok := err.(*pgconn.PgError); pqErr != nil && ok && pqErr.Code == uniqueViolation {
			return false, nil
		} else {
			return false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	go r.publishEvent(domain.ExternalScriptEvent{
		EventType: domain.ExternalScriptDeleted,
		Info:      domain.AddressInfo{Account: scriptHash},
	})

	return true, nil
}

func (r *scriptRepositoryPg) publishEvent(event domain.ExternalScriptEvent) {
	r.chLock.Lock()
	defer r.chLock.Unlock()

	r.chEvents <- event
}

func (r *scriptRepositoryPg) close() {}

func (r *scriptRepositoryPg) getScripts(
	ctx context.Context,
) ([]domain.AddressInfo, error) {
	rows, err := r.querier.GetAllScripts(ctx)
	if err != nil {
		return nil, err
	}

	scripts := make([]domain.AddressInfo, 0, len(rows))
	for _, r := range rows {
		scripts = append(scripts, domain.AddressInfo{
			Account:     r.Account,
			Script:      r.Script,
			BlindingKey: r.BlindingKey,
		})
	}

	return scripts, nil
}

func (r *scriptRepositoryPg) reset(
	querier *queries.Queries, ctx context.Context,
) {
	querier.ResetScripts(ctx)
}

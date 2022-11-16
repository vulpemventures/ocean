package postgresdb

import (
	"context"
	"database/sql"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"sync"
)

type utxoRepositoryPg struct {
	db               *sql.DB
	chLock           *sync.Mutex
	chEvents         chan domain.UtxoEvent
	externalChEvents chan domain.UtxoEvent
}

func NewUtxoRepositoryPgImpl(db *sql.DB) domain.UtxoRepository {
	return &utxoRepositoryPg{
		db:               db,
		chLock:           &sync.Mutex{},
		chEvents:         make(chan domain.UtxoEvent),
		externalChEvents: make(chan domain.UtxoEvent),
	}
}

func (u *utxoRepositoryPg) AddUtxos(
	ctx context.Context,
	utxos []*domain.Utxo,
) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) GetUtxosByKey(
	ctx context.Context,
	utxoKeys []domain.UtxoKey,
) ([]*domain.Utxo, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) GetAllUtxos(ctx context.Context) []*domain.Utxo {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) GetSpendableUtxos(
	ctx context.Context,
) ([]*domain.Utxo, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) GetAllUtxosForAccount(
	ctx context.Context,
	account string,
) ([]*domain.Utxo, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) GetSpendableUtxosForAccount(
	ctx context.Context,
	account string,
) ([]*domain.Utxo, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) GetLockedUtxosForAccount(
	ctx context.Context,
	account string,
) ([]*domain.Utxo, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) GetBalanceForAccount(
	ctx context.Context,
	account string,
) (map[string]*domain.Balance, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) SpendUtxos(
	ctx context.Context,
	utxoKeys []domain.UtxoKey,
	status domain.UtxoStatus,
) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) ConfirmUtxos(
	ctx context.Context,
	utxoKeys []domain.UtxoKey,
	status domain.UtxoStatus,
) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) LockUtxos(
	ctx context.Context,
	utxoKeys []domain.UtxoKey,
	timestamp int64,
) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) UnlockUtxos(
	ctx context.Context,
	utxoKeys []domain.UtxoKey,
) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) DeleteUtxosForAccount(
	ctx context.Context,
	accountName string,
) error {
	//TODO implement me
	panic("implement me")
}

func (u *utxoRepositoryPg) GetEventChannel() chan domain.UtxoEvent {
	return u.externalChEvents
}

func (u *utxoRepositoryPg) publishEvent(event domain.UtxoEvent) {
	u.chLock.Lock()
	defer u.chLock.Unlock()

	u.chEvents <- event
	// send over channel without blocking in case nobody is listening.
	select {
	case u.externalChEvents <- event:
	default:
	}
}

func (u *utxoRepositoryPg) close() {
	close(u.chEvents)
	close(u.externalChEvents)
}

package postgresdb

import (
	"context"
	"database/sql"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"sync"
)

type walletRepositoryPg struct {
	db               *sql.DB
	chLock           *sync.Mutex
	chEvents         chan domain.WalletEvent
	externalChEvents chan domain.WalletEvent
}

func NewWalletRepositoryPgImpl(db *sql.DB) domain.WalletRepository {
	return &walletRepositoryPg{
		db:               db,
		chLock:           &sync.Mutex{},
		chEvents:         make(chan domain.WalletEvent),
		externalChEvents: make(chan domain.WalletEvent),
	}
}

func (w *walletRepositoryPg) CreateWallet(
	ctx context.Context,
	wallet *domain.Wallet,
) error {
	//TODO implement me
	panic("implement me")
}

func (w *walletRepositoryPg) GetWallet(
	ctx context.Context,
) (*domain.Wallet, error) {
	//TODO implement me
	panic("implement me")
}

func (w *walletRepositoryPg) UnlockWallet(
	ctx context.Context,
	password string,
) error {
	//TODO implement me
	panic("implement me")
}

func (w *walletRepositoryPg) LockWallet(
	ctx context.Context,
	password string,
) error {
	//TODO implement me
	panic("implement me")
}

func (w *walletRepositoryPg) UpdateWallet(
	ctx context.Context,
	updateFn func(v *domain.Wallet) (*domain.Wallet, error),
) error {
	//TODO implement me
	panic("implement me")
}

func (w *walletRepositoryPg) CreateAccount(
	ctx context.Context,
	accountName string,
	birthdayBlock uint32,
) (*domain.AccountInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (w *walletRepositoryPg) DeriveNextExternalAddressesForAccount(
	ctx context.Context,
	accountName string,
	numOfAddresses uint64,
) ([]domain.AddressInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (w *walletRepositoryPg) DeriveNextInternalAddressesForAccount(
	ctx context.Context,
	accountName string,
	numOfAddresses uint64,
) ([]domain.AddressInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (w *walletRepositoryPg) DeleteAccount(
	ctx context.Context,
	accountName string,
) error {
	//TODO implement me
	panic("implement me")
}

func (w *walletRepositoryPg) GetEventChannel() chan domain.WalletEvent {
	return w.externalChEvents
}

func (w *walletRepositoryPg) publishEvent(event domain.WalletEvent) {
	w.chLock.Lock()
	defer w.chLock.Unlock()

	w.chEvents <- event
	// send over channel without blocking in case nobody is listening.
	select {
	case w.externalChEvents <- event:
	default:
	}
}

func (w *walletRepositoryPg) close() {
	close(w.chEvents)
	close(w.externalChEvents)
}

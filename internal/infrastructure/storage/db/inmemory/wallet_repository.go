package inmemory

import (
	"context"
	"fmt"
	"sync"

	"github.com/vulpemventures/ocean/internal/core/domain"
)

var (
	ErrWalletAlreadyExisting = fmt.Errorf("wallet already existing")
)

type walletInmemoryStore struct {
	wallet *domain.Wallet
	lock   *sync.RWMutex
}

type walletRepository struct {
	store            *walletInmemoryStore
	chEvents         chan domain.WalletEvent
	externalChEvents chan domain.WalletEvent
	chLock           *sync.Mutex
}

func NewWalletRepository() domain.WalletRepository {
	return newWalletRepository()
}

func newWalletRepository() *walletRepository {
	return &walletRepository{
		store: &walletInmemoryStore{
			lock: &sync.RWMutex{},
		},
		chEvents:         make(chan domain.WalletEvent),
		externalChEvents: make(chan domain.WalletEvent),
		chLock:           &sync.Mutex{},
	}
}

func (r *walletRepository) CreateWallet(
	ctx context.Context, wallet *domain.Wallet,
) error {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	if r.store.wallet != nil {
		return ErrWalletAlreadyExisting
	}

	r.store.wallet = wallet

	go r.publishEvent(domain.WalletEvent{
		EventType: domain.WalletCreated,
	})

	return nil
}

func (r *walletRepository) GetWallet(ctx context.Context) (*domain.Wallet, error) {
	r.store.lock.RLock()
	defer r.store.lock.RUnlock()

	if r.store.wallet == nil {
		return nil, fmt.Errorf("wallet is not initialized")
	}
	return r.store.wallet, nil
}

func (r *walletRepository) UnlockWallet(
	ctx context.Context, password string,
) error {
	if err := r.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			if err := w.Unlock(password); err != nil {
				return nil, err
			}
			return w, nil
		},
	); err != nil {
		return err
	}

	go r.publishEvent(domain.WalletEvent{
		EventType: domain.WalletUnlocked,
	})

	return nil
}

func (r *walletRepository) LockWallet(
	ctx context.Context, password string,
) error {
	if err := r.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			if err := w.Lock(password); err != nil {
				return nil, err
			}
			return w, nil
		},
	); err != nil {
		return err
	}

	go r.publishEvent(domain.WalletEvent{
		EventType: domain.WalletLocked,
	})

	return nil
}

func (r *walletRepository) UpdateWallet(
	ctx context.Context, updateFn func(*domain.Wallet) (*domain.Wallet, error),
) error {
	wallet, err := r.GetWallet(ctx)
	if err != nil {
		return err
	}

	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	updatedWallet, err := updateFn(wallet)
	if err != nil {
		return err
	}

	r.store.wallet = updatedWallet
	return nil
}

func (r *walletRepository) CreateAccount(
	ctx context.Context,
	label string,
	birthdayBlock uint32,
) (*domain.AccountInfo, error) {
	var accountInfo *domain.AccountInfo

	if err := r.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			account, err := w.CreateAccount(label, birthdayBlock)
			if err != nil {
				return nil, err
			}

			accountInfo = &account.Info
			return w, nil
		},
	); err != nil {
		return nil, err
	}

	go r.publishEvent(domain.WalletEvent{
		EventType:        domain.WalletAccountCreated,
		AccountNamespace: accountInfo.Namespace,
	})

	return accountInfo, nil
}

func (r *walletRepository) DeriveNextExternalAddressesForAccount(
	ctx context.Context, namespace string, numOfAddresses uint64,
) ([]domain.AddressInfo, error) {
	addressesInfo := make([]domain.AddressInfo, 0)

	if err := r.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			for i := 0; i < int(numOfAddresses); i++ {
				addrInfo, err := w.DeriveNextExternalAddressForAccount(namespace)
				if err != nil {
					return nil, err
				}
				addressesInfo = append(addressesInfo, *addrInfo)
			}
			return w, nil
		},
	); err != nil {
		return nil, err
	}

	go r.publishEvent(domain.WalletEvent{
		EventType:        domain.WalletAccountAddressesDerived,
		AccountNamespace: namespace,
		AccountAddresses: addressesInfo,
	})

	return addressesInfo, nil
}

func (r *walletRepository) DeriveNextInternalAddressesForAccount(
	ctx context.Context, namespace string, numOfAddresses uint64,
) ([]domain.AddressInfo, error) {
	addressesInfo := make([]domain.AddressInfo, 0)

	if err := r.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			for i := 0; i < int(numOfAddresses); i++ {
				addrInfo, err := w.DeriveNextInternalAddressForAccount(namespace)
				if err != nil {
					return nil, err
				}
				addressesInfo = append(addressesInfo, *addrInfo)
			}
			return w, nil
		},
	); err != nil {
		return nil, err
	}

	go r.publishEvent(domain.WalletEvent{
		EventType:        domain.WalletAccountAddressesDerived,
		AccountNamespace: namespace,
		AccountAddresses: addressesInfo,
	})

	return addressesInfo, nil
}

func (r *walletRepository) DeleteAccount(ctx context.Context, namespace string) error {
	if err := r.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			if err := w.DeleteAccount(namespace); err != nil {
				return nil, err
			}
			return w, nil
		},
	); err != nil {
		return err
	}

	go r.publishEvent(domain.WalletEvent{
		EventType:        domain.WalletAccountDeleted,
		AccountNamespace: namespace,
	})

	return nil
}

func (r *walletRepository) GetEventChannel() chan domain.WalletEvent {
	return r.externalChEvents
}

func (r *walletRepository) publishEvent(event domain.WalletEvent) {
	r.chLock.Lock()
	defer r.chLock.Unlock()

	r.chEvents <- event
	// send over channel without blocking in case nobody is listening.
	select {
	case r.externalChEvents <- event:
	default:
	}
}

func (r *walletRepository) close() {
	close(r.chEvents)
	close(r.externalChEvents)
}

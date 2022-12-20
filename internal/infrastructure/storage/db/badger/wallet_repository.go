package dbbadger

import (
	"context"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/equitas-foundation/bamp-ocean/internal/core/domain"
	log "github.com/sirupsen/logrus"
	"github.com/timshannon/badgerhold/v4"
)

const (
	//since there can be only 1 wallet in database,
	//key is hardcoded for easier retrival
	walletKey = "wallet"
)

type walletRepository struct {
	store            *badgerhold.Store
	chEvents         chan domain.WalletEvent
	externalChEvents chan domain.WalletEvent
	lock             *sync.Mutex

	log func(format string, a ...interface{})
}

func NewWalletRepository(store *badgerhold.Store) domain.WalletRepository {
	return newWalletRepository(store)
}

func newWalletRepository(store *badgerhold.Store) *walletRepository {
	chEvents := make(chan domain.WalletEvent, 10)
	extrernalChEvents := make(chan domain.WalletEvent, 10)
	lock := &sync.Mutex{}
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("wallet repository: %s", format)
		log.Debugf(format, a...)
	}
	return &walletRepository{store, chEvents, extrernalChEvents, lock, logFn}
}

func (r *walletRepository) CreateWallet(
	ctx context.Context, wallet *domain.Wallet,
) error {
	if err := r.insertWallet(ctx, wallet); err != nil {
		return err
	}

	go r.publishEvent(domain.WalletEvent{
		EventType: domain.WalletCreated,
	})

	return nil
}

func (r *walletRepository) GetWallet(
	ctx context.Context,
) (*domain.Wallet, error) {
	return r.getWallet(ctx)
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
	ctx context.Context, updateFn func(v *domain.Wallet) (*domain.Wallet, error),
) error {
	wallet, err := r.getWallet(ctx)
	if err != nil {
		return err
	}

	updatedWallet, err := updateFn(wallet)
	if err != nil {
		return err
	}

	return r.updateWallet(ctx, updatedWallet)
}

func (r *walletRepository) CreateAccount(
	ctx context.Context, accountName, xpub string, birthdayBlock uint32,
) (*domain.AccountInfo, error) {
	var accountInfo *domain.AccountInfo
	if err := r.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			var account *domain.Account
			var err error
			if len(xpub) > 0 {
				account, err = w.CreateMSAccount(accountName, xpub, birthdayBlock)
			} else {
				account, err = w.CreateAccount(accountName, birthdayBlock)
			}
			if err != nil {
				return nil, err
			}
			if account == nil {
				return nil, fmt.Errorf("account %s already existing", accountName)
			}
			accountInfo = &account.Info
			return w, nil
		},
	); err != nil {
		return nil, err
	}

	go r.publishEvent(domain.WalletEvent{
		EventType:            domain.WalletAccountCreated,
		AccountName:          accountName,
		AccountBirthdayBlock: birthdayBlock,
	})

	return accountInfo, nil
}

func (r *walletRepository) DeriveNextExternalAddressesForAccount(
	ctx context.Context, accountName string, numOfAddress uint64,
) ([]domain.AddressInfo, error) {
	if numOfAddress == 0 {
		numOfAddress = 1
	}

	var addressesInfo []domain.AddressInfo
	if err := r.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			for i := 0; i < int(numOfAddress); i++ {
				addrInfo, err := w.DeriveNextExternalAddressForAccount(accountName)
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
		AccountName:      accountName,
		AccountAddresses: addressesInfo,
	})

	return addressesInfo, nil
}

func (r *walletRepository) DeriveNextInternalAddressesForAccount(
	ctx context.Context, accountName string, numOfAddress uint64,
) ([]domain.AddressInfo, error) {
	if numOfAddress == 0 {
		numOfAddress = 1
	}

	var addressesInfo []domain.AddressInfo
	if err := r.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			for i := 0; i < int(numOfAddress); i++ {
				addrInfo, err := w.DeriveNextInternalAddressForAccount(accountName)
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
		AccountName:      accountName,
		AccountAddresses: addressesInfo,
	})

	return addressesInfo, nil
}

func (r *walletRepository) DeleteAccount(
	ctx context.Context, accountName string,
) error {
	if err := r.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			if err := w.DeleteAccount(accountName); err != nil {
				return nil, err
			}
			return w, nil
		},
	); err != nil {
		return err
	}

	go r.publishEvent(domain.WalletEvent{
		EventType:   domain.WalletAccountDeleted,
		AccountName: accountName,
	})

	return nil
}

func (r *walletRepository) GetEventChannel() chan domain.WalletEvent {
	return r.externalChEvents
}

func (r *walletRepository) insertWallet(
	ctx context.Context, wallet *domain.Wallet,
) error {
	var err error

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxInsert(tx, walletKey, *wallet)
	} else {
		err = r.store.Insert(walletKey, *wallet)
	}
	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return fmt.Errorf("wallet is already initialized")
		}
		return err
	}

	return nil
}

func (r *walletRepository) getWallet(ctx context.Context) (*domain.Wallet, error) {
	var err error
	var wallet domain.Wallet

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxGet(tx, walletKey, &wallet)
	} else {
		err = r.store.Get(walletKey, &wallet)
	}

	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, fmt.Errorf("wallet is not initialized")
		}
		return nil, err
	}

	return &wallet, nil
}

func (r *walletRepository) updateWallet(
	ctx context.Context, wallet *domain.Wallet,
) error {
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		return r.store.TxUpdate(tx, walletKey, *wallet)
	}
	return r.store.Update(walletKey, *wallet)
}

func (r *walletRepository) publishEvent(event domain.WalletEvent) {
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

func (r *walletRepository) close() {
	r.store.Close()
	close(r.chEvents)
	close(r.externalChEvents)
}

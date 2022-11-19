package postgresdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres/sqlc/queries"
	"sync"
)

const (
	//since there can be only 1 wallet in database,
	//key is hardcoded for easier retrival
	walletKey = "wallet"
	//uniqueViolation is a postgres error code for unique constraint violation
	uniqueViolation = "23505"
	pgxNoRows       = "no rows in result set"
)

var (
	ErrorWalletNotFound     = errors.New("wallet not found")
	ErrWalletAlreadyCreated = errors.New("wallet already created")
	ErrAccountNotFound      = errors.New("account not found")
)

type walletRepositoryPg struct {
	pgxPool          *pgxpool.Pool
	querier          *queries.Queries
	chLock           *sync.Mutex
	chEvents         chan domain.WalletEvent
	externalChEvents chan domain.WalletEvent
}

func NewWalletRepositoryPgImpl(pgxPool *pgxpool.Pool) domain.WalletRepository {
	return &walletRepositoryPg{
		pgxPool:          pgxPool,
		querier:          queries.New(pgxPool),
		chLock:           &sync.Mutex{},
		chEvents:         make(chan domain.WalletEvent),
		externalChEvents: make(chan domain.WalletEvent),
	}
}

func (w *walletRepositoryPg) CreateWallet(
	ctx context.Context,
	wallet *domain.Wallet,
) error {
	if _, err := w.querier.InsertWallet(ctx, queries.InsertWalletParams{
		ID:                  walletKey,
		EncryptedMnemonic:   wallet.EncryptedMnemonic,
		PasswordHash:        wallet.PasswordHash,
		BirthdayBlockHeight: int32(wallet.BirthdayBlockHeight),
		RootPath:            wallet.RootPath,
		NetworkName:         wallet.NetworkName,
		NextAccountIndex:    int32(wallet.NextAccountIndex),
	},
	); err != nil {
		if pqErr, ok := err.(*pgconn.PgError); pqErr != nil && ok && pqErr.Code == uniqueViolation {
			return ErrWalletAlreadyCreated
		} else {
			return err
		}
	}

	go w.publishEvent(domain.WalletEvent{
		EventType: domain.WalletCreated,
	})

	return nil
}

func (w *walletRepositoryPg) GetWallet(
	ctx context.Context,
) (*domain.Wallet, error) {
	return w.getWallet(ctx)
}

func (w *walletRepositoryPg) UnlockWallet(
	ctx context.Context,
	password string,
) error {
	wallet, err := w.getWallet(ctx)
	if err != nil {
		return err
	}

	if err := wallet.Unlock(password); err != nil {
		return err
	}

	go w.publishEvent(domain.WalletEvent{
		EventType: domain.WalletUnlocked,
	})

	return nil
}

func (w *walletRepositoryPg) LockWallet(
	ctx context.Context,
	password string,
) error {
	wallet, err := w.getWallet(ctx)
	if err != nil {
		return err
	}

	if err := wallet.Lock(password); err != nil {
		return err
	}

	go w.publishEvent(domain.WalletEvent{
		EventType: domain.WalletLocked,
	})

	return nil
}

// UpdateWallet updates 3 tables in database: wallet, account, account_script_info
func (w *walletRepositoryPg) UpdateWallet(
	ctx context.Context,
	updateFn func(v *domain.Wallet) (*domain.Wallet, error),
) error {
	wallet, err := w.getWallet(ctx)
	if err != nil {
		return err
	}

	updatedWallet, err := updateFn(wallet)
	if err != nil {
		return err
	}

	conn, err := w.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	querierWithTx := w.querier.WithTx(tx)

	//update wallet table
	if _, err := querierWithTx.UpdateWallet(
		ctx,
		queries.UpdateWalletParams{
			ID:                  walletKey,
			EncryptedMnemonic:   updatedWallet.EncryptedMnemonic,
			PasswordHash:        updatedWallet.PasswordHash,
			BirthdayBlockHeight: int32(updatedWallet.BirthdayBlockHeight),
			RootPath:            updatedWallet.RootPath,
			NetworkName:         updatedWallet.NetworkName,
			NextAccountIndex:    int32(updatedWallet.NextAccountIndex),
		},
	); err != nil {
		return err
	}

	// loop over accounts and update account table if it is existing
	// or insert new account if it is not existing
	// insert account scripts as well
	for _, account := range updatedWallet.AccountsByKey {
		newAccount := false
		_, err := querierWithTx.GetAccount(ctx, account.Info.Key.Name)
		if err != nil {
			if err == sql.ErrNoRows {
				newAccount = true
			} else {
				return err
			}
		}

		if newAccount {
			if _, err := querierWithTx.InsertAccount(ctx, queries.InsertAccountParams{
				Name:              account.Info.Key.Name,
				Index:             int32(account.Info.Key.Index),
				Xpub:              account.Info.Xpub,
				DerivationPath:    account.Info.DerivationPath,
				NextExternalIndex: int32(account.NextExternalIndex),
				NextInternalIndex: int32(account.NextInternalIndex),
				FkWalletID:        walletKey,
			}); err != nil {
				return err
			}
		} else {
			if _, err := querierWithTx.UpdateAccountIndexes(
				ctx,
				queries.UpdateAccountIndexesParams{
					NextExternalIndex: int32(account.NextExternalIndex),
					NextInternalIndex: int32(account.NextInternalIndex),
					Name:              account.Info.Key.Name,
				},
			); err != nil {
				return err
			}
		}

		req := make([]queries.InsertAccountScriptsParams, 0)
		for k, v := range account.DerivationPathByScript {
			req = append(req, queries.InsertAccountScriptsParams{
				Script:         k,
				DerivationPath: v,
				FkAccountName:  account.Info.Key.Name,
			})
		}
		if len(req) > 0 {
			if err := querierWithTx.DeleteAccountScripts(ctx, account.Info.Key.Name); err != nil {
				return err
			}

			if _, err := querierWithTx.InsertAccountScripts(
				ctx,
				req,
			); err != nil {
				if pqErr, ok := err.(*pgconn.PgError); pqErr != nil && ok && pqErr.Code == uniqueViolation {
					continue
				} else {
					return err
				}
			}
		}
	}

	return tx.Commit(ctx)
}

func (w *walletRepositoryPg) CreateAccount(
	ctx context.Context,
	accountName string,
	birthdayBlock uint32,
) (*domain.AccountInfo, error) {
	wallet, err := w.getWallet(ctx)
	if err != nil {
		return nil, err
	}

	account, err := wallet.CreateAccount(accountName, birthdayBlock)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, fmt.Errorf("account %s already existing", accountName)
	}

	if _, err := w.querier.InsertAccount(ctx, queries.InsertAccountParams{
		Name:              account.Info.Key.Name,
		Index:             int32(account.Info.Key.Index),
		Xpub:              account.Info.Xpub,
		DerivationPath:    account.Info.DerivationPath,
		NextExternalIndex: int32(account.NextExternalIndex),
		NextInternalIndex: int32(account.NextInternalIndex),
		FkWalletID:        walletKey,
	}); err != nil {
		return nil, err
	}

	go w.publishEvent(domain.WalletEvent{
		EventType:   domain.WalletAccountCreated,
		AccountName: accountName,
	})

	return &account.Info, nil
}

func (w *walletRepositoryPg) DeriveNextExternalAddressesForAccount(
	ctx context.Context,
	accountName string,
	numOfAddresses uint64,
) ([]domain.AddressInfo, error) {
	addressesInfo := make([]domain.AddressInfo, 0)

	if err := w.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			for i := 0; i < int(numOfAddresses); i++ {
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

	go w.publishEvent(domain.WalletEvent{
		EventType:        domain.WalletAccountAddressesDerived,
		AccountName:      accountName,
		AccountAddresses: addressesInfo,
	})

	return addressesInfo, nil
}

func (w *walletRepositoryPg) DeriveNextInternalAddressesForAccount(
	ctx context.Context,
	accountName string,
	numOfAddresses uint64,
) ([]domain.AddressInfo, error) {
	addressesInfo := make([]domain.AddressInfo, 0)

	if err := w.UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			for i := 0; i < int(numOfAddresses); i++ {
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

	go w.publishEvent(domain.WalletEvent{
		EventType:        domain.WalletAccountAddressesDerived,
		AccountName:      accountName,
		AccountAddresses: addressesInfo,
	})

	return addressesInfo, nil
}

func (w *walletRepositoryPg) DeleteAccount(
	ctx context.Context,
	accountName string,
) error {
	conn, err := w.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	querierWithTx := w.querier.WithTx(tx)

	_, err = querierWithTx.GetAccount(ctx, accountName)
	if err != nil {
		if err.Error() == pgxNoRows {
			return ErrAccountNotFound
		}
		return err
	}

	if err := querierWithTx.DeleteAccountScripts(ctx, accountName); err != nil {
		return err
	}

	if err := querierWithTx.DeleteAccount(ctx, accountName); err != nil {
		return err
	}

	go w.publishEvent(domain.WalletEvent{
		EventType:   domain.WalletAccountDeleted,
		AccountName: accountName,
	})

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
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

//getWallet recreates wallet based on 3 tables: wallet, account, account_script_info
func (w *walletRepositoryPg) getWallet(
	ctx context.Context,
) (*domain.Wallet, error) {
	walletAccounts, err := w.querier.GetWalletAccountsAndScripts(ctx, walletKey)
	if err != nil {
		return nil, err
	}

	if len(walletAccounts) == 0 {
		return nil, ErrorWalletNotFound
	}

	wallet := &domain.Wallet{
		EncryptedMnemonic:   walletAccounts[0].EncryptedMnemonic,
		PasswordHash:        walletAccounts[0].PasswordHash,
		BirthdayBlockHeight: uint32(walletAccounts[0].BirthdayBlockHeight),
		RootPath:            walletAccounts[0].RootPath,
		NetworkName:         walletAccounts[0].NetworkName,
		NextAccountIndex:    uint32(walletAccounts[0].NextAccountIndex),
		AccountsByKey:       make(map[string]*domain.Account),
		AccountKeysByIndex:  make(map[uint32]string),
		AccountKeysByName:   make(map[string]string),
	}

	accounts := make(map[string]domain.Account, 0)
	if walletAccounts[0].Name.Valid {
		for _, v := range walletAccounts {
			if _, ok := accounts[v.Name.String]; !ok {
				derivationPathByScript := make(map[string]string)
				if v.ScriptDerivationPath.Valid {
					derivationPathByScript[v.Script.String] = v.ScriptDerivationPath.String
				}

				accounts[v.Name.String] = domain.Account{
					Info: domain.AccountInfo{
						Key: domain.AccountKey{
							Name:  v.Name.String,
							Index: uint32(v.Index.Int32),
						},
						Xpub:           v.Xpub.String,
						DerivationPath: v.AccountDerivationPath.String,
					},
					BirthdayBlock:          uint32(v.BirthdayBlockHeight),
					NextExternalIndex:      uint(v.NextExternalIndex.Int32),
					NextInternalIndex:      uint(v.NextInternalIndex.Int32),
					DerivationPathByScript: derivationPathByScript,
				}
			} else {
				if v.ScriptDerivationPath.Valid {
					accounts[v.Name.String].DerivationPathByScript[v.Script.String] =
						v.ScriptDerivationPath.String
				}
			}
		}
	}

	if len(accounts) > 0 {
		accountsByKey := make(map[string]*domain.Account)
		accountKeysByIndex := make(map[uint32]string)
		accountKeysByName := make(map[string]string)
		for k, v := range accounts {
			accountKey := domain.AccountKey{
				Name:  k,
				Index: wallet.NextAccountIndex,
			}
			accountsByKey[accountKey.String()] = &v

			accountKeysByIndex[v.Info.Key.Index] = accountKey.String()

			accountKeysByName[v.Info.Key.Name] = accountKey.String()
		}

		wallet.AccountsByKey = accountsByKey
		wallet.AccountKeysByIndex = accountKeysByIndex
		wallet.AccountKeysByName = accountKeysByName
	}

	return wallet, nil
}

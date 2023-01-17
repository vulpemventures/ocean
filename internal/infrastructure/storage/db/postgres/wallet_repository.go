package postgresdb

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres/sqlc/queries"
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
	for _, account := range updatedWallet.AccountsByNamespace {
		newAccount := false
		_, err := querierWithTx.GetAccount(ctx, account.Info.Namespace)
		if err != nil {
			if err.Error() == pgxNoRows {
				newAccount = true
			} else {
				return err
			}
		}

		label := sql.NullString{}
		if account.Info.Label != "" {
			label = sql.NullString{
				String: account.Info.Label,
				Valid:  true,
			}
		}

		if newAccount {
			if _, err := querierWithTx.InsertAccount(ctx, queries.InsertAccountParams{
				Namespace:         account.Info.Namespace,
				Index:             int32(account.Info.Index),
				Xpub:              account.Info.Xpub,
				DerivationPath:    account.Info.DerivationPath,
				NextExternalIndex: int32(account.NextExternalIndex),
				NextInternalIndex: int32(account.NextInternalIndex),
				FkWalletID:        walletKey,
				Label:             label,
			}); err != nil {
				return err
			}
		} else {
			if _, err := querierWithTx.UpdateAccountIndexes(
				ctx,
				queries.UpdateAccountIndexesParams{
					NextExternalIndex: int32(account.NextExternalIndex),
					NextInternalIndex: int32(account.NextInternalIndex),
					Label:             label,
					Namespace:         account.Info.Namespace,
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
				FkAccountNamespace: sql.NullString{
					String: account.Info.Namespace,
					Valid:  true,
				},
			})
		}
		if len(req) > 0 {
			if err := querierWithTx.DeleteAccountScripts(
				ctx,
				sql.NullString{
					String: account.Info.Namespace,
					Valid:  true,
				},
			); err != nil {
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
	label string,
	birthdayBlock uint32,
) (*domain.AccountInfo, error) {
	var accountInfo *domain.AccountInfo
	if err := w.UpdateWallet(
		ctx, func(wallet *domain.Wallet) (*domain.Wallet, error) {
			account, err := wallet.CreateAccount(label, birthdayBlock)
			if err != nil {
				return nil, err
			}

			accountInfo = &account.Info
			return wallet, nil
		},
	); err != nil {
		return nil, err
	}

	go w.publishEvent(domain.WalletEvent{
		EventType:        domain.WalletAccountCreated,
		AccountNamespace: accountInfo.Namespace,
	})

	return accountInfo, nil
}

func (w *walletRepositoryPg) DeriveNextExternalAddressesForAccount(
	ctx context.Context,
	accountName string,
	numOfAddresses uint64,
) ([]domain.AddressInfo, error) {
	addressesInfo := make([]domain.AddressInfo, 0)

	account, err := w.querier.GetAccount(ctx, accountName)
	if err != nil {
		return nil, err
	}

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
		AccountNamespace: account.Namespace,
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

	account, err := w.querier.GetAccount(ctx, accountName)
	if err != nil {
		return nil, err
	}

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
		AccountNamespace: account.Namespace,
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

	account, err := querierWithTx.GetAccount(ctx, accountName)
	if err != nil {
		if err.Error() == pgxNoRows {
			return ErrAccountNotFound
		}
		return err
	}

	if err := querierWithTx.DeleteAccountScripts(
		ctx,
		sql.NullString{
			String: account.Namespace,
			Valid:  true,
		},
	); err != nil {
		return err
	}

	if err := querierWithTx.DeleteAccount(ctx, account.Namespace); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	go w.publishEvent(domain.WalletEvent{
		EventType:        domain.WalletAccountDeleted,
		AccountNamespace: account.Namespace,
	})

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

// getWallet recreates wallet based on 3 tables: wallet, account, account_script_info
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

	accounts := make(map[string]domain.Account, 0)
	for _, v := range walletAccounts {
		if v.Namespace.Valid {
			if _, ok := accounts[v.Namespace.String]; !ok {
				derivationPathByScript := make(map[string]string)
				if v.ScriptDerivationPath.Valid {
					derivationPathByScript[v.Script.String] = v.ScriptDerivationPath.String
				}

				label := ""
				if v.Label.Valid {
					label = v.Label.String
				}
				accounts[v.Namespace.String] = domain.Account{
					Info: domain.AccountInfo{
						Namespace:      v.Namespace.String,
						Index:          uint32(v.Index.Int32),
						Label:          label,
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
					accounts[v.Namespace.String].DerivationPathByScript[v.Script.String] =
						v.ScriptDerivationPath.String
				}
			}
		}
	}

	accountsNamespaceByLabel := make(map[string]string)
	accountsByNamespace := make(map[string]*domain.Account)
	for k := range accounts {
		v := accounts[k]
		accountsNamespaceByLabel[v.Info.Label] = v.Info.Namespace
		accountsByNamespace[v.Info.Namespace] = &v
	}

	return &domain.Wallet{
		EncryptedMnemonic:        walletAccounts[0].EncryptedMnemonic,
		PasswordHash:             walletAccounts[0].PasswordHash,
		BirthdayBlockHeight:      uint32(walletAccounts[0].BirthdayBlockHeight),
		RootPath:                 walletAccounts[0].RootPath,
		NetworkName:              walletAccounts[0].NetworkName,
		AccountsNamespaceByLabel: accountsNamespaceByLabel,
		AccountsByNamespace:      accountsByNamespace,
		NextAccountIndex:         uint32(walletAccounts[0].NextAccountIndex),
	}, nil
}

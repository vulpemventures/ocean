package application

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
)

const (
	bip84RootPathPurpose = "84"
)

// AccountService is responsible for operations related to wallet accounts:
//   - Create a new account.
//   - Derive addresses for an existing account.
//   - List derived addresses for an existing account.
//   - Get balance of an existing account.
//   - List utxos of an existing account.
//   - Delete an existing account.
//
// The service registers 3 handlers related to the following wallet events:
//   - domain.WalletAccountCreated - whenever an account is created, the service initializes a dedicated blockchain scanner and starts listening for its reports.
//   - domain.WalletAccountAddressesDerived - whenever one or more addresses are derived for an account, they are added to the list of those watched by the account's scanner.
//   - domain.WalletAccountDeleted - whenever an account is deleted, the relative scanner is stopped and removed.
//
// The service guarantees to be always listening to notifications coming from
// each of its blockchain scanners in order to keep updated the utxo set of the
// relative accounts, ie. at startup it takes care of initializing a scanner
// for any existing account in case the wallet is already initialized and was
// just restarted.
type AccountService struct {
	repoManager ports.RepoManager
	bcScanner   ports.BlockchainScanner

	log  func(format string, a ...interface{})
	warn func(err error, format string, a ...interface{})
}

func NewAccountService(
	repoManager ports.RepoManager, bcScanner ports.BlockchainScanner,
) *AccountService {
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("account service: %s", format)
		log.Debugf(format, a...)
	}
	warnFn := func(err error, format string, a ...interface{}) {
		format = fmt.Sprintf("account service: %s", format)
		log.WithError(err).Warnf(format, a...)
	}

	svc := &AccountService{repoManager, bcScanner, logFn, warnFn}
	svc.registerHandlerForWalletEvents()
	return svc
}

func (as *AccountService) CreateAccountBIP44(
	ctx context.Context, label string,
) (*AccountInfo, error) {
	_, birthdayBlockHeight, err := as.bcScanner.GetLatestBlock()
	if err != nil {
		return nil, err
	}
	accountInfo, err := as.repoManager.WalletRepository().CreateAccount(
		ctx, bip84RootPathPurpose, label, birthdayBlockHeight,
	)
	if err != nil {
		return nil, err
	}
	return (*AccountInfo)(accountInfo), nil
}

func (as *AccountService) DeriveAddressesForAccount(
	ctx context.Context, namespace string, numOfAddresses uint64,
) (AddressesInfo, error) {
	if numOfAddresses == 0 {
		numOfAddresses = 1
	}

	addressesInfo, err := as.repoManager.WalletRepository().
		DeriveNextExternalAddressesForAccount(ctx, namespace, numOfAddresses)
	if err != nil {
		return nil, err
	}
	return AddressesInfo(addressesInfo), nil
}

func (as *AccountService) DeriveChangeAddressesForAccount(
	ctx context.Context, namespace string, numOfAddresses uint64,
) (AddressesInfo, error) {
	if numOfAddresses == 0 {
		numOfAddresses = 1
	}

	addressesInfo, err := as.repoManager.WalletRepository().
		DeriveNextInternalAddressesForAccount(ctx, namespace, numOfAddresses)
	if err != nil {
		return nil, err
	}
	return addressesInfo, nil
}

func (as *AccountService) ListAddressesForAccount(
	ctx context.Context, namespace string,
) (AddressesInfo, error) {
	w, err := as.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}

	addressesInfo, err := w.AllDerivedAddressesForAccount(namespace)
	if err != nil {
		return nil, err
	}
	return AddressesInfo(addressesInfo), nil
}

func (as *AccountService) GetBalanceForAccount(
	ctx context.Context, namespace string,
) (BalanceInfo, error) {
	w, err := as.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}

	account, err := w.GetAccount(namespace)
	if err != nil {
		return nil, err
	}

	return as.repoManager.UtxoRepository().GetBalanceForAccount(
		ctx, account.Info.Key.Namespace,
	)
}

func (as *AccountService) ListUtxosForAccount(
	ctx context.Context, namespace string,
) (*UtxoInfo, error) {
	w, err := as.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}

	account, err := w.GetAccount(namespace)
	if err != nil {
		return nil, err
	}

	spendableUtxos, err := as.repoManager.UtxoRepository().GetSpendableUtxosForAccount(
		ctx, account.Info.Key.Namespace,
	)
	if err != nil {
		return nil, err
	}

	lockedUtxos, err := as.repoManager.UtxoRepository().GetLockedUtxosForAccount(
		ctx, account.Info.Key.Namespace,
	)
	if err != nil {
		return nil, err
	}

	return &UtxoInfo{spendableUtxos, lockedUtxos}, nil
}

func (as *AccountService) DeleteAccount(
	ctx context.Context, namespace string,
) (err error) {
	balance, err := as.GetBalanceForAccount(ctx, namespace)
	if err != nil {
		return
	}
	if len(balance) > 0 {
		err = fmt.Errorf(
			"account %s must have zero balance to be deleted", namespace,
		)
		return
	}

	defer func() {
		if err == nil {
			if err := as.repoManager.UtxoRepository().DeleteUtxosForAccount(
				ctx, namespace,
			); err != nil {
				as.warn(
					err, "account service: error while deleting utxos for account %s",
					namespace,
				)
			}
		}
	}()

	err = as.repoManager.WalletRepository().DeleteAccount(ctx, namespace)
	return
}

func (as *AccountService) SetAccountLabel(
	ctx context.Context,
	namespace string,
	label string,
) error {
	return as.repoManager.WalletRepository().UpdateWallet(
		ctx,
		func(w *domain.Wallet) (*domain.Wallet, error) {
			if err := w.UpdateAccountsLabel(namespace, label); err != nil {
				return nil, err
			}
			return w, nil
		},
	)
}

func (as *AccountService) registerHandlerForWalletEvents() {
	// Start watching all existing accounts' addresses as soon as wallet is unlocked.
	as.repoManager.RegisterHandlerForWalletEvent(
		domain.WalletUnlocked, func(event domain.WalletEvent) {
			w, _ := as.repoManager.WalletRepository().GetWallet(context.Background())

			for namespace := range w.AccountKeysByNamespace {
				accountKey := w.AccountKeysByNamespace[namespace]
				account := w.AccountsByKey[accountKey]
				addressesInfo, _ := w.AllDerivedAddressesForAccount(namespace)
				if len(addressesInfo) > 0 {
					as.log("start watching addresses for account %s", namespace)
					as.bcScanner.WatchForAccount(namespace, account.BirthdayBlock, addressesInfo)
				}
				go as.listenToUtxoChannel(namespace, as.bcScanner.GetUtxoChannel(namespace))
				go as.listenToTxChannel(namespace, as.bcScanner.GetTxChannel(namespace))
			}
		},
	)
	// Start watching account as soon as it is created.
	as.repoManager.RegisterHandlerForWalletEvent(
		domain.WalletAccountCreated, func(event domain.WalletEvent) {
			as.bcScanner.WatchForAccount(
				event.AccountNamespace, event.AccountBirthdayBlock, event.AccountAddresses,
			)
			chUtxos := as.bcScanner.GetUtxoChannel(event.AccountNamespace)
			chTxs := as.bcScanner.GetTxChannel(event.AccountNamespace)
			go as.listenToUtxoChannel(event.AccountNamespace, chUtxos)
			go as.listenToTxChannel(event.AccountNamespace, chTxs)
		},
	)
	// Start watching account address as soon as it's derived.
	as.repoManager.RegisterHandlerForWalletEvent(
		domain.WalletAccountAddressesDerived, func(event domain.WalletEvent) {
			as.bcScanner.WatchForAccount(
				event.AccountNamespace, event.AccountBirthdayBlock, event.AccountAddresses,
			)
		},
	)
	// Stop watching account and all its addresses as soon as it's deleted.
	as.repoManager.RegisterHandlerForWalletEvent(
		domain.WalletAccountDeleted, func(event domain.WalletEvent) {
			as.bcScanner.StopWatchForAccount(event.AccountNamespace)
		},
	)
	// Start watching for when utxos are spent as soon as they are added to the storage.
	as.repoManager.RegisterHandlerForUtxoEvent(
		domain.UtxoAdded, func(event domain.UtxoEvent) {
			namespace := event.Utxos[0].FkAccountNamespace
			as.bcScanner.WatchForUtxos(namespace, event.Utxos)
		},
	)

	// In background, make sure to watch for all utxos to get notified when they are spent.
	go func() {
		utxos, err := as.repoManager.UtxoRepository().GetAllUtxos(
			context.Background(),
		)
		if err != nil {
			as.warn(err, "account service: error while getting all utxos")
			return
		}
		for _, u := range utxos {
			if !u.IsSpent() {
				as.bcScanner.WatchForUtxos(u.FkAccountNamespace, []domain.UtxoInfo{u.Info()})
			}
		}
	}()
}

func (as *AccountService) listenToUtxoChannel(
	namespace string, chUtxos chan []*domain.Utxo,
) {
	as.log("start listening to utxo channel for account %s", namespace)

	for utxos := range chUtxos {
		time.Sleep(time.Millisecond)

		as.log(
			"received %d utxo(s) from channel for account %s",
			len(utxos), namespace,
		)

		utxoKeys := make([]domain.UtxoKey, 0, len(utxos))
		for _, u := range utxos {
			utxoKeys = append(utxoKeys, u.Key())
		}
		if utxos[0].IsSpent() {
			count, err := as.repoManager.UtxoRepository().SpendUtxos(
				context.Background(), utxoKeys, utxos[0].SpentStatus,
			)
			if err != nil {
				as.warn(
					err, "error while updating utxos status to spent for account %s",
					namespace,
				)
			}
			if count > 0 {
				as.log("spent %d utxos for account %s", count, namespace)
			}
			continue
		}

		if utxos[0].IsConfirmed() {
			count, err := as.repoManager.UtxoRepository().ConfirmUtxos(
				context.Background(), utxoKeys, utxos[0].ConfirmedStatus,
			)
			if err != nil {
				as.warn(
					err, "error while updating utxos status to confirmed for account %s",
					namespace,
				)
			}
			if count > 0 {
				as.log("confirmed %d utxo(s) for account %s", count, namespace)
				continue
			}
		}

		count, err := as.repoManager.UtxoRepository().AddUtxos(
			context.Background(), utxos,
		)
		if err != nil {
			as.warn(err, "error while adding new utxos for account %s", namespace)
		}
		if count > 0 {
			as.log("added %d utxo(s) for account %s", count, namespace)
		}
	}
}

func (as *AccountService) listenToTxChannel(
	namespace string, chTxs chan *domain.Transaction,
) {
	as.log("start listening to tx channel for account %s", namespace)

	ctx := context.Background()
	txRepo := as.repoManager.TransactionRepository()
	for tx := range chTxs {
		time.Sleep(time.Millisecond)

		as.log("received new tx %s from channel", tx.TxID)

		gotTx, _ := txRepo.GetTransaction(ctx, tx.TxID)
		if gotTx == nil {
			if _, err := txRepo.AddTransaction(ctx, tx); err != nil {
				as.warn(
					err, "error while adding new transaction %s for account %s",
					tx.TxID, namespace,
				)
				continue
			}
			as.log("added new transaction %s for account %s", tx.TxID, namespace)
			continue
		}
		if !gotTx.IsConfirmed() && tx.IsConfirmed() {
			if _, err := txRepo.ConfirmTransaction(
				ctx, tx.TxID, tx.BlockHash, tx.BlockHeight,
			); err != nil {
				as.warn(
					err, "error while confirming transaction %s for account %s",
					tx.TxID, namespace,
				)
			}
			as.log("confirmed transaction %s for account %s", tx.TxID, namespace)
		}

		if !gotTx.HasAccounts(tx) {
			if err := txRepo.UpdateTransaction(
				ctx, tx.TxID, func(t *domain.Transaction) (*domain.Transaction, error) {
					for _, account := range tx.GetAccounts() {
						t.AddAccount(account)
					}
					return t, nil
				},
			); err != nil {
				as.warn(err, "error while updating accounts to transaction %s", tx.TxID)
				continue
			}
			as.log("updated accounts for transaction %s", tx.TxID)
		}
	}
}

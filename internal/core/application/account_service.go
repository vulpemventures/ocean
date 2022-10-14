package application

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
)

// AccountService is responsible for operations related to wallet accounts:
// 	* Create a new account.
// 	* Derive addresses for an existing account.
// 	* List derived addresses for an existing account.
// 	* Get balance of an existing account.
// 	* List utxos of an existing account.
// 	* Delete an existing account.
//
// The service registers 3 handlers related to the following wallet events:
//	* domain.WalletAccountCreated - whenever an account is created, the service initializes a dedicated blockchain scanner and starts listening for its reports.
//	* domain.WalletAccountAddressesDerived - whenever one or more addresses are derived for an account, they are added to the list of those watched by the account's scanner.
//	* domain.WalletAccountDeleted - whenever an account is deleted, the relative scanner is stopped and removed.
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
	ctx context.Context, accountName string,
) (*AccountInfo, error) {
	_, birthdayBlockHeight, err := as.bcScanner.GetLatestBlock()
	if err != nil {
		return nil, err
	}
	accountInfo, err := as.repoManager.WalletRepository().CreateAccount(
		ctx, accountName, birthdayBlockHeight,
	)
	if err != nil {
		return nil, err
	}
	return (*AccountInfo)(accountInfo), nil
}

func (as *AccountService) DeriveAddressesForAccount(
	ctx context.Context, accountName string, numOfAddresses uint64,
) (AddressesInfo, error) {
	if numOfAddresses == 0 {
		numOfAddresses = 1
	}

	addressesInfo, err := as.repoManager.WalletRepository().
		DeriveNextExternalAddressesForAccount(ctx, accountName, numOfAddresses)
	if err != nil {
		return nil, err
	}
	return AddressesInfo(addressesInfo), nil
}

func (as *AccountService) DeriveChangeAddressesForAccount(
	ctx context.Context, accountName string, numOfAddresses uint64,
) (AddressesInfo, error) {
	if numOfAddresses == 0 {
		numOfAddresses = 1
	}

	addressesInfo, err := as.repoManager.WalletRepository().
		DeriveNextInternalAddressesForAccount(ctx, accountName, numOfAddresses)
	if err != nil {
		return nil, err
	}
	return AddressesInfo(addressesInfo), nil
}

func (as *AccountService) ListAddressesForAccount(
	ctx context.Context, accountName string,
) (AddressesInfo, error) {
	w, err := as.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}

	addressesInfo, err := w.AllDerivedExternalAddressesForAccount(accountName)
	if err != nil {
		return nil, err
	}
	return AddressesInfo(addressesInfo), nil
}

func (as *AccountService) GetBalanceForAccount(
	ctx context.Context, accountName string,
) (BalanceInfo, error) {
	w, err := as.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}

	account, err := w.GetAccount(accountName)
	if err != nil {
		return nil, err
	}

	return as.repoManager.UtxoRepository().GetBalanceForAccount(
		ctx, account.Info.Key.Name,
	)
}

func (as *AccountService) ListUtxosForAccount(
	ctx context.Context, accountName string,
) (*UtxoInfo, error) {
	w, err := as.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}

	account, err := w.GetAccount(accountName)
	if err != nil {
		return nil, err
	}

	spendableUtxos, err := as.repoManager.UtxoRepository().GetSpendableUtxosForAccount(
		ctx, account.Info.Key.Name,
	)
	if err != nil {
		return nil, err
	}

	lockedUtxos, err := as.repoManager.UtxoRepository().GetLockedUtxosForAccount(
		ctx, account.Info.Key.Name,
	)
	if err != nil {
		return nil, err
	}

	return &UtxoInfo{spendableUtxos, lockedUtxos}, nil
}

func (as *AccountService) DeleteAccount(
	ctx context.Context, accountName string,
) (err error) {
	balance, err := as.GetBalanceForAccount(ctx, accountName)
	if err != nil {
		return
	}
	if len(balance) > 0 {
		err = fmt.Errorf(
			"account %s must have zero balance to be deleted", accountName,
		)
		return
	}

	defer func() {
		if err == nil {
			if err := as.repoManager.UtxoRepository().DeleteUtxosForAccount(
				ctx, accountName,
			); err != nil {
				as.warn(
					err, "account service: error while deleting utxos for account %s",
					accountName,
				)
			}
		}
	}()

	err = as.repoManager.WalletRepository().DeleteAccount(ctx, accountName)
	return
}

func (as *AccountService) registerHandlerForWalletEvents() {
	// Start watching all existing accounts' addresses as soon as wallet is unlocked.
	as.repoManager.RegisterHandlerForWalletEvent(
		domain.WalletUnlocked, func(event domain.WalletEvent) {
			w, _ := as.repoManager.WalletRepository().GetWallet(context.Background())
			for key, acc := range w.AccountsByKey {
				fmt.Printf("%s %+v\n", key, acc)
			}

			for accountName := range w.AccountKeysByName {
				accountKey := w.AccountKeysByName[accountName]
				account := w.AccountsByKey[accountKey]
				addressesInfo, _ := w.AllDerivedAddressesForAccount(accountName)
				if len(addressesInfo) > 0 {
					as.log("start watching addresses for account %s", accountName)
					as.bcScanner.WatchForAccount(accountName, account.BirthdayBlock, addressesInfo)
				}
				go as.listenToUtxoChannel(accountName, as.bcScanner.GetUtxoChannel(accountName))
				go as.listenToTxChannel(accountName, as.bcScanner.GetTxChannel(accountName))
			}
		},
	)
	// Start watching account as soon as it is created.
	as.repoManager.RegisterHandlerForWalletEvent(
		domain.WalletAccountCreated, func(event domain.WalletEvent) {
			as.bcScanner.WatchForAccount(
				event.AccountName, event.AccountBirthdayBlock, event.AccountAddresses,
			)
			chUtxos := as.bcScanner.GetUtxoChannel(event.AccountName)
			chTxs := as.bcScanner.GetTxChannel(event.AccountName)
			go as.listenToUtxoChannel(event.AccountName, chUtxos)
			go as.listenToTxChannel(event.AccountName, chTxs)
		},
	)
	// Start watching account address as soon as it's derived.
	as.repoManager.RegisterHandlerForWalletEvent(
		domain.WalletAccountAddressesDerived, func(event domain.WalletEvent) {
			as.bcScanner.WatchForAccount(
				event.AccountName, event.AccountBirthdayBlock, event.AccountAddresses,
			)
		},
	)
	// Stop watching account and all its addresses as soon as it's deleted.
	as.repoManager.RegisterHandlerForWalletEvent(
		domain.WalletAccountDeleted, func(event domain.WalletEvent) {
			as.bcScanner.StopWatchForAccount(event.AccountName)
		},
	)
}

func (as *AccountService) listenToUtxoChannel(
	accountName string, chUtxos chan []*domain.Utxo,
) {
	as.log("start listening to utxo channel for account %s", accountName)

	for utxos := range chUtxos {
		time.Sleep(time.Millisecond)

		as.log(
			"received %d utxo(s) from channel for account %s",
			len(utxos), accountName,
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
					accountName,
				)
			}
			if count > 0 {
				as.log("spent %d utxos for account %s", count, accountName)
			}
			continue
		}

		if utxos[0].IsConfirmed() {
			count, err := as.repoManager.UtxoRepository().ConfirmUtxos(
				context.Background(), utxoKeys, utxos[0].ConfirmedStatus,
			)
			if err != nil {
				as.warn(
					err, "error while updating utxos status to spent for account %s",
					accountName,
				)
			}
			if count > 0 {
				as.log("spent %d utxos for account %s", count, accountName)
				continue
			}
		}

		count, err := as.repoManager.UtxoRepository().AddUtxos(
			context.Background(), utxos,
		)
		if err != nil {
			as.warn(err, "error while adding new utxos for account %s", accountName)
		}
		if count > 0 {
			as.log("added %d utxos for account %s", count, accountName)
		}
	}
}

func (as *AccountService) listenToTxChannel(
	accountName string, chTxs chan *domain.Transaction,
) {
	as.log("start listening to tx channel for account %s", accountName)

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
					tx.TxID, accountName,
				)
				continue
			}
			as.log("added new transaction %s for account %s", tx.TxID, accountName)
			continue
		}
		if !gotTx.IsConfirmed() && tx.IsConfirmed() {
			if _, err := txRepo.ConfirmTransaction(
				ctx, tx.TxID, tx.BlockHash, tx.BlockHeight,
			); err != nil {
				as.warn(
					err, "error while confirming transaction %s for account %s",
					tx.TxID, accountName,
				)
			}
			as.log("confirmed transaction %s for account %s", tx.TxID, accountName)
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

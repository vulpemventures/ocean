package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
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
	txQueue     *transactionQueue

	log  func(format string, a ...interface{})
	warn func(err error, format string, a ...interface{})
}

func NewAccountService(
	repoManager ports.RepoManager, bcScanner ports.BlockchainScanner,
) *AccountService {
	txQueue := newTransactionQueue()
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("account service: %s", format)
		log.Debugf(format, a...)
	}
	warnFn := func(err error, format string, a ...interface{}) {
		format = fmt.Sprintf("account service: %s", format)
		log.WithError(err).Warnf(format, a...)
	}

	svc := &AccountService{repoManager, bcScanner, txQueue, logFn, warnFn}
	svc.registerHandlerForWalletEvents()
	return svc
}

func (as *AccountService) CreateAccountBIP44(
	ctx context.Context, label string, unconf bool,
) (*AccountInfo, error) {
	_, birthdayBlockHeight, err := as.bcScanner.GetLatestBlock()
	if err != nil {
		return nil, err
	}
	accountInfo, err := as.repoManager.WalletRepository().CreateAccount(
		ctx, label, birthdayBlockHeight, unconf,
	)
	if err != nil {
		return nil, err
	}
	return &AccountInfo{*accountInfo}, nil
}

func (as *AccountService) SetAccountLabel(
	ctx context.Context, accountName, label string,
) (*AccountInfo, error) {
	w, err := as.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}

	if err := w.SetLabelForAccount(accountName, label); err != nil {
		return nil, err
	}

	if err := as.repoManager.WalletRepository().UpdateWallet(
		ctx, func(_ *domain.Wallet) (*domain.Wallet, error) {
			return w, nil
		},
	); err != nil {
		return nil, err
	}

	account, _ := w.GetAccount(label)
	return &AccountInfo{account.AccountInfo}, nil
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

	addressesInfo, err := w.AllDerivedAddressesForAccount(accountName)
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
		ctx, account.Namespace,
	)
}

func (as *AccountService) ListUtxosForAccount(
	ctx context.Context, accountName string, scripts [][]byte,
) (*UtxoInfo, error) {
	w, err := as.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}

	account, err := w.GetAccount(accountName)
	if err != nil {
		return nil, err
	}

	utxos, err := as.repoManager.UtxoRepository().GetAllUtxosForAccount(
		ctx, account.Namespace, scripts,
	)
	if err != nil {
		return nil, err
	}

	spendableUtxos := make([]*domain.Utxo, 0, len(utxos))
	unconfirmedUtxos := make([]*domain.Utxo, 0, len(utxos))
	lockedUtxos := make([]*domain.Utxo, 0, len(utxos))

	for _, u := range utxos {
		if u.IsLocked() {
			lockedUtxos = append(lockedUtxos, u)
		} else {
			if u.IsConfirmed() {
				spendableUtxos = append(spendableUtxos, u)
			} else {
				unconfirmedUtxos = append(unconfirmedUtxos, u)
			}
		}
	}

	return &UtxoInfo{spendableUtxos, lockedUtxos, unconfirmedUtxos}, nil
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

			for _, account := range w.Accounts {
				addressesInfo, _ := w.AllDerivedAddressesForAccount(account.Namespace)
				if len(addressesInfo) > 0 {
					as.log("start watching addresses for account %s", account.Namespace)
					as.bcScanner.WatchForAccount(
						account.Namespace, account.BirthdayBlock, addressesInfo,
					)
				}
				go as.listenToUtxoChannel(
					account.Namespace, as.bcScanner.GetUtxoChannel(account.Namespace),
				)
				go as.listenToTxChannel(
					account.Namespace, as.bcScanner.GetTxChannel(account.Namespace),
				)
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

		utxoKeys := make([]domain.UtxoKey, 0, len(utxos))
		for _, u := range utxos {
			utxoKeys = append(utxoKeys, u.Key())
		}

		if utxos[0].IsConfirmedSpent() {
			count, err := as.repoManager.UtxoRepository().ConfirmSpendUtxos(
				context.Background(), utxoKeys, utxos[0].SpentStatus,
			)
			if err != nil {
				as.warn(
					err, "error while updating utxos status to confirmed spend for account %s",
					accountName,
				)
			}
			if count > 0 {
				as.log("confirmed spend of %d utxos for account %s", count, accountName)
			}
			continue
		}

		if utxos[0].IsSpent() {
			count, err := as.repoManager.UtxoRepository().SpendUtxos(
				context.Background(), utxoKeys, utxos[0].SpentStatus.Txid,
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
					err, "error while updating utxos status to confirmed for account %s",
					accountName,
				)
			}
			if count > 0 {
				as.log("confirmed %d utxo(s) for account %s", count, accountName)
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
			as.log("added %d utxo(s) for account %s", count, accountName)
		}
	}
}

func (as *AccountService) listenToTxChannel(
	accountName string, chTxs chan *domain.Transaction,
) {
	as.log("start listening to tx channel for account %s", accountName)

	// Every tx received from the blockchain scanner is pushed to a queue that is
	// emptied 1 second after the first elem is added. All the queued txs are
	// then persisted in the repository. This because it can happen to receive
	// here the same tx on 2 different channels in case the user moves
	// funds from one account to another.
	// In such cases, the queue takes care of updating a tx if it's already
	// queued, instead of doing this operation against the repo (can be slower).
	for tx := range chTxs {
		if as.txQueue.len() <= 0 {
			go func() {
				time.Sleep(time.Second)
				as.storeQueuedTransactions()
			}()
		}
		as.txQueue.pushBack(tx)
	}
}

func (as *AccountService) storeQueuedTransactions() {
	txs := as.txQueue.pop()
	ctx := context.Background()
	txRepo := as.repoManager.TransactionRepository()
	for _, tx := range txs {
		gotTx, _ := txRepo.GetTransaction(ctx, tx.TxID)
		accounts := strings.Join(tx.GetAccounts(), ", ")
		if gotTx == nil {
			as.log("received new tx %s from channel", tx.TxID)

			if _, err := txRepo.AddTransaction(ctx, tx); err != nil {
				as.warn(err, "error while adding new transaction %s", tx.TxID)
				continue
			}
			as.log("added new transaction %s for account(s) %s", tx.TxID, accounts)
			continue
		}

		if !gotTx.IsConfirmed() && tx.IsConfirmed() {
			as.log("received confirmed tx %s from channel", tx.TxID)

			if _, err := txRepo.ConfirmTransaction(
				ctx, tx.TxID, tx.BlockHash, tx.BlockHeight, tx.BlockTime,
			); err != nil {
				as.warn(
					err, "error while confirming transaction %s for account(s) %s",
					tx.TxID, accounts,
				)
				continue
			}
			as.log("confirmed transaction %s for account(s) %s", tx.TxID, accounts)
		}
	}
}

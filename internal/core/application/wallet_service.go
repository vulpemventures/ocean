package application

import (
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	path "github.com/vulpemventures/ocean/pkg/wallet/derivation-path"
	"github.com/vulpemventures/ocean/pkg/wallet/mnemonic"
	singlesig "github.com/vulpemventures/ocean/pkg/wallet/single-sig"
)

const (
	defaultEmptyAccountThreshold    = 3
	defaultUnusedAddressesThreshold = 100
)

// WalletService is responsible for operations related to the managment of the
// wallet:
//   - Generate a new random 24-words mnemonic.
//   - Create a new wallet from scratch with given mnemonic and locked with the given password.
//   - Unlock the wallet with a password.
//   - Change the wallet password. It requires the wallet to be locked.
//   - Get the status of the wallet (initialized, unlocked, inSync).
//   - Get non-sensiive (network, native asset) and possibly sensitive info (root path, master blinding key and basic accounts' info) about the wallet. Sensitive info are returned only if the wallet is unlocked.
//
// This service doesn't register any handler for wallet events, rather it
// allows its users to register their handler to manage situations like the
// unlocking of the wallet (for example, check how the grpc service uses this
// feature).
type WalletService struct {
	repoManager ports.RepoManager
	bcScanner   ports.BlockchainScanner
	rootPath    string
	network     *network.Network
	buildInfo   BuildInfo

	initialized bool
	unlocked    bool
	synced      bool
	lock        *sync.RWMutex

	log func(format string, a ...interface{})
}

func NewWalletService(
	repoManager ports.RepoManager, bcScanner ports.BlockchainScanner,
	rootPath string, net *network.Network, buildInfo BuildInfo,
) *WalletService {
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("wallet service: %s", format)
		log.Debugf(format, a...)
	}
	ws := &WalletService{
		repoManager: repoManager,
		bcScanner:   bcScanner,
		rootPath:    rootPath,
		network:     net,
		buildInfo:   buildInfo,
		lock:        &sync.RWMutex{},
		log:         logFn,
	}
	w, _ := ws.repoManager.WalletRepository().GetWallet(context.Background())
	if w != nil {
		ws.setInitialized()
		ws.setSynced()
	}
	return ws
}

func (ws *WalletService) GenSeed(ctx context.Context) ([]string, error) {
	return mnemonic.NewMnemonic(mnemonic.NewMnemonicArgs{})
}

func (ws *WalletService) CreateWallet(
	ctx context.Context, mnemonic []string, passphrase string,
) (err error) {
	defer func() {
		if err == nil {
			ws.setInitialized()
			ws.setSynced()
		}
	}()

	if ws.isInitialized() {
		return fmt.Errorf("wallet is already initialized")
	}

	_, birthdayBlockHeight, err := ws.bcScanner.GetLatestBlock()
	if err != nil {
		return
	}

	newWallet, err := domain.NewWallet(
		mnemonic, passphrase, ws.rootPath, ws.network.Name,
		birthdayBlockHeight, nil,
	)
	if err != nil {
		return
	}
	newWallet.Lock(passphrase)

	return ws.repoManager.WalletRepository().CreateWallet(ctx, newWallet)
}

func (ws *WalletService) Unlock(
	ctx context.Context, password string,
) (err error) {
	if ws.isUnlocked() {
		return nil
	}

	defer func() {
		if err == nil {
			ws.setUnlocked()
		}
	}()

	return ws.repoManager.WalletRepository().UnlockWallet(ctx, password)
}

func (ws *WalletService) Lock(
	ctx context.Context, password string,
) (err error) {
	if !ws.isUnlocked() {
		return nil
	}

	defer func() {
		if err == nil {
			ws.setLocked()
		}
	}()

	return ws.repoManager.WalletRepository().LockWallet(ctx, password)
}

func (ws *WalletService) ChangePassword(
	ctx context.Context, currentPassword, newPassword string,
) error {
	return ws.repoManager.WalletRepository().UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			if err := w.ChangePassword(
				currentPassword, newPassword,
			); err != nil {
				return nil, err
			}
			return w, nil
		},
	)
}

func (ws *WalletService) RestoreWallet(
	ctx context.Context, chMessages chan WalletRestoreMessage,
	mnemonic []string, rootPath, passpharse string,
	birthdayBlockHeight, emptyAccountsThreshold, unusedAddressesThreshold uint32,
) {
	defer close(chMessages)

	canceled := false
	c, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(c, ctx context.Context, b *bool) {
		for {
			select {
			case <-ctx.Done():
				*b = true
				ws.log("process aborted")
				ws.repoManager.Reset()
				ws.setNotInitialized()
				return
			case <-c.Done():
				return
			}
		}
	}(c, ctx, &canceled)

	if ws.isInitialized() {
		sendMessage(canceled, chMessages, WalletRestoreMessage{
			Err: fmt.Errorf("wallet is already initialized"),
		})
		return
	}

	walletRootPath := rootPath
	if walletRootPath == "" {
		walletRootPath = ws.rootPath
	}
	if emptyAccountsThreshold == 0 {
		emptyAccountsThreshold = defaultEmptyAccountThreshold
	}
	if unusedAddressesThreshold == 0 {
		unusedAddressesThreshold = defaultUnusedAddressesThreshold
	}
	accountIndex := uint32(0)
	emptyAccountCounter := uint32(0)
	accounts := make([]domain.Account, 0)
	w, _ := singlesig.NewWalletFromMnemonic(singlesig.NewWalletFromMnemonicArgs{
		RootPath: walletRootPath,
		Mnemonic: mnemonic,
	})

	if !sendMessage(canceled, chMessages, WalletRestoreMessage{
		Message: "start restoring wallet accounts...",
	}) {
		return
	}

	addressesByAccount := make(map[uint32][]domain.AddressInfo)
	accountByScript := make(map[string]string)
	for {
		if emptyAccountCounter == emptyAccountsThreshold {
			break
		}

		accountName := fmt.Sprintf("account%d", accountIndex)

		msg := fmt.Sprintf("restoring account %d...", accountIndex)
		if !sendMessage(canceled, chMessages, WalletRestoreMessage{
			Message: msg,
		}) {
			return
		}
		ws.log(msg)
		xpub, _ := w.AccountExtendedPublicKey(singlesig.ExtendedKeyArgs{
			Account: accountIndex,
		})
		masterBlidningKeyStr, _ := w.MasterBlindingKey()
		masterBlidningKey, _ := hex.DecodeString(masterBlidningKeyStr)
		externalAddresses, internalAddresses, err := ws.bcScanner.RestoreAccount(
			accountIndex, xpub, masterBlidningKey, birthdayBlockHeight, unusedAddressesThreshold,
		)
		if err != nil {
			sendMessage(canceled, chMessages, WalletRestoreMessage{Err: err})
			return
		}

		if len(externalAddresses) <= 0 && len(internalAddresses) <= 0 {
			if !sendMessage(canceled, chMessages, WalletRestoreMessage{
				Message: fmt.Sprintf("account %d empty", accountIndex),
			}) {
				return
			}
			ws.log("account %d empty", accountIndex)
			emptyAccountCounter++
			accountIndex++
			continue
		}

		msg = fmt.Sprintf(
			"found %d external address(es) for account %d",
			len(externalAddresses), accountIndex,
		)
		if !sendMessage(canceled, chMessages, WalletRestoreMessage{
			Message: msg,
		}) {
			return
		}
		ws.log(msg)

		msg = fmt.Sprintf(
			"found %d internal address(es) for account %d",
			len(internalAddresses), accountIndex,
		)
		if !sendMessage(canceled, chMessages, WalletRestoreMessage{
			Message: msg,
		}) {
			return
		}
		ws.log(msg)

		addressesByAccount[accountIndex] = append(
			addressesByAccount[accountIndex], externalAddresses...,
		)
		addressesByAccount[accountIndex] = append(
			addressesByAccount[accountIndex], internalAddresses...,
		)

		// sort addresses by derivation path (desc order) to facilitate retrieving
		// the last derived index.
		sort.SliceStable(externalAddresses, func(i, j int) bool {
			path1, _ := path.ParseDerivationPath(externalAddresses[i].DerivationPath)
			path2, _ := path.ParseDerivationPath(externalAddresses[j].DerivationPath)
			return path1[len(path1)-1] > path2[len(path2)-1]
		})
		sort.SliceStable(internalAddresses, func(i, j int) bool {
			path1, _ := path.ParseDerivationPath(internalAddresses[i].DerivationPath)
			path2, _ := path.ParseDerivationPath(internalAddresses[j].DerivationPath)
			return path1[len(path1)-1] > path2[len(path2)-1]
		})

		derivationPaths := make(map[string]string)
		for _, i := range externalAddresses {
			accountByScript[i.Script] = accountName
			derivationPaths[i.Script] = i.DerivationPath
		}
		for _, i := range internalAddresses {
			accountByScript[i.Script] = accountName
			derivationPaths[i.Script] = i.DerivationPath
		}

		var nextExternalIndex, nextInternalIndex uint
		if len(externalAddresses) > 0 {
			p, _ := path.ParseDerivationPath(externalAddresses[0].DerivationPath)
			nextExternalIndex = uint(p[len(p)-1] + 1)
		}
		if len(internalAddresses) > 0 {
			p, _ := path.ParseDerivationPath(internalAddresses[0].DerivationPath)
			nextInternalIndex = uint(p[len(p)-1] + 1)
		}

		// TODO: maybe take name from function args? Something like a mapping <index, name>
		accounts = append(accounts, domain.Account{
			Info: domain.AccountInfo{
				Key: domain.AccountKey{
					Index: accountIndex,
					Name:  accountName,
				},
				Xpub:           xpub,
				DerivationPath: fmt.Sprintf("%s/%d'", walletRootPath, accountIndex),
			},
			BirthdayBlock:          birthdayBlockHeight,
			NextExternalIndex:      uint(nextExternalIndex),
			NextInternalIndex:      uint(nextInternalIndex),
			DerivationPathByScript: derivationPaths,
		})
		accountIndex++
		emptyAccountCounter = 0
	}

	if !sendMessage(canceled, chMessages, WalletRestoreMessage{
		Message: "wallet accounts restored",
	}) {
		return
	}

	if !sendMessage(canceled, chMessages, WalletRestoreMessage{
		Message: "initializing wallet...",
	}) {
		return
	}

	newWallet, err := domain.NewWallet(
		mnemonic, passpharse, walletRootPath, ws.network.Name,
		birthdayBlockHeight, accounts,
	)
	if err != nil {
		sendMessage(canceled, chMessages, WalletRestoreMessage{Err: err})
		return
	}

	if rootPath != "" {
		ws.rootPath = rootPath
	}

	if err := ws.repoManager.WalletRepository().CreateWallet(
		ctx, newWallet,
	); err != nil {
		sendMessage(canceled, chMessages, WalletRestoreMessage{Err: err})
		return
	}

	ws.setInitialized()

	if !sendMessage(canceled, chMessages, WalletRestoreMessage{
		Message: "wallet initialized",
	}) {
		return
	}

	if !sendMessage(canceled, chMessages, WalletRestoreMessage{
		Message: "restoring wallet utxo pool...",
	}) {
		return
	}

	addresses := make([]domain.AddressInfo, 0)
	for _, accountAddresses := range addressesByAccount {
		addresses = append(addresses, accountAddresses...)
	}
	utxos, err := ws.bcScanner.GetUtxosForAddresses(addresses)
	if err != nil {
		sendMessage(canceled, chMessages, WalletRestoreMessage{Err: err})
		return
	}

	accountsBalance := make(map[string]map[string]uint64)
	for i := range utxos {
		utxo := utxos[i]
		if utxo.IsSpent() {
			continue
		}
		utxos[i].AccountName = accountByScript[hex.EncodeToString(utxo.Script)]
		if _, ok := accountsBalance[utxos[i].AccountName]; !ok {
			accountsBalance[utxos[i].AccountName] = make(map[string]uint64)
		}
		accountsBalance[utxos[i].AccountName][utxo.Asset] += utxo.Value
	}

	count, err := ws.repoManager.UtxoRepository().AddUtxos(context.Background(), utxos)
	if err != nil {
		sendMessage(canceled, chMessages, WalletRestoreMessage{Err: err})
		return
	}
	if count > 0 {
		ws.log("added %d utxo(s)", count)
	}
	if !sendMessage(canceled, chMessages, WalletRestoreMessage{
		Message: "restored wallet utxo pool",
	}) {
		return
	}

	ws.setSynced()

	sendMessage(canceled, chMessages, WalletRestoreMessage{
		Message: "wallet restored",
	})
}

func (ws *WalletService) GetStatus(_ context.Context) WalletStatus {
	return WalletStatus{
		IsInitialized: ws.isInitialized(),
		IsUnlocked:    ws.isUnlocked(),
		IsSynced:      ws.isSynced(),
	}
}

func (ws *WalletService) GetInfo(ctx context.Context) (*WalletInfo, error) {
	w, _ := ws.repoManager.WalletRepository().GetWallet(ctx)

	if w == nil || w.IsLocked() {
		return &WalletInfo{
			Network:     ws.network.Name,
			NativeAsset: ws.network.AssetID,
			BuildInfo:   ws.buildInfo,
		}, nil
	}

	birthdayBlock, _ := ws.bcScanner.GetBlockHash(w.BirthdayBlockHeight)
	accounts := make([]AccountInfo, 0, len(w.Accounts))
	for _, a := range w.Accounts {
		accounts = append(accounts, AccountInfo{a.AccountInfo})
	}
	return &WalletInfo{
		Network:             w.NetworkName,
		NativeAsset:         ws.network.AssetID,
		RootPath:            w.RootPath,
		BirthdayBlockHash:   elementsutil.TxIDFromBytes(birthdayBlock),
		BirthdayBlockHeight: w.BirthdayBlockHeight,
		Accounts:            accounts,
		BuildInfo:           ws.buildInfo,
	}, nil
}

func (ws *WalletService) Auth(
	ctx context.Context,
	password string,
) (bool, error) {
	wallet, err := ws.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return false, err
	}

	return wallet.IsValidPassword(password), nil
}

func (ws *WalletService) RegisterHandlerForWalletEvent(
	eventType domain.WalletEventType, handler ports.WalletEventHandler,
) {
	ws.repoManager.RegisterHandlerForWalletEvent(eventType, handler)
}

func (ws *WalletService) setInitialized() {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	ws.initialized = true
}

func (ws *WalletService) setNotInitialized() {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	ws.initialized = false
}

func (ws *WalletService) isInitialized() bool {
	ws.lock.RLock()
	defer ws.lock.RUnlock()

	return ws.initialized
}

func (ws *WalletService) setUnlocked() {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	ws.unlocked = true
}

func (ws *WalletService) setLocked() {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	ws.unlocked = false
}

func (ws *WalletService) isUnlocked() bool {
	ws.lock.RLock()
	defer ws.lock.RUnlock()

	return ws.unlocked
}

func (ws *WalletService) setSynced() {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	ws.synced = true
}

func (ws *WalletService) isSynced() bool {
	ws.lock.RLock()
	defer ws.lock.RUnlock()

	return ws.synced
}

func sendMessage(
	canceled bool, ch chan WalletRestoreMessage, msg WalletRestoreMessage,
) bool {
	if canceled {
		return false
	}
	ch <- msg
	return true
}

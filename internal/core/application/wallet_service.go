package application

import (
	"context"
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/vulpemventures/go-elements/block"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	wallet "github.com/vulpemventures/ocean/pkg/single-key-wallet"
)

// WalletService is responsible for operations related to the managment of the
// wallet:
// 	* Generate a new random 24-words mnemonic.
// 	* Create a new wallet from scratch with given mnemonic and locked with the given password.
// 	* Unlock the wallet with a password.
// 	* Change the wallet password. It requires the wallet to be locked.
// 	* Get the status of the wallet (initialized, unlocked, inSync).
// 	* Get non-sensiive (network, native asset) and possibly sensitive info (root path, master blinding key and basic accounts' info) about the wallet. Sensitive info are returned only if the wallet is unlocked.
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

	initialized bool
	unlocked    bool
	synced      bool
	lock        *sync.RWMutex
}

func NewWalletService(
	repoManager ports.RepoManager, bcScanner ports.BlockchainScanner,
	rootPath string, net *network.Network,
) *WalletService {
	ws := &WalletService{
		repoManager: repoManager,
		bcScanner:   bcScanner,
		rootPath:    rootPath,
		network:     net,
		lock:        &sync.RWMutex{},
	}
	w, _ := ws.repoManager.WalletRepository().GetWallet(context.Background())
	if w != nil {
		ws.setInitialized()
		ws.setSynced()
	}
	return ws
}

func (ws *WalletService) GenSeed(ctx context.Context) ([]string, error) {
	return wallet.NewMnemonic(wallet.NewMnemonicArgs{})
}

func (ws *WalletService) CreateWallet(
	ctx context.Context, mnemonic []string, passpharse string,
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

	birthdayBlock, err := ws.bcScanner.GetLatestBlock()
	if err != nil {
		return
	}

	newWallet, err := domain.NewWallet(
		mnemonic, passpharse, ws.rootPath, ws.network.Name,
		birthdayBlock.Height, nil,
	)
	if err != nil {
		return
	}
	newWallet.Lock()

	return ws.repoManager.WalletRepository().CreateWallet(ctx, newWallet)
}

func (ws *WalletService) Unlock(
	ctx context.Context, password string,
) (err error) {
	defer func() {
		if err == nil {
			ws.setUnlocked()
		}
	}()

	return ws.repoManager.WalletRepository().UnlockWallet(ctx, password)
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
	ctx context.Context, mnemonic []string, passpharse string,
	birthdayBlockHash []byte,
) (err error) {
	defer func() {
		if err == nil {
			ws.setInitialized()
			ws.setSynced()
		}
	}()

	birthdayBlock, err := ws.getBlockByHash(birthdayBlockHash)
	if err != nil {
		return
	}
	// TODO: implement restoration

	newWallet, err := domain.NewWallet(
		mnemonic, passpharse, ws.rootPath, ws.network.Name,
		birthdayBlock.Height, nil,
	)
	if err != nil {
		return
	}

	return ws.repoManager.WalletRepository().CreateWallet(ctx, newWallet)
}

func (ws *WalletService) GetStatus(_ context.Context) WalletStatus {
	return WalletStatus{
		IsInitialized: ws.isInitialized(),
		IsUnlocked:    ws.isUnlocked(),
		IsSynced:      ws.isSynced(),
	}
}

func (ws *WalletService) GetInfo(ctx context.Context) (*WalletInfo, error) {
	w, err := ws.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}
	if w.IsLocked() {
		return &WalletInfo{
			Network:     w.NetworkName,
			NativeAsset: ws.network.AssetID,
		}, nil
	}

	birthdayBlock, _ := ws.getBlockByHeight(w.BirthdayBlockHeight)
	var birthdayBlockHash string
	var birthdayBlockHeight uint32
	if birthdayBlock != nil {
		if hash, err := birthdayBlock.Hash(); err == nil {
			birthdayBlockHash = hash.String()
		}
		birthdayBlockHeight = birthdayBlock.Height
	}
	masterBlingingKey, _ := w.GetMasterBlindingKey()
	accounts := make([]AccountInfo, 0, len(w.AccountsByKey))
	for _, a := range w.AccountsByKey {
		accounts = append(accounts, AccountInfo(a.Info))
	}
	return &WalletInfo{
		Network:             w.NetworkName,
		NativeAsset:         ws.network.AssetID,
		RootPath:            w.RootPath,
		MasterBlindingKey:   masterBlingingKey,
		BirthdayBlockHash:   birthdayBlockHash,
		BirthdayBlockHeight: birthdayBlockHeight,
		Accounts:            accounts,
	}, nil
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

func (ws *WalletService) getBlockByHash(blockHash []byte) (*block.Header, error) {
	hash, err := chainhash.NewHash(blockHash)
	if err != nil {
		return nil, err
	}
	return ws.bcScanner.GetBlockHeader(*hash)
}

func (ws *WalletService) getBlockByHeight(blockHeight uint32) (*block.Header, error) {
	hash, err := ws.bcScanner.GetBlockHash(blockHeight)
	if err != nil {
		return nil, err
	}
	return ws.bcScanner.GetBlockHeader(*hash)
}

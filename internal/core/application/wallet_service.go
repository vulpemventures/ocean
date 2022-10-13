package application

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/neutrino-elements/pkg/scanner"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vulpemventures/go-elements/elementsutil"
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
	buildInfo   BuildInfo

	initialized bool
	unlocked    bool
	synced      bool
	lock        *sync.RWMutex
}

func NewWalletService(
	repoManager ports.RepoManager, bcScanner ports.BlockchainScanner,
	rootPath string, net *network.Network, buildInfo BuildInfo,
) *WalletService {
	ws := &WalletService{
		repoManager: repoManager,
		bcScanner:   bcScanner,
		rootPath:    rootPath,
		network:     net,
		buildInfo:   buildInfo,
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

	_, birthdayBlockHeight, err := ws.bcScanner.GetLatestBlock()
	if err != nil {
		return
	}

	newWallet, err := domain.NewWallet(
		mnemonic, passpharse, ws.rootPath, ws.network.Name,
		birthdayBlockHeight, nil,
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

func (ws *WalletService) Lock(ctx context.Context) (err error) {
	if !ws.isUnlocked() {
		return nil
	}

	defer func() {
		if err == nil {
			ws.setLocked()
		}
	}()

	return ws.repoManager.WalletRepository().LockWallet(ctx)
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

	birthdayBlockHeight, err := ws.bcScanner.GetBlockHeight(birthdayBlockHash)
	if err != nil {
		return
	}
	// TODO: implement restoration

	newWallet, err := domain.NewWallet(
		mnemonic, passpharse, ws.rootPath, ws.network.Name,
		birthdayBlockHeight, nil,
	)
	if err != nil {
		return
	}

	return ws.repoManager.WalletRepository().CreateWallet(ctx, newWallet)
}

func (ws *WalletService) restore(
	wallet domain.Wallet,
	numOfAccounts int,
	numOfAddressesPerAccount int,
	birthdayBlock uint32,
	recoveryWindow int,
) ([]domain.Utxo, error) {
	result := make([]domain.Utxo, 0)
accountLoop:
	for i := 0; i < numOfAccounts; i++ {
		accountName := strconv.Itoa(i + 1)
		account, err := wallet.CreateAccount(accountName, birthdayBlock)
		if err != nil {
			return nil, err
		}

		accountAddresses := make([]domain.AddressInfo, 0, 100)
		blindingKeyPerScript := make(map[string][]byte)
		loopExternal := true
		loopInternal := true
		externalAddrNotFoundDelta := 0
		internalAddrNotFoundDelta := 0
		addressLoopCount := 1
	addressLoop:
		for ii := 0; ii < numOfAddressesPerAccount; i++ {
			if loopExternal {
				externalAddresses, err := wallet.DeriveNextExternalAddressForAccount(accountName)
				if err != nil {
					return nil, err
				}

				blindingKeyPerScript[externalAddresses.Script] = externalAddresses.BlindingKey
				accountAddresses = append(accountAddresses, *externalAddresses)
			}

			if loopInternal {
				internalAddresses, err := wallet.DeriveNextExternalAddressForAccount(accountName)
				if err != nil {
					return nil, err
				}

				blindingKeyPerScript[internalAddresses.Script] = internalAddresses.BlindingKey
				accountAddresses = append(accountAddresses, *internalAddresses)
			}
		}

		report := ws.bcScanner.WatchAddressesForAccount(
			accountName,
			birthdayBlock,
			accountAddresses,
		)

		chainLastIndex := make(map[int]int)
		select {
		case r := <-report:
			if derivationPath, ok := account.DerivationPathByScript[hex.EncodeToString(r.Request.Item.Bytes())]; ok {
				path := strings.Split(derivationPath, "/")
				chainStr := path[1]
				chain, err := strconv.Atoi(chainStr)
				if err != nil {
					return nil, err
				}

				indexStr := path[2]
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					return nil, err
				}

				if lastIndex, ok := chainLastIndex[chain]; !ok {
					chainLastIndex[chain] = index
				} else {
					if index > lastIndex {
						chainLastIndex[chain] = index
					}
				}
			}

			newUtxos, err := getUtxos(r, accountName, blindingKeyPerScript)
			if err != nil {
				return nil, err
			}
			result = append(result, newUtxos...)

			//we would need to refactor scanner in order to know if scanning is done
			//till that we set timeout
		case <-time.After(20 * time.Second):
			goto accountLoop
		}

		externalAddrNotFoundDelta += addressLoopCount*numOfAddressesPerAccount - chainLastIndex[domain.ExternalChain]
		internalAddrNotFoundDelta += addressLoopCount*numOfAddressesPerAccount - chainLastIndex[domain.InternalChain]
		loopExternal = externalAddrNotFoundDelta < recoveryWindow
		loopInternal = internalAddrNotFoundDelta < recoveryWindow
		addressLoopCount++
		if loopInternal || loopExternal {
			goto addressLoop
		}
	}

	return result, nil
}

func getUtxos(
	r scanner.Report,
	accountName string,
	blindingKeyPerScript map[string][]byte,
) ([]domain.Utxo, error) {
	result := make([]domain.Utxo, 0)
	tx := r.Transaction
	txid := tx.TxHash().String()
	var blockHash string
	var blockHeight uint64
	if r.BlockHash != nil {
		blockHash = r.BlockHash.String()
		blockHeight = uint64(r.BlockHeight)
	}

	for i, out := range tx.Outputs {
		if len(out.Script) == 0 {
			continue
		}

		script := hex.EncodeToString(out.Script)
		blindingKey, ok := blindingKeyPerScript[script]
		if !ok {
			continue
		}

		revealed, err := confidential.UnblindOutputWithKey(out, blindingKey)
		if err != nil {
			return nil, err
		}

		var assetCommitment, valueCommitment []byte
		if out.IsConfidential() {
			valueCommitment, assetCommitment = out.Value, out.Asset
		}

		result = append(result, domain.Utxo{
			UtxoKey: domain.UtxoKey{
				TxID: txid,
				VOut: uint32(i),
			},
			Value:           revealed.Value,
			Asset:           assetFromBytes(revealed.Asset),
			ValueCommitment: valueCommitment,
			AssetCommitment: assetCommitment,
			ValueBlinder:    revealed.ValueBlindingFactor,
			AssetBlinder:    revealed.AssetBlindingFactor,
			Script:          out.Script,
			Nonce:           out.Nonce,
			RangeProof:      out.RangeProof,
			SurjectionProof: out.SurjectionProof,
			AccountName:     accountName,
			ConfirmedStatus: domain.UtxoStatus{
				BlockHeight: blockHeight,
				BlockHash:   blockHash,
			},
		})
	}

	return result, nil
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
		BirthdayBlockHash:   elementsutil.TxIDFromBytes(birthdayBlock),
		BirthdayBlockHeight: w.BirthdayBlockHeight,
		Accounts:            accounts,
		BuildInfo:           ws.buildInfo,
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

func assetFromBytes(buf []byte) string {
	return hex.EncodeToString(elementsutil.ReverseBytes(buf))
}

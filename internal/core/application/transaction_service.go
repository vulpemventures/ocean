package application

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"github.com/btcsuite/btcd/txscript"
	log "github.com/sirupsen/logrus"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/psetv2"
	"github.com/vulpemventures/go-elements/transaction"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	wallet "github.com/vulpemventures/ocean/pkg/single-key-wallet"
)

var (
	ErrForbiddenUnlockedInputs = fmt.Errorf(
		"the utxos used within 'external' transactions must be coming from a " +
			"wallet's coin selection so that they can be temporary locked and " +
			"prevent to accidentally double spending them",
	)
)

// TransactionService is responsible for operations related to one or more
// accounts:
// 	* Get info about a wallet-related transaction.
// 	* Select a subset of the utxos of an existing account to cover a target amount. The selected utxos will be temporary locked to prevent double spending them.
// 	* Estimate the fee amount for a transation composed by X inputs and Y outputs. It is required that the inputs owned by the wallet are locked utxos.
// 	* Sign a raw transaction (in hex format). It is required that the inputs of the tx owned by the wallet are locked utxos.
// 	* Broadcast a raw transaction (in hex format). It is required that the inputs of the tx owned by the wallet are locked utxos.
// 	* Create a partial transaction (v2) given a list of inputs and outputs. It is required that the inputs of the tx owned by the wallet are locked utxos.
// 	* Add inputs or outputs to partial transaction (v2). It is required that the inputs of the tx owned by the wallet are locked utxos.
// 	* Blind a partial transaction (v2) either as non-last or last blinder. It is required that the inputs of the tx owned by the wallet are locked utxos.
// 	* Sign a partial transaction (v2). It is required that the inputs of the tx owned by the wallet are locked utxos.
// 	* Craft a finalized transaction to transfer some funds from an existing account to somewhere else, given a list of outputs.
//
// The service registers 1 handler for the following utxo event:
// 	* domain.UtxoLocked - whenever one or more utxos are locked, the service spawns a so-called unlocker, a goroutine wating for X seconds before unlocking them if necessary. The operation is just skipped if the utxos have been spent meanwhile.
//
// The service guarantees that any locked utxo is eventually unlocked ASAP
// after the waiting time expires.
// Therefore, at startup, it makes sure to unlock any still-locked utxo that
// can be unlocked, and to spawn the required numnber of unlockers for those
// whose waiting time didn't expire yet.
type TransactionService struct {
	repoManager        ports.RepoManager
	bcScanner          ports.BlockchainScanner
	network            *network.Network
	rootPath           string
	utxoExpiryDuration time.Duration

	log func(format string, a ...interface{})
}

func NewTransactionService(
	repoManager ports.RepoManager, bcScanner ports.BlockchainScanner,
	net *network.Network, rootPath string, utxoExpiryDuration time.Duration,
) *TransactionService {
	logFn := func(format string, a ...interface{}) {
		format = fmt.Sprintf("transaction service: %s", format)
		log.Debugf(format, a...)
	}
	svc := &TransactionService{
		repoManager, bcScanner, net, rootPath, utxoExpiryDuration, logFn,
	}
	svc.registerHandlerForUtxoEvents()

	ctx := context.Background()
	w, _ := repoManager.WalletRepository().GetWallet(ctx)
	if w == nil {
		return svc
	}

	for accountName := range w.AccountKeysByName {
		utxos, _ := repoManager.UtxoRepository().GetLockedUtxosForAccount(
			ctx, accountName,
		)
		if len(utxos) > 0 {
			utxosToUnlock := make([]domain.UtxoKey, 0, len(utxos))
			utxosToSpawnUnlocker := make([]domain.UtxoKey, 0, len(utxos))
			for _, u := range utxos {
				if u.CanUnlock() {
					utxosToUnlock = append(utxosToUnlock, u.Key())
				} else {
					utxosToSpawnUnlocker = append(utxosToSpawnUnlocker, u.Key())
				}
			}

			if len(utxosToUnlock) > 0 {
				count, _ := repoManager.UtxoRepository().UnlockUtxos(ctx, utxosToUnlock)
				svc.log(
					"unlocked %d utxo(s) for account %s (%s)",
					count, accountName, UtxoKeys(utxosToUnlock),
				)
			}
			if len(utxosToSpawnUnlocker) > 0 {
				svc.spawnUtxoUnlocker(utxosToSpawnUnlocker)
			}
		}
	}

	return svc
}

func (ts *TransactionService) GetTransactionInfo(
	ctx context.Context, txid string,
) (*TransactionInfo, error) {
	tx, err := ts.repoManager.TransactionRepository().GetTransaction(ctx, txid)
	if err != nil {
		return nil, err
	}
	return (*TransactionInfo)(tx), nil
}

func (ts *TransactionService) SelectUtxos(
	ctx context.Context, accountName, targetAsset string, targetAmount uint64,
	coinSelectionStrategy int,
) (Utxos, uint64, int64, error) {
	if _, err := ts.getAccount(ctx, accountName); err != nil {
		return nil, 0, -1, err
	}

	utxos, err := ts.repoManager.UtxoRepository().GetSpendableUtxosForAccount(
		ctx, accountName,
	)
	if err != nil {
		return nil, 0, -1, err
	}

	coinSelector := DefaultCoinSelector
	if factory, ok := coinSelectorByType[coinSelectionStrategy]; ok {
		coinSelector = factory()
	}

	utxos, change, err := coinSelector.SelectUtxos(utxos, targetAmount, targetAsset)
	if err != nil {
		return nil, 0, -1, err
	}
	now := time.Now().Unix()
	keys := Utxos(utxos).Keys()
	count, err := ts.repoManager.UtxoRepository().LockUtxos(
		ctx, keys, now,
	)
	if err != nil {
		return nil, 0, -1, err
	}
	if count > 0 {
		ts.log(
			"locked %d utxo(s) for account %s (%s)",
			count, accountName, UtxoKeys(keys),
		)
	}

	expirationDate := time.Now().Add(ts.utxoExpiryDuration).Unix()
	return utxos, change, expirationDate, nil
}

func (ts *TransactionService) EstimateFees(
	ctx context.Context, ins Inputs, outs Outputs, millisatsPerByte uint64,
) (uint64, error) {
	if _, err := ts.getWallet(ctx); err != nil {
		return 0, err
	}

	walletInputs, err := ts.getLockedInputs(ctx, ins)
	if err != nil {
		return 0, err
	}
	externalInputs, err := ts.getExternalInputs(walletInputs, ins)
	if err != nil {
		return 0, err
	}

	if millisatsPerByte == 0 {
		millisatsPerByte = MinMillisatsPerByte
	}

	inputs := append(walletInputs, externalInputs...)
	return wallet.EstimateFees(
		inputs, outs.toWalletOutputs(), millisatsPerByte,
	), nil
}

func (ts *TransactionService) SignTransaction(
	ctx context.Context, txHex string, sighashType uint32,
) (string, error) {
	w, err := ts.getWallet(ctx)
	if err != nil {
		return "", err
	}

	inputs, err := ts.findLockedInputs(ctx, txHex)
	if err != nil {
		return "", err
	}

	return w.SignTransaction(wallet.SignTransactionArgs{
		TxHex:        txHex,
		InputsToSign: inputs,
		SigHashType:  txscript.SigHashType(sighashType),
	})
}

func (ts *TransactionService) BroadcastTransaction(
	ctx context.Context, txHex string,
) (string, error) {
	keys := utxoKeysFromRawTx(txHex)
	utxos, err := ts.repoManager.UtxoRepository().GetUtxosByKey(ctx, keys)
	if err != nil {
		return "", err
	}
	if len(utxos) > 0 {
		for _, u := range utxos {
			if !u.IsLocked() {
				return "", fmt.Errorf(
					"cannot broadcast transaction containing unlocked utxos",
				)
			}
		}
	}

	return ts.bcScanner.BroadcastTransaction(txHex)
}

func (ts *TransactionService) CreatePset(
	ctx context.Context, inputs Inputs, outputs Outputs,
) (string, error) {
	w, err := ts.getWallet(ctx)
	if err != nil {
		return "", err
	}

	walletInputs, err := ts.getLockedInputs(ctx, inputs)
	if err != nil {
		return "", err
	}

	return w.CreatePset(wallet.CreatePsetArgs{
		Inputs:  walletInputs,
		Outputs: outputs.toWalletOutputs(),
	})
}

func (ts *TransactionService) UpdatePset(
	ctx context.Context, ptx string, inputs Inputs, outputs Outputs,
) (string, error) {
	w, err := ts.getWallet(ctx)
	if err != nil {
		return "", err
	}

	walletInputs, err := ts.getLockedInputs(ctx, inputs)
	if err != nil {
		return "", err
	}

	return w.UpdatePset(wallet.UpdatePsetArgs{
		PsetBase64: ptx,
		Inputs:     walletInputs,
		Outputs:    outputs.toWalletOutputs(),
	})
}

func (ts *TransactionService) BlindPset(
	ctx context.Context, ptx string, extraBlindingKeys map[string][]byte,
	lastBlinder bool,
) (string, error) {
	w, err := ts.getWallet(ctx)
	if err != nil {
		return "", err
	}

	walletInputs, err := ts.findLockedInputs(ctx, ptx)
	if err != nil {
		return "", err
	}

	return w.BlindPsetWithOwnedInputs(
		wallet.BlindPsetWithOwnedInputsArgs{
			PsetBase64:         ptx,
			OwnedInputsByIndex: walletInputs,
			ExtraBlindingKeys:  extraBlindingKeys,
			LastBlinder:        lastBlinder,
		},
	)
}

func (ts *TransactionService) SignPset(
	ctx context.Context, ptx string, sighashType uint32,
) (string, error) {
	w, err := ts.getWallet(ctx)
	if err != nil {
		return "", err
	}

	walletInputs, err := ts.findLockedInputs(ctx, ptx)
	if err != nil {
		return "", err
	}
	derivationPaths := make(map[string]string)
	for _, in := range walletInputs {
		script := hex.EncodeToString(in.Script)
		derivationPaths[script] = in.DerivationPath
	}

	return w.SignPset(wallet.SignPsetArgs{
		PsetBase64:        ptx,
		DerivationPathMap: derivationPaths,
		SigHashType:       txscript.SigHashType(sighashType),
	})
}

func (ts *TransactionService) Transfer(
	ctx context.Context, accountName string, outputs Outputs,
	millisatsPerByte uint64,
) (string, error) {
	w, err := ts.getWallet(ctx)
	if err != nil {
		return "", err
	}
	account, err := ts.getAccount(ctx, accountName)
	if err != nil {
		return "", err
	}

	utxoRepo := ts.repoManager.UtxoRepository()
	walletRepo := ts.repoManager.WalletRepository()

	utxos, err := utxoRepo.GetSpendableUtxosForAccount(
		ctx, accountName,
	)
	if err != nil {
		return "", err
	}
	if len(utxos) == 0 {
		return "", fmt.Errorf("no utxos found for account %s", accountName)
	}

	changeByAsset := make(map[string]uint64)
	selectedUtxos := make([]*domain.Utxo, 0)
	for targetAsset, targetAmount := range outputs.totalAmountByAsset() {
		utxos, change, err := DefaultCoinSelector.SelectUtxos(utxos, targetAmount, targetAsset)
		if err != nil {
			return "", err
		}
		changeByAsset[targetAsset] = change
		selectedUtxos = append(selectedUtxos, utxos...)
	}

	inputs := make([]wallet.Input, 0, len(utxos))
	inputsByIndex := make(map[uint32]wallet.Input)
	for i, u := range selectedUtxos {
		input := wallet.Input{
			TxID:            u.TxID,
			TxIndex:         u.VOut,
			Value:           u.Value,
			Asset:           u.Asset,
			Script:          u.Script,
			ValueBlinder:    u.ValueBlinder,
			AssetBlinder:    u.AssetBlinder,
			ValueCommitment: u.ValueCommitment,
			AssetCommitment: u.AssetCommitment,
			Nonce:           u.Nonce,
		}
		inputs = append(inputs, input)
		inputsByIndex[uint32(i)] = input
	}

	changeOutputs := make([]wallet.Output, 0)
	if len(changeByAsset) > 0 {
		addressesInfo, err := walletRepo.DeriveNextInternalAddressesForAccount(
			ctx, accountName, uint64(len(changeByAsset)),
		)
		if err != nil {
			return "", err
		}

		i := 0
		for asset, amount := range changeByAsset {
			changeOutputs = append(changeOutputs, wallet.Output{
				Asset:   asset,
				Amount:  amount,
				Address: addressesInfo[i].Address,
			})
			i++
		}
	}

	outs := outputs.toWalletOutputs()
	feeAmount := wallet.EstimateFees(
		inputs, append(outs, changeOutputs...), millisatsPerByte,
	)

	// If feeAmount is lower than the lbtc change, it's enough to deduct it
	// from the change amount.
	if feeAmount < changeByAsset[ts.network.AssetID] {
		for i, out := range changeOutputs {
			if out.Asset == ts.network.AssetID {
				changeOutputs[i].Amount -= feeAmount
				break
			}
		}
	}
	// If feeAmount is exactly the lbtc change, it's enough to remove the
	// change output.
	if feeAmount == changeByAsset[ts.network.AssetID] {
		var outIndex int
		for i, out := range changeOutputs {
			if out.Asset == ts.network.AssetID {
				outIndex = i
				break
			}
		}
		changeOutputs = append(
			changeOutputs[:outIndex], changeOutputs[outIndex+1:]...,
		)
	}
	// If feeAmount is greater than the lbtc change, another coin-selection round
	// is required.
	if feeAmount > changeByAsset[ts.network.AssetID] {
		targetAsset := ts.network.AssetID
		targetAmount := wallet.DummyFeeAmount
		if feeAmount > targetAmount {
			targetAmount = roundUpAmount(feeAmount)
		}

		// Coin-selection must be done over remaining utxos.
		remainingUtxos := getRemainingUtxos(utxos, selectedUtxos)
		selectedUtxos, change, err := DefaultCoinSelector.SelectUtxos(
			remainingUtxos, targetAmount, targetAsset,
		)
		if err != nil {
			return "", err
		}

		for _, u := range selectedUtxos {
			input := wallet.Input{
				TxID:            u.TxID,
				TxIndex:         u.VOut,
				Value:           u.Value,
				Asset:           u.Asset,
				Script:          u.Script,
				ValueBlinder:    u.ValueBlinder,
				AssetBlinder:    u.AssetBlinder,
				ValueCommitment: u.ValueCommitment,
				AssetCommitment: u.AssetCommitment,
				Nonce:           u.Nonce,
			}
			inputs = append(inputs, input)
			inputsByIndex[uint32(len(inputs))] = input
		}

		if change > 0 {
			// For the eventual change amount, it might be necessary to add a lbtc
			// change output to the list if it's still not in the list.
			if _, ok := changeByAsset[targetAsset]; !ok {
				addrInfo, err := walletRepo.DeriveNextInternalAddressesForAccount(
					ctx, accountName, 1,
				)
				if err != nil {
					return "", err
				}
				changeOutputs = append(changeOutputs, wallet.Output{
					Amount:  change,
					Asset:   targetAsset,
					Address: addrInfo[0].Address,
				})
			}

			// Now that we have all inputs and outputs, estimate the real fee amount.
			feeAmount = wallet.EstimateFees(
				inputs, append(outs, changeOutputs...), millisatsPerByte,
			)

			// Update the change amount by adding the delta
			// delta = targetAmount - feeAmount.
			for i, out := range changeOutputs {
				if out.Asset == targetAsset {
					// This way the delta is subtracted in case it's negative.
					changeOutputs[i].Amount = uint64(int(out.Amount) + int(targetAmount) - int(feeAmount))
					break
				}
			}
		}
	}

	outs = append(outs, changeOutputs...)
	outs = append(outs, wallet.Output{
		Asset:  ts.network.AssetID,
		Amount: feeAmount,
	})

	ptx, err := w.CreatePset(wallet.CreatePsetArgs{
		Inputs:  inputs,
		Outputs: outs,
	})
	if err != nil {
		return "", err
	}

	blindedPtx, err := w.BlindPsetWithOwnedInputs(
		wallet.BlindPsetWithOwnedInputsArgs{
			PsetBase64:         ptx,
			OwnedInputsByIndex: inputsByIndex,
			LastBlinder:        true,
		},
	)
	if err != nil {
		return "", err
	}

	signedPtx, err := w.SignPset(wallet.SignPsetArgs{
		PsetBase64:        blindedPtx,
		DerivationPathMap: account.DerivationPathByScript,
	})
	if err != nil {
		return "", err
	}

	txHex, _, err := wallet.FinalizeAndExtractTransaction(wallet.FinalizeAndExtractTransactionArgs{
		PsetBase64: signedPtx,
	})
	if err != nil {
		return "", err
	}

	keys := Utxos(selectedUtxos).Keys()
	now := time.Now().Unix()
	count, err := ts.repoManager.UtxoRepository().LockUtxos(ctx, keys, now)
	if err != nil {
		return "", err
	}
	if count > 0 {
		ts.log(
			"locked %d utxo(s) for account %s (%s) ",
			count, accountName, UtxoKeys(keys),
		)
	}

	return txHex, nil
}

func (ts *TransactionService) registerHandlerForUtxoEvents() {
	ts.repoManager.RegisterHandlerForUtxoEvent(
		domain.UtxoLocked, func(event domain.UtxoEvent) {
			keys := UtxosInfo(event.Utxos).Keys()
			ts.spawnUtxoUnlocker(keys)
		},
	)
}

// spawnUtxoUnlocker groups the locked utxos identified by the given keys by their
// locking timestamps, and then creates a goroutine for each group in order to unlock
// the utxos if they are still locked when their expiration time comes.
func (ts *TransactionService) spawnUtxoUnlocker(utxoKeys []domain.UtxoKey) {
	ctx := context.Background()
	utxos, _ := ts.repoManager.UtxoRepository().GetUtxosByKey(ctx, utxoKeys)

	utxosByLockTimestamp := make(map[int64][]domain.UtxoKey)
	for _, u := range utxos {
		utxosByLockTimestamp[u.LockTimestamp] = append(
			utxosByLockTimestamp[u.LockTimestamp], u.Key(),
		)
	}

	for timestamp := range utxosByLockTimestamp {
		keys := utxosByLockTimestamp[timestamp]
		unlockTime := ts.utxoExpiryDuration - time.Since(time.Unix(timestamp, 0))
		t := time.NewTicker(unlockTime)
		go func() {
			ts.log("spawning unlocker for utxo(s) %s", UtxoKeys(keys))
			ts.log(fmt.Sprintf(
				"utxo(s) will be eventually unlocked in ~%.0f seconds",
				math.Round(unlockTime.Seconds()/10)*10,
			))

			for range t.C {
				utxos, _ := ts.repoManager.UtxoRepository().GetUtxosByKey(ctx, keys)
				utxosToUnlock := make([]domain.UtxoKey, 0, len(utxos))
				spentUtxos := make([]domain.UtxoKey, 0, len(utxos))
				for _, u := range utxos {
					if !u.IsSpent() && u.IsLocked() {
						utxosToUnlock = append(utxosToUnlock, u.Key())
					} else {
						spentUtxos = append(spentUtxos, u.Key())
					}
				}

				if len(utxosToUnlock) > 0 {
					// In case of errors here, the ticker is possibly reset to a shortest
					// duration to keep retring to unlock the locked utxos as soon as
					// possible.
					count, err := ts.repoManager.UtxoRepository().UnlockUtxos(
						ctx, utxosToUnlock,
					)
					if err != nil {
						shortDuration := 5 * time.Second
						if shortDuration < unlockTime {
							t.Reset(shortDuration)
						}
						continue
					}

					if count > 0 {
						ts.log("unlocked %d utxo(s) %s", count, UtxoKeys(keys))
					}
					t.Stop()
				}
				if len(spentUtxos) > 0 {
					ts.log(
						"utxo(s) %s have been spent, skipping unlocking",
						UtxoKeys(spentUtxos),
					)
				}
			}
		}()
	}
}

func (ts *TransactionService) getWallet(
	ctx context.Context,
) (*wallet.Wallet, error) {
	w, err := ts.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}
	mnemonic, err := w.GetMnemonic()
	if err != nil {
		return nil, err
	}

	return wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicArgs{
		RootPath: ts.rootPath,
		Mnemonic: mnemonic,
	})
}

func (ts *TransactionService) getAccount(
	ctx context.Context, accountName string,
) (*domain.Account, error) {
	w, err := ts.repoManager.WalletRepository().GetWallet(ctx)
	if err != nil {
		return nil, err
	}
	return w.GetAccount(accountName)
}

func (ts *TransactionService) getLockedInputs(
	ctx context.Context, ins Inputs,
) ([]wallet.Input, error) {
	keys := make([]domain.UtxoKey, 0, len(ins))
	for _, in := range ins {
		keys = append(keys, domain.UtxoKey(in))
	}
	utxos, err := ts.repoManager.UtxoRepository().GetUtxosByKey(ctx, keys)
	if err != nil {
		return nil, err
	}
	if len(utxos) == 0 {
		return nil, fmt.Errorf("no utxos found with given keys")
	}

	w, _ := ts.repoManager.WalletRepository().GetWallet(ctx)
	inputs := make([]wallet.Input, 0, len(utxos))
	for _, u := range utxos {
		if !u.IsLocked() {
			return nil, ErrForbiddenUnlockedInputs
		}

		account, _ := w.GetAccount(u.AccountName)
		script := hex.EncodeToString(u.Script)
		derivationPath := account.DerivationPathByScript[script]

		inputs = append(inputs, wallet.Input{
			TxID:            u.TxID,
			TxIndex:         u.VOut,
			Value:           u.Value,
			Asset:           u.Asset,
			Script:          u.Script,
			ValueBlinder:    u.ValueBlinder,
			AssetBlinder:    u.AssetBlinder,
			ValueCommitment: u.ValueCommitment,
			AssetCommitment: u.AssetCommitment,
			Nonce:           u.Nonce,
			RangeProof:      u.RangeProof,
			SurjectionProof: u.SurjectionProof,
			DerivationPath:  derivationPath,
		})
	}

	return inputs, nil
}

func (ts *TransactionService) findLockedInputs(
	ctx context.Context, tx string,
) (map[uint32]wallet.Input, error) {
	rawTx, _ := transaction.NewTxFromHex(tx)
	var keys = make([]domain.UtxoKey, 0)
	if rawTx != nil {
		keys = utxoKeysFromRawTx(tx)
	} else {
		keys = utxoKeysFromPartialTx(tx)
	}

	utxos, err := ts.repoManager.UtxoRepository().GetUtxosByKey(ctx, keys)
	if err != nil {
		return nil, err
	}
	if len(utxos) == 0 {
		return nil, fmt.Errorf("no wallet utxos found in given transaction")
	}

	w, _ := ts.repoManager.WalletRepository().GetWallet(ctx)

	inputs := make(map[uint32]wallet.Input)
	for _, u := range utxos {
		if !u.IsLocked() {
			return nil, fmt.Errorf(
				"cannot use unlocked utxos. The utxos used within 'external' " +
					"transactions must be coming from a coin selection so that they " +
					"can be locked to prevent double spending them",
			)

		}
		account, _ := w.GetAccount(u.AccountName)
		script := hex.EncodeToString(u.Script)
		inIndex := findUtxoIndexInTx(tx, u)
		inputs[inIndex] = wallet.Input{
			TxID:            u.TxID,
			TxIndex:         u.VOut,
			Value:           u.Value,
			Asset:           u.Asset,
			Script:          u.Script,
			ValueBlinder:    u.ValueBlinder,
			AssetBlinder:    u.AssetBlinder,
			ValueCommitment: u.ValueCommitment,
			AssetCommitment: u.AssetCommitment,
			Nonce:           u.Nonce,
			DerivationPath:  account.DerivationPathByScript[script],
		}
	}
	return inputs, nil
}

func (ts *TransactionService) getExternalInputs(
	walletIns []wallet.Input, allIns Inputs,
) ([]wallet.Input, error) {
	externalInputKeys := make([]domain.UtxoKey, 0)
	for _, key := range allIns {
		isExternalInput := true
		for _, in := range walletIns {
			if in.TxID == key.TxID && in.TxIndex == key.VOut {
				isExternalInput = false
				break
			}
			if !isExternalInput {
				continue
			}
			externalInputKeys = append(externalInputKeys, domain.UtxoKey(key))
		}
	}

	if len(externalInputKeys) == 0 {
		return nil, nil
	}

	utxos, err := ts.bcScanner.GetUtxos(externalInputKeys)
	if err != nil {
		return nil, err
	}

	externalInputs := make([]wallet.Input, 0, len(utxos))
	for _, u := range utxos {
		externalInputs = append(externalInputs, wallet.Input{
			TxID:            u.TxID,
			TxIndex:         u.VOut,
			Value:           u.Value,
			Asset:           u.Asset,
			ValueCommitment: u.ValueCommitment,
			AssetCommitment: u.AssetCommitment,
			Script:          u.Script,
		})
	}
	return externalInputs, nil
}

func utxoKeysFromRawTx(txHex string) []domain.UtxoKey {
	tx, _ := transaction.NewTxFromHex(txHex)
	keys := make([]domain.UtxoKey, 0, len(tx.Inputs))
	for _, in := range tx.Inputs {
		keys = append(keys, domain.UtxoKey{
			TxID: elementsutil.TxIDFromBytes(in.Hash),
			VOut: in.Index,
		})
	}
	return keys
}

func utxoKeysFromPartialTx(psetBase64 string) []domain.UtxoKey {
	tx, _ := psetv2.NewPsetFromBase64(psetBase64)
	keys := make([]domain.UtxoKey, 0, len(tx.Inputs))
	for _, in := range tx.Inputs {
		keys = append(keys, domain.UtxoKey{
			TxID: elementsutil.TxIDFromBytes(in.PreviousTxid),
			VOut: in.PreviousTxIndex,
		})
	}
	return keys
}

func findUtxoIndexInTx(tx string, utxo *domain.Utxo) uint32 {
	rawTx, _ := transaction.NewTxFromHex(tx)
	if rawTx != nil {
		return findUtxoIndexInRawTx(tx, utxo)
	}

	return findUtxoIndexInPartialTx(tx, utxo)
}

func findUtxoIndexInRawTx(txHex string, utxo *domain.Utxo) uint32 {
	tx, _ := transaction.NewTxFromHex(txHex)

	for i, in := range tx.Inputs {
		key := domain.UtxoKey{
			TxID: elementsutil.TxIDFromBytes(in.Hash),
			VOut: in.Index,
		}
		if utxo.Key() == key {
			return uint32(i)
		}
	}
	return uint32(len(tx.Inputs) - 1)
}

func findUtxoIndexInPartialTx(psetBase64 string, utxo *domain.Utxo) uint32 {
	tx, _ := psetv2.NewPsetFromBase64(psetBase64)

	for i, in := range tx.Inputs {
		key := domain.UtxoKey{
			TxID: elementsutil.TxIDFromBytes(in.PreviousTxid),
			VOut: in.PreviousTxIndex,
		}
		if utxo.Key() == key {
			return uint32(i)
		}
	}
	return uint32(len(tx.Inputs) - 1)
}

func roundUpAmount(amount uint64) uint64 {
	roundedAmount := float64(amount)
	orderOfMagnitude := 0
	for roundedAmount > 10 {
		roundedAmount /= 10
		orderOfMagnitude++
	}
	roundedAmount = math.Ceil(roundedAmount)
	roundedAmount *= math.Pow10(orderOfMagnitude)
	return uint64(roundedAmount)
}

func getRemainingUtxos(utxos, selectedUtxos []*domain.Utxo) []*domain.Utxo {
	remainingUtxos := make([]*domain.Utxo, 0)
	for _, u := range utxos {
		isSelected := false
		for _, su := range selectedUtxos {
			if u.Key() == su.Key() {
				isSelected = true
				break
			}
		}
		if isSelected {
			continue
		}
		remainingUtxos = append(remainingUtxos, u)
	}
	return remainingUtxos
}

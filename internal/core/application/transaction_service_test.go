package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/ocean/internal/core/application"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	dbbadger "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/badger"
	wallet "github.com/vulpemventures/ocean/pkg/single-key-wallet"
)

var (
	coinSelectionStrategy = application.CoinSelectionStrategySmallestSubset
	receiverAddress       = "el1qqd926hzqpdeh33jkd2acujjvxwuyfnxpcnve7ts5rvv8w57wku53n8zumkc5ya9jejmejs92xu6gac38kup6u6yta3u4njavl"
	outputs               = []application.Output{
		{Asset: regtest.AssetID, Amount: 1000000, Address: receiverAddress},
	}
	utxoExpiryDuration = 2 * time.Minute
)

func TestTransactionService(t *testing.T) {
	testInternalTransaction(t)

	testExternalTransaction(t)
}

func testExternalTransaction(t *testing.T) {
	t.Run("craft_transaction_externally", func(t *testing.T) {
		mockedBcScanner := newMockedBcScanner()
		mockedBcScanner.On("BroadcastTransaction", mock.Anything).Return(randomHex(32), nil)
		mockedBcScanner.On("GetUtxos", mock.Anything).Return(nil, nil)
		repoManager, err := newRepoManagerForTxService()
		require.NoError(t, err)
		require.NotNil(t, repoManager)

		svc := application.NewTransactionService(
			repoManager, mockedBcScanner, regtest, rootPath, utxoExpiryDuration,
		)

		selectedUtxos, change, err := svc.SelectUtxos(
			ctx, accountName, regtest.AssetID, 1000000, coinSelectionStrategy,
		)
		require.NoError(t, err)
		require.NotEmpty(t, selectedUtxos)

		inputs := make(application.Inputs, 0, len(selectedUtxos))
		for _, u := range selectedUtxos {
			inputs = append(inputs, application.Input(u.Key()))
		}

		if change > 0 {
			addrInfo, err := repoManager.WalletRepository().
				DeriveNextInternalAddressesForAccount(ctx, accountName, 1)
			require.NoError(t, err)
			require.Len(t, addrInfo, 1)

			outputs = append(outputs, application.Output{
				Asset:   regtest.AssetID,
				Amount:  change,
				Address: addrInfo[0].Address,
			})
		}

		feeAmount, err := svc.EstimateFees(ctx, inputs, outputs, 0)
		require.NoError(t, err)
		require.NotZero(t, feeAmount)

		// This is just for sake of simplicity.
		// In real scenarios, there are 3 different situations:
		// 1. feeAmount < changeAmount -> deduct fee amount from change (like done here)
		// 2. feeAmount = changeAmount -> remove change output
		// 3. feeAmount > changeAmount -> another round of coin-selection is required
		// Take a look at how the app service handles these scenarios into "internal"
		// transactions like Transfer().
		outputs[len(outputs)-1].Amount -= feeAmount
		outputs = append(outputs, application.Output{
			Asset:  regtest.AssetID,
			Amount: feeAmount,
		})

		newPset, err := svc.CreatePset(ctx, inputs, outputs)
		require.NoError(t, err)
		require.NotEmpty(t, newPset)

		blindedPset, err := svc.BlindPset(ctx, newPset, nil, true)
		require.NoError(t, err)
		require.NotEmpty(t, blindedPset)

		signedPset, err := svc.SignPset(ctx, blindedPset)
		require.NoError(t, err)

		txHex, _, err := wallet.FinalizeAndExtractTransaction(wallet.FinalizeAndExtractTransactionArgs{
			PsetBase64: signedPset,
		})
		require.NoError(t, err)
		require.NotEmpty(t, txHex)

		txid, err := svc.BroadcastTransaction(ctx, txHex)
		require.NoError(t, err)
		require.NotEmpty(t, txid)
	})
}

func testInternalTransaction(t *testing.T) {
	t.Run("craft_transaction_internally", func(t *testing.T) {
		mockedBcScanner := newMockedBcScanner()
		mockedBcScanner.On("BroadcastTransaction", mock.Anything).Return(randomHex(32), nil)
		repoManager, err := newRepoManagerForTxService()
		require.NoError(t, err)
		require.NotNil(t, repoManager)

		svc := application.NewTransactionService(
			repoManager, mockedBcScanner, regtest, rootPath, utxoExpiryDuration,
		)

		txid, err := svc.Transfer(ctx, accountName, outputs, 0)
		require.NoError(t, err)
		require.NotEmpty(t, txid)
	})
}

func newRepoManagerForTxService() (ports.RepoManager, error) {
	rm, err := dbbadger.NewRepoManager("", nil)
	if err != nil {
		return nil, err
	}

	wallet, err := domain.NewWallet(
		mnemonic, password, rootPath, regtest.Name, birthdayBlockHeight, nil,
	)
	if err != nil {
		return nil, err
	}

	if err := rm.WalletRepository().CreateWallet(ctx, wallet); err != nil {
		return nil, err
	}

	if err := rm.WalletRepository().UpdateWallet(
		ctx, func(w *domain.Wallet) (*domain.Wallet, error) {
			w.Unlock(password)
			w.CreateAccount(accountName, 0)
			return w, nil
		},
	); err != nil {
		return nil, err
	}

	addrInfo, err := rm.WalletRepository().DeriveNextExternalAddressesForAccount(ctx, accountName, 2)
	if err != nil {
		return nil, err
	}

	addresses := application.AddressesInfo(addrInfo).Addresses()
	utxos := make([]*domain.Utxo, 0, len(addresses))
	for _, addr := range addresses {
		utxo := randomUtxo(accountName, addr)
		utxo.Value = 100000000
		utxo.Asset = regtest.AssetID
		utxos = append(utxos, utxo)
	}

	if _, err := rm.UtxoRepository().AddUtxos(ctx, utxos); err != nil {
		return nil, err
	}

	return rm, nil
}

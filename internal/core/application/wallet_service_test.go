package application_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/ocean/internal/core/application"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	dbbadger "github.com/vulpemventures/ocean/internal/infrastructure/storage/db/badger"
)

var (
	rootPath    = "m/84'/1'"
	regtest     = &network.Regtest
	ctx         = context.Background()
	password    = "password"
	newPassword = "newpassword"
	mnemonic    = []string{
		"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
		"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
		"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
	}
	encryptedMnemonic   = "8f29524ee5995c838ca6f28c7ded7da6dc51de804fd2703775989e65ddc1bb3b60122bf0f430bb3b7a267449aaeee103375737d679bfdabf172c3842048925e6f8952e214f6b900435d24cff938be78ad3bb303d305702fbf168534a45a57ac98ca940d4c3319f14d0c97a20b5bcb456d72857d48d0b4f0e0dcf71d1965b6a42aca8d84fcb66aadeabc812a9994cf66e7a75f8718a031418468f023c560312a02f46ec8e65d5dd65c968ddb93e10950e96c8e730ce7a74d33c6ddad9e12f45e534879f1605eb07fe90432f6592f7996091bbb3e3b2"
	accountName         = "test1"
	birthdayBlockHeight = uint32(randomIntInRange(1, 1000))
	birthdayBlockHash   = randomBytes(32)
)

func TestMain(m *testing.M) {
	mockedMnemonicCypher := &mockMnemonicCypher{}
	mockedMnemonicCypher.On("Encrypt", mock.Anything, mock.Anything).Return(h2b(encryptedMnemonic), nil)
	mockedMnemonicCypher.On("Decrypt", mock.Anything, []byte(password)).Return([]byte(strings.Join(mnemonic, " ")), nil)
	mockedMnemonicCypher.On("Decrypt", h2b(encryptedMnemonic), []byte(newPassword)).Return([]byte(strings.Join(mnemonic, " ")), nil)
	mockedMnemonicCypher.On("Decrypt", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("invalid password"))
	domain.MnemonicCypher = mockedMnemonicCypher
	domain.MnemonicStore = newInMemoryMnemonicStore()

	os.Exit(m.Run())
}

func TestWalletService(t *testing.T) {
	testInitWalletFromScratch(t)

	testInitWalletFromRestart(t)
}

func testInitWalletFromScratch(t *testing.T) {
	t.Run("init_wallet_from_scratch", func(t *testing.T) {
		mockedBcScanner := newMockedBcScanner()
		mockedBcScanner.On("GetLatestBlock").Return(birthdayBlockHash, birthdayBlockHeight, nil)
		mockedBcScanner.On("GetBlockHash", mock.Anything).Return(birthdayBlockHash, nil)
		mockedBcScanner.On("GetBlockHeight", mock.Anything).Return(birthdayBlockHeight, nil)
		repoManager, err := newRepoManagerForNewWallet()
		require.NoError(t, err)
		require.NotNil(t, repoManager)

		svc := application.NewWalletService(
			repoManager, mockedBcScanner, rootPath, regtest,
		)

		status := svc.GetStatus(ctx)
		require.False(t, status.IsInitialized)
		require.False(t, status.IsSynced)
		require.False(t, status.IsUnlocked)

		info, err := svc.GetInfo(ctx)
		require.Error(t, err)
		require.Nil(t, info)

		newMnemonic, err := svc.GenSeed(ctx)
		require.NoError(t, err)

		err = svc.CreateWallet(ctx, newMnemonic, password)
		require.NoError(t, err)

		status = svc.GetStatus(ctx)
		require.True(t, status.IsInitialized)
		require.True(t, status.IsSynced)
		require.False(t, status.IsUnlocked)

		info, err = svc.GetInfo(ctx)
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, regtest.Name, info.Network)
		require.Equal(t, regtest.AssetID, info.NativeAsset)
		require.Empty(t, info.RootPath)
		require.Empty(t, info.MasterBlindingKey)
		require.Empty(t, info.Accounts)

		err = svc.Unlock(ctx, password)
		require.NoError(t, err)

		status = svc.GetStatus(ctx)
		require.True(t, status.IsInitialized)
		require.True(t, status.IsSynced)
		require.True(t, status.IsUnlocked)

		info, err = svc.GetInfo(ctx)
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, regtest.Name, info.Network)
		require.Equal(t, regtest.AssetID, info.NativeAsset)
		require.Equal(t, rootPath, info.RootPath)
		require.NotEmpty(t, info.MasterBlindingKey)
		require.Empty(t, info.Accounts)
	})
}

func testInitWalletFromRestart(t *testing.T) {
	t.Run("init_wallet_from_restart", func(t *testing.T) {
		mockedBcScanner := newMockedBcScanner()
		mockedBcScanner.On("GetBlockHash", mock.Anything).Return(birthdayBlockHash, nil)
		repoManager, err := newRepoManagerForExistingWallet()
		require.NoError(t, err)
		require.NotNil(t, repoManager)

		svc := application.NewWalletService(
			repoManager, mockedBcScanner, rootPath, regtest,
		)

		status := svc.GetStatus(ctx)
		require.True(t, status.IsInitialized)
		require.True(t, status.IsSynced)
		require.False(t, status.IsUnlocked)

		info, err := svc.GetInfo(ctx)
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, regtest.Name, info.Network)
		require.Equal(t, regtest.AssetID, info.NativeAsset)
		require.Empty(t, info.RootPath)
		require.Empty(t, info.MasterBlindingKey)
		require.Empty(t, info.Accounts)

		err = svc.ChangePassword(ctx, password, newPassword)
		require.NoError(t, err)

		err = svc.Unlock(ctx, newPassword)
		require.NoError(t, err)

		status = svc.GetStatus(ctx)
		require.True(t, status.IsInitialized)
		require.True(t, status.IsSynced)
		require.True(t, status.IsUnlocked)

		info, err = svc.GetInfo(ctx)
		require.NoError(t, err)
		require.NotNil(t, info)
		require.Equal(t, regtest.Name, info.Network)
		require.Equal(t, regtest.AssetID, info.NativeAsset)
		require.Equal(t, rootPath, info.RootPath)
		require.NotEmpty(t, info.MasterBlindingKey)
		require.NotEmpty(t, info.Accounts)
	})
}

// TODO: uncomment this test once supporting restring a wallet.
// (Changes might be required)
// func testInitWalletFromRestore(t *testing.T) {
// 	t.Run("init_wallet_from_restore", func(t *testing.T) {
// 		repoManager, err := newRepoManagerForRestoredWallet()
// 		require.NoError(t, err)

// 		svc := application.NewWalletService(repoManager, rootPath, regtest)

// 		status := svc.GetStatus(ctx)
// 		require.False(t, status.IsInitialized)
// 		require.False(t, status.IsSynced)
// 		require.False(t, status.IsUnlocked)

// 		err = svc.RestoreWallet(ctx, restoreMnemonic, password)
// 		require.NoError(t, err)

// 		status = svc.GetStatus(ctx)
// 		require.True(t, status.IsInitialized)
// 		require.False(t, status.IsSynced)
// 		require.False(t, status.IsUnlocked)

// 		err = svc.Unlock(ctx, password)
// 		require.NoError(t, err)

// 		status = svc.GetStatus(ctx)
// 		require.True(t, status.IsInitialized)
// 		require.False(t, status.IsSynced)
// 		require.True(t, status.IsUnlocked)

// 		info, err := svc.GetInfo(ctx)
// 		require.NoError(t, err)
// 		require.NotNil(t, info)
// 		require.Equal(t, regtest.Name, info.Network)
// 		require.Equal(t, regtest.AssetID, info.NativeAsset)
// 		require.Equal(t, rootPath, info.RootPath)
// 		require.NotEmpty(t, info.MasterBlindingKey)
// 		require.Empty(t, info.Accounts)

// 		status = svc.GetStatus(ctx)
// 		require.True(t, status.IsInitialized)
// 		require.True(t, status.IsSynced)
// 		require.True(t, status.IsUnlocked)

// 		info, err = svc.GetInfo(ctx)
// 		require.NoError(t, err)
// 		require.NotNil(t, info)
// 		require.NotEmpty(t, info.Accounts)
// 	})
// }

func newRepoManagerForNewWallet() (ports.RepoManager, error) {
	return dbbadger.NewRepoManager("", nil)
}

func newRepoManagerForExistingWallet() (ports.RepoManager, error) {
	rm, err := dbbadger.NewRepoManager("", nil)
	if err != nil {
		return nil, err
	}

	accounts := []domain.Account{
		{
			Info: domain.AccountInfo{
				Key:            domain.AccountKey{Name: "test1", Index: 0},
				Xpub:           "xpub6CvgMkAYP4RFDuozj9Mji9ncsoTiHyf4mFVVJKAHSTeecsR9hwxKa1PkfayopR32SXJRKx1WJJkGjgndyPxhDRpBxJGwzXJCELybhPQxd8Y",
				DerivationPath: "m/84'/0'/0'",
			},
			NextExternalIndex: 2,
			NextInternalIndex: 4,
			DerivationPathByScript: map[string]string{
				"00141d124e9e47aded6bcd2bdfe86eea2ea1c4391cbe": "0'/1/1",
				"00143309416df9f260be2547e505251c4c888fa5d4fe": "0'/1/3",
				"001440d399dca3f89e937534b1624aa3b6c4167aa6d9": "0'/0/0",
				"0014488ebe5da5a52e111bf72241b11a722162f86473": "0'/1/2",
				"001474ce4fc0b2443a9cca131f187a9eb1607e35636a": "0'/1/0",
				"0014e9fe164b891a806aa3c729608cb8251e12753918": "0'/0/1",
			},
		},
	}
	wallet, err := domain.NewWallet(
		mnemonic, password, rootPath, regtest.Name, birthdayBlockHeight, accounts,
	)
	if err != nil {
		return nil, err
	}
	wallet.Lock()

	if err := rm.WalletRepository().CreateWallet(ctx, wallet); err != nil {
		return nil, err
	}
	return rm, nil
}

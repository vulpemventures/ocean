package domain_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

var (
	mnemonic = []string{
		"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
		"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
		"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
	}
	regtest           = network.Regtest.Name
	rootPath          = "m/84'/1'"
	password          = "password"
	newPassword       = "newpassword"
	wrongPassword     = "wrongpassword"
	masterBlingingKey = "9390be245db10fd2d5a1dd3b07a5cabfcc52108dd4f7bd93ee07d045ca872bda"
	encryptedMnemonic = "8f29524ee5995c838ca6f28c7ded7da6dc51de804fd2703775989e65ddc1bb3b60122bf0f430bb3b7a267449aaeee103375737d679bfdabf172c3842048925e6f8952e214f6b900435d24cff938be78ad3bb303d305702fbf168534a45a57ac98ca940d4c3319f14d0c97a20b5bcb456d72857d48d0b4f0e0dcf71d1965b6a42aca8d84fcb66aadeabc812a9994cf66e7a75f8718a031418468f023c560312a02f46ec8e65d5dd65c968ddb93e10950e96c8e730ce7a74d33c6ddad9e12f45e534879f1605eb07fe90432f6592f7996091bbb3e3b2"
	passwordHash      = "b8affdb68657a0417b09a02dd209585480f5a920"
	newPasswordHash   = "b34d0f1bcefa7d25beefec121165c765c41550f7"
	birthdayBlock     = uint32(1)
)

func TestMain(m *testing.M) {
	mockedMnemonicCypher := &mockMnemonicCypher{}
	mockedMnemonicCypher.On("Encrypt", mock.Anything, mock.Anything).Return(h2b(encryptedMnemonic), nil)
	mockedMnemonicCypher.On("Decrypt", h2b(encryptedMnemonic), []byte(password)).Return([]byte(strings.Join(mnemonic, " ")), nil)
	mockedMnemonicCypher.On("Decrypt", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("invalid password"))
	domain.MnemonicCypher = mockedMnemonicCypher
	domain.MnemonicStore = newInMemoryMnemonicStore()

	os.Exit(m.Run())
}

func TestNewWallet(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		w, err := newTestWallet()
		require.NoError(t, err)
		require.NotNil(t, w)
		require.Equal(t, "m/84'/1'", w.RootPath)
		require.Equal(t, regtest, w.NetworkName)
		require.Equal(t, encryptedMnemonic, b2h(w.EncryptedMnemonic))
		require.Equal(t, passwordHash, b2h(w.PasswordHash))
		require.Empty(t, w.AccountKeysByIndex)
		require.Empty(t, w.AccountKeysByName)
		require.Empty(t, w.AccountsByKey)
		require.Equal(t, 0, int(w.NextAccountIndex))
		require.True(t, w.IsInitialized())
		require.True(t, w.IsLocked())

		err = w.Unlock(password)
		require.NoError(t, err)

		m, err := w.GetMnemonic()
		require.NoError(t, err)
		require.Equal(t, mnemonic, m)

		masterKey, err := w.GetMasterBlindingKey()
		require.NoError(t, err)
		require.Equal(t, masterBlingingKey, masterKey)

		err = w.Lock("wrong password")
		require.EqualError(t, err, domain.ErrWalletInvalidPassword.Error())
		require.False(t, w.IsLocked())

		err = w.Lock(password)
		require.NoError(t, err)
		require.True(t, w.IsLocked())

		m, err = w.GetMnemonic()
		require.EqualError(t, domain.ErrWalletLocked, err.Error())
		require.Empty(t, m)

		masterKey, err = w.GetMasterBlindingKey()
		require.EqualError(t, domain.ErrWalletLocked, err.Error())
		require.Empty(t, masterKey)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			mnemonic      []string
			password      string
			network       string
			birthdayBlock uint32
			expectedError error
		}{
			{nil, password, regtest, birthdayBlock, domain.ErrWalletMissingMnemonic},
			{mnemonic, "", regtest, birthdayBlock, domain.ErrWalletMissingPassword},
			{mnemonic, password, "", birthdayBlock, domain.ErrWalletMissingNetwork},
			{mnemonic, password, regtest, 0, domain.ErrWalletMissingBirthdayBlock},
		}

		for _, tt := range tests {
			v, err := domain.NewWallet(
				tt.mnemonic, tt.password, "", tt.network, tt.birthdayBlock, nil,
			)
			require.Nil(t, v)
			require.EqualError(t, err, tt.expectedError.Error())
		}
	})
}

func TestLockUnlock(t *testing.T) {
	w, err := newTestWallet()
	require.NoError(t, err)

	err = w.Unlock(password)
	require.NoError(t, err)

	err = w.Lock(password)
	require.NoError(t, err)

	err = w.Unlock(wrongPassword)
	require.Error(t, err)

	err = w.Unlock(password)
	require.NoError(t, err)
}

func TestChangePassword(t *testing.T) {
	w, err := newTestWallet()
	require.NoError(t, err)

	err = w.Unlock(password)
	require.NoError(t, err)

	err = w.ChangePassword(password, newPassword)
	require.EqualError(t, domain.ErrWalletUnlocked, err.Error())

	err = w.Lock(password)
	require.NoError(t, err)

	err = w.ChangePassword(password, newPassword)
	require.NoError(t, err)
	require.Equal(t, newPasswordHash, b2h(w.PasswordHash))
}

func TestWalletAccount(t *testing.T) {
	w, err := newTestWallet()
	require.NoError(t, err)

	err = w.Lock(password)
	require.NoError(t, err)

	accountName := "test1"
	account, err := w.CreateAccount(accountName, 0)
	require.EqualError(t, domain.ErrWalletLocked, err.Error())
	require.Nil(t, account)

	err = w.Unlock(password)
	require.NoError(t, err)

	account, err = w.CreateAccount(accountName, 0)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Empty(t, account.NextExternalIndex)
	require.Empty(t, account.NextInternalIndex)
	require.Empty(t, account.DerivationPathByScript)
	require.Equal(t, 0, int(account.Info.Key.Index))
	require.Equal(t, accountName, account.Info.Key.Name)
	require.Equal(t, "m/84'/1'/0'", account.Info.DerivationPath)
	require.NotEmpty(t, account.Info.Xpub)

	err = w.Lock(password)
	require.NoError(t, err)

	gotAccount, err := w.GetAccount(accountName)
	require.EqualError(t, domain.ErrWalletLocked, err.Error())
	require.Nil(t, gotAccount)

	w.Unlock(password)

	gotAccount, err = w.GetAccount(accountName)
	require.NoError(t, err)
	require.Exactly(t, *account, *gotAccount)

	err = w.Lock(password)
	require.NoError(t, err)

	allAddrInfo, err := w.AllDerivedAddressesForAccount(accountName)
	require.EqualError(t, domain.ErrWalletLocked, err.Error())
	require.Nil(t, allAddrInfo)

	w.Unlock(password)

	allAddrInfo, err = w.AllDerivedAddressesForAccount(accountName)
	require.NoError(t, err)
	require.Empty(t, allAddrInfo)

	err = w.Lock(password)
	require.NoError(t, err)

	addrInfo, err := w.DeriveNextExternalAddressForAccount(accountName)
	require.EqualError(t, domain.ErrWalletLocked, err.Error())
	require.Nil(t, addrInfo)

	w.Unlock(password)

	addrInfo, err = w.DeriveNextExternalAddressForAccount(accountName)
	require.NoError(t, err)
	require.NotNil(t, addrInfo)
	require.NotEmpty(t, addrInfo.Address)
	require.NotEmpty(t, addrInfo.BlindingKey)
	require.NotEmpty(t, addrInfo.Script)
	require.NotEmpty(t, addrInfo.DerivationPath)
	require.NotEmpty(t, addrInfo.AccountKey.Name)

	allAddrInfo, err = w.AllDerivedAddressesForAccount(accountName)
	require.NoError(t, err)
	require.Len(t, allAddrInfo, 1)
	require.Exactly(t, *addrInfo, allAddrInfo[0])

	err = w.Lock(password)
	require.NoError(t, err)

	err = w.DeleteAccount(accountName)
	require.EqualError(t, domain.ErrWalletLocked, err.Error())

	w.Unlock(password)

	err = w.DeleteAccount(accountName)
	require.NoError(t, err)

	_, err = w.GetAccount(accountName)
	require.EqualError(t, domain.ErrAccountNotFound, err.Error())
}

func newTestWallet() (*domain.Wallet, error) {
	return domain.NewWallet(mnemonic, password, rootPath, regtest, birthdayBlock, nil)
}

func b2h(buf []byte) string {
	return hex.EncodeToString(buf)
}

func h2b(str string) []byte {
	buf, _ := hex.DecodeString(str)
	return buf
}

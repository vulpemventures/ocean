package domain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/vulpemventures/go-elements/network"
	path "github.com/vulpemventures/ocean/pkg/wallet/derivation-path"
	singlesig "github.com/vulpemventures/ocean/pkg/wallet/single-sig"
)

const (
	externalChain = 0
	internalChain = 1

	namespaceFormat = "bip%v-account%v"
)

var (
	ErrWalletMissingMnemonic         = fmt.Errorf("missing mnemonic")
	ErrWalletMissingPassword         = fmt.Errorf("missing password")
	ErrWalletMissingNetwork          = fmt.Errorf("missing network name")
	ErrWalletMissingBirthdayBlock    = fmt.Errorf("missing birthday block height")
	ErrWalletLocked                  = fmt.Errorf("wallet is locked")
	ErrWalletUnlocked                = fmt.Errorf("wallet must be locked")
	ErrWalletMaxAccountNumberReached = fmt.Errorf("reached max number of accounts")
	ErrWalletInvalidPassword         = fmt.Errorf("wrong password")
	ErrWalletInvalidNetwork          = fmt.Errorf("unknown network")
	ErrAccountNotFound               = fmt.Errorf("account not found in wallet")

	networks = map[string]*network.Network{
		"liquid":  &network.Liquid,
		"testnet": &network.Testnet,
		"regtest": &network.Regtest,
	}
)

// AddressInfo holds useful info about a derived address.
type AddressInfo struct {
	Account        string
	Address        string
	BlindingKey    []byte
	DerivationPath string
	Script         string
}

// Wallet is the data structure representing a secure HD wallet, ie. protected
// by a password that encrypts/decrypts the mnemonic seed.
type Wallet struct {
	EncryptedMnemonic   []byte
	PasswordHash        []byte
	BirthdayBlockHeight uint32
	RootPath            string
	NetworkName         string
	Accounts            map[string]*Account
	AccountsByLabel     map[string]string
	NextAccountIndex    uint32
}

// NewWallet encrypts the provided mnemonic with the passhrase and returns a new
// Wallet initialized with the encrypted mnemonic, the hash of the password,
// the given root path, network and possible a list of accounts for an already
// used one.
// The Wallet is locked by default since it is initialized without the mnemonic
// in plain text.
func NewWallet(
	mnemonic []string, password, rootPath, network string,
	birthdayBlock uint32, accounts []Account,
) (*Wallet, error) {
	if len(mnemonic) <= 0 {
		return nil, ErrWalletMissingMnemonic
	}
	if len(password) <= 0 {
		return nil, ErrWalletMissingPassword
	}
	if birthdayBlock == 0 {
		return nil, ErrWalletMissingBirthdayBlock
	}
	if network == "" {
		return nil, ErrWalletMissingNetwork
	}
	if _, ok := networks[network]; !ok {
		return nil, ErrWalletInvalidNetwork
	}

	if _, err := singlesig.NewWalletFromMnemonic(singlesig.NewWalletFromMnemonicArgs{
		RootPath: rootPath,
		Mnemonic: mnemonic,
	}); err != nil {
		return nil, err
	}

	strMnemonic := strings.Join(mnemonic, " ")
	encryptedMnemonic, err := MnemonicCypher.Encrypt(
		[]byte(strMnemonic), []byte(password),
	)
	if err != nil {
		return nil, err
	}

	accountsByNamespace := make(map[string]*Account)
	accountsByLabel := make(map[string]string)
	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].AccountInfo.DerivationPath > accounts[j].AccountInfo.DerivationPath
	})
	for i := range accounts {
		account := accounts[i]
		accountsByNamespace[account.Namespace] = &account
		if account.Label != "" {
			accountsByLabel[account.Label] = account.Namespace
		}
	}
	var nextAccountIndex uint32
	if len(accounts) > 0 {
		p, _ := path.ParseDerivationPath(accounts[0].AccountInfo.DerivationPath)
		nextAccountIndex = p[len(p)-1] - hdkeychain.HardenedKeyStart
	}

	return &Wallet{
		EncryptedMnemonic:   encryptedMnemonic,
		PasswordHash:        btcutil.Hash160([]byte(password)),
		BirthdayBlockHeight: birthdayBlock,
		RootPath:            rootPath,
		Accounts:            accountsByNamespace,
		AccountsByLabel:     accountsByLabel,
		NetworkName:         network,
		NextAccountIndex:    nextAccountIndex,
	}, nil
}

// IsInitialized returns wheter the wallet is initialized with an encrypted
// mnemonic.
func (w *Wallet) IsInitialized() bool {
	return len(w.EncryptedMnemonic) > 0
}

// IsLocked returns whether the wallet is initialized and the plaintext
// mnemonic is set in its store.
func (w *Wallet) IsLocked() bool {
	return !w.IsInitialized() || !MnemonicStore.IsSet()
}

// GetMnemonic safely returns the plaintext mnemonic.
func (w *Wallet) GetMnemonic() ([]string, error) {
	if w.IsLocked() {
		return nil, ErrWalletLocked
	}

	return MnemonicStore.Get(), nil
}

// Lock locks the Wallet by wiping the plaintext mnemonic from its store.
func (w *Wallet) Lock(password string) error {
	if w.IsLocked() {
		return nil
	}

	if !w.IsValidPassword(password) {
		return ErrWalletInvalidPassword
	}

	MnemonicStore.Unset()
	return nil
}

// Unlock attempts to decrypt the encrypted mnemonic with the provided
// password.
func (w *Wallet) Unlock(password string) error {
	if !w.IsLocked() {
		return nil
	}

	if !w.IsValidPassword(password) {
		return ErrWalletInvalidPassword
	}

	mnemonic, err := MnemonicCypher.Decrypt(w.EncryptedMnemonic, []byte(password))
	if err != nil {
		return err
	}

	MnemonicStore.Set(string(mnemonic))
	return nil
}

// ChangePassword attempts to unlock the wallet with the given currentPassword,
// then encrypts the plaintext mnemonic again with new password, stores its hash
// and, finally, locks the Wallet again.
func (w *Wallet) ChangePassword(currentPassword, newPassword string) error {
	if !w.IsLocked() {
		return ErrWalletUnlocked
	}
	if !w.IsValidPassword(currentPassword) {
		return ErrWalletInvalidPassword
	}

	mnemonic, err := MnemonicCypher.Decrypt(w.EncryptedMnemonic, []byte(currentPassword))
	if err != nil {
		return err
	}

	encryptedMnemonic, err := MnemonicCypher.Encrypt(mnemonic, []byte(newPassword))
	if err != nil {
		return err
	}

	w.EncryptedMnemonic = encryptedMnemonic
	w.PasswordHash = btcutil.Hash160([]byte(newPassword))
	return nil
}

// CreateAccount creates a new account with the given name by preventing
// collisions with existing ones. If successful, returns the Account created.
func (w *Wallet) CreateAccount(label string, birthdayBlock uint32) (*Account, error) {
	account, err := w.getAccount(label)
	if err != nil && err != ErrAccountNotFound {
		return nil, err
	}
	if account != nil {
		return nil, nil
	}
	if w.NextAccountIndex == hdkeychain.HardenedKeyStart {
		return nil, ErrWalletMaxAccountNumberReached
	}

	mnemonic := MnemonicStore.Get()
	namespace := getAccountNamespace(w.RootPath, w.NextAccountIndex)

	ww, _ := singlesig.NewWalletFromMnemonic(singlesig.NewWalletFromMnemonicArgs{
		RootPath: w.RootPath,
		Mnemonic: mnemonic,
	})
	xpub, _ := ww.AccountExtendedPublicKey(singlesig.ExtendedKeyArgs{Account: w.NextAccountIndex})

	derivationPath, _ := path.ParseDerivationPath(w.RootPath)
	derivationPath = append(derivationPath, w.NextAccountIndex+hdkeychain.HardenedKeyStart)
	bdayBlock := w.BirthdayBlockHeight
	if birthdayBlock > bdayBlock {
		bdayBlock = birthdayBlock
	}
	newAccount := &Account{
		AccountInfo: AccountInfo{
			Namespace:      namespace,
			Label:          label,
			Xpub:           xpub,
			DerivationPath: derivationPath.String(),
		},
		Index:                  w.NextAccountIndex,
		DerivationPathByScript: make(map[string]string),
		BirthdayBlock:          bdayBlock,
	}

	w.Accounts[namespace] = newAccount
	if label != "" {
		w.AccountsByLabel[label] = namespace
	}
	w.NextAccountIndex++
	return newAccount, nil
}

// GetAccount safely returns an Account identified by the given name.
func (w *Wallet) GetAccount(accountName string) (*Account, error) {
	return w.getAccount(accountName)
}

// SetLabelForAccount changes the label for the given account
func (w *Wallet) SetLabelForAccount(accountName, label string) error {
	account, err := w.getAccount(accountName)
	if err != nil {
		return err
	}

	if account.Label != "" {
		delete(w.AccountsByLabel, account.Label)
	}
	w.Accounts[account.Namespace].Label = label
	w.AccountsByLabel[label] = account.Namespace
	return nil
}

// DeleteAccount safely removes an Account and all related stored info from the
// Wallet.
func (w *Wallet) DeleteAccount(accountName string) error {
	account, err := w.getAccount(accountName)
	if err != nil {
		return err
	}

	delete(w.Accounts, account.Namespace)
	if account.Label != "" {
		delete(w.AccountsByLabel, account.Label)
	}
	return nil
}

// DeriveNextExternalAddressForAccount returns all useful info about the next
// new receiving address for the given account.
func (w *Wallet) DeriveNextExternalAddressForAccount(
	accountName string,
) (*AddressInfo, error) {
	return w.deriveNextAddressForAccount(accountName, externalChain)
}

// DeriveNextInternalAddressForAccount returns all useful info about the next
// new change address for the given account.
func (w *Wallet) DeriveNextInternalAddressForAccount(
	accountName string,
) (*AddressInfo, error) {
	return w.deriveNextAddressForAccount(accountName, internalChain)
}

// AllDerivedAddressesForAccount returns info about all derived receiving and
// change addresses derived so far for the given account.
func (w *Wallet) AllDerivedAddressesForAccount(
	accountName string,
) ([]AddressInfo, error) {
	return w.allDerivedAddressesForAccount(accountName, true)
}

// AllDerivedExternalAddressesForAccount returns info about all derived
// receiving addresses derived so far for the given account.
func (w *Wallet) AllDerivedExternalAddressesForAccount(
	accountName string,
) ([]AddressInfo, error) {
	return w.allDerivedAddressesForAccount(accountName, false)
}

func (w *Wallet) IsValidPassword(password string) bool {
	return bytes.Equal(w.PasswordHash, btcutil.Hash160([]byte(password)))
}

func (w *Wallet) getAccount(accountName string) (*Account, error) {
	if w.IsLocked() {
		return nil, ErrWalletLocked
	}

	if namespace, ok := w.AccountsByLabel[accountName]; ok {
		return w.Accounts[namespace], nil
	}

	account, ok := w.Accounts[accountName]
	if !ok {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

func (w *Wallet) deriveNextAddressForAccount(
	accountName string, chainIndex int,
) (*AddressInfo, error) {
	account, err := w.getAccount(accountName)
	if err != nil {
		return nil, err
	}

	mnemonic, _ := w.GetMnemonic()
	ww, _ := singlesig.NewWalletFromMnemonic(singlesig.NewWalletFromMnemonicArgs{
		RootPath: w.RootPath,
		Mnemonic: mnemonic,
	})

	addressIndex := account.NextExternalIndex
	if chainIndex == internalChain {
		addressIndex = account.NextInternalIndex
	}
	derivationPath := fmt.Sprintf(
		"%d'/%d/%d", account.Index, chainIndex, addressIndex,
	)
	net := networkFromName(w.NetworkName)
	addr, script, err := ww.DeriveConfidentialAddress(singlesig.DeriveConfidentialAddressArgs{
		DerivationPath: derivationPath,
		Network:        net,
	})
	if err != nil {
		return nil, err
	}

	blindingKey, _, _ := ww.DeriveBlindingKeyPair(singlesig.DeriveBlindingKeyPairArgs{
		Script: script,
	})

	account.addDerivationPath(hex.EncodeToString(script), derivationPath)
	if chainIndex == internalChain {
		account.incrementInternalIndex()
	} else {
		account.incrementExternalIndex()
	}

	return &AddressInfo{
		Account:        account.Namespace,
		Address:        addr,
		Script:         hex.EncodeToString(script),
		BlindingKey:    blindingKey.Serialize(),
		DerivationPath: derivationPath,
	}, nil
}

func (w *Wallet) allDerivedAddressesForAccount(
	accountName string, includeInternals bool,
) ([]AddressInfo, error) {
	account, err := w.getAccount(accountName)
	if err != nil {
		return nil, err
	}

	net := networkFromName(w.NetworkName)
	mnemonic, _ := w.GetMnemonic()
	ww, _ := singlesig.NewWalletFromMnemonic(singlesig.NewWalletFromMnemonicArgs{
		RootPath: w.RootPath,
		Mnemonic: mnemonic,
	})

	infoLen := account.NextExternalIndex
	if includeInternals {
		infoLen += account.NextInternalIndex
	}
	info := make([]AddressInfo, 0, infoLen)
	for i := 0; i < int(account.NextExternalIndex); i++ {
		derivationPath := fmt.Sprintf(
			"%d'/%d/%d", account.Index, externalChain, i,
		)
		addr, script, err := ww.DeriveConfidentialAddress(singlesig.DeriveConfidentialAddressArgs{
			DerivationPath: derivationPath,
			Network:        net,
		})
		if err != nil {
			return nil, err
		}
		key, _, _ := ww.DeriveBlindingKeyPair(singlesig.DeriveBlindingKeyPairArgs{
			Script: script,
		})
		info = append(info, AddressInfo{
			Account:        account.Namespace,
			Address:        addr,
			BlindingKey:    key.Serialize(),
			DerivationPath: derivationPath,
			Script:         hex.EncodeToString(script),
		})
	}
	if includeInternals {
		for i := 0; i < int(account.NextInternalIndex); i++ {
			derivationPath := fmt.Sprintf(
				"%d'/%d/%d", account.Index, internalChain, i,
			)
			addr, script, err := ww.DeriveConfidentialAddress(singlesig.DeriveConfidentialAddressArgs{
				DerivationPath: derivationPath,
				Network:        net,
			})
			if err != nil {
				return nil, err
			}
			key, _, _ := ww.DeriveBlindingKeyPair(singlesig.DeriveBlindingKeyPairArgs{
				Script: script,
			})
			info = append(info, AddressInfo{
				Account:        account.Namespace,
				Address:        addr,
				BlindingKey:    key.Serialize(),
				DerivationPath: derivationPath,
				Script:         hex.EncodeToString(script),
			})
		}
	}

	return info, nil
}

func networkFromName(net string) *network.Network {
	return networks[net]
}

func getAccountNamespace(rootPath string, index uint32) string {
	derivationPath, _ := path.ParseDerivationPath(rootPath)
	purpose := derivationPath[0] - hdkeychain.HardenedKeyStart
	return fmt.Sprintf("bip%d-account%d", purpose, index)
}

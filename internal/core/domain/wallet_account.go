package domain

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	path "github.com/vulpemventures/ocean/pkg/wallet/derivation-path"
	singlesig "github.com/vulpemventures/ocean/pkg/wallet/single-sig"
)

// AccountKey holds the unique info of an account: name and HD index.
type AccountKey struct {
	Name  string
	Index uint32
}

func (ak *AccountKey) String() string {
	key := btcutil.Hash160([]byte(fmt.Sprintf("%s%d", ak.Name, ak.Index)))
	return hex.EncodeToString(key[:6])
}

// AccountInfo holds basic info about an account.
type AccountInfo struct {
	Key            AccountKey
	Xpub           string
	DerivationPath string
}

func (i *AccountInfo) GetMasterBlindingKey() (string, error) {
	mnemonic := MnemonicStore.Get()
	rootPath, _ := path.ParseDerivationPath(i.DerivationPath)
	rootPath = rootPath[:len(rootPath)-1]
	ww, _ := singlesig.NewWalletFromMnemonic(singlesig.NewWalletFromMnemonicArgs{
		RootPath: rootPath.String(),
		Mnemonic: mnemonic,
	})
	return ww.MasterBlindingKey()
}

// Account defines the entity data struture for a derived account of the
// daemon's HD wallet
type Account struct {
	Info                   AccountInfo
	BirthdayBlock          uint32
	NextExternalIndex      uint
	NextInternalIndex      uint
	DerivationPathByScript map[string]string
}

func (a *Account) incrementExternalIndex() (next uint) {
	// restart from 0 if index has reached the its max value
	next = 0
	if a.NextExternalIndex != hdkeychain.HardenedKeyStart-1 {
		next = a.NextExternalIndex + 1
	}
	a.NextExternalIndex = next
	return
}

func (a *Account) incrementInternalIndex() (next uint) {
	next = 0
	if a.NextInternalIndex != hdkeychain.HardenedKeyStart-1 {
		next = a.NextInternalIndex + 1
	}
	a.NextInternalIndex = next
	return
}

func (a *Account) addDerivationPath(outputScript, derivationPath string) {
	if _, ok := a.DerivationPathByScript[outputScript]; !ok {
		a.DerivationPathByScript[outputScript] = derivationPath
	}
}

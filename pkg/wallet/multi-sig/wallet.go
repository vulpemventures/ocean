package multisig

import (
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	path "github.com/equitas-foundation/bamp-ocean/pkg/wallet/derivation-path"
	"github.com/equitas-foundation/bamp-ocean/pkg/wallet/mnemonic"
)

// Wallet is the data structure representing an HD wallet of an Elements based
// network.
type Wallet struct {
	mnemonic          []string
	signingMasterKey  []byte
	blindingMasterKey []byte
	xpubs             []string
}

type NewWalletArgs struct {
	RootPath string
	Xpubs    []string
}

func (a NewWalletArgs) validate() error {
	if a.RootPath == "" {
		return ErrMissingRootPath
	}
	derivationPath, err := path.ParseDerivationPath(a.RootPath)
	if err != nil {
		return err
	}
	if len(derivationPath) != 4 {
		return ErrInvalidRootPath
	}
	for _, step := range derivationPath {
		if step < hdkeychain.HardenedKeyStart {
			return ErrInvalidRootPath
		}
	}
	if len(a.Xpubs) <= 0 {
		return ErrMissingCosignerXpubs
	}
	for _, xpub := range a.Xpubs {
		if _, err := hdkeychain.NewKeyFromString(xpub); err != nil {
			return ErrInvalidXpub
		}
	}
	return nil
}

// NewWallet creates a new HD wallet with a random mnemonic
func NewWallet(args NewWalletArgs) (*Wallet, error) {
	if err := args.validate(); err != nil {
		return nil, err
	}

	mnemonic, _ := mnemonic.NewMnemonic(mnemonic.NewMnemonicArgs{
		EntropySize: 256,
	})
	seed := generateSeedFromMnemonic(mnemonic)
	rootPath, _ := path.ParseDerivationPath(args.RootPath)
	signingMasterKey, err := generateSigningMasterKey(seed, rootPath)
	if err != nil {
		return nil, err
	}
	xprv, _ := hdkeychain.NewKeyFromString(base58.Encode(signingMasterKey))
	xpub, _ := xprv.Neuter()
	xpubs := append(args.Xpubs, xpub.String())

	blindingMasterKey, err := generateBlindingMasterKey(xpubs)
	if err != nil {
		return nil, err
	}

	return &Wallet{mnemonic, signingMasterKey, blindingMasterKey, xpubs}, nil
}

type NewWalletFromMnemonicArgs struct {
	RootPath string
	Mnemonic []string
	Xpubs    []string
}

func (a NewWalletFromMnemonicArgs) validate() error {
	if a.RootPath == "" {
		return ErrMissingRootPath
	}
	derivationPath, err := path.ParseDerivationPath(a.RootPath)
	if err != nil {
		return err
	}
	if len(derivationPath) != 4 {
		return ErrInvalidRootPath
	}
	for _, step := range derivationPath {
		if step < hdkeychain.HardenedKeyStart {
			return ErrInvalidRootPath
		}
	}

	if len(a.Mnemonic) == 0 {
		return ErrMissingMnemonic
	}
	if !isMnemonicValid(a.Mnemonic) {
		return ErrInvalidMnemonic
	}

	if len(a.Xpubs) <= 0 {
		return ErrMissingCosignerXpubs
	}
	for _, xpub := range a.Xpubs {
		if _, err := hdkeychain.NewKeyFromString(xpub); err != nil {
			return ErrInvalidXpub
		}
	}
	return nil
}

func (a NewWalletFromMnemonicArgs) containsXpub(xpub string) bool {
	for _, xp := range a.Xpubs {
		if xp == xpub {
			return true
		}
	}
	return false
}

// NewWalletFromMnemonic creates a new HD wallet with the given mnemonic seed
// and root path
func NewWalletFromMnemonic(args NewWalletFromMnemonicArgs) (*Wallet, error) {
	if err := args.validate(); err != nil {
		return nil, err
	}

	seed := generateSeedFromMnemonic(args.Mnemonic)
	rootPath, _ := path.ParseDerivationPath(args.RootPath)
	signingMasterKey, err := generateSigningMasterKey(seed, rootPath)
	if err != nil {
		return nil, err
	}
	xprv, _ := hdkeychain.NewKeyFromString(base58.Encode(signingMasterKey))
	xpub, _ := xprv.Neuter()
	xpubs := make([]string, len(args.Xpubs))
	copy(xpubs, args.Xpubs)
	if !args.containsXpub(xpub.String()) {
		xpubs = append(args.Xpubs, xpub.String())
	}

	blindingMasterKey, err := generateBlindingMasterKey(xpubs)
	if err != nil {
		return nil, err
	}

	return &Wallet{args.Mnemonic, signingMasterKey, blindingMasterKey, xpubs}, nil
}

// Mnemonic returns the mnemonic of the wallet.
func (w *Wallet) Mnemonic() ([]string, error) {
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w.mnemonic, nil
}

func (w *Wallet) validate() error {
	if len(w.signingMasterKey) <= 0 {
		return ErrMissingSigningMasterKey
	}
	if len(w.xpubs) <= 0 {
		return ErrMissingCosignerXpubs
	}
	return nil
}

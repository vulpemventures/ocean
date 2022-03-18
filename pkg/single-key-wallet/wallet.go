package wallet

import "github.com/btcsuite/btcd/btcutil/hdkeychain"

var (
	DefaultRootPath = "m/84'/0'"
)

// Wallet is the data structure representing an HD wallet of an Elements based
// network.
type Wallet struct {
	mnemonic          []string
	signingMasterKey  []byte
	blindingMasterKey []byte
}

type NewWalletArgs struct {
	RootPath string
}

func (a NewWalletArgs) validate() error {
	if a.RootPath == "" {
		return nil
	}
	path, err := ParseDerivationPath(a.RootPath)
	if err != nil {
		return err
	}
	if len(path) != 2 {
		return ErrInvalidRootPathLen
	}
	if path[0] < hdkeychain.HardenedKeyStart || path[1] < hdkeychain.HardenedKeyStart {
		return ErrInvalidRootPathValue
	}
	return nil
}

func (a NewWalletArgs) rootPath() DerivationPath {
	rootPath := DefaultRootPath
	if len(a.RootPath) > 0 {
		rootPath = a.RootPath
	}
	path, _ := ParseDerivationPath(rootPath)
	return path
}

// NewWallet creates a new HD wallet with a random mnemonic
func NewWallet(args NewWalletArgs) (*Wallet, error) {
	if err := args.validate(); err != nil {
		return nil, err
	}

	mnemonic, err := NewMnemonic(NewMnemonicArgs{EntropySize: 256})
	if err != nil {
		return nil, err
	}
	seed := generateSeedFromMnemonic(mnemonic)

	signingMasterKey, err := generateSigningMasterKey(
		seed, args.rootPath(),
	)
	if err != nil {
		return nil, err
	}
	blindingMasterKey, err := generateBlindingMasterKey(seed)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		mnemonic:          mnemonic,
		signingMasterKey:  signingMasterKey,
		blindingMasterKey: blindingMasterKey,
	}, nil
}

type NewWalletFromMnemonicArgs struct {
	RootPath string
	Mnemonic []string
}

func (a NewWalletFromMnemonicArgs) validate() error {
	if a.RootPath != "" {
		path, err := ParseDerivationPath(a.RootPath)
		if err != nil {
			return err
		}
		if len(path) != 2 {
			return ErrInvalidRootPathLen
		}
		if path[0] < hdkeychain.HardenedKeyStart || path[1] < hdkeychain.HardenedKeyStart {
			return ErrInvalidRootPathValue
		}
	}
	if len(a.Mnemonic) == 0 {
		return ErrMissingMnemonic
	}
	if !isMnemonicValid(a.Mnemonic) {
		return ErrInvalidMnemonic
	}
	return nil
}

func (a NewWalletFromMnemonicArgs) rootPath() DerivationPath {
	rootPath := DefaultRootPath
	if len(a.RootPath) > 0 {
		rootPath = a.RootPath
	}
	path, _ := ParseDerivationPath(rootPath)
	return path
}

// NewWalletFromMnemonic creates a new HD wallet with the given mnemonic seed
// and root path
func NewWalletFromMnemonic(args NewWalletFromMnemonicArgs) (*Wallet, error) {
	if err := args.validate(); err != nil {
		return nil, err
	}

	seed := generateSeedFromMnemonic(args.Mnemonic)
	signingMasterKey, err := generateSigningMasterKey(
		seed, args.rootPath(),
	)
	if err != nil {
		return nil, err
	}

	blindingMasterKey, err := generateBlindingMasterKey(seed)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		mnemonic:          args.Mnemonic,
		signingMasterKey:  signingMasterKey,
		blindingMasterKey: blindingMasterKey,
	}, nil
}

func (w *Wallet) validate() error {
	if len(w.signingMasterKey) <= 0 {
		return ErrMissingSigningMasterKey
	}
	if len(w.mnemonic) <= 0 {
		return ErrMissingMnemonic
	}
	if !isMnemonicValid(w.mnemonic) {
		return ErrInvalidMnemonic
	}

	if len(w.mnemonic) > 0 {
		if !isMnemonicValid(w.mnemonic) {
			return ErrInvalidMnemonic
		}
		if len(w.blindingMasterKey) <= 0 {
			return ErrMissingBlindingMasterKey
		}
	}
	return nil
}

// Mnemonic returns the mnemonic of the wallet.
func (w *Wallet) Mnemonic() ([]string, error) {
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w.mnemonic, nil
}

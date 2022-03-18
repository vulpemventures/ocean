package wallet

import (
	"encoding/hex"
	"math"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/slip77"
)

const (
	MaxHardenedValue = math.MaxUint32 - hdkeychain.HardenedKeyStart
)

type ExtendedKeyArgs struct {
	Account uint32
}

func (a ExtendedKeyArgs) validate() error {
	if a.Account > (MaxHardenedValue) {
		return ErrOutOfRangeDerivationPathAccount
	}
	return nil
}

// AccountExtendedPrivateKey returns the extended private key in base58 format
// for the given account index.
func (w *Wallet) AccountExtendedPrivateKey(args ExtendedKeyArgs) (string, error) {
	if err := args.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	xprv, err := w.extendedPrivateKey(args.Account)
	if err != nil {
		return "", err
	}

	return xprv.String(), nil
}

// AccountExtendedPublicKey returns the extended public key in base58 format
// for the given account index.
func (w *Wallet) AccountExtendedPublicKey(args ExtendedKeyArgs) (string, error) {
	if err := args.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	xprv, err := w.extendedPrivateKey(args.Account)
	if err != nil {
		return "", err
	}

	xpub, err := xprv.Neuter()
	if err != nil {
		return "", err
	}

	return xpub.String(), nil
}

// MasterBlindingKey returns the SLIP-77 master blinding key of the wallet in
// hex format.
func (w *Wallet) MasterBlindingKey() (string, error) {
	if err := w.validate(); err != nil {
		return "", err
	}

	if len(w.mnemonic) == 0 {
		return "", nil
	}

	return hex.EncodeToString(w.blindingMasterKey), nil
}

type DeriveSigningKeyPairArgs struct {
	DerivationPath string
}

func (a DeriveSigningKeyPairArgs) validate() error {
	derivationPath, err := ParseDerivationPath(a.DerivationPath)
	if err != nil {
		return err
	}

	return checkDerivationPath(derivationPath)
}

// DeriveSigningKeyPair derives the key pair from the given derivation path.
func (w *Wallet) DeriveSigningKeyPair(args DeriveSigningKeyPairArgs) (
	*btcec.PrivateKey, *btcec.PublicKey, error,
) {
	if err := args.validate(); err != nil {
		return nil, nil, err
	}
	if err := w.validate(); err != nil {
		return nil, nil, err
	}

	hdNode, err := hdkeychain.NewKeyFromString(
		base58.Encode(w.signingMasterKey),
	)
	if err != nil {
		return nil, nil, err
	}

	derivationPath, _ := ParseDerivationPath(args.DerivationPath)
	for _, step := range derivationPath {
		hdNode, err = hdNode.Derive(step)
		if err != nil {
			return nil, nil, err
		}
	}

	privateKey, err := hdNode.ECPrivKey()
	if err != nil {
		return nil, nil, err
	}

	publicKey, err := hdNode.ECPubKey()
	if err != nil {
		return nil, nil, err
	}

	prvKey := (*btcec.PrivateKey)(privateKey.ToECDSA())
	pubKey := (*btcec.PublicKey)(publicKey.ToECDSA())
	return prvKey, pubKey, nil
}

type DeriveBlindingKeyPairArgs struct {
	Script []byte
}

func (a DeriveBlindingKeyPairArgs) validate() error {
	if len(a.Script) <= 0 {
		return ErrMissingOutputScript
	}
	return nil
}

// DeriveBlindingKeyPair derives the SLIP77 blinding key pair from the given
// output script.
func (w *Wallet) DeriveBlindingKeyPair(
	args DeriveBlindingKeyPairArgs,
) (*btcec.PrivateKey, *btcec.PublicKey, error) {
	if err := args.validate(); err != nil {
		return nil, nil, err
	}
	if err := w.validate(); err != nil {
		return nil, nil, err
	}
	if len(w.blindingMasterKey) <= 0 {
		return nil, nil, ErrMissingBlindingMasterKey
	}
	slip77Node, err := slip77.FromMasterKey(w.blindingMasterKey)
	if err != nil {
		return nil, nil, err
	}
	return slip77Node.DeriveKey(args.Script)
}

type DeriveConfidentialAddressArgs struct {
	DerivationPath string
	Network        *network.Network
}

func (a DeriveConfidentialAddressArgs) validate() error {
	derivationPath, err := ParseDerivationPath(a.DerivationPath)
	if err != nil {
		return err
	}

	err = checkDerivationPath(derivationPath)
	if err != nil {
		return err
	}

	if a.Network == nil {
		return ErrMissingNetwork
	}

	return nil
}

// DeriveConfidentialAddress derives both the HD signing and the SLIP-77
// blinding key-pairs to generate a confidential address.
func (w *Wallet) DeriveConfidentialAddress(
	args DeriveConfidentialAddressArgs,
) (string, []byte, error) {
	if err := args.validate(); err != nil {
		return "", nil, err
	}
	if err := w.validate(); err != nil {
		return "", nil, err
	}

	_, pubkey, err := w.DeriveSigningKeyPair(DeriveSigningKeyPairArgs{
		DerivationPath: args.DerivationPath,
	})
	if err != nil {
		return "", nil, err
	}

	script := payment.FromPublicKey(pubkey, args.Network, nil).WitnessScript

	_, blindingPubkey, err := w.DeriveBlindingKeyPair(DeriveBlindingKeyPairArgs{
		Script: script,
	})
	if err != nil {
		return "", nil, err
	}

	p2wpkh := payment.FromPublicKey(pubkey, args.Network, blindingPubkey)
	addr, err := p2wpkh.ConfidentialWitnessPubKeyHash()
	if err != nil {
		return "", nil, err
	}
	return addr, p2wpkh.WitnessScript, nil
}

func (w *Wallet) extendedPrivateKey(
	account uint32,
) (*hdkeychain.ExtendedKey, error) {
	masterKey, err := hdkeychain.NewKeyFromString(
		base58.Encode(w.signingMasterKey),
	)
	if err != nil {
		return nil, err
	}

	step := account + hdkeychain.HardenedKeyStart
	return masterKey.Derive(step)
}

func checkDerivationPath(path DerivationPath) error {
	if len(path) != 3 {
		return ErrInvalidDerivationPathLength
	}
	// first elem must be hardened!
	if path[0] < hdkeychain.HardenedKeyStart {
		return ErrInvalidDerivationPathAccount
	}
	return nil
}

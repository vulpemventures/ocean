package multisig

import (
	"encoding/hex"
	"sort"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	path "github.com/equitas-foundation/bamp-ocean/pkg/wallet/derivation-path"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/slip77"
)

// AccountExtendedPrivateKey returns the extended private key in base58 format
// for the given account index.
func (w *Wallet) AccountExtendedPrivateKey() (string, error) {
	if err := w.validate(); err != nil {
		return "", err
	}

	return base58.Encode(w.signingMasterKey), nil
}

// AccountExtendedPublicKey returns the extended public key in base58 format
// for the given account index.
func (w *Wallet) AccountExtendedPublicKey() (string, error) {
	if err := w.validate(); err != nil {
		return "", err
	}

	xprv, _ := hdkeychain.NewKeyFromString(
		base58.Encode(w.signingMasterKey),
	)

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

	return hex.EncodeToString(w.blindingMasterKey), nil
}

type DeriveSigningKeyPairArgs struct {
	DerivationPath string
}

func (a DeriveSigningKeyPairArgs) validate() error {
	derivationPath, err := path.ParseDerivationPath(a.DerivationPath)
	if err != nil {
		return err
	}

	return checkDerivationPath(derivationPath)
}

// DeriveSigningKeyPair derives the key pair from the given derivation path.
func (w *Wallet) DeriveSigningKeyPair(args DeriveSigningKeyPairArgs) (
	*btcec.PrivateKey, []*btcec.PublicKey, error,
) {
	if err := args.validate(); err != nil {
		return nil, nil, err
	}
	if err := w.validate(); err != nil {
		return nil, nil, err
	}

	hdNode, _ := hdkeychain.NewKeyFromString(
		base58.Encode(w.signingMasterKey),
	)
	walletXpub, _ := hdNode.Neuter()

	derivationPath, _ := path.ParseDerivationPath(args.DerivationPath)
	for _, step := range derivationPath {
		hdNode, _ = hdNode.Derive(step)
	}

	privateKey, err := hdNode.ECPrivKey()
	if err != nil {
		return nil, nil, err
	}

	xpubs := []string{walletXpub.String()}
	for _, xpub := range w.xpubs {
		if xpub != xpubs[0] {
			xpubs = append(xpubs, xpub)
		}
	}

	sort.SliceStable(xpubs, func(i, j int) bool {
		return xpubs[i] < xpubs[j]
	})

	pubKeys := make([]*btcec.PublicKey, 0, len(xpubs))
	for _, xpub := range xpubs {
		hdNode, _ := hdkeychain.NewKeyFromString(xpub)

		derivationPath, _ := path.ParseDerivationPath(args.DerivationPath)
		for _, step := range derivationPath {
			hdNode, _ = hdNode.Derive(step)
		}

		pubKey, err := hdNode.ECPubKey()
		if err != nil {
			return nil, nil, err
		}
		pubKeys = append(pubKeys, pubKey)
	}

	return privateKey, pubKeys, nil
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
	derivationPath, err := path.ParseDerivationPath(a.DerivationPath)
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
) (string, []byte, []byte, error) {
	if err := args.validate(); err != nil {
		return "", nil, nil, err
	}
	if err := w.validate(); err != nil {
		return "", nil, nil, err
	}

	_, pubkeys, err := w.DeriveSigningKeyPair(DeriveSigningKeyPairArgs{
		DerivationPath: args.DerivationPath,
	})
	if err != nil {
		return "", nil, nil, err
	}

	unconf, err := payment.FromPublicKeys(pubkeys, len(pubkeys), args.Network, nil)
	if err != nil {
		return "", nil, nil, err
	}

	_, blindingPubkey, err := w.DeriveBlindingKeyPair(DeriveBlindingKeyPairArgs{
		Script: unconf.WitnessScript,
	})
	if err != nil {
		return "", nil, nil, err
	}

	p2wsh, err := payment.FromPublicKeys(pubkeys, len(pubkeys), args.Network, blindingPubkey)
	if err != nil {
		return "", nil, nil, err
	}
	addr, err := p2wsh.ConfidentialWitnessScriptHash()
	if err != nil {
		return "", nil, nil, err
	}
	return addr, p2wsh.WitnessScript, p2wsh.Redeem.Script, nil
}

func checkDerivationPath(path path.DerivationPath) error {
	if len(path) != 2 {
		return ErrInvalidDerivationPathLength
	}
	// first 2 elems must not be hardened!
	if path[0] >= hdkeychain.HardenedKeyStart ||
		path[1] >= hdkeychain.HardenedKeyStart {
		return ErrInvalidDerivationPathAccount
	}
	return nil
}

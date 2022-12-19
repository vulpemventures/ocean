package singlesig

import (
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/vulpemventures/go-bip39"
	"github.com/vulpemventures/go-elements/slip77"
	path "github.com/vulpemventures/ocean/pkg/wallet/derivation-path"
)

/*
Utils for wallet creation.
*/
func generateSeedFromMnemonic(mnemonic []string) []byte {
	m := strings.Join(mnemonic, " ")
	return bip39.NewSeed(m, "")
}

func isMnemonicValid(mnemonic []string) bool {
	m := strings.Join(mnemonic, " ")
	return bip39.IsMnemonicValid(m)
}

func generateSigningMasterKey(
	seed []byte, derivationPath path.DerivationPath,
) ([]byte, error) {
	hdNode, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	for _, step := range derivationPath {
		hdNode, err = hdNode.Derive(step)
		if err != nil {
			return nil, err
		}
	}
	return base58.Decode(hdNode.String()), nil
}

func generateBlindingMasterKey(seed []byte) ([]byte, error) {
	slip77Node, err := slip77.FromSeed(seed)
	if err != nil {
		return nil, err
	}
	return slip77Node.MasterKey, nil
}

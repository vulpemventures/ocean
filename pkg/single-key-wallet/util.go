package wallet

import (
	"math"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/base58"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/vulpemventures/go-bip39"
	"github.com/vulpemventures/go-elements/slip77"
)

/*
	Utils for wallet creation.
*/
func generateMnemonic(entropySize int) ([]string, error) {
	entropy, err := bip39.NewEntropy(entropySize)
	if err != nil {
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}
	return strings.Split(mnemonic, " "), nil
}

func generateSeedFromMnemonic(mnemonic []string) []byte {
	m := strings.Join(mnemonic, " ")
	return bip39.NewSeed(m, "")
}

func isMnemonicValid(mnemonic []string) bool {
	m := strings.Join(mnemonic, " ")
	return bip39.IsMnemonicValid(m)
}

func generateSigningMasterKey(seed []byte, path DerivationPath) ([]byte, error) {
	hdNode, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	for _, step := range path {
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

/*
	Utils for fee estimation.
*/
func varIntSerializeSize(val uint64) int {
	// The value is small enough to be represented by itself, so it's
	// just 1 byte.
	if val < 0xfd {
		return 1
	}

	// Discriminant 1 byte plus 2 bytes for the uint16.
	if val <= math.MaxUint16 {
		return 3
	}

	// Discriminant 1 byte plus 4 bytes for the uint32.
	if val <= math.MaxUint32 {
		return 5
	}

	// Discriminant 1 byte plus 8 bytes for the uint64.
	return 9
}

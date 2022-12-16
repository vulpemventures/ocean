package multisig

import (
	"crypto/sha256"
	"errors"
	"math"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	path "github.com/equitas-foundation/bamp-ocean/pkg/wallet/derivation-path"
	"github.com/vulpemventures/go-bip39"
	"github.com/vulpemventures/go-elements/slip77"
)

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

func generateBlindingMasterKey(xPubKeys []string) ([]byte, error) {
	chainCodes := make([][]byte, 0, 3)
	for _, v := range xPubKeys {
		extendedPubKey, err := hdkeychain.NewKeyFromString(v)
		if err != nil {
			return nil, err
		}

		chainCodes = append(chainCodes, extendedPubKey.ChainCode())
	}

	blindingSeed, err := blindingKeyFromChainCode(chainCodes)
	if err != nil {
		return nil, err
	}

	masterKey, err := slip77.FromSeed(blindingSeed[:])
	if err != nil {
		return nil, err
	}

	return masterKey.MasterKey, nil
}

// blindingKeyFromChainCode returns sha256("blinding_key" + xor(chaincodes)) as a
// blinding key for multisig wallet.
// https://github.com/cryptoadvance/specter-desktop/blob/master/src/cryptoadvance/specter/liquid/wallet.py#L77-L85
// param chainCodes the co-signers xpubs chainCodes (from the first receiving address)
func blindingKeyFromChainCode(chainCodes [][]byte) ([32]byte, error) {
	prefix := []byte("blinding_key")
	chainCodesXOR := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	for _, v := range chainCodes {
		tmp, err := xor(chainCodesXOR, v)
		if err != nil {
			return [32]byte{}, err
		}
		chainCodesXOR = tmp
	}

	bk := append(prefix, chainCodesXOR...)

	return sha256.Sum256(bk), nil
}

func xor(a []byte, b []byte) ([]byte, error) {
	result := make([]byte, len(a))

	if len(a) != len(b) {
		return nil, errors.New("len(a) != len(b)")
	}

	for i := 0; i < len(a); i++ {
		result[i] = a[i] ^ b[i]
	}

	return result, nil
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

func varSliceSerializeSize(val []byte) int {
	return varIntSerializeSize(uint64(len(val))) + len(val)
}

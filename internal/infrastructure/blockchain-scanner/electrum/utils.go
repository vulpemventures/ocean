package electrum_scanner

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func calcScriptHash(script string) string {
	buf, _ := hex.DecodeString(script)
	hashedBuf := sha256.Sum256(buf)
	hash, _ := chainhash.NewHash(hashedBuf[:])
	return hash.String()
}

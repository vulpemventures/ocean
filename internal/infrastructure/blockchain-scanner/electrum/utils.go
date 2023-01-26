package electrum_scanner

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func calcScriptHash(script string) string {
	buf, _ := hex.DecodeString(script)
	hashedBuf := sha256.Sum256(buf)
	hash, _ := chainhash.NewHash(hashedBuf[:])
	return hash.String()
}

func getConn(addr string) (net.Conn, error) {
	split := strings.Split(addr, "://")
	proto, url := split[0], split[1]
	switch proto {
	case "tcp":
		return net.Dial(proto, url)
	case "ssl":
		return tls.Dial("tcp", url, nil)
	default:
		return nil, fmt.Errorf("invalid address: unknown prototocol")
	}
}

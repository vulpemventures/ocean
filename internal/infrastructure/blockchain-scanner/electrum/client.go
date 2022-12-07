package electrum_scanner

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/vulpemventures/go-elements/transaction"
	"github.com/vulpemventures/ocean/internal/core/domain"
)

type electrumClient interface {
	listen()
	close()

	subscribeForBlocks()
	subscribeForAccount(
		account string, addresses []domain.AddressInfo,
	) (chan accountReport, map[string][]txInfo)
	unsubscribeForAccount(account string)

	getLatestBlock() ([]byte, uint32, error)
	getBlockInfo(height uint32) (*chainhash.Hash, int64, error)
	getScriptHashHistory(scriptHash string) ([]txInfo, error)
	getTx(txid string) (*transaction.Transaction, error)
	getUtxos(outpoints []domain.Utxo) ([]domain.Utxo, error)
	broadcastTx(txHex string) (string, error)
}

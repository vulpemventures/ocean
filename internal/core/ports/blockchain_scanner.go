package ports

import (
	"github.com/vulpemventures/ocean/internal/core/domain"
)

// BlockchainScanner is the abstraction for any kind of service representing an
// Elements node. It gives info about txs and utxos related to one or more HD
// accounts in a aync way (via channels), and lets broadcast transactions over
// the Liquid network.
type BlockchainScanner interface {
	// Start starts the service.
	Start()
	// Stop stops the service.
	Stop()

	// WatchForAccount instructs the scanner to start notifying about txs/utxos
	// related to the given list of addresses belonging to the given HD account.
	WatchForAccount(
		namespace string, startingBlockHeight uint32,
		addresses []domain.AddressInfo,
	)
	WatchForUtxos(
		namespace string, utxos []domain.UtxoInfo,
	)
	// StopWatchForAccount instructs the scanner to stop notifying about
	// txs/utxos related to any address belonging to the given HD account.
	StopWatchForAccount(namespace string)

	// GetUtxoChannel returns the channel where notification about utxos realated
	// to the given HD account are sent.
	GetUtxoChannel(namespace string) chan []*domain.Utxo
	// GetTxChannel returns the channel where notification about txs realated to
	// the given HD account are sent.
	GetTxChannel(namespace string) chan *domain.Transaction

	// GetLatestBlock returns the header of the latest block of the blockchain.
	GetLatestBlock() ([]byte, uint32, error)
	// GetBlockHash returns the hash of the block identified by its height.
	GetBlockHash(height uint32) ([]byte, error)
	// GetUtxos is a sync function to get info about the utxos represented by
	// given outpoints (UtxoKeys).
	GetUtxos(utxos []domain.Utxo) ([]domain.Utxo, error)
	// BroadcastTransaction sends the given raw tx (in hex string) over the
	// network in order to be included in a later block of the Liquid blockchain.
	BroadcastTransaction(txHex string) (string, error)
}

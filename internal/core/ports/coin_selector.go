package ports

import "github.com/equitas-foundation/bamp-ocean/internal/core/domain"

// CoinSelector is the abstraction for any kind of service intended to return a
// subset of the given utxos with target asset hash, covering the target amount
// based on a specific strategy.
type CoinSelector interface {
	// SelectUtxos implements a certain coin selection strategy.
	SelectUtxos(
		utxos []*domain.Utxo, targetAmount uint64, targetAsset string,
	) (selectedUtxos []*domain.Utxo, change uint64, err error)
}

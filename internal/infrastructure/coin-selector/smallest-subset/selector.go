package smallestsubset_selector

import (
	"fmt"
	"sort"

	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
)

var (
	ErrBlindedUtxos           = fmt.Errorf("error on utxos: all confidential utxos must be already revealed")
	ErrTargetAmountNotReached = fmt.Errorf("not found enough utxos to cover target amount")
)

type selector struct{}

func NewSmallestSubsetCoinSelector() ports.CoinSelector {
	return &selector{}
}

func (s *selector) SelectUtxos(
	utxos []*domain.Utxo, targetAmount uint64, targetAsset string,
) ([]*domain.Utxo, uint64, error) {
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Value > utxos[j].Value
	})

	targetUtxos := make([]*domain.Utxo, 0)
	totalAmount := uint64(0)
	for i := range utxos {
		utxo := utxos[i]
		if utxo.IsConfidential() && !utxo.IsRevealed() {
			return nil, 0, ErrBlindedUtxos
		}
		if utxo.Asset == targetAsset {
			targetUtxos = append(targetUtxos, utxo)
		}
	}

	indexes := selectUtxos(targetAmount, targetUtxos)
	if len(indexes) <= 0 {
		return nil, 0, ErrTargetAmountNotReached
	}

	selectedUtxos := make([]*domain.Utxo, 0)
	for _, v := range indexes {
		totalAmount += targetUtxos[v].Value
		selectedUtxos = append(selectedUtxos, targetUtxos[v])
	}

	change := totalAmount - targetAmount
	return selectedUtxos, change, nil
}

// selectUtxos returns the index of the utxos that are going to be selected.
// The goal of this strategy is to select as less utxos as possible covering
// the target amount.
func selectUtxos(targetAmount uint64, utxos []*domain.Utxo) []int {
	utxoValues := []uint64{}
	for _, u := range utxos {
		utxoValues = append(utxoValues, u.Value)
	}

	list := getBestCombination(utxoValues, targetAmount)

	//since list variable contains values,
	//indexes holding those values needs to be calculated
	indexes := findIndexes(list, utxoValues)

	return indexes
}

func findIndexes(list []uint64, utxosValues []uint64) []int {
	var indexes []int
loop:
	for _, v := range list {
		for i, v1 := range utxosValues {
			if v == v1 {
				if isIndexOccupied(i, indexes) {
					continue
				} else {
					indexes = append(indexes, i)
					continue loop
				}
			}
		}
	}
	return indexes
}

func isIndexOccupied(i int, list []int) bool {
	for _, v := range list {
		if v == i {
			return true
		}
	}
	return false
}

// getBestCombination attempts to select as less items as possible
// covering the given target amount.
// The strategy here is to try finding exactly 1 utxo covering the given target
// amount or, otherwise, progressively increase the number of utxos until
// finding a combination that satisfies the criteria.
// If a combination exceeds the target amount, it is returned straightaway if
// its total amount is lower than 10 times the target one.
// Otherwise, if no combination satisfies this last criteria, the very first
// one found is returned.
func getBestCombination(items []uint64, target uint64) []uint64 {
	combinations := [][]uint64{}
	for i := 1; i < len(items)+1; i++ {
		combinations = append(combinations, getCombination(items, i, 0)...)
		for j := 0; j < len(combinations); j++ {
			total := sum(combinations[j])
			if total < target {
				continue
			}
			if total == target {
				return combinations[j]
			}
			if total <= target*10 {
				return combinations[j]
			}
		}
	}

	for _, combo := range combinations {
		if totalAmount := sum(combo); totalAmount >= target {
			return combo
		}
	}

	return []uint64{}
}

var combination = []uint64{}

// getCombination returns all combinations of size elements from the src slice.
func getCombination(src []uint64, size int, offset int) [][]uint64 {
	result := [][]uint64{}
	if size == 0 {
		temp := make([]uint64, len(combination))
		copy(temp, combination)
		return append(result, temp)
	}
	for i := offset; i <= len(src)-size; i++ {
		combination = append(combination, src[i])
		temp := getCombination(src, size-1, i+1)
		result = append(result, temp...)
		combination = combination[:len(combination)-1]
	}
	return result[:]
}

func sum(items []uint64) uint64 {
	var total uint64
	for _, v := range items {
		total += v
	}
	return total
}

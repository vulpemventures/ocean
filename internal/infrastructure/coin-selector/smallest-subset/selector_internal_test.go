package smallestsubset_selector

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBestPairs(t *testing.T) {
	type args struct {
		items  []uint64
		target uint64
	}
	tests := []struct {
		name     string
		args     args
		expected []uint64
	}{
		{
			name: "1",
			args: args{
				items:  []uint64{61, 61, 61, 38, 61, 61, 61, 1, 1, 1, 3},
				target: 6,
			},
			expected: []uint64{38},
		},
		{
			name: "2",
			args: args{
				items:  []uint64{61, 61, 61, 61, 61, 61, 1, 1, 1, 3},
				target: 6,
			},
			expected: []uint64{3, 1, 1, 1},
		},
		{
			name: "3",
			args: args{
				items:  []uint64{61, 61},
				target: 6,
			},
			expected: []uint64{61},
		},
		{
			name: "4",
			args: args{
				items:  []uint64{2, 2},
				target: 6,
			},
			expected: []uint64{},
		},
		{
			name: "5",
			args: args{
				items:  []uint64{61, 1, 1, 1, 3, 56},
				target: 6,
			},
			expected: []uint64{56},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Slice(tt.args.items, func(i, j int) bool {
				return tt.args.items[i] > tt.args.items[j]
			})
			combos := getBestCombination(tt.args.items, tt.args.target)
			require.Equal(t, tt.expected, combos)
		})
	}
}

func TestFindIndexes(t *testing.T) {
	type args struct {
		list                 []uint64
		unblindedUtxosValues []uint64
	}
	tests := []struct {
		name     string
		args     args
		expected []int
	}{
		{
			name: "1",
			args: args{
				list:                 []uint64{1000},
				unblindedUtxosValues: []uint64{1000, 1000, 1000},
			},
			expected: []int{0},
		},
		{
			name: "2",
			args: args{
				list:                 []uint64{1000, 1000},
				unblindedUtxosValues: []uint64{1000, 2000, 1000},
			},
			expected: []int{0, 2},
		},
		{
			name: "3",
			args: args{
				list: []uint64{2000, 2000},
				unblindedUtxosValues: []uint64{1000, 2000, 1000, 2000, 2000,
					2000},
			},
			expected: []int{1, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexes := findIndexes(tt.args.list, tt.args.unblindedUtxosValues)
			require.Equal(t, tt.expected, indexes)
		})
	}
}

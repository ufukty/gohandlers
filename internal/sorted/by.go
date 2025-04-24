package sorted

import (
	"cmp"
	"iter"
	"maps"
	"slices"
)

func ByValues[K comparable, V cmp.Ordered](m map[K]V) iter.Seq2[K, V] {
	sorted := slices.SortedFunc(maps.Keys(m), func(a, b K) int {
		return cmp.Compare(m[a], m[b])
	})
	return func(yield func(K, V) bool) {
		for _, k := range sorted {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}

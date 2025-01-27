package sorted

import (
	"cmp"
	"iter"
	"slices"

	"golang.org/x/exp/maps"
)

func ByValues[K comparable, V cmp.Ordered](m map[K]V) iter.Seq2[K, V] {
	o := maps.Keys(m)
	slices.SortFunc(o, func(a, b K) int {
		va := m[a]
		vb := m[b]
		if va < vb {
			return -1
		} else if va > vb {
			return 1
		} else {
			return 0
		}
	})
	return func(yield func(K, V) bool) {
		for _, k := range o {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}

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
		return cmp.Compare(m[a], m[b])
	})
	return func(yield func(K, V) bool) {
		for _, k := range o {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}

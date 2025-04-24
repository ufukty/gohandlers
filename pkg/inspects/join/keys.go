package join

import (
	"maps"
	"slices"
	"strings"
)

func Keys[V any](m map[string]V, sep string) string {
	return strings.Join(slices.Collect(maps.Keys(m)), sep)
}

func Values[K comparable](m map[K]string, sep string) string {
	return strings.Join(slices.Collect(maps.Values(m)), sep)
}

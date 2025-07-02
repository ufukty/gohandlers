package sort

import (
	"cmp"
	"go/ast"
	"slices"
	"strings"
)

func thirdparty(s string) bool {
	return strings.Contains(strings.Split(s, "/")[0], ".")
}

func localFirst(a, b string) int {
	if !thirdparty(a) && thirdparty(b) {
		return -1
	} else if thirdparty(a) && !thirdparty(b) {
		return 1
	}
	return 0
}

func Imports(specs []ast.Spec) {
	slices.SortFunc(specs, func(a, b ast.Spec) int {
		sa := a.(*ast.ImportSpec).Path.Value
		sb := b.(*ast.ImportSpec).Path.Value
		return cmp.Or(localFirst(sa, sb), cmp.Compare(sa, sb))
	})
}

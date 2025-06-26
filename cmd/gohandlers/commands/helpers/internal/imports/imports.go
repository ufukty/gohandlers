package imports

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"github.com/ufukty/gohandlers/pkg/inspects"
)

// bq.Parse and bs.Parse needs for content type check
// bq.Build needs for join in url building
// bq.Build needs for route parameter replacement
func needsStrings(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, infos := range infoss {
		for _, info := range infos {
			if info.RequestType != nil || info.ResponseType != nil {
				return true
			}
		}
	}
	return false
}

func needsJson(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, infos := range infoss {
		for _, info := range infos {
			if info.RequestType != nil && len(info.RequestType.Params.Json) > 0 {
				return true
			}
			if info.ResponseType != nil && len(info.ResponseType.Params.Json) > 0 {
				return true
			}
		}
	}
	return false
}

func needsBytes(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, infos := range infoss {
		for _, info := range infos {
			if info.RequestType != nil && info.RequestType.ContainsBody {
				return true
			}
			if info.ResponseType != nil && info.ResponseType.ContainsBody {
				return true
			}
		}
	}
	return false
}

func thirdparty(s string) bool {
	return strings.Contains(strings.Split(s, "/")[0], ".")
}

func sortImports(specs []ast.Spec) {
	slices.SortFunc(specs, func(a, b ast.Spec) int {
		sa := a.(*ast.ImportSpec).Path.Value
		sb := b.(*ast.ImportSpec).Path.Value
		if !thirdparty(sa) && thirdparty(sb) {
			return -1
		} else if thirdparty(sa) && !thirdparty(sb) {
			return 1
		} else if sa < sb {
			return -1
		} else if sa == sb {
			return 0
		} else {
			return 1
		}
	})
}

func List(infoss map[inspects.Receiver]map[string]inspects.Info) []ast.Spec {
	imports := []ast.Spec{
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`}},
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/ufukty/gohandlers/pkg/gohandlers"`}},
	}
	if needsBytes(infoss) {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"bytes"`}},
		)
	}
	if needsJson(infoss) {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"encoding/json"`}},
		)
	}
	if needsStrings(infoss) {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"strings"`}},
		)
	}
	sortImports(imports)
	return imports
}

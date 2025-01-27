package bindings

import (
	"go/ast"
	"go/token"
	"gohandlers/pkg/inspects"
	"slices"
)

func needsStrings(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, infos := range infoss {
		for _, info := range infos {
			if info.RequestType != nil {
				return true
			}
		}
	}
	return false
}

func containsHandlerWithJsonBody(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
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

func containsHandlerWithMultipartBody(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, infos := range infoss {
		for _, info := range infos {
			if info.RequestType != nil && (len(info.RequestType.Params.Part) > 0 || len(info.RequestType.Params.File) > 0) {
				return true
			}
			if info.ResponseType != nil && (len(info.ResponseType.Params.Part) > 0 || len(info.ResponseType.Params.File) > 0) {
				return true
			}
		}
	}
	return false
}

func containsHandlerNeedsBytes(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
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

func imports(infoss map[inspects.Receiver]map[string]inspects.Info) []ast.Spec {
	imports := []ast.Spec{
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`}},
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
	}
	if containsHandlerNeedsBytes(infoss) {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"bytes"`}},
		)
	}
	if containsHandlerWithJsonBody(infoss) {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"encoding/json"`}},
		)
	}
	if containsHandlerWithMultipartBody(infoss) {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"mime/multipart"`}},
		)
	}
	if needsStrings(infoss) {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"strings"`}},
		)
	}
	slices.SortFunc(imports, func(a, b ast.Spec) int {
		av := a.(*ast.ImportSpec).Path.Value
		bv := b.(*ast.ImportSpec).Path.Value
		if av < bv {
			return -1
		} else if av == bv {
			return 0
		} else {
			return 1
		}
	})
	return imports
}

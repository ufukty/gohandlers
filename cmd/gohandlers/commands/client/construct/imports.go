package construct

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/ufukty/gohandlers/pkg/inspects"
)

func importNetHttp(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, infos := range infoss {
		for _, hi := range infos {
			if hi.ResponseType == nil {
				return true
			}
		}
	}
	return false
}

func imports(importpkg string) ast.Decl {
	imports := []ast.Spec{
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`}},
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
	}
	if importpkg != "" {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", importpkg)}},
		)
	}
	slices.SortFunc(imports, func(a, b ast.Spec) int {
		va := a.(*ast.ImportSpec).Path.Value
		vb := b.(*ast.ImportSpec).Path.Value
		if va < vb {
			return -1
		} else if va == vb {
			return 0
		} else {
			return 1
		}
	})
	return &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: imports,
	}
}

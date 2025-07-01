package construct

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
)

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

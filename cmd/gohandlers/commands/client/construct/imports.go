package construct

import (
	"fmt"
	"go/ast"
	"go/token"

	"go.ufukty.com/gohandlers/cmd/gohandlers/internal/pretty/sort"
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
	sort.Imports(imports)
	return &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: imports,
	}
}

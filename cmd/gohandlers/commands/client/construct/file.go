package construct

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/ufukty/gohandlers/pkg/inspects"
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

func File(infoss map[inspects.Receiver]map[string]inspects.Info, pkgdst, pkgsrc, importpkg string) *ast.File {
	f := &ast.File{
		Name: &ast.Ident{Name: pkgdst},
		Decls: []ast.Decl{
			imports(importpkg),
			pool(),
			client(),
			clientConstructor(),
		},
	}

	f.Decls = append(f.Decls, clientMethods(infoss, pkgsrc, importpkg)...)

	return f
}

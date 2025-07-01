package construct

import (
	"go/ast"

	"github.com/ufukty/gohandlers/pkg/inspects"
)

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

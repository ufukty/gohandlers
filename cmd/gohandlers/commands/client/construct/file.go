package construct

import (
	"go/ast"

	"github.com/ufukty/gohandlers/pkg/inspects"
)

func File(infoss map[inspects.Receiver]map[string]inspects.Info, pkgdst, pkgsrc, importpkg string) *ast.File {
	f := &ast.File{
		Name:  &ast.Ident{Name: pkgdst},
		Decls: []ast.Decl{imports(importpkg)},
	}
	f.Decls = append(f.Decls,
		iface(infoss, pkgsrc, importpkg != ""),
		mockstruct(infoss, pkgsrc, importpkg != ""),
	)
	f.Decls = append(f.Decls,
		mockmethods(infoss, pkgsrc, importpkg != "")...,
	)
	f.Decls = append(f.Decls,
		pool(),
		client(),
		clientConstructor(),
	)
	f.Decls = append(f.Decls,
		clientMethods(infoss, pkgsrc, importpkg)...,
	)

	return f
}

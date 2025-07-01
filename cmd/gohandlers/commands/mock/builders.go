package mock

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

func imports(infoss map[inspects.Receiver]map[string]inspects.Info, importpkg string) *ast.GenDecl {
	imports := []ast.Spec{}
	if importNetHttp(infoss) {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
		)
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

	gd := &ast.GenDecl{Tok: token.IMPORT, Specs: imports}
	return gd
}

func methodtype(hi inspects.Info, pkgsrc string, imported bool) *ast.FuncType {
	var bq ast.Expr
	var bn ast.Expr

	if hi.RequestType == nil {
		// TODO:
	} else if imported {
		bq = &ast.SelectorExpr{X: &ast.Ident{Name: pkgsrc}, Sel: &ast.Ident{Name: hi.RequestType.Typename}}
	} else {
		bq = &ast.Ident{Name: hi.RequestType.Typename}
	}

	if hi.ResponseType == nil {
		bn = &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Response"}}
	} else if imported {
		bn = &ast.SelectorExpr{X: &ast.Ident{Name: pkgsrc}, Sel: &ast.Ident{Name: hi.ResponseType.Typename}}
	} else {
		bn = &ast.Ident{Name: hi.ResponseType.Typename}
	}

	ft := &ast.FuncType{
		Params: &ast.FieldList{List: []*ast.Field{{Type: &ast.StarExpr{X: bq}}}},
		Results: &ast.FieldList{List: []*ast.Field{
			{Type: &ast.StarExpr{X: bn}},
			{Type: &ast.Ident{Name: "error"}},
		}},
	}

	return ft
}

func iface(infoss map[inspects.Receiver]map[string]inspects.Info, pkgsrc string, imported bool) *ast.GenDecl {
	list := []*ast.Field{}

	for _, infos := range infoss {
		for hn, hi := range infos {
			if hi.RequestType == nil {
				continue // TODO:
			}
			list = append(list,
				&ast.Field{
					Names: []*ast.Ident{{Name: hn}},
					Type:  methodtype(hi, pkgsrc, imported),
				},
			)
		}
	}

	slices.SortFunc(list, func(a, b *ast.Field) int {
		na := a.Names[0].Name
		nb := b.Names[0].Name
		if na < nb {
			return -1
		} else if na == nb {
			return 0
		} else {
			return 1
		}
	})

	gd := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: "Interface"},
				Type: &ast.InterfaceType{Methods: &ast.FieldList{List: list}},
			},
		},
	}

	return gd
}

func mockstruct() *ast.GenDecl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: "Mock"},
				Type: &ast.StructType{Fields: &ast.FieldList{}},
			},
		},
	}
}

func mockmethods(infoss map[inspects.Receiver]map[string]inspects.Info, pkgsrc string, imported bool) []ast.Decl {
	ds := []ast.Decl{}
	for _, infos := range infoss {
		for hn, hi := range infos {
			if hi.RequestType == nil {
				continue // TODO:
			}
			ds = append(ds, &ast.FuncDecl{
				Recv: &ast.FieldList{List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "m"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: "Mock"}}},
				}},
				Name: &ast.Ident{Name: hn},
				Type: methodtype(hi, pkgsrc, imported),
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "nil"}, &ast.Ident{Name: "nil"}}},
				}},
			})
		}
	}

	slices.SortFunc(ds, func(a, b ast.Decl) int {
		fa := a.(*ast.FuncDecl).Name.Name
		fb := b.(*ast.FuncDecl).Name.Name

		if fa < fb {
			return -1
		} else if fa == fb {
			return 0
		} else {
			return 1
		}
	})
	return ds
}

func file(infoss map[inspects.Receiver]map[string]inspects.Info, pkgdst, pkgsrc, importpkg string) *ast.File {
	f := &ast.File{
		Name: &ast.Ident{Name: pkgdst},
		Decls: []ast.Decl{
			imports(infoss, importpkg),
			iface(infoss, pkgsrc, importpkg != ""),
			mockstruct(),
		},
	}

	f.Decls = append(f.Decls, mockmethods(infoss, pkgsrc, importpkg != "")...)

	return f
}

package construct

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/ufukty/gohandlers/pkg/inspects"
)

func imports(importpkg string) []ast.Spec {
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
	return imports
}

func File(infoss map[inspects.Receiver]map[string]inspects.Info, pkgdst, pkgsrc, importpkg string) *ast.File {
	f := &ast.File{
		Name:  &ast.Ident{Name: pkgdst},
		Decls: []ast.Decl{},
	}

	f.Decls = append(f.Decls,
		&ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: imports(importpkg),
		},
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{Name: "Pool"},
					Type: &ast.InterfaceType{Methods: &ast.FieldList{List: []*ast.Field{{
						Names: []*ast.Ident{{Name: "Host"}},
						Type: &ast.FuncType{
							Params:  &ast.FieldList{},
							Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "string"}}, {Type: &ast.Ident{Name: "error"}}}},
						},
					}}}},
				},
			},
		},
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{Name: "Client"},
					Type: &ast.StructType{Fields: &ast.FieldList{List: []*ast.Field{
						{Names: []*ast.Ident{{Name: "p"}}, Type: &ast.Ident{Name: "Pool"}},
					}}},
				},
			},
		},
		&ast.FuncDecl{
			Name: &ast.Ident{Name: "NewClient"},
			Type: &ast.FuncType{
				Params: &ast.FieldList{List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "p"}}, Type: &ast.Ident{Name: "Pool"}},
				}},
				Results: &ast.FieldList{List: []*ast.Field{
					{Type: &ast.StarExpr{X: &ast.Ident{Name: "Client"}}},
				}},
			},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{
					&ast.UnaryExpr{
						Op: token.AND,
						X: &ast.CompositeLit{
							Type: &ast.Ident{Name: "Client"},
							Elts: []ast.Expr{
								&ast.KeyValueExpr{Key: &ast.Ident{Name: "p"}, Value: &ast.Ident{Name: "p"}},
							},
						},
					},
				}},
			}},
		},
	)

	fds := []ast.Decl{}
	for _, infos := range infoss {
		for hn, hi := range infos {
			if hi.RequestType == nil {
				continue
			}
			fds = append(fds, clientMethod(hn, hi, pkgsrc, importpkg != ""))
		}
	}
	slices.SortFunc(fds, func(a, b ast.Decl) int {
		if a.(*ast.FuncDecl).Name.Name < b.(*ast.FuncDecl).Name.Name {
			return -1
		} else if a.(*ast.FuncDecl).Name.Name == b.(*ast.FuncDecl).Name.Name {
			return 0
		} else {
			return 1
		}
	})

	f.Decls = append(f.Decls, fds...)

	return f
}

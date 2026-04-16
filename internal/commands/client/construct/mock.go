package construct

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"go.ufukty.com/gohandlers/pkg/inspects"
)

func methodtype(hi inspects.Info, pkgsrc string, imported, namedparams bool) *ast.FuncType {
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

	param1 := &ast.Field{Type: &ast.StarExpr{X: bq}}
	if namedparams {
		param1.Names = append(param1.Names, &ast.Ident{Name: "bq"})
	}

	ft := &ast.FuncType{
		Params: &ast.FieldList{List: []*ast.Field{param1}},
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
					Type:  methodtype(hi, pkgsrc, imported, false),
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

func mockstruct(infoss map[inspects.Receiver]map[string]inspects.Info, pkgsrc string, imported bool) *ast.GenDecl {
	fs := []*ast.Field{}
	for _, infos := range infoss {
		for hn, hi := range infos {
			if hi.RequestType == nil {
				continue // TODO:
			}
			fs = append(fs, &ast.Field{
				Names: []*ast.Ident{{Name: hn + "Func"}},
				Type:  methodtype(hi, pkgsrc, imported, false),
			})
		}
	}
	slices.SortFunc(fs, func(a, b *ast.Field) int {
		return cmp.Compare(string(a.Names[0].Name), string(b.Names[0].Name))
	})
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: "Mock"},
				Type: &ast.StructType{Fields: &ast.FieldList{List: fs}},
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
				Type: methodtype(hi, pkgsrc, imported, true),
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{
							X:  &ast.SelectorExpr{X: &ast.Ident{Name: "m"}, Sel: &ast.Ident{Name: hn + "Func"}},
							Op: token.EQL,
							Y:  &ast.Ident{Name: "nil"},
						},
						Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{
							&ast.Ident{Name: "nil"},
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: fmt.Sprintf(`"not implemented: %s"`, hn),
									},
								},
							},
						}}}},
					},
					&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "m"}, Sel: &ast.Ident{Name: hn + "Func"}},
						Args: []ast.Expr{&ast.Ident{Name: "bq"}},
					}}},
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

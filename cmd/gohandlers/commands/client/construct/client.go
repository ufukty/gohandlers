package construct

import (
	"go/ast"
	"go/token"
	"slices"

	"github.com/ufukty/gohandlers/pkg/inspects"
)

func ternary[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

func pool() ast.Decl {
	return &ast.GenDecl{
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
	}
}

func client() ast.Decl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: "Client"},
				Type: &ast.StructType{Fields: &ast.FieldList{List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "p"}}, Type: &ast.Ident{Name: "Pool"}},
				}}},
			},
		},
	}
}

func clientConstructor() ast.Decl {
	return &ast.FuncDecl{
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
	}
}

func clientMethod(hn string, hi inspects.Info, pkgsrc string, imported bool) *ast.FuncDecl {
	fd := &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{{Name: "c"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: "Client"}}},
		}},
		Name: &ast.Ident{Name: hn},
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "bq"}},
					Type: &ast.StarExpr{X: ternary[ast.Expr](
						imported,
						&ast.SelectorExpr{X: &ast.Ident{Name: pkgsrc}, Sel: &ast.Ident{Name: hi.RequestType.Typename}},
						&ast.Ident{Name: hi.RequestType.Typename},
					)},
				},
			}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}

	if hi.ResponseType != nil {
		rt := ternary[ast.Expr](
			imported,
			&ast.SelectorExpr{X: &ast.Ident{Name: pkgsrc}, Sel: &ast.Ident{Name: hi.ResponseType.Typename}},
			&ast.Ident{Name: hi.ResponseType.Typename},
		)
		fd.Type.Results = &ast.FieldList{List: []*ast.Field{
			{Type: &ast.StarExpr{X: rt}},
			{Type: &ast.Ident{Name: "error"}},
		}}
	} else {
		fd.Type.Results = &ast.FieldList{List: []*ast.Field{
			{Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Response"}}}},
			{Type: &ast.Ident{Name: "error"}},
		}}
	}

	fd.Body.List = append(fd.Body.List,
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "h"}, &ast.Ident{Name: "err"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.SelectorExpr{X: &ast.Ident{Name: "c"}, Sel: &ast.Ident{Name: "p"}},
						Sel: &ast.Ident{Name: "Host"},
					},
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
					&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
						Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"selecting host: %w"`}, &ast.Ident{Name: "err"}},
					},
				}},
			}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "rq"}, &ast.Ident{Name: "err"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: "Build"}},
					Args: []ast.Expr{&ast.Ident{Name: "h"}},
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
					&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
						Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"building request: %w"`}, &ast.Ident{Name: "err"}},
					},
				}},
			}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "rs"}, &ast.Ident{Name: "err"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "DefaultClient"}},
					Sel: &ast.Ident{Name: "Do"},
				},
				Args: []ast.Expr{
					&ast.Ident{Name: "rq"},
				},
			}},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{Name: "nil"},
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: `"sending: %w"`},
									&ast.Ident{Name: "err"},
								},
							},
						},
					},
				},
			},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  &ast.SelectorExpr{X: &ast.Ident{Name: "rs"}, Sel: &ast.Ident{Name: "StatusCode"}},
				Op: token.NEQ,
				Y:  &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "StatusOK"}},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{Name: ternary(hi.ResponseType != nil, "nil", "rs")},
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: `"non-200 status code: %d (%s)"`},
									&ast.SelectorExpr{X: &ast.Ident{Name: "rs"}, Sel: &ast.Ident{Name: "StatusCode"}},
									&ast.CallExpr{
										Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "StatusText"}},
										Args: []ast.Expr{&ast.SelectorExpr{X: &ast.Ident{Name: "rs"}, Sel: &ast.Ident{Name: "StatusCode"}}},
									},
								},
							},
						},
					},
				},
			},
		},
	)

	if hi.ResponseType != nil {
		rt := ternary[ast.Expr](
			imported,
			&ast.SelectorExpr{X: &ast.Ident{Name: pkgsrc}, Sel: &ast.Ident{Name: hi.ResponseType.Typename}},
			&ast.Ident{Name: hi.ResponseType.Typename},
		)

		fd.Body.List = append(fd.Body.List,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "bs"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.UnaryExpr{Op: token.AND, X: &ast.CompositeLit{Type: rt}}},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "bs"}, Sel: &ast.Ident{Name: "Parse"}},
						Args: []ast.Expr{&ast.Ident{Name: "rs"}},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ReturnStmt{Results: []ast.Expr{
						&ast.Ident{Name: "nil"},
						&ast.CallExpr{
							Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
							Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"parsing response: %w"`}, &ast.Ident{Name: "err"}},
						},
					}},
				}},
			},
		)
	}

	fd.Body.List = append(fd.Body.List,
		&ast.ReturnStmt{
			Results: []ast.Expr{&ast.Ident{Name: ternary(hi.ResponseType != nil, "bs", "rs")}, &ast.Ident{Name: "nil"}},
		},
	)

	return fd
}

func clientMethods(infoss map[inspects.Receiver]map[string]inspects.Info, pkgsrc, importpkg string) []ast.Decl {
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
	return fds
}

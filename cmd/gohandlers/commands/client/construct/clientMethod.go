package construct

import (
	"go/ast"
	"go/token"

	"github.com/ufukty/gohandlers/pkg/inspects"
)

func ternary[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
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
						Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"Host: %w"`}, &ast.Ident{Name: "err"}},
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
						Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"Build: %w"`}, &ast.Ident{Name: "err"}},
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
									&ast.BasicLit{Kind: token.STRING, Value: `"Do: %w"`},
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
							Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"Parse: %w"`}, &ast.Ident{Name: "err"}},
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

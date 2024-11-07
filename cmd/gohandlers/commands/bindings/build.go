package bindings

import (
	"fmt"
	"go/ast"
	"go/token"
	"gohandlers/pkg/inspects"
	"slices"
)

// produces the bqtn.Build method
func build(info inspects.Info) *ast.FuncDecl {
	fd := &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{{Name: "bq"}}, Type: &ast.Ident{Name: info.RequestType.Typename}},
		}},
		Name: &ast.Ident{Name: "Build"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "lb"}},
					Type:  &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "balancer"}, Sel: &ast.Ident{Name: "LoadBalancer"}}},
				},
			}},
			Results: &ast.FieldList{List: []*ast.Field{
				{Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Request"}}}},
				{Type: &ast.Ident{Name: "error"}},
			}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}

	fd.Body.List = append(fd.Body.List,
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "uri"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.Path)}},
		},
	)
	replacements := []ast.Stmt{}
	for routeparam, fieldname := range info.RequestType.RouteParams {
		replacements = append(replacements,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "uri"}},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "strings"}, Sel: &ast.Ident{Name: "Replace"}},
						Args: []ast.Expr{
							&ast.Ident{Name: "uri"},
							&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", fmt.Sprintf("{%s}", routeparam))},
							&ast.CallExpr{
								Fun:  &ast.Ident{Name: "string"},
								Args: []ast.Expr{&ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fieldname}}},
							},
							&ast.BasicLit{Kind: token.INT, Value: "1"},
						},
					},
				},
			},
		)
	}
	slices.SortFunc(replacements, func(a, b ast.Stmt) int {
		va := a.(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Args[1].(*ast.BasicLit).Value
		vb := b.(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Args[1].(*ast.BasicLit).Value
		if va < vb {
			return -1
		} else if va == vb {
			return 0
		} else {
			return 1
		}
	})
	fd.Body.List = append(fd.Body.List, replacements...)

	if info.RequestType.ContainsBody {
		fd.Body.List = append(fd.Body.List,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "body"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "bytes"}, Sel: &ast.Ident{Name: "NewBuffer"}},
					Args: []ast.Expr{&ast.CompositeLit{Type: &ast.ArrayType{Elt: &ast.Ident{Name: "byte"}}}},
				}},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.CallExpr{
							Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "json"}, Sel: &ast.Ident{Name: "NewEncoder"}},
							Args: []ast.Expr{&ast.Ident{Name: "body"}},
						},
						Sel: &ast.Ident{Name: "Encode"},
					},
					Args: []ast.Expr{&ast.Ident{Name: "bq"}},
				}},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
						Args: []ast.Expr{
							&ast.BasicLit{Kind: token.STRING, Value: `"json.Encoder.Encode: %w"`},
							&ast.Ident{Name: "err"},
						},
					},
				}}}},
			},
		)
	}

	fd.Body.List = append(fd.Body.List,
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "h"}, &ast.Ident{Name: "err"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "lb"}, Sel: &ast.Ident{Name: "Next"}}, Args: nil}},
		},
		&ast.IfStmt{
			Init: nil,
			Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{
				&ast.Ident{Name: "nil"},
				&ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
					Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"lb.Next: %w"`}, &ast.Ident{Name: "err"}},
				},
			}}}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "r"}, &ast.Ident{Name: "err"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "NewRequest"}},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.Method)},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "urls"}, Sel: &ast.Ident{Name: "Join"}},
						Args: []ast.Expr{
							&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "h"}, Sel: &ast.Ident{Name: "String"}}, Args: nil},
							&ast.Ident{Name: "uri"},
						},
					},
					&ast.Ident{Name: ternary(info.RequestType.ContainsBody, "body", "nil")},
				},
			}},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
					&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
						Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"http.NewRequest: %w"`}, &ast.Ident{Name: "err"}},
					},
				}},
			}},
		})

	if info.RequestType.ContainsBody {
		fd.Body.List = append(fd.Body.List,
			&ast.ExprStmt{X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "r"}, Sel: &ast.Ident{Name: "Header"}},
					Sel: &ast.Ident{Name: "Set"},
				},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"Content-Type"`},
					&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "mime"}, Sel: &ast.Ident{Name: "TypeByExtension"}},
						Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"json"`}},
					},
				},
			}},
			&ast.ExprStmt{X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "r"}, Sel: &ast.Ident{Name: "Header"}},
					Sel: &ast.Ident{Name: "Set"},
				},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"Content-Length"`},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Sprintf"}},
						Args: []ast.Expr{
							&ast.BasicLit{Kind: token.STRING, Value: `"%d"`},
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "body"}, Sel: &ast.Ident{Name: "Len"}},
							},
						},
					},
				},
			}},
		)
	}

	fd.Body.List = append(fd.Body.List,
		&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "r"}, &ast.Ident{Name: "nil"}}},
	)

	return fd
}

func buildfuncs(infoss map[inspects.Receiver]map[string]inspects.Info) []ast.Decl {
	fds := []ast.Decl{}
	for _, infos := range infoss {
		for _, info := range infos {
			if info.RequestType != nil {
				fds = append(fds, build(info))
			}
		}
	}
	slices.SortFunc(fds, func(a, b ast.Decl) int {
		na := a.(*ast.FuncDecl).Recv.List[0].Type.(*ast.Ident).Name
		nb := b.(*ast.FuncDecl).Recv.List[0].Type.(*ast.Ident).Name
		if na < nb {
			return -1
		} else if na == nb {
			return 0
		} else {
			return 1
		}
	})
	return fds
}

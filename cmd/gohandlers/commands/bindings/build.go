package bindings

import (
	"fmt"
	"go/ast"
	"go/token"
	"gohandlers/pkg/inspects"
	"slices"

	"golang.org/x/exp/maps"
)

// produces the bqtn.Build method
func bqBuild(info inspects.Info) *ast.FuncDecl {
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

	type symboltable struct {
		err     bool
		encoded bool
	}
	symbols := symboltable{}

	fd.Body.List = append(fd.Body.List,
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "uri"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes(info.Path)}},
		},
	)
	replacements := []ast.Stmt{}
	o := maps.Keys(info.RequestType.RouteParams)
	slices.SortFunc(o, func(a, b string) int {
		va := info.RequestType.RouteParams[a]
		vb := info.RequestType.RouteParams[b]
		if va < vb {
			return -1
		} else if va > vb {
			return 1
		} else {
			return 0
		}
	})
	for _, routeparam := range o {
		fieldname := info.RequestType.RouteParams[routeparam]
		replacements = append(replacements,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "encoded"}, &ast.Ident{Name: "err"}},
				Tok: ternary(symbols.encoded && symbols.err, token.ASSIGN, token.DEFINE),
				Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: "Root"}}, Sel: &ast.Ident{Name: "ToRoute"}}}},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ReturnStmt{Results: []ast.Expr{
						&ast.Ident{Name: "nil"},
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: quotes(fmt.Sprintf("%s.%s.ToRoute: %%w", info.RequestType.Typename, fieldname))},
								&ast.Ident{Name: "err"},
							},
						},
					}},
				}},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "uri"}},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "strings"}, Sel: &ast.Ident{Name: "Replace"}},
						Args: []ast.Expr{
							&ast.Ident{Name: "uri"},
							&ast.BasicLit{Kind: token.STRING, Value: quotes(fmt.Sprintf("{%s}", routeparam))},
							&ast.Ident{Name: "encoded"},
							&ast.BasicLit{Kind: token.INT, Value: "1"},
						},
					},
				},
			},
		)
		symbols.encoded = true
		symbols.err = true
	}
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

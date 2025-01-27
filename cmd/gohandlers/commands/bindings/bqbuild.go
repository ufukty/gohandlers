package bindings

import (
	"fmt"
	"go/ast"
	"go/token"
	"gohandlers/internal/sorted"
	"gohandlers/pkg/inspects"
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
				{Names: []*ast.Ident{{Name: "host"}}, Type: &ast.Ident{Name: "string"}},
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
		ok      bool
	}
	symbols := symboltable{}

	fd.Body.List = append(fd.Body.List,
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "uri"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes(info.Path)}},
		},
	)

	for rp, fn := range sorted.ByValues(info.RequestType.Params.Route) {
		fd.Body.List = append(fd.Body.List,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "encoded"}, &ast.Ident{Name: "err"}},
				Tok: ternary(symbols.encoded && symbols.err, token.ASSIGN, token.DEFINE),
				Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}},
					Sel: &ast.Ident{Name: "ToRoute"},
				}}},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ReturnStmt{Results: []ast.Expr{
						&ast.Ident{Name: "nil"},
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: quotes(fmt.Sprintf("%s.%s.ToRoute: %%w", info.RequestType.Typename, fn))},
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
							&ast.BasicLit{Kind: token.STRING, Value: quotes(fmt.Sprintf("{%s}", rp))},
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

	if len(info.RequestType.Params.Query) > 0 {
		fd.Body.List = append(fd.Body.List,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "q"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CompositeLit{Type: &ast.ArrayType{Elt: &ast.Ident{Name: "string"}}}},
			},
		)

		for qp, fn := range sorted.ByValues(info.RequestType.Params.Query) {
			fd.Body.List = append(fd.Body.List,
				&ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "encoded"}, &ast.Ident{Name: "ok"}, &ast.Ident{Name: "err"}},
					Tok: ternary(symbols.encoded && symbols.ok && symbols.err, token.ASSIGN, token.DEFINE),
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}},
								Sel: &ast.Ident{Name: "ToQuery"},
							},
						},
					},
				},
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
					Body: &ast.BlockStmt{List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.Ident{Name: "nil"},
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: quotes(fmt.Sprintf("%s.%s.ToQuery: %%w", info.RequestType.Typename, fn))},
										&ast.Ident{Name: "err"},
									},
								},
							},
						},
					}},
				},
				&ast.IfStmt{
					Cond: &ast.Ident{Name: "ok"},
					Body: &ast.BlockStmt{List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{&ast.Ident{Name: "q"}},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.Ident{Name: "append"},
									Args: []ast.Expr{
										&ast.Ident{Name: "q"},
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Sprintf"}},
											Args: []ast.Expr{
												&ast.BasicLit{Kind: token.STRING, Value: quotes(fmt.Sprintf("%s=%%s", qp))},
												&ast.Ident{Name: "encoded"},
											},
										},
									},
								},
							},
						},
					}},
				},
			)
			symbols.ok = true
			symbols.encoded = true
			symbols.err = true
		}

		fd.Body.List = append(fd.Body.List,
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  &ast.CallExpr{Fun: &ast.Ident{Name: "len"}, Args: []ast.Expr{&ast.Ident{Name: "q"}}},
					Op: token.GTR,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
				},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: "uri"}},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Sprintf"}},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: `"%s?%s"`},
									&ast.Ident{Name: "uri"},
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "strings"}, Sel: &ast.Ident{Name: "Join"}},
										Args: []ast.Expr{
											&ast.Ident{Name: "q"},
											&ast.BasicLit{Kind: token.STRING, Value: quotes("&")},
										},
									},
								},
							},
						},
					},
				}},
			},
		)
	}

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
				Tok: ternary(symbols.err, token.ASSIGN, token.DEFINE),
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
		symbols.err = true
	}

	fd.Body.List = append(fd.Body.List,
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "r"}, &ast.Ident{Name: "err"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "NewRequest"}},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: quotes(info.Method)},
					&ast.CallExpr{
						Fun:  &ast.Ident{Name: "join"},
						Args: []ast.Expr{&ast.Ident{Name: "host"}, &ast.Ident{Name: "uri"}},
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
		},
	)

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
						Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `".json"`}},
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

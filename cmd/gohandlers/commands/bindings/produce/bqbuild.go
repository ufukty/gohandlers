package produce

import (
	"fmt"
	"go/ast"
	"go/token"
	"gohandlers/internal/sorted"
	"gohandlers/pkg/inspects"
)

type bqBuildSymbolTable struct {
	err     bool
	encoded bool
	ok      bool
}

type bqBuild struct {
	table bqBuildSymbolTable
}

func (p *bqBuild) route(info inspects.Info) []ast.Stmt {
	stmts := []ast.Stmt{}
	for rp, fn := range sorted.ByValues(info.RequestType.Params.Route) {
		stmts = append(stmts,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "encoded"}, &ast.Ident{Name: "err"}},
				Tok: ternary(p.table.encoded && p.table.err, token.ASSIGN, token.DEFINE),
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
		p.table.encoded = true
		p.table.err = true
	}
	return stmts
}

func (p *bqBuild) query(info inspects.Info) []ast.Stmt {
	stmts := []ast.Stmt{}
	if len(info.RequestType.Params.Query) > 0 {
		stmts = append(stmts,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "q"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CompositeLit{Type: &ast.ArrayType{Elt: &ast.Ident{Name: "string"}}}},
			},
		)

		for qp, fn := range sorted.ByValues(info.RequestType.Params.Query) {
			stmts = append(stmts,
				&ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "encoded"}, &ast.Ident{Name: "ok"}, &ast.Ident{Name: "err"}},
					Tok: ternary(p.table.encoded && p.table.ok && p.table.err, token.ASSIGN, token.DEFINE),
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
			p.table.ok = true
			p.table.encoded = true
			p.table.err = true
		}

		stmts = append(stmts,
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

	return stmts
}

func (p *bqBuild) body(info inspects.Info) []ast.Stmt {
	stmts := []ast.Stmt{}
	if info.RequestType.ContentType != "" {
		stmts = append(stmts, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "body"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "bytes"}, Sel: &ast.Ident{Name: "NewBuffer"}},
				Args: []ast.Expr{&ast.CompositeLit{Type: &ast.ArrayType{Elt: &ast.Ident{Name: "byte"}}}},
			}},
		})
	}
	return stmts
}

func (p *bqBuild) json(info inspects.Info) []ast.Stmt {
	stmts := []ast.Stmt{}
	if len(info.RequestType.Params.Json) > 0 {
		stmts = append(stmts,
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
					Tok: ternary(p.table.err, token.ASSIGN, token.DEFINE),
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
		p.table.err = true
	}
	return stmts
}

func (p *bqBuild) multipartFormData(info inspects.Info) []ast.Stmt {
	stmts := []ast.Stmt{}
	if info.RequestType.ContentType == "multipart/form-data" {
		stmts = append(stmts,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "ct"}, &ast.Ident{Name: "err"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: "marshalMultipartFormData"}},
					Args: []ast.Expr{&ast.Ident{Name: "body"}},
				}},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
						Args: []ast.Expr{
							&ast.BasicLit{Kind: token.STRING, Value: `"marshal multipart/form-data body: %w"`},
							&ast.Ident{Name: "err"},
						},
					},
				}}}},
			},
		)
	}
	return stmts

}

func (p *bqBuild) request(info inspects.Info) []ast.Stmt {
	stmts := []ast.Stmt{}
	stmts = append(stmts,
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
	return stmts
}

func (p *bqBuild) postRequest(info inspects.Info) []ast.Stmt {
	stmts := []ast.Stmt{}
	if info.RequestType.ContainsBody {
		stmts = append(stmts,
			&ast.ExprStmt{X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "r"}, Sel: &ast.Ident{Name: "Header"}},
					Sel: &ast.Ident{Name: "Set"},
				},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"Content-Type"`},
					ternary[ast.Expr](
						info.RequestType.ContentType == "multipart/form-data",
						&ast.Ident{Name: "ct"},
						&ast.BasicLit{Kind: token.STRING, Value: quotes(info.RequestType.ContentType)},
					),
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
	return stmts
}

// produces the bqtn.Build method
func (p *bqBuild) Produce(info inspects.Info) *ast.FuncDecl {
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

	fd.Body.List = append(fd.Body.List,
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "uri"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes(info.Path)}},
		},
	)

	fd.Body.List = append(fd.Body.List, p.route(info)...)
	fd.Body.List = append(fd.Body.List, p.query(info)...)
	fd.Body.List = append(fd.Body.List, p.body(info)...)
	fd.Body.List = append(fd.Body.List, p.json(info)...)
	fd.Body.List = append(fd.Body.List, p.multipartFormData(info)...)
	fd.Body.List = append(fd.Body.List, p.request(info)...)
	fd.Body.List = append(fd.Body.List, p.postRequest(info)...)

	fd.Body.List = append(fd.Body.List,
		&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "r"}, &ast.Ident{Name: "nil"}}},
	)

	return fd
}

func BqBuild(i inspects.Info) *ast.FuncDecl {
	p := &bqBuild{}
	return p.Produce(i)
}

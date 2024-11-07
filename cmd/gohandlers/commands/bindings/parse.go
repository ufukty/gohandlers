package bindings

import (
	"fmt"
	"go/ast"
	"go/token"
	"gohandlers/pkg/inspects"
)

func bqParse(info inspects.Info) *ast.FuncDecl {
	fd := &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{{Name: "bq"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: info.RequestType.Typename}}},
		}},
		Name: &ast.Ident{Name: "Parse"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "rq"}},
				Type:  &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Request"}}},
			}}},
			Results: &ast.FieldList{List: []*ast.Field{
				{Type: &ast.Ident{Name: "error"}},
			}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}

	type symboltable struct {
		err bool
	}
	symbols := symboltable{}

	for rp, fn := range info.RequestType.RouteParams {
		fd.Body.List = append(fd.Body.List,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
				Tok: ternary(symbols.err, token.ASSIGN, token.DEFINE),
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}},
							Sel: &ast.Ident{Name: "FromRoute"},
						},
						Args: []ast.Expr{&ast.CallExpr{
							Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "rq"}, Sel: &ast.Ident{Name: "PathValue"}},
							Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes(rp)}},
						}},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
						Args: []ast.Expr{
							&ast.BasicLit{Kind: token.STRING, Value: quotes(fmt.Sprintf("%s.%s.FromRoute: %%w", info.RequestType.Typename, fn))},
							&ast.Ident{Name: "err"}},
					}}},
				}},
			},
		)
		symbols.err = true
	}

	for qp, fn := range info.RequestType.QueryParams {
		fd.Body.List = append(fd.Body.List,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "q"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "rq"}, Sel: &ast.Ident{Name: "URL"}},
					Sel: &ast.Ident{Name: "Query"},
				}}},
			},
			&ast.IfStmt{
				Cond: &ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "q"}, Sel: &ast.Ident{Name: "Has"}},
					Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes(qp)}},
				},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}}, Sel: &ast.Ident{Name: "FromQuery"},
								},
								Args: []ast.Expr{&ast.CallExpr{
									Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "q"}, Sel: &ast.Ident{Name: "Get"}},
									Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes(qp)}},
								}},
							},
						},
					},
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
						Body: &ast.BlockStmt{List: []ast.Stmt{
							&ast.ReturnStmt{Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: quotes(fmt.Sprintf("%s.%s.FromQuery: %%w", info.RequestType.Typename, fn))},
										&ast.Ident{Name: "err"},
									},
								},
							}},
						}},
					},
				}},
			},
		)
	}

	if info.RequestType.ContainsBody {
		fd.Body.List = append(fd.Body.List,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
				Tok: ternary(symbols.err, token.ASSIGN, token.DEFINE),
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.CallExpr{
								Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "json"}, Sel: &ast.Ident{Name: "NewDecoder"}},
								Args: []ast.Expr{&ast.SelectorExpr{X: &ast.Ident{Name: "rq"}, Sel: &ast.Ident{Name: "Body"}}},
							},
							Sel: &ast.Ident{Name: "Decode"},
						},
						Args: []ast.Expr{&ast.Ident{Name: "bq"}},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ReturnStmt{Results: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: `"decoding body: %w"`},
								&ast.Ident{Name: "err"},
							},
						},
					}},
				}},
			},
		)
		symbols.err = true
	}

	fd.Body.List = append(fd.Body.List,
		&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "nil"}}},
	)

	return fd
}

package produce

import (
	"go/ast"
	"go/token"
	"gohandlers/pkg/inspects"
)

type bsParse struct{}

func (p *bsParse) contentTypeCheck(info inspects.Info) []ast.Stmt {
	return []ast.Stmt{
		&ast.IfStmt{
			Cond: &ast.UnaryExpr{
				Op: token.NOT,
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "strings"}, Sel: &ast.Ident{Name: "HasPrefix"}},
					Args: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "rs"},
									Sel: &ast.Ident{Name: "Header"},
								},
								Sel: &ast.Ident{Name: "Get"},
							},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: `"Content-Type"`},
							},
						},
						&ast.BasicLit{Kind: token.STRING, Value: quotes(info.ResponseType.ContentType)},
					},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "fmt"},
									Sel: &ast.Ident{Name: "Errorf"},
								},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: `"invalid content type for request: %s"`},
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "rs"},
												Sel: &ast.Ident{Name: "Header"},
											},
											Sel: &ast.Ident{Name: "Get"},
										},
										Args: []ast.Expr{
											&ast.BasicLit{Kind: token.STRING, Value: `"Content-Type"`},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (p *bsParse) json(info inspects.Info) []ast.Stmt {
	stmts := []ast.Stmt{}
	if len(info.ResponseType.Params.Json) > 0 {
		stmts = append(stmts,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.CallExpr{
								Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "json"}, Sel: &ast.Ident{Name: "NewDecoder"}},
								Args: []ast.Expr{&ast.SelectorExpr{X: &ast.Ident{Name: "rs"}, Sel: &ast.Ident{Name: "Body"}}},
							},
							Sel: &ast.Ident{Name: "Decode"},
						},
						Args: []ast.Expr{&ast.Ident{Name: "bs"}},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ReturnStmt{Results: []ast.Expr{
						&ast.CallExpr{
							Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
							Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"decoding the body: %w"`}, &ast.Ident{Name: "err"}},
						},
					}},
				}},
			},
		)
	}
	return stmts
}

func (p *bsParse) Produce(info inspects.Info) *ast.FuncDecl {
	fd := &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{{Name: "bs"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: info.ResponseType.Typename}}},
		}},
		Name: &ast.Ident{Name: "Parse"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "rs"}},
				Type:  &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Response"}}},
			}}},
			Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "error"}}}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}

	fd.Body.List = append(fd.Body.List, p.contentTypeCheck(info)...)
	fd.Body.List = append(fd.Body.List, p.json(info)...)

	fd.Body.List = append(fd.Body.List,
		&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "nil"}}},
	)

	return fd
}

func BsParse(i inspects.Info) *ast.FuncDecl {
	p := &bsParse{}
	return p.Produce(i)
}

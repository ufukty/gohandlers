package bindings

import (
	"go/ast"
	"go/token"
	"gohandlers/pkg/inspects"
)

func needsjoin(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, handlers := range infoss {
		for _, info := range handlers {
			if info.RequestType != nil {
				return true
			}
		}
	}
	return false
}

func join() *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: &ast.Ident{Name: "join"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "segments"}},
				Type:  &ast.Ellipsis{Elt: &ast.Ident{Name: "string"}},
			}}},
			Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "string"}}}},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "url"}},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes("")}},
				},
				&ast.RangeStmt{
					Key:   &ast.Ident{Name: "i"},
					Value: &ast.Ident{Name: "segment"},
					Tok:   token.DEFINE,
					X:     &ast.Ident{Name: "segments"},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.IfStmt{
								Cond: &ast.BinaryExpr{
									X:  &ast.BinaryExpr{X: &ast.Ident{Name: "i"}, Op: token.NEQ, Y: &ast.BasicLit{Kind: token.INT, Value: "0"}},
									Op: token.LAND,
									Y: &ast.UnaryExpr{
										Op: token.NOT,
										X: &ast.CallExpr{
											Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "strings"}, Sel: &ast.Ident{Name: "HasPrefix"}},
											Args: []ast.Expr{&ast.Ident{Name: "segment"}, &ast.BasicLit{Kind: token.STRING, Value: quotes("")}},
										},
									},
								},
								Body: &ast.BlockStmt{List: []ast.Stmt{
									&ast.AssignStmt{
										Lhs: []ast.Expr{&ast.Ident{Name: "url"}},
										Tok: token.ADD_ASSIGN,
										Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes("")}},
									},
								}},
							},
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "url"}},
								Tok: token.ADD_ASSIGN,
								Rhs: []ast.Expr{&ast.Ident{Name: "segment"}},
							},
						},
					},
				},
				&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "url"}}},
			},
		},
	}
}

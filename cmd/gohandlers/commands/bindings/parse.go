package bindings

import (
	"go/ast"
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

	return fd
}

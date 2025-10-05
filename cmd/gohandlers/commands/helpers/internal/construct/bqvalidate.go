package construct

import (
	"go/ast"
	"go/token"

	"go.ufukty.com/gohandlers/internal/sorted"
	"go.ufukty.com/gohandlers/pkg/inspects"
)

func merge[K comparable, V any](maps ...map[K]V) map[K]V {
	t := 0
	for _, m := range maps {
		t += len(m)
	}
	m := make(map[K]V, t)
	for _, mp := range maps {
		for k, v := range mp {
			m[k] = v
		}
	}
	return m
}

func BqValidate(bti *inspects.BindingTypeInfo) *ast.FuncDecl {
	var fd = &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "bq"}}, Type: &ast.Ident{Name: bti.Typename}}}},
		Name: &ast.Ident{Name: "Validate"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{List: []*ast.Field{
				{Names: []*ast.Ident{{Name: "issues"}}, Type: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "any"}}},
			}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("issues")},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{&ast.CompositeLit{Type: &ast.MapType{Key: ast.NewIdent("string"), Value: ast.NewIdent("any")}}},
		}}},
	}
	params := merge(
		bti.Params.Form,
		bti.Params.Json,
		bti.Params.Query,
		bti.Params.Route,
	)
	for p, fn := range sorted.ByValues(params) {
		fd.Body.List = append(fd.Body.List, &ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "issue"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}},
					Sel: &ast.Ident{Name: "Validate"},
				}}},
			},
			Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "issue"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.IndexExpr{X: &ast.Ident{Name: "issues"}, Index: &ast.BasicLit{Kind: token.STRING, Value: quotes(p)}}},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{&ast.Ident{Name: "issue"}},
			}}},
		})
	}
	fd.Body.List = append(fd.Body.List, &ast.ReturnStmt{})
	return fd
}

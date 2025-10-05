package construct

import (
	"fmt"
	"go/ast"
	"go/token"

	"go.ufukty.com/gohandlers/internal/sorted"
	"go.ufukty.com/gohandlers/pkg/inspects"
)

func BqUnmarshalFormData(i inspects.Info) *ast.FuncDecl {
	fd := &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{{Name: "bq"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: i.RequestType.Typename}}},
		}},
		Name: &ast.Ident{Name: "unmarshalFormData"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "rq"}},
				Type:  &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Request"}}},
			}}},
			Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "error"}}}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "rq"}, Sel: &ast.Ident{Name: "ParseForm"}}}},
				},
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
					Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"ParseForm: %w"`}, &ast.Ident{Name: "err"}},
				}}}}},
			},
		}},
	}

	for p, fn := range sorted.ByValues(i.RequestType.Params.Form) {
		fd.Body.List = append(fd.Body.List,
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}},
								Sel: &ast.Ident{Name: "FromForm"},
							},
							Args: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.Ident{Name: "firstOrZero"},
									Args: []ast.Expr{
										&ast.IndexExpr{
											X:     &ast.SelectorExpr{X: &ast.Ident{Name: "rq"}, Sel: &ast.Ident{Name: "PostForm"}},
											Index: &ast.BasicLit{Kind: token.STRING, Value: quotes(p)},
										},
									},
								},
							},
						},
					},
				},
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
					Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s: FromForm: %%w"`, fn)},
						&ast.Ident{Name: "err"},
					},
				}}}}},
			},
		)
	}

	fd.Body.List = append(fd.Body.List,
		&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "nil"}}},
	)

	return fd
}

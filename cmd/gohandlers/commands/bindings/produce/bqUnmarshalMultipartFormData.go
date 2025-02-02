package produce

import (
	"fmt"
	"go/ast"
	"go/token"
	"gohandlers/internal/sorted"
	"gohandlers/pkg/inspects"
)

func ResponseUnmarshalMultipartFormData(i inspects.Info) *ast.FuncDecl {
	fd := &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{{
			Names: []*ast.Ident{{Name: "bq"}},
			Type:  &ast.StarExpr{X: &ast.Ident{Name: i.RequestType.Typename}},
		}}},
		Name: &ast.Ident{Name: "unmarshalMultipartFormData"},
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

	fd.Body.List = append(fd.Body.List,
		&ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "rq"}, Sel: &ast.Ident{Name: "ParseMultipartForm"}},
					Args: []ast.Expr{&ast.BinaryExpr{
						X: &ast.BinaryExpr{
							X:  &ast.BasicLit{Kind: token.INT, Value: "10"},
							Op: token.MUL,
							Y:  &ast.BasicLit{Kind: token.INT, Value: "1024"},
						},
						Op: token.MUL,
						Y:  &ast.BasicLit{Kind: token.INT, Value: "1024"},
					}},
				}},
			},
			Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
				Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"ParseMultipartForm: %w"`},
					&ast.Ident{Name: "err"},
				},
			}}}}},
		},
	)

	for p, fn := range sorted.ByValues(i.RequestType.Params.Part) {
		fd.Body.List = append(fd.Body.List,
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: &ast.Ident{Name: "fromPart"},
						Args: []ast.Expr{
							&ast.UnaryExpr{Op: token.AND, X: &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}}},
							&ast.Ident{Name: "rq"},
							&ast.BasicLit{Kind: token.STRING, Value: quotes(p)},
						},
					}},
				},
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
									Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s: %%w"`, fn)}, &ast.Ident{Name: "err"}},
								},
							},
						},
					},
				},
			},
		)
	}

	for p, fn := range sorted.ByValues(i.RequestType.Params.File) {
		fd.Body.List = append(fd.Body.List,
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.Ident{Name: "fromFileHeader"},
							Args: []ast.Expr{
								&ast.UnaryExpr{
									Op: token.AND,
									X:  &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}},
								},
								&ast.Ident{Name: "rq"},
								&ast.BasicLit{Kind: token.STRING, Value: quotes(p)},
							},
						},
					},
				},
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
					Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s: %%w"`, fn)}, &ast.Ident{Name: "err"}},
				}}}}},
			},
		)
	}

	fd.Body.List = append(fd.Body.List, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.Ident{Name: "nil"},
		},
	})

	return fd
}

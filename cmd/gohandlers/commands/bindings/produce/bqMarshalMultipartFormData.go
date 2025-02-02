package produce

import (
	"fmt"
	"go/ast"
	"go/token"
	"gohandlers/internal/sorted"
	"gohandlers/pkg/inspects"
)

func ResponseMarshalMultipartFormData(i inspects.Info) *ast.FuncDecl {
	fd := &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{{Name: "bq"}}, Type: &ast.Ident{Name: i.RequestType.Typename}},
		}},
		Name: &ast.Ident{Name: "marshalMultipartFormData"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "body"}},
				Type:  &ast.SelectorExpr{X: &ast.Ident{Name: "io"}, Sel: &ast.Ident{Name: "Writer"}},
			}}},
			Results: &ast.FieldList{
				List: []*ast.Field{{Type: &ast.Ident{Name: "string"}}, {Type: &ast.Ident{Name: "error"}}},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "mp"}},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "multipart"}, Sel: &ast.Ident{Name: "NewWriter"}},
						Args: []ast.Expr{&ast.Ident{Name: "body"}},
					}},
				},
				&ast.DeferStmt{Call: &ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "mp"}, Sel: &ast.Ident{Name: "Close"}}}},
			},
		},
	}

	for p, fn := range sorted.ByValues(i.RequestType.Params.Part) {
		fd.Body.List = append(fd.Body.List,
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.Ident{Name: "toPart"},
							Args: []ast.Expr{
								&ast.Ident{Name: "mp"},
								&ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}},
								&ast.BasicLit{Kind: token.STRING, Value: quotes(p)},
							},
						},
					},
				},
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: `""`},
						&ast.CallExpr{
							Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
							Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s: %%w"`, fn)}, &ast.Ident{Name: "err"}},
						},
					},
				}}},
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
							Fun: &ast.Ident{Name: "toFile"},
							Args: []ast.Expr{
								&ast.Ident{Name: "mp"},
								&ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}},
								&ast.BasicLit{Kind: token.STRING, Value: quotes(p)},
							},
						},
					},
				},
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `""`},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
						Args: []ast.Expr{
							&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s: %%w"`, fn)},
							&ast.Ident{Name: "err"},
						},
					},
				}}}},
			},
		)
	}

	fd.Body.List = append(fd.Body.List, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "mp"}, Sel: &ast.Ident{Name: "FormDataContentType"}},
			},
			&ast.Ident{Name: "nil"},
		},
	})
	return fd
}

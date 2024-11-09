package bindings

import (
	"go/ast"
	"go/token"
	"gohandlers/pkg/inspects"
)

func bsWrite(info inspects.Info) *ast.FuncDecl {
	fd := &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{{Name: "bs"}}, Type: &ast.Ident{Name: info.ResponseType.Typename}},
		}},
		Name: &ast.Ident{Name: "Write"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "w"}}, Type: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "ResponseWriter"}},
			}}},
			Results: &ast.FieldList{List: []*ast.Field{
				{Type: &ast.Ident{Name: "error"}},
			}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}

	if info.ResponseType.ContainsBody {
		fd.Body.List = append(fd.Body.List,
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "w"}, Sel: &ast.Ident{Name: "Header"}}},
						Sel: &ast.Ident{Name: "Set"},
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: `"Content-Type"`},
						&ast.CallExpr{
							Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "mime"}, Sel: &ast.Ident{Name: "TypeByExtension"}},
							Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `".json"`}},
						},
					},
				},
			},
		)
	}

	fd.Body.List = append(fd.Body.List,
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "w"}, Sel: &ast.Ident{Name: "WriteHeader"}},
				Args: []ast.Expr{&ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "StatusOK"}}},
			},
		},
	)

	if info.ResponseType.ContainsBody {
		fd.Body.List = append(fd.Body.List,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.CallExpr{
								Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "json"}, Sel: &ast.Ident{Name: "NewEncoder"}},
								Args: []ast.Expr{&ast.Ident{Name: "w"}},
							},
							Sel: &ast.Ident{Name: "Encode"},
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
							Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"encoding the body: %w"`}, &ast.Ident{Name: "err"}},
						},
					}},
				}},
			},
			&ast.ReturnStmt{
				Results: []ast.Expr{&ast.Ident{Name: "nil"}},
			},
		)
	}

	return fd
}

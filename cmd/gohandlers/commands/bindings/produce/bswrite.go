package produce

import (
	"go/ast"
	"go/token"
	"gohandlers/pkg/inspects"
)

type bsWrite struct{}

func (p *bsWrite) headers(info inspects.Info) []ast.Stmt {
	stmts := []ast.Stmt{}
	if info.ResponseType.ContentType != "" {
		stmts = append(stmts,
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "w"}, Sel: &ast.Ident{Name: "Header"}}},
						Sel: &ast.Ident{Name: "Set"},
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: `"Content-Type"`},
						&ast.BasicLit{Kind: token.STRING, Value: quotes(info.ResponseType.ContentType)},
					},
				},
			},
		)
	}
	stmts = append(stmts,
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "w"}, Sel: &ast.Ident{Name: "WriteHeader"}},
				Args: []ast.Expr{&ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "StatusOK"}}},
			},
		},
	)
	return stmts
}

func (p *bsWrite) json(info inspects.Info) []ast.Stmt {
	stmts := []ast.Stmt{}
	if len(info.RequestType.Params.Json) > 0 {
		stmts = append(stmts,
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
		)
	}
	return stmts
}

func (p *bsWrite) Produce(info inspects.Info) *ast.FuncDecl {
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

	fd.Body.List = append(fd.Body.List, p.headers(info)...)
	fd.Body.List = append(fd.Body.List, p.json(info)...)

	fd.Body.List = append(fd.Body.List, &ast.ReturnStmt{
		Results: []ast.Expr{&ast.Ident{Name: "nil"}},
	})

	return fd
}

func BsWrite(i inspects.Info) *ast.FuncDecl {
	p := &bsWrite{}
	return p.Produce(i)
}

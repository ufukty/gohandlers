package utilities

import (
	"fmt"
	"go/ast"
	"go/token"

	"go.ufukty.com/gohandlers/pkg/inspects"
)

func quotes(s string) string {
	return fmt.Sprintf("%q", s)
}

func needsJoin(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, handlers := range infoss {
		for _, info := range handlers {
			if info.RequestType != nil {
				return true
			}
		}
	}
	return false
}

var join = &ast.FuncDecl{
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
										Args: []ast.Expr{&ast.Ident{Name: "segment"}, &ast.BasicLit{Kind: token.STRING, Value: quotes("/")}},
									},
								},
							},
							Body: &ast.BlockStmt{List: []ast.Stmt{
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "url"}},
									Tok: token.ADD_ASSIGN,
									Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: quotes("/")}},
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

var firstOrZero = &ast.FuncDecl{
	Name: &ast.Ident{Name: "firstOrZero"},
	Type: &ast.FuncType{
		TypeParams: &ast.FieldList{
			List: []*ast.Field{{Names: []*ast.Ident{{Name: "E"}}, Type: &ast.Ident{Name: "any"}}},
		},
		Params: &ast.FieldList{
			List: []*ast.Field{{Names: []*ast.Ident{{Name: "s"}}, Type: &ast.ArrayType{Elt: &ast.Ident{Name: "E"}}}},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{{Type: &ast.Ident{Name: "E"}}},
		},
	},
	Body: &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok:   token.VAR,
					Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{{Name: "e"}}, Type: &ast.Ident{Name: "E"}}},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  &ast.CallExpr{Fun: &ast.Ident{Name: "len"}, Args: []ast.Expr{&ast.Ident{Name: "s"}}},
					Op: token.GTR,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{&ast.Ident{Name: "e"}},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{&ast.IndexExpr{X: &ast.Ident{Name: "s"}, Index: &ast.BasicLit{Kind: token.INT, Value: "0"}}},
						},
					},
				},
			},
			&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "e"}}},
		},
	},
}

func needsFirstOrZero(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, infos := range infoss {
		for _, info := range infos {
			if info.RequestType != nil && len(info.RequestType.Params.Form) > 0 {
				return true
			}
			if info.ResponseType != nil && len(info.ResponseType.Params.Form) > 0 {
				return true
			}
		}
	}
	return false
}

func Produce(infoss map[inspects.Receiver]map[string]inspects.Info) []ast.Decl {
	decls := []ast.Decl{}
	if needsJoin(infoss) {
		decls = append(decls, join)
	}
	if needsFirstOrZero(infoss) {
		decls = append(decls, firstOrZero)
	}
	return decls
}

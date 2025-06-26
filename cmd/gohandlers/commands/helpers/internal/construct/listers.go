package construct

import (
	"go/ast"
	"go/token"
	"slices"

	"github.com/ufukty/gohandlers/pkg/inspects"
)

func Listers(infoss map[inspects.Receiver]map[string]inspects.Info) []ast.Decl {
	var handlerinfo ast.Expr = &ast.SelectorExpr{
		X:   ast.NewIdent("gohandlers"),
		Sel: ast.NewIdent("HandlerInfo"),
	}

	fds := []ast.Decl{}
	for recvt, infos := range infoss {
		elts := []ast.Expr{}
		for hn, info := range infos {
			kv := &ast.KeyValueExpr{
				Key: &ast.BasicLit{Kind: token.STRING, Value: quotes(hn)},
				Value: &ast.CompositeLit{Elts: []ast.Expr{
					&ast.KeyValueExpr{Key: &ast.Ident{Name: "Method"}, Value: &ast.BasicLit{Kind: token.STRING, Value: quotes(info.Method)}},
					&ast.KeyValueExpr{Key: &ast.Ident{Name: "Path"}, Value: &ast.BasicLit{Kind: token.STRING, Value: quotes(info.Path)}},
					&ast.KeyValueExpr{Key: &ast.Ident{Name: "Ref"}, Value: info.Ref},
				}},
			}
			elts = append(elts, kv)
		}

		slices.SortFunc(elts, func(a, b ast.Expr) int {
			ka := a.(*ast.KeyValueExpr).Key.(*ast.BasicLit).Value
			kb := b.(*ast.KeyValueExpr).Key.(*ast.BasicLit).Value
			if ka < kb {
				return -1
			} else if ka == kb {
				return 0
			} else {
				return 1
			}
		})

		fd := &ast.FuncDecl{
			Name: &ast.Ident{Name: "ListHandlers"},
			Type: &ast.FuncType{
				Params: &ast.FieldList{List: []*ast.Field{}},
				Results: &ast.FieldList{List: []*ast.Field{
					{Type: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: handlerinfo}},
				}},
			},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{&ast.CompositeLit{
					Type: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: handlerinfo},
					Elts: elts,
				}}},
			}},
		}

		if recvt.Type != "" {
			fd.Recv = &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{{Name: recvt.Name}},
				Type:  &ast.StarExpr{X: &ast.Ident{Name: recvt.Type}},
			}}}
		}

		fds = append(fds, fd)
	}

	slices.SortFunc(fds, func(a, b ast.Decl) int {
		if a.(*ast.FuncDecl).Recv == nil {
			return -1
		}
		if b.(*ast.FuncDecl).Recv == nil {
			return 1
		}

		at := a.(*ast.FuncDecl).Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
		bt := b.(*ast.FuncDecl).Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name

		if at < bt {
			return -1
		} else if at == bt {
			return 0
		} else {
			return 1
		}
	})

	return fds
}

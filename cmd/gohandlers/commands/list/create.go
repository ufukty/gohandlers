package list

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"io"
	"os"
	"strings"

	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/version"
	"github.com/ufukty/gohandlers/pkg/inspects"

	"golang.org/x/exp/slices"
)

func quotes(src string) string {
	return fmt.Sprintf("%q", src)
}

func addnewlines(f string) string {
	f = strings.ReplaceAll(f, "}\nfunc", "}\n\nfunc")
	hit := "HandlerInfo"
	f = strings.ReplaceAll(f, fmt.Sprintf("%s{", hit), fmt.Sprintf("%s{\n", hit)) // beginning composite literal
	f = strings.ReplaceAll(f, "}, \"", "},\n\"")                                  // after each line
	f = strings.ReplaceAll(f, "}}", "},\n}")                                      // ending composite literal
	return f
}

func create(dst string, infoss map[inspects.Receiver]map[string]inspects.Info, pkgname string, hi HandlerInfo) error {
	f := &ast.File{
		Name:  ast.NewIdent(pkgname),
		Decls: []ast.Decl{},
	}

	imports := []ast.Spec{}
	if hi.ImportPath != "" {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: quotes(hi.ImportPath)}},
		)
	} else {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: quotes("net/http")}},
		)
	}

	f.Decls = append(f.Decls,
		&ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: imports,
		},
	)

	if hi.Typename == "" {
		f.Decls = append(f.Decls,
			&ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: &ast.Ident{Name: "HandlerInfo"},
						Type: &ast.StructType{Fields: &ast.FieldList{
							List: []*ast.Field{
								{Names: []*ast.Ident{{Name: "Method"}}, Type: &ast.Ident{Name: "string"}},
								{Names: []*ast.Ident{{Name: "Path"}}, Type: &ast.Ident{Name: "string"}},
								{Names: []*ast.Ident{{Name: "Ref"}}, Type: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "HandlerFunc"}}},
							},
						}},
					},
				},
			},
		)
	}

	var handlerinfo ast.Expr
	if hi.Typename == "" {
		handlerinfo = &ast.Ident{Name: "HandlerInfo"}
	} else {
		fgrs := strings.Split(hi.Typename, ".")
		switch len(fgrs) {
		case 2:
			handlerinfo = &ast.SelectorExpr{X: ast.NewIdent(fgrs[0]), Sel: ast.NewIdent(fgrs[1])}
		case 1:
			handlerinfo = ast.NewIdent(hi.Typename)
		default:
			return fmt.Errorf("unexpected number of dots in the value for HandlerInfo substitution: %s", hi.Typename)
		}
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
	f.Decls = append(f.Decls, fds...)

	b := bytes.NewBufferString("")
	fmt.Fprint(b, version.Top())
	err := printer.Fprint(b, token.NewFileSet(), f)
	if err != nil {
		return fmt.Errorf("printing: %w", err)
	}
	o, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer o.Close()

	bt, err := format.Source([]byte(addnewlines(b.String())))
	if err != nil {
		return fmt.Errorf("formatting output file: %w", err)
	}
	io.Copy(o, bytes.NewBuffer(bt))

	return nil
}

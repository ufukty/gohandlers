package build

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"gohandlers/pkg/inspects"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func ternary[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

// produces the bqtn.Build method
func build(info inspects.Info) *ast.FuncDecl {
	fd := &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{{Name: "bq"}}, Type: &ast.Ident{Name: info.RequestType.Typename}},
		}},
		Name: &ast.Ident{Name: "Build"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "lb"}},
					Type:  &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "balancer"}, Sel: &ast.Ident{Name: "LoadBalancer"}}},
				},
			}},
			Results: &ast.FieldList{List: []*ast.Field{
				{Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "Request"}}}},
				{Type: &ast.Ident{Name: "error"}},
			}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}

	fd.Body.List = append(fd.Body.List,
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "uri"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.Path)}},
		},
	)
	replacements := []ast.Stmt{}
	for routeparam, fieldname := range info.RequestType.RouteParams {
		replacements = append(replacements,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "uri"}},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "strings"}, Sel: &ast.Ident{Name: "Replace"}},
						Args: []ast.Expr{
							&ast.Ident{Name: "uri"},
							&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", fmt.Sprintf("{%s}", routeparam))},
							&ast.CallExpr{
								Fun:  &ast.Ident{Name: "string"},
								Args: []ast.Expr{&ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fieldname}}},
							},
							&ast.BasicLit{Kind: token.INT, Value: "1"},
						},
					},
				},
			},
		)
	}
	slices.SortFunc(replacements, func(a, b ast.Stmt) int {
		va := a.(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Args[1].(*ast.BasicLit).Value
		vb := b.(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Args[1].(*ast.BasicLit).Value
		if va < vb {
			return -1
		} else if va == vb {
			return 0
		} else {
			return 1
		}
	})
	fd.Body.List = append(fd.Body.List, replacements...)

	if info.RequestType.ContainsBody {
		fd.Body.List = append(fd.Body.List,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "body"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "bytes"}, Sel: &ast.Ident{Name: "NewBuffer"}},
					Args: []ast.Expr{&ast.CompositeLit{Type: &ast.ArrayType{Elt: &ast.Ident{Name: "byte"}}}},
				}},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.CallExpr{
							Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "json"}, Sel: &ast.Ident{Name: "NewEncoder"}},
							Args: []ast.Expr{&ast.Ident{Name: "body"}},
						},
						Sel: &ast.Ident{Name: "Encode"},
					},
					Args: []ast.Expr{&ast.Ident{Name: "bq"}},
				}},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
						Args: []ast.Expr{
							&ast.BasicLit{Kind: token.STRING, Value: `"json.Encoder.Encode: %w"`},
							&ast.Ident{Name: "err"},
						},
					},
				}}}},
			},
		)
	}

	fd.Body.List = append(fd.Body.List,
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "h"}, &ast.Ident{Name: "err"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "lb"}, Sel: &ast.Ident{Name: "Next"}}, Args: nil}},
		},
		&ast.IfStmt{
			Init: nil,
			Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{
				&ast.Ident{Name: "nil"},
				&ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
					Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"lb.Next: %w"`}, &ast.Ident{Name: "err"}},
				},
			}}}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "r"}, &ast.Ident{Name: "err"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "NewRequest"}},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.Method)},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "urls"}, Sel: &ast.Ident{Name: "Join"}},
						Args: []ast.Expr{
							&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "h"}, Sel: &ast.Ident{Name: "String"}}, Args: nil},
							&ast.Ident{Name: "uri"},
						},
					},
					&ast.Ident{Name: ternary(info.RequestType.ContainsBody, "body", "nil")},
				},
			}},
		},
		&ast.IfStmt{
			Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
					&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
						Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"http.NewRequest: %w"`}, &ast.Ident{Name: "err"}},
					},
				}},
			}},
		})

	if info.RequestType.ContainsBody {
		fd.Body.List = append(fd.Body.List,
			&ast.ExprStmt{X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "r"}, Sel: &ast.Ident{Name: "Header"}},
					Sel: &ast.Ident{Name: "Set"},
				},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"Content-Type"`},
					&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "mime"}, Sel: &ast.Ident{Name: "TypeByExtension"}},
						Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"json"`}},
					},
				},
			}},
			&ast.ExprStmt{X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "r"}, Sel: &ast.Ident{Name: "Header"}},
					Sel: &ast.Ident{Name: "Set"},
				},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"Content-Length"`},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Sprintf"}},
						Args: []ast.Expr{
							&ast.BasicLit{Kind: token.STRING, Value: `"%d"`},
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "body"}, Sel: &ast.Ident{Name: "Len"}},
							},
						},
					},
				},
			}},
		)
	}

	fd.Body.List = append(fd.Body.List,
		&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "r"}, &ast.Ident{Name: "nil"}}},
	)

	return fd
}

func pathvarimports(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, infos := range infoss {
		for _, info := range infos {
			if info.RequestType != nil && len(info.RequestType.RouteParams) > 0 {
				return true
			}
		}
	}
	return false
}

func bodyimports(infoss map[inspects.Receiver]map[string]inspects.Info) bool {
	for _, infos := range infoss {
		for _, info := range infos {
			if info.RequestType != nil && info.RequestType.ContainsBody {
				return true
			}
		}
	}
	return false
}

func imports(infoss map[inspects.Receiver]map[string]inspects.Info) []ast.Spec {
	imports := []ast.Spec{
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`}},
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"logbook/internal/utils/urls"`}},
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"logbook/internal/web/balancer"`}},
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
	}
	if bodyimports(infoss) {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"bytes"`}},
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"encoding/json"`}},
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"mime"`}},
		)
	}
	if pathvarimports(infoss) {
		imports = append(imports,
			&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"strings"`}},
		)
	}
	slices.SortFunc(imports, func(a, b ast.Spec) int {
		av := a.(*ast.ImportSpec).Path.Value
		bv := b.(*ast.ImportSpec).Path.Value
		if av < bv {
			return -1
		} else if av == bv {
			return 0
		} else {
			return 1
		}
	})
	return imports
}

func buildfuncs(infoss map[inspects.Receiver]map[string]inspects.Info) []ast.Decl {
	fds := []ast.Decl{}
	for _, infos := range infoss {
		for _, info := range infos {
			if info.RequestType != nil {
				fds = append(fds, build(info))
			}
		}
	}
	slices.SortFunc(fds, func(a, b ast.Decl) int {
		na := a.(*ast.FuncDecl).Recv.List[0].Type.(*ast.Ident).Name
		nb := b.(*ast.FuncDecl).Recv.List[0].Type.(*ast.Ident).Name
		if na < nb {
			return -1
		} else if na == nb {
			return 0
		} else {
			return 1
		}
	})
	return fds
}

type Args struct {
	Dir  string
	Out  string
	Recv string
}

func post(src string) string {
	src = strings.ReplaceAll(src, "}\nfunc", "}\n\nfunc")
	return src
}

func Main() error {
	args := &Args{}
	flag.StringVar(&args.Dir, "dir", "", "the directory contains Go files")
	flag.StringVar(&args.Out, "out", "", "the output file")
	flag.StringVar(&args.Recv, "recv", "", "only use request types that is prefixed with handlers defined on this type")
	flag.Parse()

	if args.Dir == "" {
		flag.PrintDefaults()
		return fmt.Errorf("bad arguments")
	}

	infoss, pkg, err := inspects.Dir(args.Dir)
	if err != nil {
		return fmt.Errorf("inspecting the directory: %w", err)
	}

	if args.Recv != "" {
		for recv := range infoss {
			if args.Recv == recv.Type {
				infoss = map[inspects.Receiver]map[string]inspects.Info{
					recv: infoss[recv],
				}
				break
			}
		}
	}

	f := &ast.File{
		Name: ast.NewIdent(pkg),
		Decls: []ast.Decl{
			&ast.GenDecl{Tok: token.IMPORT, Specs: imports(infoss)},
		},
	}

	f.Decls = append(f.Decls, buildfuncs(infoss)...)

	if args.Out == "" {
		args.Out = "build.gh.go"
	}
	fh, err := os.Create(filepath.Join(args.Dir, args.Out))
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer fh.Close()
	b := bytes.NewBuffer([]byte{})
	err = printer.Fprint(b, token.NewFileSet(), f)
	if err != nil {
		return fmt.Errorf("printing: %w", err)
	}
	fmt.Fprint(fh, post(b.String()))

	return nil
}

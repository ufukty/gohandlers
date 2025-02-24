package validate

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"gohandlers/cmd/gohandlers/commands/version"
	"gohandlers/pkg/inspects"
	"os"
	"path/filepath"
	"strings"
)

type Args struct {
	Dir     string
	Out     string
	Recv    string
	Verbose bool
}

func post(src string) string {
	src = strings.ReplaceAll(src, "}\nfunc", "}\n\nfunc")
	return src
}

func filterByRecv(infoss map[inspects.Receiver]map[string]inspects.Info, recvt string) (map[inspects.Receiver]map[string]inspects.Info, error) {
	for recv := range infoss {
		if recvt == recv.Type {
			return map[inspects.Receiver]map[string]inspects.Info{
				recv: infoss[recv],
			}, nil
		}
	}
	return nil, fmt.Errorf("receiver not found: %s", recvt)
}

func quotes(s string) string {
	return fmt.Sprintf("%q", s)
}

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

func produce(bti *inspects.BindingTypeInfo) *ast.FuncDecl {
	var fd = &ast.FuncDecl{
		Recv: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "bq"}}, Type: &ast.Ident{Name: bti.Typename}}}},
		Name: &ast.Ident{Name: "Validate"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{List: []*ast.Field{
				{Names: []*ast.Ident{{Name: "errs"}}, Type: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "error"}}},
			}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}
	params := merge(
		bti.Params.Form,
		bti.Params.Json,
		bti.Params.Query,
		bti.Params.Route,
	)
	for p, fn := range params {
		fd.Body.List = append(fd.Body.List, &ast.IfStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "bq"}, Sel: &ast.Ident{Name: fn}},
					Sel: &ast.Ident{Name: "Validate"},
				}}},
			},
			Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.IndexExpr{X: &ast.Ident{Name: "errs"}, Index: &ast.BasicLit{Kind: token.STRING, Value: quotes(p)}}},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{&ast.Ident{Name: "err"}},
			}}},
		})
	}
	fd.Body.List = append(fd.Body.List, &ast.ReturnStmt{})
	return fd
}

func Main() error {
	args := &Args{}
	flag.StringVar(&args.Dir, "dir", "", "the directory contains Go files")
	flag.StringVar(&args.Out, "out", "", "the output file")
	flag.StringVar(&args.Recv, "recv", "", "only use request types that is prefixed with handlers defined on this type")
	flag.BoolVar(&args.Verbose, "v", false, "prints additional information")
	flag.Parse()

	if args.Dir == "" {
		flag.PrintDefaults()
		return fmt.Errorf("bad arguments")
	}

	infoss, pkg, err := inspects.Dir(args.Dir, args.Verbose)
	if err != nil {
		return fmt.Errorf("inspecting the directory: %w", err)
	}

	if args.Recv != "" {
		infoss, err = filterByRecv(infoss, args.Recv)
		if err != nil {
			return fmt.Errorf("filtering binding types based on the receiver type of handlers: %w", err)
		}
	}

	f := &ast.File{
		Name:  ast.NewIdent(pkg),
		Decls: []ast.Decl{},
	}

	for _, o := range ordered(infoss) {
		i := infoss[o.receiver][o.handler]
		if i.RequestType != nil {
			f.Decls = append(f.Decls, produce(i.RequestType))
		}
	}

	if args.Out == "" {
		args.Out = "validate.gh.go"
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
	fmt.Fprint(fh, version.Top())
	fmt.Fprint(fh, post(b.String()))

	return nil
}

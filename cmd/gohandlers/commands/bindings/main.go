package bindings

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

func ternary[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
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

	f.Decls = append(f.Decls, bqBuildFuncs(infoss)...)

	if args.Out == "" {
		args.Out = "bindings.gh.go"
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

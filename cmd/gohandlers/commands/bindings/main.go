package bindings

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"gohandlers/cmd/gohandlers/commands/bindings/imports"
	"gohandlers/cmd/gohandlers/commands/bindings/produce"
	"gohandlers/cmd/gohandlers/commands/bindings/utilities"
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
		Name: ast.NewIdent(pkg),
		Decls: []ast.Decl{
			&ast.GenDecl{Tok: token.IMPORT, Specs: imports.List(infoss)},
		},
	}

	f.Decls = append(f.Decls, utilities.Produce(infoss)...)

	for _, o := range ordered(infoss) {
		i := infoss[o.receiver][o.handler]
		if i.RequestType != nil {
			f.Decls = append(f.Decls, produce.BqBuild(i))
			f.Decls = append(f.Decls, produce.BqUmarshal(i)...)
			f.Decls = append(f.Decls, produce.BqParse(i))
		}
		if i.ResponseType != nil {
			f.Decls = append(f.Decls, produce.BsWrite(i))
			f.Decls = append(f.Decls, produce.BsParse(i))
		}
	}

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

package helpers

import (
	"bytes"
	"cmp"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"os"
	"slices"

	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/helpers/internal/construct"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/helpers/internal/imports"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/helpers/internal/post"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/helpers/internal/utilities"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/version"
	"github.com/ufukty/gohandlers/pkg/inspects"
)

type Args struct {
	Dir     string
	Out     string
	Recv    string
	PkgName string
	Verbose bool
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

// just used for stable ordering of methods
type funcrecv struct {
	receiver inspects.Receiver // receiver of the handler. not the receiver of Parse and Build methods
	handler  string
}

func ordered(infoss map[inspects.Receiver]map[string]inspects.Info) []funcrecv {
	o := []funcrecv{}
	for recv, handlers := range infoss {
		for handler := range handlers {
			o = append(o, funcrecv{recv, handler})
		}
	}
	slices.SortFunc(o, func(a, b funcrecv) int {
		return cmp.Or(cmp.Compare(a.receiver.Type, b.receiver.Type), cmp.Compare(a.handler, b.handler))
	})
	return o
}

func pretty(f *ast.File) (io.Reader, error) {
	b := bytes.NewBuffer([]byte{})
	err := printer.Fprint(b, token.NewFileSet(), f)
	if err != nil {
		return nil, fmt.Errorf("printing: %w", err)
	}
	fmt.Fprint(b, version.Top())
	fmt.Fprint(b, post.Process(b.String()))
	return b, nil
}

func Main() error {
	args := &Args{}
	flag.StringVar(&args.Dir, "dir", ".", "the source directory contains Go files for handlers and binding types")
	flag.StringVar(&args.Out, "out", "gh.go", "the path for output file")
	flag.StringVar(&args.PkgName, "pkg", "", "override the package name resolved from Go files")
	flag.StringVar(&args.Recv, "recv", "", "ignore handlers defined on other receivers")
	flag.BoolVar(&args.Verbose, "v", false, "prints additional information")
	flag.Parse()

	if args.Dir == "" {
		flag.PrintDefaults()
		return fmt.Errorf("bad arguments")
	}

	infoss, pkgName, err := inspects.Dir(args.Dir, args.Verbose)
	if err != nil {
		return fmt.Errorf("inspecting the directory: %w", err)
	}

	if args.PkgName != "" {
		pkgName = args.PkgName
	}

	if args.Recv != "" {
		infoss, err = filterByRecv(infoss, args.Recv)
		if err != nil {
			return fmt.Errorf("filtering binding types based on the receiver type of handlers: %w", err)
		}
	}
	f := &ast.File{
		Name: ast.NewIdent(pkgName),
		Decls: []ast.Decl{
			&ast.GenDecl{Tok: token.IMPORT, Specs: imports.List(infoss)},
		},
	}

	f.Decls = append(f.Decls, construct.Listers(infoss)...)
	f.Decls = append(f.Decls, utilities.Produce(infoss)...)
	for _, o := range ordered(infoss) {
		i := infoss[o.receiver][o.handler]
		if i.RequestType != nil {
			f.Decls = append(f.Decls, construct.BqBuild(i))
			if len(i.RequestType.Params.Form) > 0 {
				f.Decls = append(f.Decls, construct.BqUnmarshalFormData(i))
			}
			f.Decls = append(f.Decls, construct.BqParse(i))
			f.Decls = append(f.Decls, construct.BqValidate(i.RequestType))
		}
		if i.ResponseType != nil {
			f.Decls = append(f.Decls, construct.BsWrite(i))
			f.Decls = append(f.Decls, construct.BsParse(i))
		}
	}

	print, err := pretty(f)
	if err != nil {
		return fmt.Errorf("pretty printing: %w", err)
	}
	fh, err := os.Create(args.Out)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer fh.Close()
	_, err = io.Copy(fh, print)
	if err != nil {
		return fmt.Errorf("writing to output file: %w", err)
	}

	return nil
}

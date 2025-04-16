package client

import (
	"bytes"
	"flag"
	"fmt"
	"go/printer"
	"go/token"
	"os"
	"strings"

	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/version"
	"github.com/ufukty/gohandlers/pkg/inspects"
)

type Args struct {
	Dir     string
	Out     string
	Pkg     string
	Import  string
	Verbose bool
}

func post(src string) string {
	src = strings.ReplaceAll(src, "}\nfunc", "}\n\nfunc")
	src = strings.ReplaceAll(src, "}\ntype", "}\n\ntype")
	return src
}

func Main() error {
	args := Args{}
	flag.StringVar(&args.Dir, "dir", "", "input directory")
	flag.StringVar(&args.Out, "out", "", "output directory")
	flag.StringVar(&args.Pkg, "pkg", "", "package name for the generated file")
	flag.StringVar(&args.Import, "import", "", "the import path of package declares binding types")
	flag.BoolVar(&args.Verbose, "v", false, "prints additional information")
	flag.Parse()

	if args.Dir == "" || args.Out == "" || args.Pkg == "" {
		flag.PrintDefaults()
		return fmt.Errorf("invalid arguments")
	}

	infoss, pkgsrc, err := inspects.Dir(args.Dir, args.Verbose)
	if err != nil {
		return fmt.Errorf("inspecting files: %w", err)
	}

	f := file(infoss, args.Pkg, pkgsrc, args.Import)

	fh, err := os.Create(args.Out)
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

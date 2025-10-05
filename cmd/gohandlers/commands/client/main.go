package client

import (
	"flag"
	"fmt"
	"io"
	"os"

	"go.ufukty.com/gohandlers/cmd/gohandlers/commands/client/construct"
	"go.ufukty.com/gohandlers/cmd/gohandlers/internal/pretty"
	"go.ufukty.com/gohandlers/pkg/inspects"
)

type Args struct {
	Dir     string
	Out     string
	Pkg     string
	Import  string
	Verbose bool
}

func Main() error {
	args := Args{}
	flag.StringVar(&args.Dir, "dir", "", "input directory")
	flag.StringVar(&args.Out, "out", "", "output file (probably in a \"client\" folder)")
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

	f := construct.File(infoss, args.Pkg, pkgsrc, args.Import)

	print, err := pretty.Print(f)
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

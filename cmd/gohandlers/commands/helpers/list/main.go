package list

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/helpers/internal/construct"
	"github.com/ufukty/gohandlers/pkg/inspects"
)

type Args struct {
	Dir     string
	Out     string
	Verbose bool
}

func Main() error {
	args := Args{}
	flag.StringVar(&args.Dir, "dir", "", "the directory contains Go files")
	flag.StringVar(&args.Out, "out", "list.gh.go", "output file that will be generated in the 'dir'")
	flag.BoolVar(&args.Verbose, "v", false, "prints additional information")
	flag.Parse()

	if args.Dir == "" {
		flag.PrintDefaults()
		return fmt.Errorf("missing arguments")
	}

	infoss, pkgname, err := inspects.Dir(args.Dir, args.Verbose)
	if err != nil {
		return fmt.Errorf("inspecting directory and handlers: %w", err)
	}

	err = construct.Listers(filepath.Join(args.Dir, args.Out), infoss, pkgname)
	if err != nil {
		return fmt.Errorf("creating the main file: %w", err)
	}

	return nil
}

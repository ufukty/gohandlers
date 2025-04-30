package list

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/ufukty/gohandlers/pkg/inspects"
)

type HandlerInfo struct {
	Typename   string
	ImportPath string
}

type Args struct {
	Dir     string
	Out     string
	Hi      HandlerInfo
	Verbose bool
}

func Main() error {
	args := Args{}
	flag.StringVar(&args.Dir, "dir", "", "the directory contains Go files")
	flag.StringVar(&args.Out, "out", "list.gh.go", "output file that will be generated in the 'dir'")
	flag.StringVar(&args.Hi.Typename, "hi-type", "", "the string to be substituted with mentions of HandlerInfo")
	flag.StringVar(&args.Hi.ImportPath, "hi-import", "", "the package to import for custom implementation of HandlerInfo")
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

	err = create(filepath.Join(args.Dir, args.Out), infoss, pkgname, args.Hi)
	if err != nil {
		return fmt.Errorf("creating the main file: %w", err)
	}

	return nil
}

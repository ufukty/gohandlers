package list

import (
	"flag"
	"fmt"
	"gohandlers/cmd/gohandlers/commands/version"
	"gohandlers/pkg/implements"
	"gohandlers/pkg/inspects"
	"path/filepath"
)

type Args struct {
	Dir    string
	Out    string
	Type   string // the type to substitude with HandlerInfo
	Import string // import path of the package contains '-type' declaration
}

func Main() error {
	args := Args{}
	flag.StringVar(&args.Dir, "dir", "", "the directory contains Go files. one handler and a request binding type is allowed per file")
	flag.StringVar(&args.Out, "out", "list.gh.go", "output file that will be generated in the 'dir'")
	flag.StringVar(&args.Type, "type", "", "the type substituded with HandlerInfo")
	flag.StringVar(&args.Import, "import", "", "the package contains the hit declaration")
	flag.Parse()

	if args.Dir == "" {
		flag.PrintDefaults()
		return fmt.Errorf("missing arguments")
	}

	infoss, pkgname, err := inspects.Dir(args.Dir)
	if err != nil {
		return fmt.Errorf("inspecting directory and handlers: %w", err)
	}

	err = implements.ListFile(filepath.Join(args.Dir, args.Out), infoss, pkgname, args.Type, args.Import, version.Version)
	if err != nil {
		return fmt.Errorf("creating the main file: %w", err)
	}

	return nil
}

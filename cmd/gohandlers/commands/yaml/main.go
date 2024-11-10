package yaml

import (
	"flag"
	"fmt"
	"gohandlers/pkg/inspects"
	"path/filepath"
)

type Args struct {
	Dir     string
	Out     string
	Verbose bool
}

func Main() error {
	args := Args{}
	flag.StringVar(&args.Dir, "dir", "", "the directory contains Go files. one handler and a request binding type is allowed per file")
	flag.StringVar(&args.Out, "yaml", "gh.yml", "yaml file that will be generated in the 'dir'")
	flag.BoolVar(&args.Verbose, "v", false, "prints additional information")
	flag.Parse()

	if args.Dir == "" {
		flag.PrintDefaults()
		return fmt.Errorf("missing arguments")
	}

	infoss, _, err := inspects.Dir(args.Dir, args.Verbose)
	if err != nil {
		return fmt.Errorf("inspecting directory and handlers: %w", err)
	}

	err = create(filepath.Join(args.Dir, args.Out), infoss)
	if err != nil {
		return fmt.Errorf("creating the yaml file: %w", err)
	}

	return nil
}

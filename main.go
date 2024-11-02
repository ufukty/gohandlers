package main

import (
	"flag"
	"fmt"
	"gohandlers/pkg/implements"
	"gohandlers/pkg/inspects"
	"os"
	"path/filepath"
)

var Version string

type Args struct {
	Dir      string
	Out      string
	Yaml     string
	HiType   string // the type substituded with HandlerInfo
	HiImport string // the package contains the hit declaration
}

func Main() error {
	args := Args{}
	flag.StringVar(&args.Dir, "dir", "", "the directory contains Go files. one handler and a request binding type is allowed per file")
	flag.StringVar(&args.Out, "out", "handlers.go", "output file that will be generated in the 'dir'")
	flag.StringVar(&args.HiType, "hit", "", "the type substituded with HandlerInfo")
	flag.StringVar(&args.HiImport, "hii", "", "the package contains the hit declaration")
	flag.StringVar(&args.Yaml, "yaml", "", "yaml file that will be generated in the 'dir'")
	flag.Parse()

	if args.Dir == "" {
		flag.PrintDefaults()
		return fmt.Errorf("missing arguments")
	}

	infoss, pkgname, err := inspects.Dir(args.Dir)
	if err != nil {
		return fmt.Errorf("inspecting directory and handlers: %w", err)
	}

	err = implements.HandlersFile(filepath.Join(args.Dir, args.Out), infoss, pkgname, args.HiType, args.HiImport, Version)
	if err != nil {
		return fmt.Errorf("creating the main file: %w", err)
	}

	if args.Yaml != "" {
		err = implements.YamlFile(filepath.Join(args.Dir, args.Yaml), infoss)
		if err != nil {
			return fmt.Errorf("creating the yaml file: %w", err)
		}
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// dd is documentation website development server
// it is only meant to be used to test the produced
// files in developer machine.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

type Args struct {
	Dir string
}

func Main() error {
	args := &Args{}
	flag.StringVar(&args.Dir, "dir", "", "path to the root of website, which is the one contain index.html")
	flag.Parse()

	if args.Dir == "" {
		flag.PrintDefaults()
		return fmt.Errorf("arg not found: dir")
	}

	if err := http.ListenAndServe(":8080", http.FileServer(http.Dir(args.Dir))); err != nil {
		return fmt.Errorf("http.ListenAndServe: %w", err)
	}
	return nil
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

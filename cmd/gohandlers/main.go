package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/bindings"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/client"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/list"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/mock"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/validate"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/version"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/yaml"
)

func listcmds(commands map[string]func() error) string {
	return strings.Join(slices.Sorted(maps.Keys(commands)), ", ")
}

func Main() error {
	commands := map[string]func() error{
		"bindings": bindings.Main,
		"client":   client.Main,
		"list":     list.Main,
		"mock":     mock.Main,
		"validate": validate.Main,
		"version":  version.Main,
		"yaml":     yaml.Main,
	}

	if len(os.Args) < 2 {
		return fmt.Errorf("subcommands: %s", listcmds(commands))
	}

	cmd := os.Args[1]
	command, ok := commands[cmd]
	if !ok {
		return fmt.Errorf("available subcommands: %s", listcmds(commands))
	}

	os.Args = os.Args[1:]
	err := command()
	if err != nil {
		return fmt.Errorf("%s: %w", cmd, err)
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

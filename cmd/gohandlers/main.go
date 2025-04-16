package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/bindings"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/clients"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/list"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/mock"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/validate"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/version"
	"github.com/ufukty/gohandlers/cmd/gohandlers/commands/yaml"

	"golang.org/x/exp/maps"
)

func Main() error {
	commands := map[string]func() error{
		"bindings": bindings.Main,
		"clients":  clients.Main,
		"list":     list.Main,
		"mock":     mock.Main,
		"validate": validate.Main,
		"version":  version.Main,
		"yaml":     yaml.Main,
	}

	if len(os.Args) < 2 {
		return fmt.Errorf("subcommands: %s", strings.Join(maps.Keys(commands), ", "))
	}

	cmd := os.Args[1]
	command, ok := commands[cmd]
	if !ok {
		return fmt.Errorf("available subcommands: %s", strings.Join(maps.Keys(commands), ", "))
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

package main

import (
	"fmt"
	"gohandlers/cmd/gohandlers/commands/bindings"
	"gohandlers/cmd/gohandlers/commands/clients"
	"gohandlers/cmd/gohandlers/commands/list"
	"gohandlers/cmd/gohandlers/commands/mock"
	"gohandlers/cmd/gohandlers/commands/version"
	"gohandlers/cmd/gohandlers/commands/yaml"
	"os"
	"strings"

	"golang.org/x/exp/maps"
)

func Main() error {
	commands := map[string]func() error{
		"bindings": bindings.Main,
		"clients":  clients.Main,
		"list":     list.Main,
		"mock":     mock.Main,
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

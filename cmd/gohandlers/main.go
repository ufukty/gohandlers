package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"go.ufukty.com/gohandlers/cmd/gohandlers/commands/client"
	"go.ufukty.com/gohandlers/cmd/gohandlers/commands/helpers"
	"go.ufukty.com/gohandlers/cmd/gohandlers/commands/version"
	"go.ufukty.com/gohandlers/cmd/gohandlers/commands/yaml"
)

func listcmds(commands map[string]func() error) string {
	return strings.Join(slices.Sorted(maps.Keys(commands)), ", ")
}

func Main() error {
	commands := map[string]func() error{
		"client":  client.Main,
		"helpers": helpers.Main,
		"version": version.Main,
		"yaml":    yaml.Main,
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

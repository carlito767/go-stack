package clp

import (
	"fmt"
	"os"
)

type Handler func([]string) error

// Parses the command-line arguments from os.Args[1:]
// and handles the corresponding command.
func HandleCommands(handlers map[string]Handler) error {
	return HandleCommandsFromArgs(handlers, os.Args[1:])
}

// Parses the given arguments
// and handles the corresponding command.
func HandleCommandsFromArgs(handlers map[string]Handler, args []string) error {
	if _, ok := handlers[""]; ok {
		isDefaultCommand := true
		if len(args) > 0 {
			_, ok = handlers[args[0]]
			isDefaultCommand = !ok
		}
		if isDefaultCommand {
			args = append([]string{""}, args...)
		}
	}

	if len(args) == 0 {
		return fmt.Errorf("missing command")
	}

	name := args[0]
	if handler, found := handlers[name]; found {
		if err := handler(args[1:]); err != nil {
			return fmt.Errorf("command '%s': %w", name, err)
		}
		return nil
	}

	return fmt.Errorf("unknown command: %s", name)
}

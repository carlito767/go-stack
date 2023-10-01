/*
Package clp implements command-line parsing.

Inspired by:

	https://pkg.go.dev/flag
	https://pkg.go.dev/github.com/jessevdk/go-flags

# Supported features

The following features are supported:

	Options with short names (-v)
	Options with long names (--verbose)
	Options with and without values
	Supports multiple short options (-fr is equivalent to -f -r)
	Supports the following types: bool, int{8..64}, uint{8..64}, float{32..64}, string
	Supports slices

# Usage

	err := clp.HandleCommands(commands)
	err := clp.HandleCommandsFromArgs(commands, args)
	err := clp.ParseOptions(&options)
	err := clp.ParseOptionsFromArgs(&options, args)

See cmd/server/main.go for example.

# Command-line option syntax

The following forms are permitted:

	-o              // short option without value (boolean option only)
	-o value        // short option with value (non-boolean option only)
	-o=value        // short option with value
	-op             // multiple short options without value (is equivalent to -op=true)
	-op=value       // multiple short options with value
	--option        // long option without value (boolean option only)
	--option value  // long option with value (non-boolean option only)
	--option=value  // long option with value

Option parsing stops ather the terminator "--".
*/
package clp

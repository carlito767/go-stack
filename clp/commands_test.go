package clp_test

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/carlito767/go-stack/clp"
)

// Command Add

type OptionsAdd struct {
	Operands []int
}

func HandleAdd(args []string) error {
	var o OptionsAdd
	if err := clp.ParseOptionsFromArgs(&o, args); err != nil {
		return err
	}

	total := 0
	var operands []string
	for _, operand := range o.Operands {
		total += operand
		operands = append(operands, strconv.Itoa(operand))
	}
	fmt.Printf("%s = %d\n", strings.Join(operands, " + "), total)
	return nil
}

// Command Hello

type OptionsHello struct {
	Language string `name:"lang,l"`
	Name     string
}

func HandleHello(args []string) error {
	o := OptionsHello{Name: "World"}
	if err := clp.ParseOptionsFromArgs(&o, args); err != nil {
		return err
	}

	var format string
	switch o.Language {
	case "fr":
		format = "Bonjour %v !\n"
	default:
		format = "Hello %v!\n"
	}

	fmt.Printf(format, o.Name)
	return nil
}

func TestParseCommands(t *testing.T) {
	t.Run("parse commands", func(t *testing.T) {
		defer func(old []string) { os.Args = old }(os.Args)
		os.Args = []string{"app", "hello", "World"}

		commands := map[string]clp.Handler{"hello": HandleHello}
		if err := clp.HandleCommands(commands); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestHandleCommandsFromArgs(t *testing.T) {
	commands := map[string]clp.Handler{"add": HandleAdd, "hello": HandleHello}

	tests := []struct {
		name     string
		commands map[string]clp.Handler
		args     []string
		wantErr  error
	}{
		{
			name:     "command add",
			commands: commands,
			args:     []string{"add", "--", "-1", "-2", "3", "4", "5", "6"},
			wantErr:  nil,
		},
		{
			name:     "command hello",
			commands: commands,
			args:     []string{"hello", "--lang=fr", "mon ami"},
			wantErr:  nil,
		},
		{
			name:     "default command",
			commands: map[string]clp.Handler{"add": HandleAdd, "": HandleHello},
			args:     []string{"--lang=fr", "mon ami"},
			wantErr:  nil,
		},
		{
			name:     "missing command",
			commands: commands,
			args:     []string{},
			wantErr:  errors.New("missing command"),
		},
		{
			name:     "options parsing failed",
			commands: commands,
			args:     []string{"add", "--", "1", "2", "3", "sun"},
			wantErr:  errors.New("command 'add': options parsing failed: strconv.ParseInt: parsing \"sun\": invalid syntax"),
		},
		{
			name:     "unknown command",
			commands: commands,
			args:     []string{"foo", "bar"},
			wantErr:  errors.New("unknown command: foo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clp.HandleCommandsFromArgs(tt.commands, tt.args)

			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("error: %v, wantErr: %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("error: %v, wantErr: %v", err, tt.wantErr)
			}
		})
	}
}

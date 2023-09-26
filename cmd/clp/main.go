package main

import (
	"fmt"
	"strconv"
	"strings"

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

func main() {
	handlers := map[string]clp.Handler{
		"add": HandleAdd,
		"":    HandleHello,
	}
	if err := clp.HandleCommands(handlers); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
}

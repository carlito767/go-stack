package clp

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type ParsingError struct {
	err error
}

func (e ParsingError) Error() string {
	return fmt.Sprintf("options parsing failed: %s", e.err.Error())
}

func (e *ParsingError) Unwrap() error {
	return e.err
}

// ParseOptions parses the command-line arguments from os.Args[1:]
// and updates the struct fields with the corresponding values.
func ParseOptions(data interface{}) error {
	return ParseOptionsFromArgs(data, os.Args[1:])
}

// ParseOptionsFromArgs parses the given arguments
// and updates the struct fields with the corresponding values.
func ParseOptionsFromArgs(data interface{}, args []string) error {
	if err := parse(data, args); err != nil {
		return &ParsingError{err}
	}
	return nil
}

func parse(data interface{}, args []string) error {
	if data == nil {
		return fmt.Errorf("uninitialized options")
	}
	// verify that data is a pointer to a struct
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("options must be a pointer to a struct")
	}

	// check data fields
	tp := reflect.TypeOf(data).Elem()
	options := make(map[string]reflect.StructField)
	remainings := []reflect.StructField{}
	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("name")
		if tag == "" {
			// it's not an option
			remainings = append(remainings, field)
		} else {
			// it's an option
			for _, v := range strings.Split(tag, ",") {
				if v == "" {
					continue
				}
				if _, ok := options[v]; ok {
					return fmt.Errorf("duplicated option name in fields '%s' and '%s': '%s'", options[v].Name, field.Name, v)
				}
				options[v] = field
			}
		}
	}

	// parse the arguments
	checkOptions := true
	seen := make(map[string]bool)
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if checkOptions && arg == "--" {
			// "--" terminates the options
			checkOptions = false
			continue
		}

		d := dashes(arg)
		if !checkOptions || d == 0 {
			// it's not an option
			if len(remainings) == 0 {
				return fmt.Errorf("unhandled argument: '%s'", arg)
			}

			field := remainings[0]
			vfield := val.Elem().FieldByName(field.Name)
			if err := convert(arg, vfield); err != nil {
				return err
			}
			if field.Type.Kind() != reflect.Slice {
				remainings = remainings[1:]
			}
			continue
		}

		// it's an option
		name := arg[d:]
		value := ""
		for j := 1; j < len(name); j++ { // equal sign cannot be first
			if name[j] == '=' {
				value = name[j+1:]
				name = name[0:j]
				break
			}
		}

		if d > 1 && len(name) == 1 {
			return fmt.Errorf("invalid syntax: '%s' (use single dash for short option)", arg)
		}

		multipleShortOptions := (d == 1 && len(name) > 1)
		names := []string{name}
		if multipleShortOptions {
			names = strings.Split(name, "")
		}

		missingValue := (name == arg[d:])
		for _, id := range names {
			field, ok := options[id]
			if !ok {
				return fmt.Errorf("invalid option: '%s'", id)
			}
			if seen[field.Name] && field.Type.Kind() != reflect.Slice {
				return fmt.Errorf("redefined field: '%s'", field.Name)
			}

			// missing value, try to find it
			if missingValue {
				if multipleShortOptions || field.Type.Kind() == reflect.Bool {
					// it's a boolean field (we assume it for mutiple short options)
					value = "true"
				} else {
					// it's not a boolean field
					if i >= len(args)-1 {
						return fmt.Errorf("missing value for field '%s'", field.Name)
					}
					i++
					value = args[i]
				}
				missingValue = false
			}

			seen[field.Name] = true
			vfield := val.Elem().FieldByName(field.Name)
			if err := convert(value, vfield); err != nil {
				return err
			}
		}
	}

	return nil
}

func convert(v string, field reflect.Value) error {
	tp := field.Type()

	switch tp.Kind() {
	case reflect.Bool:
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		field.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(v, 10, tp.Bits())
		if err != nil {
			return err
		}
		field.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(v, 10, tp.Bits())
		if err != nil {
			return err
		}
		field.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(v, tp.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(parsed)
	case reflect.Slice:
		elemtp := tp.Elem()
		elemvalptr := reflect.New(elemtp)
		elemval := reflect.Indirect(elemvalptr)
		if err := convert(v, elemval); err != nil {
			return err
		}
		field.Set(reflect.Append(field, elemval))
	case reflect.String:
		field.SetString(v)
	default:
		return fmt.Errorf("unsupported field type: '%s'", tp.Kind())
	}
	return nil
}

func dashes(arg string) int {
	// single dash option
	if len(arg) > 1 && arg[0] == '-' && arg[1] != '-' && arg[1] != '=' {
		return 1
	}

	// double dash option
	if len(arg) > 2 && arg[0] == '-' && arg[1] == '-' && arg[2] != '-' && arg[2] != '=' {
		return 2
	}

	// not an option
	return 0
}

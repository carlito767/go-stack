package clp_test

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/carlito767/go-stack/clp"
)

type DuplicatedOptionName struct {
	Foo1 string `name:"foo"`
	Foo2 string `name:"foo"`
}

type EmptyTag struct {
	V bool `name:"value,,,,v"`
}

type MultipleShortOptions struct {
	A bool   `name:"a"`
	B bool   `name:"b"`
	C bool   `name:"c"`
	S string `name:"s"`
}

type Person struct {
	Surname string `name:"surname,s"`
	Name    string `name:"name,n"`
	Age     uint
	Address []string
}

type SliceOption struct {
	Numbers []int `name:"n"`
}

type Supported struct {
	Bool    bool    `name:"bool,b"`
	Int     int     `name:"i"`
	Int8    int8    `name:"int8"`
	Int16   int     `name:"int16"`
	Int32   int     `name:"int32"`
	Int64   int     `name:"int64"`
	Uint    uint    `name:"u"`
	Uint8   int8    `name:"uint8"`
	Uint16  int     `name:"uint16"`
	Uint32  int     `name:"uint32"`
	Uint64  int     `name:"uint64"`
	Float32 float32 `name:"float32"`
	Float64 float32 `name:"float64"`
	String  string  `name:"s"`
}

type UnexportedField struct {
	Exported   bool `name:"e"`
	unexported bool `name:"u"`
}

type UnsupportedArray struct {
	Foo [3]int `name:"f"`
}

type UnsupportedChan struct {
	Foo chan string `name:"f"`
}

type UnsupportedComplex64 struct {
	Foo complex64 `name:"f"`
}

type UnsupportedComplex128 struct {
	Foo complex128 `name:"f"`
}

type UnsupportedFunc struct {
	Foo func() `name:"f"`
}

type UnsupportedInterface struct {
	Foo interface{} `name:"f"`
}

type UnsupportedMap struct {
	Foo map[string]string `name:"f"`
}

type UnsupportedPointer struct {
	Foo *int `name:"f"`
}

type UnsupportedSlice struct {
	Foo []bool `name:"f"`
}

type UnsupportedStruct struct {
	Foo struct{} `name:"f"`
}

type UnsupportedUintprt struct {
	Foo uintptr `name:"f"`
}

func unsupported(t string) error {
	return fmt.Errorf("unsupported field type: '%s'", t)
}

func TestParseOptions(t *testing.T) {
	t.Run("parse options", func(t *testing.T) {
		defer func(old []string) { os.Args = old }(os.Args)
		os.Args = []string{"app", "--surname=Doe", "--name=John"}

		var p Person
		if err := clp.ParseOptions(&p); err != nil {
			t.Errorf("unexpected parsing error: %v", err)
		}

		if p.Surname != "Doe" || p.Name != "John" {
			t.Error("unexpected parsing error")
		}
	})
}

func TestParseOptionsFromArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    interface{}
		wantErr error
	}{
		{
			name:    "duplicated option name",
			args:    []string{"--foo=duplicated"},
			want:    &DuplicatedOptionName{},
			wantErr: errors.New("duplicated option name in fields 'Foo1' and 'Foo2': 'foo'"),
		},
		{
			name:    "empty tag",
			args:    []string{"-v=0"},
			want:    &EmptyTag{},
			wantErr: nil,
		},
		{
			name:    "invalid option",
			args:    []string{"--invalid-option"},
			want:    &Supported{},
			wantErr: errors.New("invalid option: 'invalid-option'"),
		},
		{
			name:    "invalid syntax",
			args:    []string{"--b"},
			want:    &Supported{},
			wantErr: errors.New("invalid syntax: '--b' (use single dash for short option)"),
		},
		{
			name:    "invalid bool value",
			args:    []string{"-b=not-bool"},
			want:    &Supported{},
			wantErr: errors.New("strconv.ParseBool: parsing \"not-bool\": invalid syntax"),
		},
		{
			name:    "invalid float32 value",
			args:    []string{"--float32=not-float"},
			want:    &Supported{},
			wantErr: errors.New("strconv.ParseFloat: parsing \"not-float\": invalid syntax"),
		},
		{
			name:    "invalid int value",
			args:    []string{"-i=not-int"},
			want:    &Supported{},
			wantErr: errors.New("strconv.ParseInt: parsing \"not-int\": invalid syntax"),
		},
		{
			name:    "invalid uint value",
			args:    []string{"-u=not-uint"},
			want:    &Supported{},
			wantErr: errors.New("strconv.ParseUint: parsing \"not-uint\": invalid syntax"),
		},
		{
			name:    "missing value",
			args:    []string{"-s"},
			want:    &Supported{},
			wantErr: errors.New("missing value for field 'String'"),
		},
		{
			name:    "multiple short options",
			args:    []string{"-abc"},
			want:    &MultipleShortOptions{A: true, B: true, C: true},
			wantErr: nil,
		},
		{
			name:    "multiple short options of different types",
			args:    []string{"-asc"},
			want:    &MultipleShortOptions{A: true, C: true, S: "true"},
			wantErr: nil,
		},
		{
			name:    "options must be a pointer to a struct",
			args:    []string{},
			want:    []int{1, 2, 3},
			wantErr: errors.New("options must be a pointer to a struct"),
		},
		{
			name:    "redefined field",
			args:    []string{"--surname=Doe", "--name=John", "-n=Jane"},
			want:    &Person{},
			wantErr: errors.New("redefined field: 'Name'"),
		},
		{
			name:    "remainings",
			args:    []string{"--surname=Doe", "--name=John", "30", "LA", "USA"},
			want:    &Person{Surname: "Doe", Name: "John", Age: 30, Address: []string{"LA", "USA"}},
			wantErr: nil,
		},
		{
			name:    "slice option",
			args:    []string{"-n=1", "-n=2", "-n=3"},
			want:    &SliceOption{Numbers: []int{1, 2, 3}},
			wantErr: nil,
		},
		{
			name:    "supported, syntax with equals",
			args:    []string{"-b", "-s=Hello", "-i=42", "--int8=-41", "--int16=40", "--int32=-39", "--int64=38", "-u=37", "--uint8=36", "--uint16=35", "--uint32=34", "--uint64=33", "--float32=-12.34", "--float64=56.789"},
			want:    &Supported{Bool: true, Int: 42, Int8: -41, Int16: 40, Int32: -39, Int64: 38, Uint: 37, Uint8: 36, Uint16: 35, Uint32: 34, Uint64: 33, Float32: -12.34, Float64: 56.789, String: "Hello"},
			wantErr: nil,
		},
		{
			name:    "supported, syntax with spaces",
			args:    []string{"-b", "-s", "Hello", "-i", "42", "--int8", "-41", "--int16", "40", "--int32", "-39", "--int64", "38", "-u", "37", "--uint8", "36", "--uint16", "35", "--uint32", "34", "--uint64", "33", "--float32", "-12.34", "--float64", "56.789"},
			want:    &Supported{Bool: true, Int: 42, Int8: -41, Int16: 40, Int32: -39, Int64: 38, Uint: 37, Uint8: 36, Uint16: 35, Uint32: 34, Uint64: 33, Float32: -12.34, Float64: 56.789, String: "Hello"},
			wantErr: nil,
		},
		{
			name:    "terminates the options",
			args:    []string{"--surname=Doe", "--name=John", "--", "30", "-LA-", "USA"},
			want:    &Person{Surname: "Doe", Name: "John", Age: 30, Address: []string{"-LA-", "USA"}},
			wantErr: nil,
		},
		{
			name:    "unexported field",
			args:    []string{"-e=1", "-u=1"},
			want:    &UnexportedField{},
			wantErr: errors.New("invalid option: 'u'"),
		},
		{
			name:    "unhandled argument",
			args:    []string{"unhandled-argument"},
			want:    &Supported{},
			wantErr: errors.New("unhandled argument: 'unhandled-argument'"),
		},
		{
			name:    "uninitialized arguments",
			args:    nil,
			want:    &Supported{},
			wantErr: nil,
		},
		{
			name:    "uninitialized options",
			args:    []string{},
			want:    nil,
			wantErr: errors.New("uninitialized options"),
		},
		{
			name:    "unsupported array",
			args:    []string{"-f="},
			want:    &UnsupportedArray{},
			wantErr: unsupported("array"),
		},
		{
			name:    "unsupported chan",
			args:    []string{"-f="},
			want:    &UnsupportedChan{},
			wantErr: unsupported("chan"),
		},
		{
			name:    "unsupported complex64",
			args:    []string{"-f="},
			want:    &UnsupportedComplex64{},
			wantErr: unsupported("complex64"),
		},
		{
			name:    "unsupported complex128",
			args:    []string{"-f="},
			want:    &UnsupportedComplex128{},
			wantErr: unsupported("complex128"),
		},
		{
			name:    "unsupported func",
			args:    []string{"-f="},
			want:    &UnsupportedFunc{},
			wantErr: unsupported("func"),
		},
		{
			name:    "unsupported interface",
			args:    []string{"-f="},
			want:    &UnsupportedInterface{},
			wantErr: unsupported("interface"),
		},
		{
			name:    "unsupported map",
			args:    []string{"-f="},
			want:    &UnsupportedMap{},
			wantErr: unsupported("map"),
		},
		{
			name:    "unsupported pointer",
			args:    []string{"-f="},
			want:    &UnsupportedPointer{},
			wantErr: unsupported("ptr"),
		},
		{
			name:    "unsupported struct",
			args:    []string{"-f="},
			want:    &UnsupportedStruct{},
			wantErr: unsupported("struct"),
		},
		{
			name:    "unsupported uintptr",
			args:    []string{"-f="},
			want:    &UnsupportedUintprt{},
			wantErr: unsupported("uintptr"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			if tt.want != nil {
				got = reflect.New(reflect.TypeOf(tt.want).Elem()).Interface()
			}

			err := clp.ParseOptionsFromArgs(got, tt.args)

			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("error: %v, wantErr: %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr != nil {
				var perr *clp.ParsingError
				if errors.As(err, &perr) {
					uerr := errors.Unwrap(err)
					if uerr != nil && uerr.Error() != tt.wantErr.Error() {
						t.Errorf("error: %v, wantErr: %v", uerr, tt.wantErr)
					}
				} else {
					t.Errorf("unexpected parsing error: %v", err)
				}
			}
		})
	}
}

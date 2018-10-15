package main

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/reconquest/karma-go"
)

type (
	fn  func()
	any interface{}
)

type (
	fnString   func(string)
	fnFlag     func(string, fn) *flag
	fnHandle   func(callback any) func(args ...any)
	fnAny      func(any)
	fnRequired func(*flag)
	fnCommand  func(string, fn)
)

var (
	Version     fnString
	Description fnString
	Flag        fnFlag
	Handle      fnHandle
	Default     fnAny
	Value       fnAny
	Command     fnCommand
	Required    fnRequired

	parentRequired    = []fnRequired{}
	parentVersion     = []fnString{}
	parentDescription = []fnString{}
	parentFlag        = []fnFlag{}
	parentHandle      = []fnHandle{}
	parentDefault     = []fnAny{}
	parentValue       = []fnAny{}
	parentCommand     = []fnCommand{}
)

type flag struct {
	name         string
	description  string
	defaultValue any
	value        any
	handler      handler
}

type command struct {
	name        string
	description string
	required    []any
	handler     handler

	flags    []*flag
	commands []command
}

type handler struct {
	parent        string
	callback      any
	setArgsCalled bool
	args          []any
}

var (
	globalVersion     string
	globalDescription string

	globalFlags    []*flag
	globalCommands []command

	globalHandler handler
	globalCalls   []fn
)

func setVersion(value string) {
	globalVersion = value
}

func setString(to *string) func(string) {
	return func(value string) {
		*to = value
	}
}

func setAny(to *any) func(any) {
	return func(value any) {
		*to = value
	}
}

func PrintVersion() {
	fmt.Println(globalVersion)
}

func PrintUsage() {
	fmt.Println(globalVersion)
	fmt.Println(globalDescription)
	fmt.Println("some usage")
}

func Cli(call fn) fn {
	Version = setVersion
	Description = setString(&globalDescription)

	Flag = addFlag(&globalFlags)
	Command = addCommand(&globalCommands)
	Handle = setHandle(&globalHandler)

	call()

	return func() {
		validate()
		//spew.Dump(globalFlags)
		//spew.Dump(globalCommands)
	}
}

func validate() {

}

func addFlag(to *[]*flag) fnFlag {
	return func(name string, callback fn) *flag {
		if to == nil {
			*to = []*flag{}
		}

		unit := newFlag(name, callback)

		*to = append(*to, unit)

		return unit
	}
}

func addCommand(to *[]command) func(string, fn) {
	return func(name string, callback fn) {
		if to == nil {
			*to = []command{}
		}

		*to = append(*to, newCommand(name, callback))
	}
}

func newFlag(name string, callback fn) *flag {
	result := flag{}
	result.name = name

	remember()

	Description = setString(&result.description)
	Default = setAny(&result.defaultValue)
	Value = setAny(&result.value)
	Handle = setHandle(&result.handler)

	callback()

	restore()

	return &result
}

func newCommand(name string, callback fn) command {
	result := command{}
	result.name = name
	result.handler.parent = `command "` + name + `"`

	remember()

	Description = setString(&result.description)
	Handle = setHandle(&result.handler)

	Required = func(value *flag) {
		result.required = append(result.required, value)
	}

	Flag = addFlag(&result.flags)

	callback()

	restore()

	err := validateHandler(&result.handler)
	if err != nil {
		panic(karma.Format(err, result.handler.parent))
	}

	return result
}

func validateHandler(handler *handler) error {
	if handler.callback == nil {
		return fmt.Errorf(
			"Command() section is declared " +
				"but method Handle() was not invoked in there",
		)
	}

	if !handler.setArgsCalled {
		return fmt.Errorf(
			"Handle() for %s was invoked, but no arguments "+
				"were specified for this handler",
			getFuncName(reflect.ValueOf(handler.callback)),
		)
	}

	return nil
}

func setHandle(handler *handler) fnHandle {
	return func(callback any) func(args ...any) {
		err := validateCallback(callback)
		if err != nil {
			if handler.parent != "" {
				panic(karma.Format(err, handler.parent))
			} else {
				panic(err)
			}
		}

		handler.callback = callback

		return func(args ...any) {
			err := validateCallbackArgs(callback, args)
			if err != nil {
				if handler.parent != "" {
					panic(karma.Format(err, handler.parent))
				} else {
					panic(err)
				}
			}

			handler.setArgsCalled = true
			handler.args = args
		}
	}
}

func validateCallback(callback any) error {
	kind := reflect.TypeOf(callback).Kind()

	if kind != reflect.Func {
		// TODO: we can extract line+number from stack
		return fmt.Errorf(
			"argument to Handle() must be a function, but got %T",
			callback,
		)
	}

	return nil
}

func validateCallbackArgs(callback any, args []any) error {
	value := reflect.ValueOf(callback)

	numIn := value.Type().NumIn()
	if numIn != len(args) {
		message := fmt.Sprintf(
			"unable to call %s: expected %d args but got %d",
			getFuncName(value),
			numIn, len(args),
		)

		return fmt.Errorf(
			"%s",
			message,
		)
	}

	return nil
}

func getFuncArgs(value reflect.Value) string {
	items := []string{}
	for i := 0; i < value.Type().NumIn(); i++ {
		items = append(items, value.Type().In(i).String())
	}

	return strings.Join(items, ", ")
}

func getArgsTypes(args []any) string {
	items := []string{}
	for i := 0; i < len(args); i++ {
		items = append(items, getFuncName(reflect.ValueOf(args[i])))
	}

	return strings.Join(items, ", ")
}

func getFuncName(value reflect.Value) string {
	name := runtime.FuncForPC(value.Pointer()).Name()
	if name == "" {
		name = value.Type().String()
	}

	return name
}

func remember() {
	parentRequired = append(parentRequired, Required)
	parentVersion = append(parentVersion, Version)
	parentDescription = append(parentDescription, Description)
	parentFlag = append(parentFlag, Flag)
	parentHandle = append(parentHandle, Handle)
	parentDefault = append(parentDefault, Default)
	parentValue = append(parentValue, Value)
	parentCommand = append(parentCommand, Command)
}

func restore() {
	size := len(parentRequired)

	Required = parentRequired[size-1]
	Version = parentVersion[size-1]
	Description = parentDescription[size-1]
	Flag = parentFlag[size-1]
	Handle = parentHandle[size-1]
	Default = parentDefault[size-1]
	Value = parentValue[size-1]
	Command = parentCommand[size-1]

	parentRequired = parentRequired[:size-1]
	parentVersion = parentVersion[:size-1]
	parentDescription = parentDescription[:size-1]
	parentFlag = parentFlag[:size-1]
	parentHandle = parentHandle[:size-1]
	parentDefault = parentDefault[:size-1]
	parentValue = parentValue[:size-1]
	parentCommand = parentCommand[:size-1]
}

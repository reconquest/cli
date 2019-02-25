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
	Name        fnString
	Description fnString
	Version     fnString
	Flag        fnFlag
	Option      fnFlag
	Argument    fnFlag
	Handle      fnHandle
	Default     fnAny
	Value       fnAny
	Command     fnCommand
	Required    fnRequired

	stackName        = []fnString{}
	stackDescription = []fnString{}
	stackVersion     = []fnString{}
	stackFlag        = []fnFlag{}
	stackOption      = []fnFlag{}
	stackArgument    = []fnFlag{}
	stackHandle      = []fnHandle{}
	stackDefault     = []fnAny{}
	stackValue       = []fnAny{}
	stackCommand     = []fnCommand{}
	stackRequired    = []fnRequired{}
)

type flag struct {
	name         string
	option       bool
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
	globalName        string

	global = command{}
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
	fmt.Printf(
		"%s - %s\n\n%s\n\nUsage:\n%s\n\nOptions:\n%s\n",
		globalName,
		globalVersion,
		globalDescription,
		getUsage(),
		getOptions(global),
	)
}

func getUsage() string {
	return "\tusage"
}

func getOptions(cmd command) string {
	lines := []string{}
	for _, flag := range cmd.flags {
		var line string
		if flag.option {
			line = flag.name + " <value> "

			withoutDesc := len(line)

			line += flag.description

			if flag.defaultValue != nil {
				line += fmt.Sprintf(
					"\n%s[default: %v]",
					strings.Repeat(" ", withoutDesc),
					flag.defaultValue,
				)
			}
		} else {
			line = flag.name + " " + flag.description
		}

		lines = append(lines, line)
	}

	for _, command := range cmd.commands {
		line := fmt.Sprintf("%s %s", command.name, command.description)
		lines = append(lines, line)
		lines = append(lines, getOptions(command))
	}

	return strings.Join(lines, "\n")
}

//func indent(lines []string, shift string) []string {
//    for key, value := range lines {
//        if strings.Contains(value, "\n") {
//            values := strings.Split(value, "\n")
//            values = indent(values, shift)
//        }
//        lines[key] = shift + value
//    }
//    return lines
//}

func Cli(call fn) fn {
	Version = setVersion
	Description = setString(&globalDescription)
	Name = setString(&globalName)

	Flag = addFlag(&global.flags, false)
	Option = addFlag(&global.flags, true)
	Command = addCommand(&global.commands)
	Handle = setHandle(&global.handler)

	call()

	return func() {
		PrintUsage()
		//spew.Dump(globalFlags)
		//spew.Dump(globalCommands)
	}
}

func addFlag(to *[]*flag, option bool) fnFlag {
	return func(name string, callback fn) *flag {
		if to == nil {
			*to = []*flag{}
		}

		unit := newFlag(name, callback, option)

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

func newFlag(name string, callback fn, option bool) *flag {
	result := flag{}
	result.name = name
	result.option = option

	pushStack()

	Description = setString(&result.description)
	Default = setAny(&result.defaultValue)
	Value = setAny(&result.value)
	Handle = setHandle(&result.handler)

	callback()

	popStack()

	return &result
}

func newCommand(name string, callback fn) command {
	result := command{}
	result.name = name
	result.handler.parent = `command "` + name + `"`

	pushStack()

	Description = setString(&result.description)
	Handle = setHandle(&result.handler)

	Required = func(value *flag) {
		result.required = append(result.required, value)
	}

	Flag = addFlag(&result.flags, false)
	Option = addFlag(&result.flags, true)

	callback()

	popStack()

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

func pushStack() {
	stackName = append(stackName, Name)
	stackDescription = append(stackDescription, Description)
	stackVersion = append(stackVersion, Version)
	stackFlag = append(stackFlag, Flag)
	stackHandle = append(stackHandle, Handle)
	stackDefault = append(stackDefault, Default)
	stackValue = append(stackValue, Value)
	stackCommand = append(stackCommand, Command)
	stackRequired = append(stackRequired, Required)
}

func popStack() {
	size := len(stackName)

	Name = stackName[size-1]
	Description = stackDescription[size-1]
	Version = stackVersion[size-1]
	Flag = stackFlag[size-1]
	Handle = stackHandle[size-1]
	Default = stackDefault[size-1]
	Value = stackValue[size-1]
	Command = stackCommand[size-1]
	Required = stackRequired[size-1]

	stackName = stackName[:size-1]
	stackDescription = stackDescription[:size-1]
	stackVersion = stackVersion[:size-1]
	stackFlag = stackFlag[:size-1]
	stackHandle = stackHandle[:size-1]
	stackDefault = stackDefault[:size-1]
	stackValue = stackValue[:size-1]
	stackCommand = stackCommand[:size-1]
	stackRequired = stackRequired[:size-1]
}

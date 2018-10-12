package main

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
)

type (
	fn  func()
	any interface{}
)

var (
	Version     func(string)
	Description func(string)
	Flag        func(string, fn)
	Call        func(fn)
	Handle      func(fn)
	Default     func(any)
	Value       func(any)
	Command     func(string, fn)
	Required    func(any)

	parentRequired    = Required
	parentVersion     = Version
	parentDescription = Description
	parentFlag        = Flag
	parentCall        = Call
	parentHandle      = Handle
	parentDefault     = Default
	parentValue       = Value
	parentCommand     = Command
)

type flag struct {
	name         string
	description  string
	defaultValue any
	value        any
	call         fn
	handle       fn
}

type command struct {
	name        string
	description string
	required    []any
	handle      fn
	call        fn

	flags    []flag
	commands []command
}

var (
	globalVersion     string
	globalDescription string

	globalFlags    []flag
	globalCommands []command

	globalHandle fn
	globalCalls  []fn
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

func setFn(to *fn) func(fn) {
	return func(value fn) {
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
	Handle = setFn(&globalHandle)

	call()

	return func() {
		spew.Dump(globalFlags)
		spew.Dump(globalCommands)
	}
}

func addFlag(to *[]flag) func(string, fn) {
	return func(name string, callback fn) {
		if to == nil {
			*to = []flag{}
		}

		*to = append(*to, newFlag(name, callback))
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

func newFlag(name string, callback fn) flag {
	result := flag{}
	result.name = name

	remember()

	Description = setString(&result.description)
	Default = setAny(&result.defaultValue)
	Value = setAny(&result.value)
	Call = setFn(&result.call)
	Handle = setFn(&result.handle)

	callback()

	restore()

	return result
}

func newCommand(name string, callback fn) command {
	result := command{}
	result.name = name

	remember()

	Description = setString(&result.description)
	Call = setFn(&result.call)
	Handle = setFn(&result.handle)

	Required = func(value any) {
		if reflect.ValueOf(value).Kind() != reflect.Ptr {
			panic("Required() accepts only pointer to variable")
		}

		result.required = append(result.required, value)
	}

	Flag = addFlag(&result.flags)

	callback()

	restore()

	return result
}

func remember() {
	parentRequired = Required
	parentVersion = Version
	parentDescription = Description
	parentFlag = Flag
	parentCall = Call
	parentHandle = Handle
	parentDefault = Default
	parentValue = Value
	parentCommand = Command
}

func restore() {
	Required = parentRequired
	Version = parentVersion
	Description = parentDescription
	Flag = parentFlag
	Call = parentCall
	Handle = parentHandle
	Default = parentDefault
	Value = parentValue
	Command = parentCommand
}

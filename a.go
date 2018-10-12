package main

import (
	"log"

	"github.com/kovetskiy/toml"
)

func main() {
	Cli(func() {
		Version("1.0")
		Description("Example of cli with DSL")

		Flag("-v --version", func() {
			Description("Show version of program")

			Handle(PrintVersion)
		})

		Flag("-h --help", func() {
			Description("Show program help")

			Handle(PrintUsage)
		})

		var path string
		var config Config
		Flag("-c --config", func() {
			Description("Read specified configuration file")

			Default("example.conf")

			Value(&path)

			Call(func() {
				_, err := toml.DecodeFile(path, &config)
				if err != nil {
					log.Fatalf("unable to load toml config: %s", err)
				}
			})
		})

		var program []string
		Flag("<program>", func() {
			Description("Program name or path to start/stop")

			Value(&program)
		})

		Command("start", func() {
			Description("Start specified program")

			Required(&program)

			Handle(func() {
				handleStart(
					config,
					program,
				)
			})
		})

		Command("stop", func() {
			var signal int
			Flag("-s --signal", func() {
				Description("Signal to send when killing process")

				Default(9)

				Value(&signal)
			})

			Description(
				"Stop specified program or stop all " +
					"programs if no <program> specified",
			)

			Required(&signal)

			Handle(func() {
				handleStop(
					config,
					signal,
					program,
				)
			})
		})
	})()
}

type (
	Config struct {
		Foo string
	}
)

func handleVersion() {

}

func handleHelp() {

}

func handleStart(config Config, program []string) {

}

func handleStop(config Config, signal int, program []string) {

}

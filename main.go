package main

func main() {
	Cli(func() {
		Version("1.0")
		Description("Example of cli with DSL")

		Flag("-v --version", func() {
			Description("Show version of program")

			Handle(PrintVersion)()
		})

		Flag("-h --help", func() {
			Description("Show program help")

			Handle(PrintUsage)()
		})

		config := Flag("-c --config", func() {
			Description("Read specified configuration file")

			Default("example.conf")
		})

		program := Flag("<program>", func() {
			Description("Program name or path to start/stop")
		})

		Command("start", func() {
			Description("Start specified program")

			Required(program)

			Handle(handleStart)(config, program)
		})

		Command("stop", func() {
			Description(
				"Stop specified program or stop all " +
					"programs if no <program> specified",
			)

			signal := Flag("-s --signal", func() {
				Description("Signal to send when killing process")

				Default(9)
			})

			Required(program)

			_ = signal
			Handle(handleStop)(config, program, signal)
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

package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/teh-hippo/foxess-exporter/foxess"
)

var foxessAPI foxess.FoxessAPI

type Runner interface {
	Register(parser *flags.Parser)
}

func main() {
	parser := flags.NewParser(&foxessAPI, flags.Default)
	commands := []Runner{
		&APIUsageCommand{},
		&DevicesCommand{},
		&HistoryCommand{},
		&RealTimeCommand{},
		&ServeCommand{},
		&VariablesCommand{},
	}

	for _, command := range commands {
		command.Register(parser)
	}

	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}

			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}

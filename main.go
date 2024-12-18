package main

import (
	"errors"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/teh-hippo/foxess-exporter/foxess"
)

const (
	FormatTable = "table"
	FormatJSON  = "json"
)

var foxessAPI foxess.Config

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
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) {
			if flagsErr.Type == flags.ErrCommandRequired {
				parser.WriteHelp(os.Stdout)
			} else if flagsErr.Type == flags.ErrHelp {
				return
			}
		}

		os.Exit(1)
	}
}

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

type Runner interface {
	Register(parser *flags.Parser, config *foxess.Config)
}

func main() {
	foxessAPI := foxess.Config{} //nolint:exhaustruct
	parser := flags.NewParser(&foxessAPI, flags.Default)
	commands := []Runner{
		&APIUsageCommand{},  //nolint:exhaustruct
		&DevicesCommand{},   //nolint:exhaustruct
		&HistoryCommand{},   //nolint:exhaustruct
		&RealTimeCommand{},  //nolint:exhaustruct
		&ServeCommand{},     //nolint:exhaustruct
		&VariablesCommand{}, //nolint:exhaustruct
	}

	for _, command := range commands {
		command.Register(parser, &foxessAPI)
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

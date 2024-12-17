package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/teh-hippo/foxess-exporter/foxess"
)

var (
	foxessApi foxess.FoxessApi
	parser    = flags.NewParser(&foxessApi, flags.Default)
)

func main() {
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

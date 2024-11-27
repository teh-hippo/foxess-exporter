package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

var options Options

type Options struct {
	ApiKey string `short:"k" long:"api-key" description:"FoxESS API Key" required:"true"`
	Debug  bool   `short:"d" long:"debug" description:"Enable debug output"`
}

var parser = flags.NewParser(&options, flags.Default)

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

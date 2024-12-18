package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type RealTimeCommand struct {
	Inverters []string `short:"i" long:"inverter" description:"Inverter serial numbers." required:"true"`
	Variables []string `short:"p" long:"variable" description:"Variables to retrieve"`
	Format    string   `short:"o" long:"output"   description:"Output format"            default:"table" choices:"table,json"`
	config    *foxess.Config
}

func (x *RealTimeCommand) Register(parser *flags.Parser, config *foxess.Config) {
	if _, err := parser.AddCommand("realtime", "Get real-time data", "Get the current real-time data for an inverter.", x); err != nil {
		panic(err)
	}

	x.config = config
}

func (x *RealTimeCommand) Execute(_ []string) error {
	data, err := x.config.GetRealTimeData(x.Inverters, x.Variables)
	if err != nil {
		return fmt.Errorf("unable to retrieve real-time data from FoxESS: %w", err)
	}

	switch x.Format {
	case FormatTable:
		tbl := table.New("Device", "Time", "Variable", "Name", "Unit", "Value")

		for _, item := range data {
			for _, variable := range item.Variables {
				tbl.AddRow(item.DeviceSN, item.Time, variable.Variable, variable.Name, variable.Unit, variable.Value.Number)
			}
		}

		tbl.Print()

		return nil
	case FormatJSON:
		err := util.JSONToStdOut(data)
		if err != nil {
			return fmt.Errorf("unable to output JSON: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedFormat, x.Format)
	}
}

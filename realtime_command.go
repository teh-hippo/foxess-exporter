package main

import (
	"fmt"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/util"
)

type RealTimeCommand struct {
	Inverters []string `short:"i" long:"inverter" description:"Inverter serial numbers." required:"true"`
	Variables []string `short:"p" long:"variable" description:"Variables to retrieve" required:"false"`
	Format    string   `short:"o" long:"output" description:"Output format" default:"table" choices:"table,json" required:"false"`
}

var realTimeCommand RealTimeCommand

func init() {
	if _, err := parser.AddCommand("realtime", "Get real-time data", "Get the current real-time data for an inverter.", &realTimeCommand); err != nil {
		panic(err)
	}
}

func (x *RealTimeCommand) Execute(args []string) error {
	data, err := foxessApi.GetRealTimeData(x.Inverters, x.Variables)
	if err != nil {
		return err
	}

	switch x.Format {
	case "table":
		tbl := table.New("Device", "Time", "Variable", "Name", "Unit", "Value")
		for _, item := range data {
			for _, variable := range item.Variables {
				tbl.AddRow(item.DeviceSN, item.Time, variable.Variable, variable.Name, variable.Unit, variable.Value.Number)
			}
		}
		tbl.Print()
		return nil
	case "json":
		return util.JsonToStdOut(data)
	default:
		return fmt.Errorf("unsupported output format: %s", x.Format)
	}
}

package main

import (
	"fmt"
	"maps"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/util"
)

type VariablesCommand struct {
	GridOnly bool   `short:"g" long:"grid-only" description:"Only show variables related to a grid tied inverter"`
	Format   string `short:"o" long:"output" description:"Output format" default:"table" choices:"table,json" required:"false"`
}

var variablesCommand VariablesCommand

func init() {
	if _, err := parser.AddCommand("variables", "List of supported variables", "Retrieves from FoxESS all variables that can be requested for history.", &variablesCommand); err != nil {
		panic(err)
	}
}

func (x *VariablesCommand) Execute(args []string) error {
	variables, err := foxessApi.GetVariables(x.GridOnly)
	if err != nil {
		return nil
	}

	switch x.Format {
	case "table":
		tbl := table.New("Variable Name", "Unit", "Grid Tied", "Energy Storage")
		for _, variable := range *variables {
			for key := range maps.Keys(variable) {
				item := variable[key]
				tbl.AddRow(key, item.Unit, item.GridTiedInverter, item.EnergyStorageInverter)
			}
		}
		tbl.Print()
		return nil
	case "json":
		return util.JsonToStdOut(variables)
	default:
		return fmt.Errorf("unsupported output format: %s", x.Format)
	}
}

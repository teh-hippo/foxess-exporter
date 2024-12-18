package main

import (
	"fmt"
	"maps"

	"github.com/jessevdk/go-flags"
	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type VariablesCommand struct {
	GridOnly bool   `short:"g" long:"grid-only" description:"Only show variables related to a grid tied inverter"`
	Format   string `short:"o" long:"output"    description:"Output format"                                       default:"table" choices:"table,json"`
	config   *foxess.Config
}

func (x *VariablesCommand) Register(parser *flags.Parser, config *foxess.Config) {
	if _, err := parser.AddCommand("variables", "List of supported variables", "Retrieve FoxESS variables for use with history or real-time data.", x); err != nil {
		panic(err)
	}

	x.config = config
}

func (x *VariablesCommand) Execute(_ []string) error {
	variables, err := x.config.GetVariables(x.GridOnly)
	if err != nil {
		return fmt.Errorf("failed to retrieve variables: %w", err)
	}

	switch x.Format {
	case FormatTable:
		tbl := table.New("Variable Name", "Unit", "Grid Tied", "Energy Storage")

		for _, variable := range *variables {
			for key := range maps.Keys(variable) {
				item := variable[key]
				tbl.AddRow(key, item.Unit, item.GridTiedInverter, item.EnergyStorageInverter)
			}
		}

		tbl.Print()

		return nil
	case FormatJSON:
		err := util.JSONToStdOut(variables)
		if err != nil {
			return fmt.Errorf("failed to output variables: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", x.Format)
	}
}

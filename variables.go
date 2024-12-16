package main

import (
	"fmt"
	"maps"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type Variable struct {
	Unit                  string `json:"unit"`
	GridTiedInverter      bool   `json:"Grid-tied inverter"`
	EnergyStorageInverter bool   `json:"Energy-storage inverter"`
}

// Define the structure for the response
type VariablesResponse struct {
	ErrorNumber int                   `json:"errno"`
	Message     string                `json:"msg"`
	Result      []map[string]Variable `json:"result"`
}

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
	variables, err := getVariables(x.GridOnly)
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

func getVariables(gridOnly bool) (*[]map[string]Variable, error) {
	response := &VariablesResponse{}
	if err := foxessApi.NewRequest("GET", "/op/v0/device/variable/get", nil, response); err != nil {
		return nil, err
	} else if err = foxess.IsError(response.ErrorNumber, response.Message); err != nil {
		return nil, err
	} else if !gridOnly {
		return &response.Result, nil
	} else {
		gridOnlyVariables := make([]map[string]Variable, 0)
		for _, variable := range response.Result {
			for key := range maps.Keys(variable) {
				item := variable[key]
				if item.GridTiedInverter {
					gridOnlyVariables = append(gridOnlyVariables, variable)
				}
			}
		}
		return &gridOnlyVariables, nil
	}
}

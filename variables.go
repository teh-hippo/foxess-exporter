package main

import (
	"maps"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
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
	GridOnly bool `short:"g" long:"grid-only" description:"Only show variables related to a grid tied inverter"`
}

var variablesCommand VariablesCommand

func init() {
	parser.AddCommand("variables", "List of supported variables", "Retrieves from FoxESS all variables that can be requested for history.", &variablesCommand)
}

func (vc *VariablesCommand) Execute(args []string) error {
	response := &VariablesResponse{}
	err := foxess.NewRequest(options.ApiKey, "GET", "/op/v0/device/variable/get", nil, response, options.Debug)
	if err != nil {
		return err
	}

	if err = foxess.IsError(response.ErrorNumber, response.Message); err != nil {
		return err
	}

	tbl := table.New("Variable Name", "Unit", "Grid Tied", "Energy Storage")
	for _, variable := range response.Result {
		for key := range maps.Keys(variable) {
			item := variable[key]
			if vc.GridOnly && !item.GridTiedInverter {
				continue
			}
			tbl.AddRow(key, item.Unit, item.GridTiedInverter, item.EnergyStorageInverter)
		}
	}
	tbl.Print()
	return nil
}

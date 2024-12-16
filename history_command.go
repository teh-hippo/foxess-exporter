package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type HistoryCommand struct {
	Inverter  string   `short:"i" long:"inverter" description:"Inverter serial number" required:"true"`
	Begin     int64    `short:"b" long:"begin" description:"Begin time for request in milliseconds"`
	End       int64    `short:"e" long:"end" description:"End time for request in milliseconds"`
	Variables []string `short:"V" long:"variable" description:"Variables to retrieve" required:"true"`
	Format    string   `short:"o" long:"output" description:"Output format" default:"table" choices:"table,json" required:"false"`
}

var historyCommand HistoryCommand

func init() {
	if _, err := parser.AddCommand("history", "Get the history", "Get the history of a variable", &historyCommand); err != nil {
		panic(err)
	}
}

func (x *HistoryCommand) Execute(args []string) error {
	if x.Begin == 0 || x.End == 0 {
		if x.Begin != x.End {
			return errors.New("provide both begin and end, or neither")
		}

		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		yesterday := startOfDay.Add(-24 * time.Hour)
		x.Begin = yesterday.UnixMilli()
		x.End = startOfDay.UnixMilli()
	}

	response, err := foxessApi.GetVariableHistory(x.Inverter, x.Begin, x.End, x.Variables)
	if err != nil {
		return fmt.Errorf("failed to retrieve history for %s from %d -> %d: %w", x.Inverter, x.Begin, x.End, err)
	}

	if err = x.writeResult(response); err != nil {
		return fmt.Errorf("failed to output result: %w", err)
	}

	return nil
}

func (x *HistoryCommand) writeResult(history []foxess.VariableHistory) error {
	switch x.Format {
	case "table":
		tbl := table.New("Variable", "Name", "Unit", "Time", "Value")
		for _, variable := range history {
			for _, point := range variable.DataPoints {
				tbl.AddRow(variable.Variable, variable.Name, variable.Unit, point.Time, point.Value)
			}
			tbl.Print()
		}
	case "json":
		return util.JsonToStdOut(history)
	}
	return nil
}

package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type HistoryRequest struct {
	SerialNumber string   `json:"sn"`
	Begin        int64    `json:"begin"`
	End          int64    `json:"end"`
	Variables    []string `json:"variables"`
}

type HistoryCommand struct {
	Inverter  string   `short:"i" long:"inverter" description:"Inverter serial number" required:"true"`
	Begin     int64    `short:"b" long:"begin" description:"Begin time for request in milliseconds"`
	End       int64    `short:"e" long:"end" description:"End time for request in milliseconds"`
	Variables []string `short:"V" long:"variable" description:"Variables to retrieve" required:"true"`
	Format    string   `short:"o" long:"output" description:"Output format" default:"table" choices:"table,json" required:"false"`
}

type DataPoint struct {
	Time  CustomTime `json:"time"`
	Value float64    `json:"value"`
}
type CustomTime struct {
	time.Time
}

func (t *CustomTime) UnmarshalJSON(b []byte) (err error) {
	value := string(b)
	const format string = `"2006-01-02 15:04:05 MST-0700"`
	date, err := time.Parse(format, value)
	if err != nil {
		return fmt.Errorf("failed to parse '%s' as date of format '%s': %w", value, format, err)
	}
	t.Time = date
	return
}

type HistoryResponse struct {
	ErrorNumber int    `json:"errno"`
	Message     string `json:"msg"`
	Result      []struct {
		Variables []VariableHistory `json:"datas"`
		DeviceSN  string            `json:"deviceSN"`
	} `json:"result"`
}

type VariableHistory struct {
	Unit       string      `json:"unit"`
	DataPoints []DataPoint `json:"data"`
	Name       string      `json:"name"`
	Variable   string      `json:"variable"`
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

	response, err := GetVariableHistory(x.Inverter, x.Begin, x.End, x.Variables)
	if err != nil {
		return fmt.Errorf("failed to retrieve history for %s from %d -> %d: %w", x.Inverter, x.Begin, x.End, err)
	}

	if err = x.writeResult(response); err != nil {
		return fmt.Errorf("failed to output result: %w", err)
	}

	return nil
}

func GetVariableHistory(inverter string, begin, end int64, variables []string) ([]VariableHistory, error) {
	request := &HistoryRequest{
		Begin:        begin,
		End:          end,
		SerialNumber: inverter,
	}

	if variables != nil {
		request.Variables = variables
	}
	response := &HistoryResponse{}
	err := foxess.NewRequest(options.ApiKey, "POST", "/op/v0/device/history/query", request, response, options.Debug)
	if err != nil {
		return nil, err
	} else if err = foxess.IsError(response.ErrorNumber, response.Message); err != nil {
		return nil, err
	}

	result := make([]VariableHistory, len(response.Result))
	for i, r := range response.Result {
		result[i] = r.Variables[i]
	}
	return result, nil
}

func (x *HistoryCommand) writeResult(history []VariableHistory) error {
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

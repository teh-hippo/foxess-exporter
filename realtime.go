package main

import (
	"encoding/json"
	"fmt"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type RealTimeRequest struct {
	SerialNumbers []string `json:"sns"`
	Variables     []string `json:"variables"`
}

type RealTimeCommand struct {
	Inverters []string `short:"i" long:"inverter" description:"Inverter serial numbers." required:"true"`
	Variables []string `short:"p" long:"variable" description:"Variables to retrieve" required:"false"`
	Format    string   `short:"o" long:"output" description:"Output format" default:"table" choices:"table,json" required:"false"`
}

type NumberAsNil struct {
	Number float64
}

func (t *NumberAsNil) UnmarshalJSON(b []byte) (err error) {
	if len(b) >= 2 && b[0] == '"' {
		b = b[1 : len(b)-1]
		if len(b) == 0 {
			return nil
		}
	}

	if err = json.Unmarshal(b, &t.Number); err != nil {
		return fmt.Errorf("failed to parse '%s': %w", b, err)
	}

	return nil
}

type RealTimeResponse struct {
	ErrorNumber int            `json:"errno"`
	Message     string         `json:"msg"`
	Result      []RealTimeData `json:"result"`
}

type RealTimeData struct {
	Variables []struct {
		Variable string      `json:"variable"`
		Unit     string      `json:"unit"`
		Name     string      `json:"name"`
		Value    NumberAsNil `json:"value"`
	} `json:"datas"`
	DeviceSN string     `json:"deviceSN"`
	Time     CustomTime `json:"time"`
}

var realTimeCommand RealTimeCommand

func init() {
	if _, err := parser.AddCommand("realtime", "Get real-time data", "Get the current real-time data for an inverter.", &realTimeCommand); err != nil {
		panic(err)
	}
}

func (x *RealTimeCommand) Execute(args []string) error {
	data, err := GetRealTimeData(x.Inverters, x.Variables)
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

func GetRealTimeData(inverters []string, variables []string) ([]RealTimeData, error) {
	request := &RealTimeRequest{
		SerialNumbers: inverters,
	}
	if variables != nil {
		request.Variables = variables
	}
	response := &RealTimeResponse{}
	if err := foxess.NewRequest(options.ApiKey, "POST", "/op/v1/device/real/query", request, response, options.Debug); err != nil {
		return nil, err
	} else if err = foxess.IsError(response.ErrorNumber, response.Message); err != nil {
		return nil, err
	} else {
		return response.Result, nil
	}
}

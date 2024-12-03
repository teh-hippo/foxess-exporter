package main

import (
	"encoding/json"
	"fmt"

	"github.com/teh-hippo/foxess-exporter/foxess"
)

type RealTimeRequest struct {
	SerialNumbers []string `json:"sns"`
	Variables     []string `json:"variables"`
}

type RealTimeCommand struct {
	Inverter  string   `short:"i" long:"inverter" description:"Inverter serial number" required:"true"`
	Variables []string `short:"v" long:"variables" description:"Variables to retrieve" required:"false"`
}

type NumberAsNil struct {
	Number float64
}

func (t *NumberAsNil) UnmarshalJSON(b []byte) (err error) {
	val := string(b)
	if len(b) >= 2 && val[0] == '"' {
		b = b[1 : len(b)-1]
		if len(b) == 0 {
			return nil
		}
	}
	err = json.Unmarshal(b, &t.Number)
	if err != nil {
		fmt.Print(val)
		return err
	}
	return err
}

type RealTimeResponse struct {
	ErrorNumber int    `json:"errno"`
	Message     string `json:"msg"`
	Result      []struct {
		Variables []struct {
			Variable string      `json:"variable"`
			Unit     string      `json:"unit"`
			Name     string      `json:"name"`
			Value    NumberAsNil `json:"value"`
		} `json:"datas"`
		DeviceSN string     `json:"deviceSN"`
		Time     CustomTime `json:"time"`
	} `json:"result"`
}

var realTimeCommand RealTimeCommand

func init() {
	parser.AddCommand("realtime", "Get real-time data", "Get the current real-time data for an inverter.", &realTimeCommand)
}

func (x *RealTimeCommand) Execute(args []string) error {
	request := &RealTimeRequest{
		SerialNumbers: []string{x.Inverter},
	}
	if x.Variables != nil {
		request.Variables = x.Variables
	}
	response := &RealTimeResponse{}
	err := foxess.NewRequest(options.ApiKey, "POST", "/op/v1/device/real/query", request, response, options.Debug)
	if err != nil {
		return err
	}

	if err = foxess.IsError(response.ErrorNumber, response.Message); err != nil {
		return err
	}

	return nil
}

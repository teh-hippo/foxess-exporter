package main

import (
	"encoding/json"
	"fmt"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
)

type ApiUsageCommand struct {
}

type AccessCountResponse struct {
	ErrorNumber int `json:"errno"`
	Result      struct {
		Total     json.Number `json:"total"`
		Remaining json.Number `json:"remaining"`
	} `json:"result"`
}

var apiUsageCommand ApiUsageCommand

func init() {
	parser.AddCommand("api-usage", "Show FoxESS API usage", "Show FoxESS API usage", &apiUsageCommand)
}

func (x *ApiUsageCommand) Execute(args []string) error {
	response := &AccessCountResponse{}
	err := foxess.NewRequest(options.ApiKey, "GET", "/op/v0/user/getAccessCount", nil, response, options.Debug)
	if err != nil {
		return err
	}

	if err = foxess.IsError(response.ErrorNumber, ""); err != nil {
		return err
	}

	tbl := table.New("Total", "Remaining", "Used")
	var total, remaining float64
	if total, err = response.Result.Total.Float64(); err != nil {
		return err
	}
	if remaining, err = response.Result.Remaining.Float64(); err != nil {
		return err
	}
	percentage := (total - remaining) / total * 100
	tbl.AddRow(response.Result.Total, response.Result.Remaining, fmt.Sprintf("%.2f%%", percentage))
	tbl.Print()

	return nil
}

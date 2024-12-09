package main

import (
	"encoding/json"
	"fmt"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type ApiUsageCommand struct {
	Format string `short:"o" long:"output" description:"Output format" default:"table" choices:"table,json" required:"false"`
}

type AccessCountResponse struct {
	ErrorNumber int `json:"errno"`
	Result      struct {
		Total     json.Number `json:"total"`
		Remaining json.Number `json:"remaining"`
	} `json:"result"`
}

type ApiUsage struct {
	Total          float64
	Remaining      float64
	PercentageUsed float64
}

var apiUsageCommand ApiUsageCommand

func init() {
	parser.AddCommand("api-usage", "Show FoxESS API usage", "Show FoxESS API usage", &apiUsageCommand)
}

func (x *ApiUsageCommand) Execute(args []string) error {
	if apiUsage, err := GetApiUsage(); err != nil {
		return err
	} else if x.Format == "table" {
		tbl := table.New("Total", "Remaining", "Used")
		tbl.AddRow(apiUsage.Total, apiUsage.Remaining, fmt.Sprintf("%.2f%%", apiUsage.PercentageUsed))
		tbl.Print()
		return nil
	} else if x.Format == "json" {
		return util.JsonToStdOut(apiUsage)
	} else {
		return fmt.Errorf("unsupported output format: %s", x.Format)
	}
}

func GetApiUsage() (*ApiUsage, error) {
	response := &AccessCountResponse{}
	err := foxess.NewRequest(options.ApiKey, "GET", "/op/v0/user/getAccessCount", nil, response, options.Debug)
	if err != nil {
		return nil, err
	}

	if err = foxess.IsError(response.ErrorNumber, ""); err != nil {
		return nil, err
	}

	var total, remaining float64
	if total, err = response.Result.Total.Float64(); err != nil {
		return nil, err
	} else if remaining, err = response.Result.Remaining.Float64(); err != nil {
		return nil, err
	}

	percentage := (total - remaining) / total * 100
	return &ApiUsage{
		Total:          total,
		Remaining:      remaining,
		PercentageUsed: percentage,
	}, nil
}

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
	if _, err := parser.AddCommand("api-usage", "Show FoxESS API usage", "Show FoxESS API usage", &apiUsageCommand); err != nil {
		panic(err)
	}
}

func (x *ApiUsageCommand) Execute(args []string) error {
	apiUsage, err := GetApiUsage()
	if err != nil {
		return fmt.Errorf("failed to retrieve the latest api usage: %w", err)
	}

	switch x.Format {
	case "table":
		tbl := table.New("Total", "Remaining", "Used")
		tbl.AddRow(apiUsage.Total, apiUsage.Remaining, fmt.Sprintf("%.2f%%", apiUsage.PercentageUsed))
		tbl.Print()
		return nil
	case "json":
		return util.JsonToStdOut(apiUsage)
	default:
		return fmt.Errorf("unsupported output format: %s", x.Format)
	}
}

func GetApiUsage() (*ApiUsage, error) {
	response := &AccessCountResponse{}
	err := foxessApi.NewRequest("GET", "/op/v0/user/getAccessCount", nil, response)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest api usage: %w", err)
	}

	if err = foxess.IsError(response.ErrorNumber, ""); err != nil {
		return nil, err
	}

	total, err := response.Result.Total.Float64()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to float '%v': %w", response.Result.Total, err)
	}

	remaining, err := response.Result.Remaining.Float64()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to float '%v': %w", response.Result.Remaining, err)
	}

	percentage := (total - remaining) / total * 100
	return &ApiUsage{
		Total:          total,
		Remaining:      remaining,
		PercentageUsed: percentage,
	}, nil
}

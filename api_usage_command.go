package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/util"
)

type APIUsageCommand struct {
	Format string `short:"o" long:"output" description:"Output format" default:"table" choices:"table,json" required:"false"`
}

func (x *APIUsageCommand) Register(parser *flags.Parser) {
	if _, err := parser.AddCommand("api-usage", "Show FoxESS API usage", "Show FoxESS API usage", x); err != nil {
		panic(err)
	}
}

func (x *APIUsageCommand) Execute(_ []string) error {
	apiUsage, err := foxessAPI.GetAPIUsage()
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

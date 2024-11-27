package main

import (
	"fmt"
	"time"

	"github.com/teh-hippo/foxess-exporter/foxess"
)

type HistoryRequest struct {
	SerialNumber string `json:"sn"`
	Begin        int64  `json:"begin"`
	End          int64  `json:"end"`
}

type HistoryCommand struct {
	Inverter string `short:"i" long:"inverter" description:"Inverter serial number"`
	Begin    int64  `short:"b" long:"begin" description:"Begin time for request in milliseconds" required:"false"`
	End      int64  `short:"e" long:"end" description:"End time for request in milliseconds" required:"false"`
}

var historyCommand HistoryCommand

func init() {
	parser.AddCommand("history", "Get the history", "Get the history of a variable", &historyCommand)
}

func (x *HistoryCommand) Execute(args []string) error {
	if len(x.Inverter) == 0 {
		return fmt.Errorf("missing inverter serial number")
	}

	if x.Begin == 0 || x.End == 0 {
		if x.Begin != x.End {
			return fmt.Errorf("when using begin/end, must provided both values")
		}

		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		yesterday := startOfDay.Add(-24 * time.Hour)
		x.Begin = yesterday.UnixMilli()
		x.End = startOfDay.UnixMilli()
	}
	request := &HistoryRequest{
		Begin:        x.Begin,
		End:          x.End,
		SerialNumber: x.Inverter,
	}
	return foxess.NewRequest(options.ApiKey, "POST", "/op/v0/device/history/query", request, nil, options.Debug)
}

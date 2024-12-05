package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/teh-hippo/foxess-exporter/foxess"
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
	Variables []string `short:"q" long:"variable" description:"Variables to retrieve" required:"true"`
}

type DataPoint struct {
	Time  CustomTime `json:"time"`
	Value float64    `json:"value"`
}
type CustomTime struct {
	time.Time
}

func (t *CustomTime) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02 15:04:05 MST-0700"`, string(b))
	if err != nil {
		return err
	}
	t.Time = date
	return
}

type HistoryResponse struct {
	ErrorNumber int    `json:"errno"`
	Message     string `json:"msg"`
	Result      []struct {
		Variables []struct {
			Unit       string      `json:"unit"`
			DataPoints []DataPoint `json:"data"`
			Name       string      `json:"name"`
			Variable   string      `json:"variable"`
		} `json:"datas"`
		DeviceSN string `json:"deviceSN"`
	} `json:"result"`
}

var historyCommand HistoryCommand

func init() {
	parser.AddCommand("history", "Get the history", "Get the history of a variable", &historyCommand)
}

func (x *HistoryCommand) Execute(args []string) error {
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
	if x.Variables != nil {
		request.Variables = x.Variables
	}
	response := &HistoryResponse{}
	err := foxess.NewRequest(options.ApiKey, "POST", "/op/v0/device/history/query", request, response, options.Debug)
	if err != nil {
		return err
	}

	if err = foxess.IsError(response.ErrorNumber, response.Message); err != nil {
		return err
	}

	for _, inverter := range response.Result {
		for _, variable := range inverter.Variables {

			slices.SortFunc(variable.DataPoints, func(i, j DataPoint) int {
				result := i.Time.Time.Compare(j.Time.Time)
				if result != 1 {
					return result
				}
				return result
			})
		}
	}
	return nil
}

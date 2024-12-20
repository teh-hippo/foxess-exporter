package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/prometheus/prompb"
	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type HistoryCommand struct {
	Inverter          string   `short:"i" long:"inverter"            description:"Inverter serial number" required:"true"`
	Date              string   `short:"d" long:"date"                description:"Date for the request"`
	End               string   `short:"e" long:"end-date"            description:"End date (range)"`
	Variables         []string `short:"V" long:"variable"            description:"Variables to retrieve"`
	Format            string   `short:"o" long:"output"              description:"Output format"          default:"table"                              choices:"table,json,remote-write"`
	RemoteWriteTarget string   `short:"t" long:"remote-write-target" description:"Remote write target"    default:"http://127.0.0.1:9090/api/v1/write"`
	config            *foxess.Config
	beginDate         time.Time
	endDate           time.Time
	SkipOutOfBounds   bool `short:"I" long:"skip-out-of-bounds" description:"Skip over dates that report back out of bounds"`
}

func (x *HistoryCommand) Register(parser *flags.Parser, config *foxess.Config) {
	if _, err := parser.AddCommand("history", "Get the history", "Get the history of a variable", x); err != nil {
		panic(err)
	}

	x.config = config
}

var ErrRemoteWrite = errors.New("failed to perform remote write operation")

const OneDay = 24 * time.Hour

func (x *HistoryCommand) Execute(_ []string) error {
	if err := x.validateArguments(); err != nil {
		return err
	}

	date := x.beginDate

	for {
		if err := x.retrieveDate(date); err != nil {
			return fmt.Errorf("failed to retrieve history for %s: %w", date.Format(time.DateOnly), err)
		}

		date = date.Add(OneDay)
		if date.UnixMilli() >= x.endDate.UnixMilli() {
			break
		}
	}

	return nil
}

func (x *HistoryCommand) validateArguments() error {
	// Work out the start date
	if x.Date == "" {
		if x.End != "" {
			return fmt.Errorf("%w: end date can only be provided when a date has been provided", ErrInvalidArgument)
		}

		now := time.Now()
		x.beginDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	} else {
		begin, err := time.Parse(time.DateOnly, x.Date)
		if err != nil {
			return fmt.Errorf("%w: unable to parse '%s': %w", ErrInvalidArgument, x.Date, err)
		}

		x.beginDate = begin
	}

	// Work out the end date
	if x.End == "" {
		x.endDate = x.beginDate.Add(OneDay)
	} else {
		end, err := time.Parse(time.DateOnly, x.End)
		if err != nil {
			return fmt.Errorf("%w: unable to parse '%s': %w", ErrInvalidArgument, x.End, err)
		}

		x.endDate = end
	}

	if x.Format == FormatRemoteWrite && x.RemoteWriteTarget == "" {
		return fmt.Errorf("%w: missing remote write target", ErrInvalidArgument)
	}

	return nil
}

func (x *HistoryCommand) retrieveDate(date time.Time) error {
	endDate := date.Add(OneDay)
	log.Printf("Retrieving history of %s for %s", x.Inverter, date.Format(time.DateOnly))

	response, err := x.config.GetVariableHistory(x.Inverter, date, endDate, x.Variables)
	if err != nil {
		return fmt.Errorf("failed to retrieve history of %s for %s: %w", x.Inverter, x.beginDate.Format(time.DateOnly), err)
	}

	if err = x.writeResult(date, response); err != nil {
		return fmt.Errorf("failed to output result: %w", err)
	}

	return nil
}

func (x *HistoryCommand) writeResult(date time.Time, inverterHistories []foxess.InverterHistory) error {
	switch x.Format {
	case FormatTable:
		createTable(inverterHistories)
	case FormatJSON:
		if err := util.JSONToStdOut(inverterHistories); err != nil {
			return fmt.Errorf("failed to write json output: %w", err)
		}

		return nil
	case FormatRemoteWrite:
		if err := x.remoteWrite(date, inverterHistories); err != nil {
			return fmt.Errorf("failed to write tsdb output: %w", err)
		}
	}

	return nil
}

func createTable(inverterHistories []foxess.InverterHistory) {
	tbl := table.New("Inverter", "Variable", "Name", "Unit", "Time", "Value")

	for _, inverter := range inverterHistories {
		for _, variable := range inverter.Variables {
			for _, point := range variable.DataPoints {
				tbl.AddRow(inverter.DeviceSN, variable.Variable, variable.Name, variable.Unit, point.Time, point.Value.Number)
			}
		}
	}

	tbl.Print()
}

func (x *HistoryCommand) remoteWrite(date time.Time, inverterHistories []foxess.InverterHistory) error {
	httpClient := &http.Client{ //nolint:exhaustruct
		Timeout: Ten * time.Second,
	}

	marshalled, err := proto.Marshal(&prompb.WriteRequest{ //nolint:exhaustruct
		Timeseries: convertToTimeSeries(inverterHistories),
	})
	if err != nil {
		return fmt.Errorf("%w: failed to marshall variables to time series: %w", ErrRemoteWrite, err)
	}

	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, x.RemoteWriteTarget, bytes.NewBuffer(snappy.Encode(nil, marshalled)))
	if err != nil {
		return fmt.Errorf("%w: write request failed: %w", ErrRemoteWrite, err)
	}

	request.Header.Add("X-Prometheus-Remote-Write-Version", "0.1.0")
	request.Header.Add("Content-Encoding", "snappy")
	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("User-Agent", "foxess-exporter 1.0")

	// Send http request.
	httpResp, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("%w: failed to complete request '%s': %w", ErrRemoteWrite, x.RemoteWriteTarget, err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		response, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return fmt.Errorf("%w: %d response, and unable to read response: %w", ErrRemoteWrite, httpResp.StatusCode, err)
		}

		message := strings.Trim(string(response), "\n")
		if httpResp.StatusCode == http.StatusBadRequest && message == "out of bounds" {
			log.Printf("Ignoring failed remote-write for %s: %s - %s", date.Format(time.DateOnly), httpResp.Status, message)

			return nil
		}

		return fmt.Errorf("%w: %d: %s", ErrRemoteWrite, httpResp.StatusCode, message)
	}

	return nil
}

func convertToTimeSeries(inverterHistories []foxess.InverterHistory) []prompb.TimeSeries {
	const endOfSeries = 0x7ff0000000000002

	var timeSeries []prompb.TimeSeries

	for _, inverter := range inverterHistories {
		bulk := make([]prompb.TimeSeries, len(inverter.Variables))

		for variableIndex, variable := range inverter.Variables {
			samples := make([]prompb.Sample, len(variable.DataPoints)+1)

			for dataIndex, dataPoint := range variable.DataPoints {
				samples[dataIndex] = prompb.Sample{ //nolint:exhaustruct
					Timestamp: dataPoint.Time.UnixNano() / int64(time.Millisecond),
					Value:     dataPoint.Value.Number,
				}
			}

			samples[len(samples)-1] = prompb.Sample{ //nolint:exhaustruct
				Timestamp: variable.DataPoints[len(variable.DataPoints)-1].Time.UnixNano()/int64(time.Millisecond) + 1,
				Value:     endOfSeries,
			}

			bulk[variableIndex] = prompb.TimeSeries{ //nolint:exhaustruct
				Labels: []prompb.Label{
					{ //nolint:exhaustruct
						Name:  "__name__",
						Value: "foxess_realtime_data",
					},
					{ //nolint:exhaustruct
						Name:  "inverter",
						Value: inverter.DeviceSN,
					},
					{ //nolint:exhaustruct
						Name:  "variable",
						Value: variable.Variable,
					},
				},
				Samples: samples,
			}
		}

		timeSeries = append(timeSeries, bulk...)
	}

	return timeSeries
}

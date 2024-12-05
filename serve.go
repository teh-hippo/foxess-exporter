package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServeCommand struct {
	Port      int      `short:"p" long:"port" description:"Port to listen on" default:"2112" required:"true" env:"PORT"`
	Inverters []string `short:"i" long:"inverter" description:"Inverter serial numbers" required:"true" env:"INVERTERS" env-delim:","`
	Variables []string `short:"v" long:"variable" description:"Variables to retrieve" required:"false" env:"VARIABLES" env-delim:","`
	Frequency int64    `short:"f" long:"frequency" description:"Frequency of updates (in seconds)." env:"FREQUENCY" default:"60" required:"true"`
}

var serveCommand ServeCommand
var reg = prometheus.NewRegistry()
var gauges = make(map[string]prometheus.Gauge)

func init() {
	parser.AddCommand("serve", "Start the exporter", "Start the exporter", &serveCommand)
}

func (x *ServeCommand) Execute(args []string) error {
	// Ensure the frequency is at least 60 seconds.
	x.Frequency = int64(math.Max(float64(60), float64(x.Frequency)))

	// Start polling FoxESS.
	x.realtime()

	// Server metrics on the standard endpoint.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	return http.ListenAndServe(":"+fmt.Sprint(x.Port), nil)
}

func (x *ServeCommand) setupGauge(variable string, inverter string) prometheus.Gauge {
	log.Printf("Creating '%s' gauge for '%s'.\n", variable, inverter)
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foxess_realtime_data",
		Help: "Data from the FoxESS platform.",
		ConstLabels: prometheus.Labels{
			"inverter": inverter,
			"variable": variable,
		},
	})
	gauges[variable] = gauge
	reg.MustRegister(gauge)
	return gauge
}

func (x *ServeCommand) realtime() {
	go func() {
		for {
			log.Printf("Retrieving device real-time data")
			response, err := GetRealTimeData(serveCommand.Inverters, serveCommand.Variables)
			if err != nil {
				fmt.Println(err)
			} else {
				for _, result := range response.Result {
					for _, variable := range result.Variables {
						gauge := gauges[variable.Variable]
						if gauge == nil {
							gauge = x.setupGauge(variable.Variable, result.DeviceSN)
						}
						log.Printf("Setting '%s' for '%s' to: %f", variable.Variable, result.DeviceSN, variable.Value.Number)
						gauge.Set(variable.Value.Number)
					}
				}
			}
			time.Sleep(time.Duration(x.Frequency) * time.Second)
		}
	}()
}

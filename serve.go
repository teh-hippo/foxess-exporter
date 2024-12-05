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
	Port      int      `short:"p" long:"port" description:"Port to listen on" default:"2112"`
	Inverter  string   `short:"i" long:"inverter" description:"Inverter serial number" required:"true"`
	Variables []string `short:"q" long:"variable" description:"Variables to retrieve" required:"true"`
	Frequency int64    `short:"f" long:"frequency" description:"Frequency of updates (in seconds)." default:"60" required:"true"`
}

var serveCommand ServeCommand

var gauges = make(map[string]prometheus.Gauge)

func init() {
	parser.AddCommand("serve", "Start the exporter", "Start the exporter", &serveCommand)
}

func (x *ServeCommand) Execute(args []string) error {
	reg := prometheus.NewRegistry()
	if len(x.Variables) == 0 {
		return fmt.Errorf("no variables specified")
	}

	x.Frequency = int64(math.Max(float64(60), float64(x.Frequency)))

	for _, variable := range x.Variables {
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "foxess_metrics",
			Help: "Displays historical value for custom FoxESS variables.",
			ConstLabels: prometheus.Labels{
				"inverter": x.Inverter,
				"variable": variable,
			},
		})
		gauges[variable] = gauge
		reg.MustRegister(gauge)
	}

	x.realtime()

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	return http.ListenAndServe(":"+fmt.Sprint(x.Port), nil)
}

func (x *ServeCommand) realtime() {
	go func() {
		log.Printf("Retrieving %d current real-time value(s)", len(x.Variables))
		response, err := GetRealTime(serveCommand.Inverter, serveCommand.Variables)
		if err != nil {
			fmt.Println(err)
		} else {
			for _, variable := range response.Result[0].Variables {
				log.Println("Setting gauge for", variable.Variable, "to", variable.Value.Number)
				gauge := gauges[variable.Variable]
				gauge.Set(variable.Value.Number)
			}
		}
		time.Sleep(time.Duration(x.Frequency) * time.Second)
	}()
}

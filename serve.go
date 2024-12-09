package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type DiscoveryTarget struct {
	Targets []string `json:"targets"`
	Labels  []Device `json:"labels"`
}

type DiscoveryResponse struct {
	Items []DiscoveryTarget
}

type ServeCommand struct {
	Port          int      `short:"p" long:"port" description:"Port to listen on" default:"2112" required:"true" env:"PORT"`
	Inverters     []string `short:"i" long:"inverter" description:"Inverter serial numbers" required:"true" env:"INVERTERS" env-delim:","`
	Variables     []string `short:"v" long:"variable" description:"Variables to retrieve" required:"false" env:"VARIABLES" env-delim:","`
	Frequency     int64    `short:"f" long:"frequency" description:"Frequency of updates (in seconds)." env:"FREQUENCY" default:"60" required:"true"`
	Discovery     string   `short:"d" long:"discovery" description:"Configure discovery behaviour." required:"false" choices:"off,on,only" default:"off"`
	ApiUsageBlock float64  `short:"l" long:"limit" description:"Block further API calls being made once usage crosses the provided percentage." required:"false" default:"90" env:"USAGE_LIMIT"`
}

var serveCommand ServeCommand
var reg = prometheus.NewRegistry()
var gauges = make(map[string]prometheus.Gauge)
var devicesChan = make(chan *[]Device, 1)
var deviceSerialNumbersChan = make(chan *[]string, 1)
var apiUsageChan = make(chan *ApiUsage, 1)

func init() {
	parser.AddCommand("serve", "Start the exporter", "Start the exporter", &serveCommand)
}

func (x *ServeCommand) Execute(args []string) error {
	// Ensure the frequency is at least 60 seconds.
	x.Frequency = int64(math.Max(float64(60), float64(x.Frequency)))

	// Regular API usage.
	x.reportusage()

	// Server metrics on the standard endpoint.
	if x.Discovery != "only" {
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		// Start polling FoxESS.
		x.realtime()
	}
	if x.Discovery != "off" {
		x.updatediscovery()
		http.Handle("/discovery", http.HandlerFunc(Discovery))
	}
	http.Handle("/favicon.ico", http.RedirectHandler("https://www.foxesscloud.com/favicon.ico", http.StatusMovedPermanently))
	return http.ListenAndServe(":"+fmt.Sprint(x.Port), nil)
}

func Discovery(w http.ResponseWriter, r *http.Request) {
	log.Printf("Discovery request received.")
	devices := <-devicesChan
	response := &DiscoveryResponse{}
	for _, device := range *devices {
		response.Items = append(response.Items,
			DiscoveryTarget{
				Targets: []string{device.DeviceSerialNumber},
				Labels:  []Device{device}})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func (x *ServeCommand) reportusage() {
	go func() {
		for {
			if apiUsage, err := GetApiUsage(); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Usage: %.0f/%.0f (%.2f%%)\n", apiUsage.Remaining, apiUsage.Total, apiUsage.PercentageUsed)
				apiUsageChan <- apiUsage
			}
			time.Sleep(10 * time.Minute)
		}
	}()
}

func GetInverters() *[]string {
	if len(serveCommand.Inverters) == 0 {
		return <-deviceSerialNumbersChan
	} else {
		return &serveCommand.Inverters
	}
}

func (x *ServeCommand) realtime() {
	go func() {
		for {
			log.Printf("Retrieving device real-time data")
			apiUsage := <-apiUsageChan
			if apiUsage.PercentageUsed >= x.ApiUsageBlock {
				log.Printf("Usage is over the limit.")
			} else if data, err := GetRealTimeData(*GetInverters(), serveCommand.Variables); err != nil {
				fmt.Println(err)
			} else {
				x.processResponse(data)
			}
			time.Sleep(time.Duration(x.Frequency) * time.Second)
		}
	}()
}

func (x *ServeCommand) processResponse(data []RealTimeData) {
	for _, result := range data {
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

func (x *ServeCommand) updatediscovery() {
	go func() {
		for {
			log.Print("Updating discovery data")
			if devices, err := GetDeviceList(); err != nil {
				fmt.Println(err)
			} else {
				deviceSerialNumbers := make([]string, len(devices))
				for i, device := range devices {
					deviceSerialNumbers[i] = device.DeviceSerialNumber
				}
				devicesChan <- &devices
				deviceSerialNumbersChan <- &deviceSerialNumbers
			}
			time.Sleep(24 * time.Hour)
		}
	}()
}

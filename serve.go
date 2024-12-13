package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/teh-hippo/foxess-exporter/util"
)

type DiscoveryTarget struct {
	Targets []string `json:"targets"`
	Labels  []Device `json:"labels"`
}

type DiscoveryResponse struct {
	Items []DiscoveryTarget
}

type ServeCommand struct {
	Port                int      `short:"p" long:"port" description:"Port to listen on" default:"2112" required:"true" env:"PORT"`
	Inverters           []string `short:"i" long:"inverter" description:"Inverter serial numbers" required:"true" env:"INVERTERS" env-delim:","`
	Variables           []string `short:"V" long:"variable" description:"Variables to retrieve" required:"false" env:"VARIABLES" env-delim:","`
	RealTimeIntervalSec int64    `short:"R" long:"realtime-interval" description:"Frequency of updating real-time data (in seconds)." env:"REAL_TIME_INTERVAL" default:"120" required:"true"`
	StatusIntervalSec   int64    `short:"S" long:"status-interval" description:"Frequency of updating the status of devices (in seconds)." env:"STATUS_INTERVAL" default:"900" required:"true"`
	Discovery           string   `short:"d" long:"discovery" description:"Configure discovery behaviour." required:"false" choices:"off,on,only" default:"off" env:"DISCOVERY"`
	Verbose             bool     `short:"v" long:"verbose" description:"Enable verbose logging." required:"false"`
}

var serveCommand ServeCommand
var reg = prometheus.NewRegistry()
var metrics = make(map[string]prometheus.Gauge)
var last_reported_time = make(map[string]time.Time)
var devicesChan = make(chan *[]Device, 1)
var deviceSerialNumbersChan = make(chan *[]string, 1)
var metricsLock = sync.Mutex{}
var apiQuota = NewApiQuota()

func init() {
	if _, err := parser.AddCommand("serve", "Serve FoxESS metrics", "Creates a Prometheus endpoint where metrics can be provided.", &serveCommand); err != nil {
		panic(err)
	}
}

func (x *ServeCommand) validateIntervals() error {
	x.RealTimeIntervalSec = util.Clamp(x.RealTimeIntervalSec, 60)
	x.StatusIntervalSec = util.Clamp(x.StatusIntervalSec, 60)

	const dayInSeconds float64 = 24 * 60 * 60
	available := float64(1440)
	available -= dayInSeconds / float64(x.RealTimeIntervalSec)
	available -= dayInSeconds / float64(x.StatusIntervalSec)
	if available < 0 {
		return errors.New("current intervals would result in API usage exceeding the maximum daily allowance")
	}
	return nil
}

func (x *ServeCommand) startDeviceStatusMetric() {
	go func() {
		for {
			if x.isApiQuotaAvailable() {
				x.verbose("Retrieving device status")
				if devices, err := GetDeviceList(); err != nil {
					fmt.Println(err)
				} else {
					for _, device := range devices {
						log.Printf("Setting status of %s to: %d (%s)", device.DeviceSerialNumber, device.Status, device.status())
						x.statusMetric(device.DeviceSerialNumber).Set(float64(device.Status))
					}
				}
			}
			time.Sleep(time.Duration(x.StatusIntervalSec) * time.Second)
		}
	}()
}

func (x *ServeCommand) Execute(args []string) error {
	x.startApiQuotaManagement()

	if x.Discovery != "only" {
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		x.startRealTimeMetrics()
		x.startDeviceStatusMetric()
	}
	if x.Discovery != "off" {
		x.startDiscovery()
		http.Handle("/discovery", http.HandlerFunc(discovery))
	}
	http.Handle("/favicon.ico", http.RedirectHandler("https://www.foxesscloud.com/favicon.ico", http.StatusMovedPermanently))
	server := &http.Server{Addr: ":" + fmt.Sprint(x.Port), ReadHeaderTimeout: 3 * time.Second}
	return server.ListenAndServe()
}

func (x *ServeCommand) startApiQuotaManagement() {
	go func() {
		for {
			if err := apiQuota.update(); err != nil {
				fmt.Println(err)
			} else {
				apiUsage := apiQuota.current()
				log.Printf("Usage: %.0f/%.0f (%.2f%%)\n", apiUsage.Total-apiUsage.Remaining, apiUsage.Total, apiUsage.PercentageUsed)
			}
			time.Sleep(10 * time.Minute)
		}
	}()
}

func (x *ServeCommand) startRealTimeMetrics() {
	go func() {
		for {
			if x.isApiQuotaAvailable() {
				x.verbose("Retrieving device real-time data")
				if data, err := GetRealTimeData(*inverters(), serveCommand.Variables); err != nil {
					fmt.Println(err)
				} else {
					x.handleRealTimeData(data)
				}
			}
			time.Sleep(time.Duration(x.RealTimeIntervalSec) * time.Second)
		}
	}()
}

func (x *ServeCommand) startDiscovery() {
	go func() {
		for {
			log.Print("Updating discovery data")
			if devices, err := GetDeviceList(); err != nil {
				fmt.Fprintln(os.Stderr, err)
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

func discovery(w http.ResponseWriter, r *http.Request) {
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

func inverters() *[]string {
	if len(serveCommand.Inverters) == 0 {
		return <-deviceSerialNumbersChan
	} else {
		return &serveCommand.Inverters
	}
}

func (x *ServeCommand) statusMetric(inverter string) prometheus.Gauge {
	metricsLock.Lock()
	defer metricsLock.Unlock()
	metric := metrics[inverter]
	if metric != nil {
		return metric
	}

	x.verbose("Creating 'status' gauge for '%s'.\n", inverter)
	metric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foxess_device_status",
		Help: "Status of the inverter.",
		ConstLabels: prometheus.Labels{
			"inverter": inverter,
		},
	})
	metrics[inverter] = metric
	reg.MustRegister(metric)
	return metric
}

func (x *ServeCommand) handleRealTimeData(data []RealTimeData) {
	for _, result := range data {
		if last_reported_time[result.DeviceSN].Equal(result.Time.Time) {
			x.verbose("No update for %s.", result.DeviceSN)
			continue
		}
		log.Printf("Updating %d metric%s for %s, recorded:%v.", len(result.Variables), util.Pluralise(len(result.Variables)), result.DeviceSN, result.Time.Time)
		last_reported_time[result.DeviceSN] = result.Time.Time
		for _, variable := range result.Variables {
			x.verbose("Setting '%s' for '%s' to: %f", variable.Variable, result.DeviceSN, variable.Value.Number)
			x.realTimeMetric(variable.Variable, result.DeviceSN).Set(variable.Value.Number)
		}
	}
}

func (x *ServeCommand) realTimeMetric(variable string, inverter string) prometheus.Gauge {
	metricsLock.Lock()
	defer metricsLock.Unlock()
	metric := metrics[variable]
	if metric != nil {
		return metric
	}
	x.verbose("Creating '%s' gauge for '%s'.\n", variable, inverter)
	metric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foxess_realtime_data",
		Help: "Data from the FoxESS platform.",
		ConstLabels: prometheus.Labels{
			"inverter": inverter,
			"variable": variable,
		},
	})
	metrics[variable] = metric
	reg.MustRegister(metric)
	return metric
}

func (x *ServeCommand) isApiQuotaAvailable() bool {
	return apiQuota.current().Remaining > 0
}

func (x *ServeCommand) verbose(format string, v ...any) {
	if x.Verbose {
		log.Printf(format, v...)
	}
}

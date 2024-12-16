package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/teh-hippo/foxess-exporter/serve"
	"github.com/teh-hippo/foxess-exporter/util"
)

type ServeCommand struct {
	Port             int             `short:"p" long:"port" description:"Port to listen on" default:"2112" required:"true" env:"PORT"`
	Inverters        map[string]bool `short:"i" long:"inverter" description:"Inverter serial numbers" env:"INVERTERS" env-delim:","`
	Variables        []string        `short:"V" long:"variable" description:"Variables to retrieve" env:"VARIABLES" env-delim:","`
	RealTimeInterval time.Duration   `short:"R" long:"realtime-interval" description:"Frequency of updating real-time data." env:"REAL_TIME_INTERVAL" default:"3m" required:"true"`
	StatusInterval   time.Duration   `short:"S" long:"status-interval" description:"Frequency of updating the status of devices." env:"STATUS_INTERVAL" default:"15m" required:"true"`
	Verbose          bool            `short:"v" long:"verbose" description:"Enable verbose logging." env:"VERBOSE"`
}

var serveCommand ServeCommand
var deviceCache *serve.DeviceCache
var apiCache *serve.ApiQuota
var metrics *serve.Metrics

func init() {
	if _, err := parser.AddCommand("serve", "Serve FoxESS metrics", "Creates a Prometheus endpoint where metrics can be provided.", &serveCommand); err != nil {
		panic(err)
	}
	deviceCache = serve.NewDeviceCache()
	apiCache = serve.NewApiCache()
	metrics = serve.NewMetrics()
}

func (x *ServeCommand) validateIntervals() error {
	const oneDay time.Duration = 24 * time.Hour
	const oneMinute = time.Minute
	x.RealTimeInterval = util.Clamp(x.RealTimeInterval, oneMinute, oneDay)
	x.StatusInterval = util.Clamp(x.StatusInterval, oneMinute, oneDay)

	apiCallsPerDay := float64(oneDay.Minutes())
	realTimeCalls := oneDay / x.RealTimeInterval
	apiCallsPerDay -= float64(realTimeCalls)
	statusCalls := oneDay / x.StatusInterval
	apiCallsPerDay -= float64(statusCalls)
	if apiCallsPerDay < 0 {
		return errors.New("current intervals would result in API usage exceeding the maximum daily allowance")
	}
	return nil
}

func (x *ServeCommand) Execute(args []string) error {
	if len(serveCommand.Inverters) > 0 {
		ids := make([]string, 0, len(serveCommand.Inverters))
		for deviceId := range serveCommand.Inverters {
			ids = append(ids, deviceId)
		}
		deviceCache.Set(ids)
	}

	run(10*time.Minute, false, x.updateApiQuota)
	run(x.StatusInterval, true, x.updateDeviceStatus)
	run(x.RealTimeInterval, true, x.updateRealTimeMetrics)

	http.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))
	http.Handle("/favicon.ico", http.RedirectHandler("https://www.foxesscloud.com/favicon.ico", http.StatusMovedPermanently))

	server := &http.Server{Addr: ":" + fmt.Sprint(x.Port), ReadHeaderTimeout: 3 * time.Second}
	return server.ListenAndServe()
}

func (x *ServeCommand) updateApiQuota() {
	a, err := foxessApi.GetApiUsage()
	if err != nil {
		fmt.Printf("failed to update API usage: %v", err)
	} else {
		x.verbose("Updating API usage")
		apiCache.Update(a)
		log.Printf("Usage: %.0f/%.0f (%.2f%%)\n", a.Total-a.Remaining, a.Total, a.PercentageUsed)
	}
}

func (x *ServeCommand) updateDeviceStatus() {
	x.verbose("Retrieving device status")
	devices, err := foxessApi.GetDeviceList()
	if err != nil {
		fmt.Printf("Unable to update device list: %v", err)
	} else {
		metrics.UpdateStatus(devices, x.Include)
		hasFilter := len(x.Inverters) > 0
		if !hasFilter {
			ids := make([]string, len(devices))
			for i, device := range devices {
				ids[i] = device.DeviceSerialNumber
			}
			deviceCache.Set(ids)
		}
	}
}

func (x *ServeCommand) updateRealTimeMetrics() {
	x.verbose("Retrieving latest real-time data")
	data, err := foxessApi.GetRealTimeData(*deviceCache.Get(), serveCommand.Variables)
	if err != nil {
		fmt.Printf("Unable to retrieve latest real-time data: %v", err)
	}

	metrics.UpdateRealTime(data)
}

func run(interval time.Duration, checkApi bool, execute func()) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			if !checkApi || apiCache.IsQuotaAvailable() {
				execute()
			}
		}
	}()
}

func (x *ServeCommand) Include(inverter string) bool {
	return len(serveCommand.Inverters) == 0 || serveCommand.Inverters[inverter]
}

func (x *ServeCommand) verbose(format string, v ...any) {
	if x.Verbose {
		log.Printf(format, v...)
	}
}

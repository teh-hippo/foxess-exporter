package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/teh-hippo/foxess-exporter/serve"
	"github.com/teh-hippo/foxess-exporter/util"
)

type ServeCommand struct {
	Port             int             `short:"p" long:"port"              description:"Port to listen on"                  env:"PORT"               required:"true" default:"2112"`
	Inverters        map[string]bool `short:"i" long:"inverter"          description:"Inverter serial numbers"            env:"INVERTERS"          env-delim:""`
	Variables        []string        `short:"V" long:"variable"          description:"Variables to retrieve"              env:"VARIABLES"          env-delim:""`
	RealTimeInterval time.Duration   `short:"R" long:"realtime-interval" description:"Update frequency of real-time data" env:"REAL_TIME_INTERVAL" required:"true" default:"3m"`
	StatusInterval   time.Duration   `short:"S" long:"status-interval"   description:"Update frequency of devices status" env:"STATUS_INTERVAL"    required:"true" default:"15m"`
	Verbose          bool            `short:"v" long:"verbose"           description:"Enable verbose logging"             env:"VERBOSE"`
}

var (
	deviceCache serve.DeviceCache
	apiCache    serve.APIQuota
	metrics     serve.Metrics
)

func (x *ServeCommand) Register(parser *flags.Parser) {
	if _, err := parser.AddCommand("serve", "Serve FoxESS metrics", "Creates a Prometheus endpoint where metrics can be provided.", x); err != nil {
		panic(err)
	}

	deviceCache = *serve.NewDeviceCache()
	apiCache = *serve.NewAPIQuota()
	metrics = *serve.NewMetrics()
}

func (x *ServeCommand) validateIntervals() error {
	const oneDay time.Duration = 24 * time.Hour
	x.RealTimeInterval = util.Clamp(x.RealTimeInterval, time.Minute, oneDay)
	x.StatusInterval = util.Clamp(x.StatusInterval, time.Minute, oneDay)

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

func (x *ServeCommand) Execute(_ []string) error {
	if len(x.Inverters) > 0 {
		ids := make([]string, 0, len(x.Inverters))
		for deviceID := range x.Inverters {
			ids = append(ids, deviceID)
		}

		deviceCache.Set(ids)
	}

	run(10*time.Minute, false, x.updateAPIQuota)
	run(x.StatusInterval, true, x.updateDeviceStatus)
	run(x.RealTimeInterval, true, x.updateRealTimeMetrics)

	http.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))
	http.Handle("/favicon.ico", http.RedirectHandler("https://www.foxesscloud.com/favicon.ico", http.StatusMovedPermanently))

	server := &http.Server{Addr: ":" + fmt.Sprint(x.Port), ReadHeaderTimeout: 3 * time.Second}
	return server.ListenAndServe()
}

func (x *ServeCommand) updateAPIQuota() {
	apiUsage, err := foxessAPI.GetAPIUsage()
	if err != nil {
		fmt.Printf("failed to update API usage: %v", err)
	} else {
		x.verbose("Updating API usage")
		apiCache.Set(apiUsage)
		log.Printf("Usage: %.0f/%.0f (%.2f%%)\n", apiUsage.Total-apiUsage.Remaining, apiUsage.Total, apiUsage.PercentageUsed)
	}
}

func (x *ServeCommand) updateDeviceStatus() {
	x.verbose("Retrieving device status")

	devices, err := foxessAPI.GetDeviceList()

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
	data, err := foxessAPI.GetRealTimeData(deviceCache.Get(), x.Variables)
	if err != nil {
		fmt.Printf("Unable to retrieve latest real-time data: %v", err)
	}

	metrics.UpdateRealTime(data)
}

func run(interval time.Duration, checkAPI bool, execute func()) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if !checkAPI || apiCache.IsQuotaAvailable() {
				execute()
			}
		}
	}()
}

func (x *ServeCommand) Include(inverter string) bool {
	return len(x.Inverters) == 0 || x.Inverters[inverter]
}

func (x *ServeCommand) verbose(format string, v ...any) {
	if x.Verbose {
		log.Printf(format, v...)
	}
}

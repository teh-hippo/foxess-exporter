package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/serve"
	"github.com/teh-hippo/foxess-exporter/util"
)

const Ten = 10

type ServeCommand struct {
	Port             int             `short:"p" long:"port"              description:"Port to listen on"                  env:"PORT"               required:"true" default:"2112"`
	Inverters        map[string]bool `short:"i" long:"inverter"          description:"Inverter serial numbers"            env:"INVERTERS"          env-delim:""`
	Variables        []string        `short:"V" long:"variable"          description:"Variables to retrieve"              env:"VARIABLES"          env-delim:""`
	RealTimeInterval time.Duration   `short:"R" long:"realtime-interval" description:"Update frequency of real-time data" env:"REAL_TIME_INTERVAL" required:"true" default:"3m"`
	StatusInterval   time.Duration   `short:"S" long:"status-interval"   description:"Update frequency of devices status" env:"STATUS_INTERVAL"    required:"true" default:"15m"`
	Verbose          bool            `short:"v" long:"verbose"           description:"Enable verbose logging"             env:"VERBOSE"`
	deviceCache      *serve.DeviceCache
	apiQuota         *serve.APIQuota
	metrics          *serve.Metrics
	config           *foxess.Config
}

func (x *ServeCommand) Register(parser *flags.Parser, config *foxess.Config) {
	if _, err := parser.AddCommand("serve", "Serve FoxESS metrics", "Creates a Prometheus endpoint where metrics can be provided.", x); err != nil {
		panic(err)
	}

	x.config = config
	x.deviceCache = serve.NewDeviceCache()
	x.apiQuota = serve.NewAPIQuota()
	x.metrics = serve.NewMetrics()
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
		return fmt.Errorf("%w: %s", ErrInvalidArgument, "current intervals would result in API usage exceeding the maximum daily allowance")
	}

	return nil
}

func (x *ServeCommand) Execute(_ []string) error {
	if len(x.Inverters) > 0 {
		ids := make([]string, 0, len(x.Inverters))
		for deviceID := range x.Inverters {
			ids = append(ids, deviceID)
		}

		x.deviceCache.Set(ids)
	}

	x.run(Ten*time.Minute, false, x.updateAPIQuota)
	x.run(x.StatusInterval, true, x.updateDeviceStatus)
	x.run(x.RealTimeInterval, true, x.updateRealTimeMetrics)

	http.Handle("/metrics", promhttp.HandlerFor(x.metrics.Registry, promhttp.HandlerOpts{ //nolint:exhaustruct
		ErrorLog: log.Default(),
	}))
	http.Handle("/favicon.ico", http.RedirectHandler("https://www.foxesscloud.com/favicon.ico", http.StatusMovedPermanently))

	server := &http.Server{Addr: ":" + strconv.Itoa(x.Port), ReadHeaderTimeout: Ten * time.Second} //nolint:exhaustruct

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (x *ServeCommand) updateAPIQuota() {
	apiUsage, err := x.config.GetAPIUsage()
	if err != nil {
		log.Printf("failed to update API usage: %v", err)
	} else {
		x.verbose("Updating API usage")
		x.apiQuota.Set(apiUsage)
		log.Printf("Usage: %.0f/%.0f (%.2f%%)\n", apiUsage.Total-apiUsage.Remaining, apiUsage.Total, apiUsage.PercentageUsed)
	}
}

func (x *ServeCommand) updateDeviceStatus() {
	x.verbose("Retrieving device status")

	devices, err := x.config.GetDeviceList()

	if err != nil {
		log.Printf("Unable to update device list: %v", err)
	} else {
		x.metrics.UpdateStatus(devices, x.Include)
		hasFilter := len(x.Inverters) > 0

		if !hasFilter {
			ids := make([]string, len(devices))
			for i, device := range devices {
				ids[i] = device.DeviceSerialNumber
			}

			x.deviceCache.Set(ids)
		}
	}
}

func (x *ServeCommand) updateRealTimeMetrics() {
	x.verbose("Retrieving latest real-time data")

	data, err := x.config.GetRealTimeData(x.deviceCache.Get(), x.Variables)
	if err != nil {
		log.Printf("Unable to retrieve latest real-time data: %v", err)
	}

	x.metrics.UpdateRealTime(data)
}

func (x *ServeCommand) run(interval time.Duration, checkAPI bool, execute func()) {
	go func() {
		for {
			if !checkAPI || x.apiQuota.IsQuotaAvailable() {
				execute()
			}

			time.Sleep(interval)
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

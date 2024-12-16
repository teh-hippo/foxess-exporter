package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/teh-hippo/foxess-exporter/serve"
	"github.com/teh-hippo/foxess-exporter/util"
)

type ApiCache struct {
	sync.Mutex
	apiUsage *ApiUsage
	cond     *sync.Cond
}

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
var apiCache ApiCache
var metrics *serve.Metrics

func init() {
	if _, err := parser.AddCommand("serve", "Serve FoxESS metrics", "Creates a Prometheus endpoint where metrics can be provided.", &serveCommand); err != nil {
		panic(err)
	}
	deviceCache = serve.NewDeviceCache()
	apiCache.cond = sync.NewCond(&apiCache)
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
	deviceCache.Initalise(serveCommand.Inverters)

	x.startApiQuotaManagement()
	x.startDeviceStatusMetric()
	x.startRealTimeMetrics()

	http.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))
	http.Handle("/favicon.ico", http.RedirectHandler("https://www.foxesscloud.com/favicon.ico", http.StatusMovedPermanently))

	server := &http.Server{Addr: ":" + fmt.Sprint(x.Port), ReadHeaderTimeout: 3 * time.Second}
	return server.ListenAndServe()
}

func (x *ServeCommand) startApiQuotaManagement() {
	x.verbose("Starting API quota management")
	go func() {
		for {
			x.verbose("Updating API usage")
			if a, err := apiCache.updater(); err != nil {
				fmt.Println(err)
			} else {
				log.Printf("Usage: %.0f/%.0f (%.2f%%)\n", a.Total-a.Remaining, a.Total, a.PercentageUsed)
			}
			time.Sleep(10 * time.Minute)
		}
	}()
}

func (x *ServeCommand) startDeviceStatusMetric() {
	x.verbose("Starting device status metric")
	go func() {
		for {
			if apiCache.isApiQuotaAvailable() {
				x.verbose("Retrieving device status")
				devices, err := foxessApi.GetDeviceList()
				if err != nil {
					fmt.Printf("Unable to update device list: %v", err)
				} else {
					metrics.UpdateStatus(devices, include)
					hasFilter := len(x.Inverters) > 0
					if !hasFilter {
						deviceCache.Set(devices)
					}
				}
			} else {
				x.verbose("No quota available to update device status")
			}
			time.Sleep(x.StatusInterval)
		}
	}()
}

func include(inverter string) bool {
	return len(serveCommand.Inverters) == 0 || serveCommand.Inverters[inverter]
}

func (x *ServeCommand) startRealTimeMetrics() {
	go func() {
		for {
			if apiCache.isApiQuotaAvailable() {
				x.verbose("Retrieving latest real-time data")
				data, err := foxessApi.GetRealTimeData(*deviceCache.Get(), serveCommand.Variables)
				if err != nil {
					fmt.Printf("Unable to retrieve latest real-time data: %v", err)
				} else {
					metrics.UpdateRealTime(data)
				}
			} else {
				x.verbose("No quota available to update real-time metrics")
			}
			time.Sleep(time.Duration(x.RealTimeInterval))
		}
	}()
}

func (x *ApiCache) updater() (*ApiUsage, error) {
	x.Lock()
	defer x.Unlock()
	a, err := GetApiUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to update API usage: %w", err)
	}
	x.apiUsage = a
	x.cond.Broadcast()
	return a, nil
}

func (x *ApiCache) isApiQuotaAvailable() bool {
	x.Lock()
	defer x.Unlock()
	if x.apiUsage == nil {
		x.cond.Wait()
	}
	return x.apiUsage.Remaining > 0
}

func (x *ServeCommand) verbose(format string, v ...any) {
	if x.Verbose {
		log.Printf(format, v...)
	}
}

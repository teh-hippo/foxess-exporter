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

type DeviceCache struct {
	sync.Mutex
	deviceIds []string
	cond      *sync.Cond
}

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
var deviceCache DeviceCache
var apiCache ApiCache
var metrics *serve.Metrics

func init() {
	if _, err := parser.AddCommand("serve", "Serve FoxESS metrics", "Creates a Prometheus endpoint where metrics can be provided.", &serveCommand); err != nil {
		panic(err)
	}
	deviceCache.cond = sync.NewCond(&deviceCache)
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

func (x *DeviceCache) updater(filtered map[string]bool) error {
	x.Lock()
	defer x.Unlock()
	if devices, err := GetDeviceList(); err != nil {
		return fmt.Errorf("Unable to update device list: %w", err)
	} else {
		var hasFilter = len(filtered) > 0
		if !hasFilter {
			x.deviceIds = make([]string, len(devices))
		}

		for i, device := range devices {
			if !hasFilter {
				x.deviceIds[i] = device.DeviceSerialNumber
			}
			if !hasFilter || filtered[device.DeviceSerialNumber] {
				log.Printf("Setting status of %s to: %d (%s)", device.DeviceSerialNumber, device.Status, device.status())
				metrics.Status.WithLabelValues(device.DeviceSerialNumber).Set(float64(device.Status))
			}
		}

		if !hasFilter {
			x.cond.Broadcast()
		}
	}
	return nil
}

func (x *DeviceCache) initalise(filtered map[string]bool) {
	if len(filtered) == 0 {
		return
	}
	x.Lock()
	defer x.Unlock()
	x.deviceIds = make([]string, 0, len(filtered))
	for deviceId := range filtered {
		x.deviceIds = append(x.deviceIds, deviceId)
	}
}

func (x *ServeCommand) Execute(args []string) error {
	deviceCache.initalise(serveCommand.Inverters)

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
				if err := deviceCache.updater(x.Inverters); err != nil {
					fmt.Println(err)
				}
			} else {
				x.verbose("No quota available to update device status")
			}
			time.Sleep(x.StatusInterval)
		}
	}()
}

func (x *ServeCommand) startRealTimeMetrics() {
	go func() {
		for {
			if apiCache.isApiQuotaAvailable() {
				x.verbose("Retrieving latest real-time data")
				if err := updateMetrics(); err != nil {
					fmt.Println(err)
				}
			} else {
				x.verbose("No quota available to update real-time metrics")
			}
			time.Sleep(time.Duration(x.RealTimeInterval))
		}
	}()
}

func updateMetrics() error {
	deviceCache.Lock()
	defer deviceCache.Unlock()
	if deviceCache.deviceIds == nil {
		deviceCache.cond.Wait()
	}
	return metrics.Updater(&foxessApi, deviceCache.deviceIds, serveCommand.Variables)
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

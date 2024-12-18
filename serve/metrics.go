package serve

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type Metrics struct {
	realtime        *prometheus.GaugeVec
	status          *prometheus.GaugeVec
	lastUpdatedTime map[string]time.Time
	Registry        *prometheus.Registry
}

func NewMetrics() *Metrics {
	metrics := &Metrics{
		status: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "foxess_device_status",
			Help: "Status of the inverter.",
		}, []string{"inverter"}),
		realtime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "foxess_realtime_data",
			Help: "Data from the FoxESS platform.",
		}, []string{"inverter", "variable"}),
		lastUpdatedTime: make(map[string]time.Time),
		Registry:        prometheus.NewRegistry(),
	}
	metrics.Registry.MustRegister(metrics.status)
	metrics.Registry.MustRegister(metrics.realtime)

	return metrics
}

func (x *Metrics) UpdateRealTime(data []foxess.RealTimeData) {
	for _, result := range data {
		if x.lastUpdatedTime[result.DeviceSN].Equal(result.Time.Time) {
			continue
		}

		log.Printf("Updating %d metric%s for %s, timestamp:%v.", len(result.Variables), util.Pluralise(len(result.Variables)), result.DeviceSN, result.Time.Time)
		x.lastUpdatedTime[result.DeviceSN] = result.Time.Time

		for _, variable := range result.Variables {
			x.realtime.WithLabelValues(result.DeviceSN, variable.Variable).Set(variable.Value.Number)
		}
	}
}

func (x *Metrics) UpdateStatus(devices []foxess.Device, include func(inverter string) bool) {
	for _, device := range devices {
		if include(device.DeviceSerialNumber) {
			x.status.WithLabelValues(device.DeviceSerialNumber).Set(float64(device.Status))
		}
	}
}

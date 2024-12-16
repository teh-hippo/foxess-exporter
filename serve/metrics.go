package serve

import (
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type Metrics struct {
	realtime        *prometheus.GaugeVec
	Status          *prometheus.GaugeVec
	lastUpdatedTime map[string]time.Time
	Registry        *prometheus.Registry
}

func NewMetrics() *Metrics {
	metrics := &Metrics{}
	metrics.Status = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "foxess_device_status",
		Help: "Status of the inverter.",
	}, []string{"inverter"})
	metrics.realtime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "foxess_realtime_data",
		Help: "Data from the FoxESS platform.",
	}, []string{"inverter", "variable"})
	metrics.lastUpdatedTime = make(map[string]time.Time)
	metrics.Registry = prometheus.NewRegistry()
	metrics.Registry.MustRegister(metrics.Status)
	metrics.Registry.MustRegister(metrics.realtime)
	return metrics
}

func (x *Metrics) Updater(foxessApi *foxess.FoxessParams, deviceIds []string, variables []string) error {
	data, err := foxessApi.GetRealTimeData(deviceIds, variables)
	if err != nil {
		return fmt.Errorf("Unable to retrieve latest real-time data: %w", err)
	}

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
	return nil
}

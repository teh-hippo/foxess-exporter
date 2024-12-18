package foxess_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teh-hippo/foxess-exporter/foxess"
)

func TestCurrentStatus(t *testing.T) {
	t.Parallel()

	testStatus := func(status int, expected string) {
		testDevice := &foxess.Device{
			Status:             status,
			DeviceSerialNumber: "1234567890",
			ModuleSerialNumber: "0987654321",
			StationID:          "1234567890",
			StationName:        "Test Station",
			HasPV:              true,
			HasBattery:         true,
			DeviceType:         "Test Device",
			ProductType:        "Test Product",
		}
		assert.Equal(t, expected, testDevice.CurrentStatus())
	}

	testStatus(foxess.StatusOnline, "Online")
	testStatus(foxess.StatusFault, "Fault")
	testStatus(foxess.StatusOffline, "Offline")
	testStatus(foxess.StatusOffline+1, "Unknown:4")
}

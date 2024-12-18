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
			Status: status,
		}
		assert.Equal(t, expected, testDevice.CurrentStatus())
	}

	testStatus(foxess.StatusOnline, "Online")
	testStatus(foxess.StatusFault, "Fault")
	testStatus(foxess.StatusOffline, "Offline")
	testStatus(foxess.StatusOffline+1, "Unknown:4")
}

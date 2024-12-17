package foxess

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentStatus(t *testing.T) {
	testStatus := func(status int, expected string) {
		testDevice := &Device{
			Status: status,
		}
		assert.Equal(t, expected, testDevice.CurrentStatus())
	}

	testStatus(DEVICES_STATUS_ONLINE, "Online")
	testStatus(DEVICES_STATUS_FAULT, "Fault")
	testStatus(DEVICES_STATUS_OFFLINE, "Offline")
	testStatus(DEVICES_STATUS_OFFLINE+1, "Unknown:4")
}

package foxess

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentStatus(t *testing.T) {
	t.Parallel()

	testStatus := func(status int, expected string) {
		testDevice := &Device{
			Status: status,
		}
		assert.Equal(t, expected, testDevice.CurrentStatus())
	}

	testStatus(StatusOnline, "Online")
	testStatus(StatusFault, "Fault")
	testStatus(StatusOffline, "Offline")
	testStatus(StatusOffline+1, "Unknown:4")
}

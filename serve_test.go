package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiUsageWontExceedAllowance(t *testing.T) {
	// Defaults
	serveCommand.RealTimeIntervalSec = 120
	serveCommand.StatusIntervalSec = 900
	actual := serveCommand.validateIntervals()
	assert.Nil(t, actual)

	// RealTimeIntervalSec too low
	serveCommand.RealTimeIntervalSec = 60
	serveCommand.StatusIntervalSec = 9999
	actual = serveCommand.validateIntervals()
	assert.NotNil(t, actual)

	// RealTimeIntervalSec too low
	serveCommand.RealTimeIntervalSec = 999999
	serveCommand.StatusIntervalSec = 60
	actual = serveCommand.validateIntervals()
	assert.NotNil(t, actual)
}

func TestUpdateIntervalsAreClamped(t *testing.T) {
	// RealTimeIntervalSec
	serveCommand.RealTimeIntervalSec = 59
	serveCommand.StatusIntervalSec = 61
	assert.NotNil(t, serveCommand.validateIntervals())
	assert.Equal(t, int64(60), serveCommand.RealTimeIntervalSec)
	assert.Equal(t, int64(61), serveCommand.StatusIntervalSec)

	// StatusIntervalSec
	serveCommand.RealTimeIntervalSec = 61
	serveCommand.StatusIntervalSec = 59
	assert.NotNil(t, serveCommand.validateIntervals())
	assert.Equal(t, int64(61), serveCommand.RealTimeIntervalSec)
	assert.Equal(t, int64(60), serveCommand.StatusIntervalSec)
}

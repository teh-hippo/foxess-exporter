package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApiUsageWontExceedAllowance(t *testing.T) {
	const longDelay time.Duration = 9999 * time.Minute

	// Defaults
	serveCommand.RealTimeInterval = 3 * time.Minute
	serveCommand.StatusInterval = 15 * time.Minute
	actual := serveCommand.validateIntervals()
	assert.Nil(t, actual)

	// RealTimeIntervalSec too low
	serveCommand.RealTimeInterval = time.Minute
	serveCommand.StatusInterval = longDelay
	actual = serveCommand.validateIntervals()
	assert.NotNil(t, actual)

	// RealTimeIntervalSec too low
	serveCommand.RealTimeInterval = longDelay
	serveCommand.StatusInterval = time.Minute
	actual = serveCommand.validateIntervals()
	assert.NotNil(t, actual)
}

func TestUpdateIntervalsAreClamped(t *testing.T) {
	const overOneMinute time.Duration = 61 * time.Second
	const underOneMinute time.Duration = 59 * time.Second

	// RealTimeInterval
	serveCommand.RealTimeInterval = underOneMinute
	serveCommand.StatusInterval = overOneMinute
	assert.NotNil(t, serveCommand.validateIntervals())
	assert.Equal(t, time.Minute, serveCommand.RealTimeInterval)
	assert.Equal(t, overOneMinute, serveCommand.StatusInterval)

	// StatusInterval
	serveCommand.RealTimeInterval = overOneMinute
	serveCommand.StatusInterval = underOneMinute
	assert.NotNil(t, serveCommand.validateIntervals())
	assert.Equal(t, overOneMinute, serveCommand.RealTimeInterval)
	assert.Equal(t, time.Minute, serveCommand.StatusInterval)
}

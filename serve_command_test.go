package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiUsageWontExceedAllowance(t *testing.T) {
	t.Parallel()
	const longDelay time.Duration = 9999 * time.Minute
	serveCommand := ServeCommand{}

	// Defaults
	serveCommand.RealTimeInterval = 3 * time.Minute
	serveCommand.StatusInterval = 15 * time.Minute
	actual := serveCommand.validateIntervals()
	require.NoError(t, actual)

	// RealTimeIntervalSec too low
	serveCommand.RealTimeInterval = time.Minute
	serveCommand.StatusInterval = longDelay
	actual = serveCommand.validateIntervals()
	require.Error(t, actual)

	// RealTimeIntervalSec too low
	serveCommand.RealTimeInterval = longDelay
	serveCommand.StatusInterval = time.Minute
	actual = serveCommand.validateIntervals()
	require.Error(t, actual)
}

func TestUpdateIntervalsAreClamped(t *testing.T) {
	t.Parallel()
	serveCommand := ServeCommand{}
	const overConfig time.Duration = 61 * time.Second
	const underConfig time.Duration = 59 * time.Second

	// RealTimeInterval
	serveCommand.RealTimeInterval = underConfig
	serveCommand.StatusInterval = overConfig
	require.Error(t, serveCommand.validateIntervals())
	assert.Equal(t, time.Minute, serveCommand.RealTimeInterval)
	assert.Equal(t, overConfig, serveCommand.StatusInterval)

	// StatusInterval
	serveCommand.RealTimeInterval = overConfig
	serveCommand.StatusInterval = underConfig
	require.Error(t, serveCommand.validateIntervals())
	assert.Equal(t, overConfig, serveCommand.RealTimeInterval)
	assert.Equal(t, time.Minute, serveCommand.StatusInterval)
}

func TestIncludeWithInverters(t *testing.T) {
	t.Parallel()
	serveCommand := ServeCommand{}
	const id1 = "1"
	const id2 = "2"
	serveCommand.Inverters = map[string]bool{
		id1: true,
	}
	assert.True(t, serveCommand.Include(id1))
	assert.False(t, serveCommand.Include(id2))
}

func TestIncludeWithoutInverters(t *testing.T) {
	t.Parallel()
	serveCommand := ServeCommand{}
	const id1 = "1"
	const id2 = "2"
	serveCommand.Inverters = map[string]bool{}
	assert.True(t, serveCommand.Include(id1))
	assert.True(t, serveCommand.Include(id2))
}

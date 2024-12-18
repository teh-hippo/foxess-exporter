package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	overConfig  time.Duration = 61 * time.Second
	underConfig time.Duration = 59 * time.Second
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

func TestRealTimeIntervalIsClamped(t *testing.T) {
	t.Parallel()

	serveCommand := ServeCommand{
		RealTimeInterval: underConfig,
		StatusInterval:   overConfig,
	}
	require.Error(t, serveCommand.validateIntervals())
	assert.Equal(t, time.Minute, serveCommand.RealTimeInterval)
	assert.Equal(t, overConfig, serveCommand.StatusInterval)
}

func TestStatusIntervalIsClamped(t *testing.T) {
	t.Parallel()

	serveCommand := ServeCommand{
		RealTimeInterval: overConfig,
		StatusInterval:   underConfig,
	}
	require.Error(t, serveCommand.validateIntervals())
	assert.Equal(t, overConfig, serveCommand.RealTimeInterval)
	assert.Equal(t, time.Minute, serveCommand.StatusInterval)
}

func TestIncludeWithInverters(t *testing.T) {
	t.Parallel()

	const (
		id1 = "1"
		id2 = "2"
	)

	serveCommand := ServeCommand{
		Inverters: map[string]bool{id1: true},
	}
	assert.True(t, serveCommand.Include(id1))
	assert.False(t, serveCommand.Include(id2))
}

func TestIncludeWithoutInverters(t *testing.T) {
	t.Parallel()

	const (
		id1 = "1"
		id2 = "2"
	)

	serveCommand := ServeCommand{
		Inverters: map[string]bool{},
	}

	assert.True(t, serveCommand.Include(id1))
	assert.True(t, serveCommand.Include(id2))
}

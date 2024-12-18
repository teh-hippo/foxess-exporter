package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/serve"
)

const (
	overConfig  time.Duration = 61 * time.Second
	underConfig time.Duration = 59 * time.Second
)

func TestApiUsageWontExceedAllowance(t *testing.T) {
	t.Parallel()

	const longDelay time.Duration = 9999 * time.Minute

	serveCommand := buildSubject()

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

	serveCommand := buildSubject()
	serveCommand.RealTimeInterval = underConfig
	serveCommand.StatusInterval = overConfig
	require.Error(t, serveCommand.validateIntervals())
	assert.Equal(t, time.Minute, serveCommand.RealTimeInterval)
	assert.Equal(t, overConfig, serveCommand.StatusInterval)
}

func TestStatusIntervalIsClamped(t *testing.T) {
	t.Parallel()

	serveCommand := buildSubject()
	serveCommand.RealTimeInterval = overConfig
	serveCommand.StatusInterval = underConfig
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

	serveCommand := buildSubject()
	serveCommand.Inverters = map[string]bool{id1: true}
	assert.True(t, serveCommand.Include(id1))
	assert.False(t, serveCommand.Include(id2))
}

func buildSubject() *ServeCommand {
	return &ServeCommand{
		Inverters:   map[string]bool{},
		deviceCache: serve.NewDeviceCache(),
		apiQuota:    serve.NewAPIQuota(),
		metrics:     serve.NewMetrics(),
		config: &foxess.Config{
			APIKey: "key",
			Debug:  false,
		},
		Port:             1234,
		Variables:        []string{},
		RealTimeInterval: 5 * time.Minute,
		StatusInterval:   10 * time.Minute,
		Verbose:          false,
	}
}

func TestIncludeWithoutInverters(t *testing.T) {
	t.Parallel()

	const (
		id1 = "1"
		id2 = "2"
	)

	serveCommand := buildSubject()

	assert.True(t, serveCommand.Include(id1))
	assert.True(t, serveCommand.Include(id2))
}

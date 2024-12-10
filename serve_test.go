package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClamp(t *testing.T) {
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

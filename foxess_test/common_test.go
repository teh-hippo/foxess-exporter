package foxess_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teh-hippo/foxess-exporter/foxess"
)

func TestCalculateSignature(t *testing.T) {
	t.Parallel()

	want := "68a007c2450d6697fbe2990f92000269"
	got := foxess.CalculateSignature("/op/v0/device/list", "abcdefghij012345689", 1705809089)

	assert.Equal(t, want, got)
}

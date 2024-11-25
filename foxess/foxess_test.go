package foxess

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelloName(t *testing.T) {
	want := "68a007c2450d6697fbe2990f92000269"
	got := CalculateSignature("/op/v0/device/list", "abcdefghij012345689", 1705809089)
	assert.Equal(t, want, got)
}

package util

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type DummyRequest struct {
	SerialNumber string `json:"sn"`
}

func TestToReader(t *testing.T) {
	want := []byte("{\"sn\":\"1234567890\"}")
	result, err := ToReader(&DummyRequest{SerialNumber: "1234567890"})
	if err != nil {
		assert.Error(t, err)
	}
	got, err := io.ReadAll(result)
	if err != nil {
		assert.Error(t, err)
	}
	assert.Equal(t, want, got)
}

func TestPlural(t *testing.T) {
	assert.Equal(t, "", Pluralise(1))
	assert.Equal(t, "s", Pluralise(0))
	assert.Equal(t, "s", Pluralise(2))
	assert.Equal(t, "s", Pluralise(789))
}

func TestClamp(t *testing.T) {
	const mid time.Duration = 5
	const max time.Duration = 10
	const min time.Duration = 1
	assert.Equal(t, mid, Clamp(mid, min, max))
	assert.Equal(t, min, Clamp(min-1, min, max))
	assert.Equal(t, max, Clamp(max+1, min, max))
}

package util

import (
	"io"
	"testing"

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
	assert.Equal(t, int64(10), Clamp(10, 0))
	assert.Equal(t, int64(0), Clamp(-1, 0))
	assert.Equal(t, int64(0), Clamp(0, 0))
}

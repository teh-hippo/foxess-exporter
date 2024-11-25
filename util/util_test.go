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

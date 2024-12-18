package util_test

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teh-hippo/foxess-exporter/util"
)

type DummyRequest struct {
	SerialNumber string `json:"sn"`
}

func TestToReader(t *testing.T) {
	t.Parallel()

	want := []byte("{\"sn\":\"1234567890\"}")

	result, err := util.ToReader(&DummyRequest{SerialNumber: "1234567890"})
	if err != nil {
		require.Error(t, err)
	}

	got, err := io.ReadAll(result)
	if err != nil {
		require.Error(t, err)
	}

	assert.Equal(t, want, got)
}

func TestPlural(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", util.Pluralise(1))
	assert.Equal(t, "s", util.Pluralise(0))
	assert.Equal(t, "s", util.Pluralise(2))
	assert.Equal(t, "s", util.Pluralise(789))
}

func TestClamp(t *testing.T) {
	t.Parallel()

	const (
		mid      time.Duration = 5
		maxLimit time.Duration = 10
		minLimit time.Duration = 1
	)

	assert.Equal(t, mid, util.Clamp(mid, minLimit, maxLimit))
	assert.Equal(t, minLimit, util.Clamp(minLimit-1, minLimit, maxLimit))
	assert.Equal(t, maxLimit, util.Clamp(maxLimit+1, minLimit, maxLimit))
}

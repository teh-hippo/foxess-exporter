package serve

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teh-hippo/foxess-exporter/foxess"
)

func TestQuotaAvailable(t *testing.T) {
	var subject = NewApiCache()
	assertThat := func(remaining float64, expected bool) {
		subject.Set(&foxess.ApiUsage{
			Remaining: remaining,
		})
		assert.Equal(t, expected, subject.IsQuotaAvailable())
	}

	assertThat(0, false)
	assertThat(1, true)
}

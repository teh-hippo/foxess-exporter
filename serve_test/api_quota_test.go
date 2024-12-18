package serve_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/serve"
)

func TestQuotaAvailable(t *testing.T) {
	t.Parallel()

	subject := serve.NewAPIQuota()
	assertThat := func(remaining float64, expected bool) {
		subject.Set(&foxess.APIUsage{
			Remaining: remaining,
		})
		assert.Equal(t, expected, subject.IsQuotaAvailable())
	}

	assertThat(0, false)
	assertThat(1, true)
}

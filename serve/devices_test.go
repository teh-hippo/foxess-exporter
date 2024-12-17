package serve

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanSet(t *testing.T) {
	t.Parallel()
	var subject = NewDeviceCache()
	const id1 = "1"
	const id2 = "2"

	var expected1 = []string{id1, id2}

	subject.Set(expected1)
	var actual1 = subject.Get()
	assert.Equal(t, expected1, actual1)

	var expected2 = []string{"3", "4"}

	subject.Set(expected2)
	var actual2 = subject.Get()
	assert.Equal(t, expected2, actual2)
}

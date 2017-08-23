package astibundler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPath(t *testing.T) {
	_, err := buildPath("/1/2")
	assert.Error(t, err)
	p, err := buildPath("/1/2/3")
	assert.NoError(t, err)
	assert.Equal(t, "1/2/3", p)
	p, err = buildPath("/1/2/3/4")
	assert.NoError(t, err)
	assert.Equal(t, "2/3/4", p)
}

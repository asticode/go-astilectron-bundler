package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLDFlags(t *testing.T) {
	l := LDFlags{}

	assert.Equal(t, l.String(), "")

	l.Set("X:main.foo=1")

	assert.Equal(t, "-X main.foo=1", l.String())

	l.Set("X:main.bar=1,main.baz=2")

	assert.Equal(t, "-X main.foo=1 -X main.bar=1 -X main.baz=2", l.String())

	l.Set("s")

	assert.Equal(t, "-X main.foo=1 -X main.bar=1 -X main.baz=2 -s", l.String())
}

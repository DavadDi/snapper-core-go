package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// IDToSource ...
func TestIDToSource(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(IDToSource("a.xx"), "android")
	assert.Equal(IDToSource("i.xx"), "ios")
	assert.Equal(IDToSource("c.xx"), "web")
}

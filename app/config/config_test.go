package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_IsProd(t *testing.T) {
	assert.True(t, (&Config{Environment: "prod"}).IsProd())
	assert.False(t, (&Config{Environment: "dev"}).IsProd())
	assert.False(t, (&Config{Environment: ""}).IsProd())
	assert.False(t, (&Config{Environment: "production"}).IsProd(), "仅精确匹配 prod")
}

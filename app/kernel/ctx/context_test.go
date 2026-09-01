package ctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/logx"
)

func TestNewContext(t *testing.T) {
	base := context.Background()
	nc := NewContext(base)

	assert.NotNil(t, nc.Value(KeyLogger), "应注入 logx.Logger")

	logger, ok := nc.Value(KeyLogger).(logx.Logger)
	assert.True(t, ok, "注入的值应为 logx.Logger 类型")
	assert.NotNil(t, logger)
}

func TestContextContainer_Logger_Injected(t *testing.T) {
	nc := NewContext(context.Background())

	logger := NewContextContainer(nc).Logger()

	assert.NotNil(t, logger)
	// 注入后取回的是同一个 logger 实例，而非回退新建
	assert.Same(t, nc.Value(KeyLogger), logger)
}

func TestContextContainer_Logger_Fallback(t *testing.T) {
	// 未注入 logger 时回退到 logx.WithContext，不 panic
	logger := NewContextContainer(context.Background()).Logger()

	assert.NotNil(t, logger)
}

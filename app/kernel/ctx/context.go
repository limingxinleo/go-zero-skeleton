package ctx

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

const KeyLogger = "ctx.logger"

type ContextContainer struct {
	ctx context.Context
}

func NewContextContainer(ctx context.Context) *ContextContainer {
	return &ContextContainer{ctx: ctx}
}

func (c *ContextContainer) Logger() logx.Logger {
	return Logger(c.ctx)
}

func NewContext(cc context.Context) context.Context {
	return context.WithValue(cc, KeyLogger, logx.WithContext(cc))
}

func Logger(cc context.Context) logx.Logger {
	if logger, ok := cc.Value(KeyLogger).(logx.Logger); ok {
		return logger
	}

	return logx.WithContext(cc)
}

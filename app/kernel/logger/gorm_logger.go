package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm/utils"
)
import gl "gorm.io/gorm/logger"

type GormLogger struct {
	gl.Config
}

func NewGormLogger() *GormLogger {
	l := &GormLogger{}

	l.LogLevel = gl.Info
	l.Colorful = false

	return l
}

func (l *GormLogger) log(ctx context.Context) logx.Logger {
	return logx.WithContext(ctx)
}

func (l *GormLogger) LogMode(level gl.LogLevel) gl.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	if l.LogLevel >= gl.Info {
		l.log(ctx).Infof("[info] "+s, i...)
	}
}

func (l *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	if l.LogLevel >= gl.Warn {
		l.log(ctx).Infof("[warn] "+s, i...)
	}
}

func (l *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	if l.LogLevel >= gl.Error {
		l.log(ctx).Errorf("[error] "+s, i...)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gl.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gl.Error && (!errors.Is(err, gl.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.log(ctx).Errorf("%s %s [%.3fms] [rows:%v] %s", utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.log(ctx).Errorf("%s %s [%.3fms] [rows:%v] %s", utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gl.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.log(ctx).Errorf("%s %s [%.3fms] [rows:%v] %s", utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.log(ctx).Errorf("%s %s [%.3fms] [rows:%v] %s", utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == gl.Info:
		sql, rows := fc()
		if rows == -1 {
			l.log(ctx).Errorf("%s [%.3fms] [rows:%v] %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.log(ctx).Errorf("%s [%.3fms] [rows:%v] %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

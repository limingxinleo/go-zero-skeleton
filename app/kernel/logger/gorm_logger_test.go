package logger

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gl "gorm.io/gorm/logger"
)

func TestNewGormLogger(t *testing.T) {
	l := NewGormLogger()

	assert.Equal(t, gl.Info, l.LogLevel)
	assert.False(t, l.Colorful)
}

func TestGormLogger_LogMode(t *testing.T) {
	l := NewGormLogger()
	silent := l.LogMode(gl.Silent)

	si, ok := silent.(*GormLogger)
	assert.True(t, ok, "LogMode 应返回 *GormLogger")
	assert.Equal(t, gl.Info, l.LogLevel, "原实例级别不应被修改")
	assert.Equal(t, gl.Silent, si.LogLevel, "应返回调整级别后的新实例")
}

func TestGormLogger_InfoWarnError(t *testing.T) {
	l := NewGormLogger()

	// Info 级别下三个方法均输出，仅验证不 panic
	l.Info(context.Background(), "info msg: %s", "x")
	l.Warn(context.Background(), "warn msg: %s", "x")
	l.Error(context.Background(), "error msg: %s", "x")

	silent := l.LogMode(gl.Silent)
	silent.Info(context.Background(), "should not log")
	silent.Warn(context.Background(), "should not log")
}

func TestGormLogger_Trace_Silent(t *testing.T) {
	l := NewGormLogger().LogMode(gl.Silent)

	called := false
	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		called = true
		return "SELECT 1", 1
	}, nil)

	assert.False(t, called, "Silent 级别下不应执行 SQL 取值函数")
}

func TestGormLogger_Trace_Error(t *testing.T) {
	l := NewGormLogger()

	// 出错分支：rows=-1 与 rows>=0 两种格式均不 panic
	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT * FROM `user`", 0
	}, errors.New("dial timeout"))

	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "INSERT INTO `user`", -1
	}, errors.New("dial timeout"))
}

func TestGormLogger_Trace_SlowSql(t *testing.T) {
	l := NewGormLogger()
	l.SlowThreshold = time.Millisecond

	// 超过慢查询阈值
	l.Trace(context.Background(), time.Now().Add(-time.Second), func() (string, int64) {
		return "SELECT SLEEP(1)", 1
	}, nil)
}

func TestGormLogger_Trace_Info(t *testing.T) {
	l := NewGormLogger()

	// 正常分支：rows>=0 与 rows=-1 两种格式均不 panic
	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "UPDATE `user`", -1
	}, nil)
}

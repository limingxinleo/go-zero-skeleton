package constants

import (
	"errors"
	"testing"

	"github.com/limingxinleo/go-zero-skeleton/app/kernel"
	"github.com/stretchr/testify/assert"
)

// 编译期断言：ErrorCode 必须满足 kernel 统一错误码契约
var _ kernel.ErrorCodeInterface = (*ErrorCode)(nil)

func TestErrorCode_GetCode(t *testing.T) {
	assert.Equal(t, 500, ServerError.GetCode())
}

func TestErrorCode_GetMessage(t *testing.T) {
	assert.Equal(t, "服务器内部错误", ServerError.GetMessage())
}

func TestErrorCode_Error(t *testing.T) {
	assert.Equal(t, "服务器内部错误", ServerError.Error())
}

func TestErrorCode_Err(t *testing.T) {
	e := &ErrorCode{Code: 1001, Message: "测试错误"}

	assert.Nil(t, e.Err())

	underlying := errors.New("connection refused")
	e.WithError(underlying)

	assert.Equal(t, underlying, e.Err())
}

func TestErrorCode_WithError(t *testing.T) {
	e := &ErrorCode{Code: 1001, Message: "测试错误"}
	underlying := errors.New("dial timeout")

	got := e.WithError(underlying)

	// 返回自身以便链式调用
	assert.Same(t, e, got)
	// 底层错误只进日志，不改变对外 Code 与 Message
	assert.Equal(t, underlying, got.Err())
	assert.Equal(t, 1001, got.GetCode())
	assert.Equal(t, "测试错误", got.GetMessage())
}

func TestErrorCode_WithMessage(t *testing.T) {
	e := &ErrorCode{Code: 1001, Message: "测试错误"}

	got := e.WithMessage("参数错误")

	assert.Same(t, e, got)
	assert.Equal(t, "参数错误", got.GetMessage())
	assert.Equal(t, "参数错误", got.Error())
	// Message 覆盖不影响 Code
	assert.Equal(t, 1001, got.GetCode())
}

func TestErrorCode_Chain(t *testing.T) {
	underlying := errors.New("sql: no rows in result set")
	e := (&ErrorCode{Code: 1002, Message: "原始提示"}).
		WithError(underlying).
		WithMessage("查询失败")

	assert.Equal(t, 1002, e.GetCode())
	assert.Equal(t, "查询失败", e.GetMessage())
	assert.Equal(t, underlying, e.Err())
}

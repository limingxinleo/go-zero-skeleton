package constants

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/limingxinleo/go-zero-skeleton/app/kernel"
	"github.com/stretchr/testify/assert"
)

// 编译期断言：ErrorCode 必须满足 kernel 统一错误码契约
var _ kernel.ErrorCode = (*ErrorCode)(nil)

func TestErrorCode_Code(t *testing.T) {
	assert.Equal(t, 500, ServerError.Code())
}

func TestErrorCode_Message(t *testing.T) {
	assert.Equal(t, "服务器内部错误", ServerError.Message())
}

func TestErrorCode_Error(t *testing.T) {
	assert.Equal(t, "服务器内部错误", ServerError.Error())
}

func TestErrorCode_Unwrap(t *testing.T) {
	e := NewErrorCode(1001, "测试错误")

	assert.NoError(t, e.Unwrap())

	underlying := errors.New("connection refused")
	got := e.WithError(underlying)

	assert.Equal(t, underlying, got.Unwrap())
	// 原错误码未被修改
	assert.NoError(t, e.Unwrap())
}

func TestErrorCode_WithError(t *testing.T) {
	e := NewErrorCode(1001, "测试错误")
	underlying := errors.New("dial timeout")

	got := e.WithError(underlying)

	// 返回新副本而非修改原错误码，全局单例在并发请求间互不覆盖
	assert.NotSame(t, e, got)
	// 底层错误只进日志，不改变对外 Code 与 Message
	assert.Equal(t, underlying, got.Unwrap())
	assert.Equal(t, 1001, got.Code())
	assert.Equal(t, "测试错误", got.Message())
}

func TestErrorCode_WithMessage(t *testing.T) {
	e := NewErrorCode(1001, "测试错误")

	got := e.WithMessage("参数错误")

	assert.NotSame(t, e, got)
	assert.Equal(t, "参数错误", got.Message())
	assert.Equal(t, "参数错误", got.Error())
	// Message 覆盖不影响 Code，且原错误码保持不变
	assert.Equal(t, 1001, got.Code())
	assert.Equal(t, "测试错误", e.Message())
}

func TestErrorCode_Chain(t *testing.T) {
	underlying := errors.New("sql: no rows in result set")
	e := NewErrorCode(1002, "原始提示").
		WithError(underlying).
		WithMessage("查询失败")

	assert.Equal(t, 1002, e.Code())
	assert.Equal(t, "查询失败", e.Message())
	assert.Equal(t, underlying, e.Unwrap())
}

// errors.Is 按错误码匹配：WithError / WithMessage 产生的副本仍识别为同一错误码，
// 不会与其他错误码或普通 error 误匹配。
func TestErrorCode_ErrorsIs(t *testing.T) {
	paramsError := NewErrorCode(1001, "参数错误")

	err := paramsError.WithMessage("id 不能为空")

	assert.ErrorIs(t, err, paramsError)
	assert.False(t, errors.Is(err, ServerError))
	assert.False(t, errors.Is(err, errors.New("id 不能为空")))
}

// 全局单例在并发 WithMessage 下不得互相覆盖（需以 go test -race 运行方可检出数据竞争）。
func TestErrorCode_ConcurrentSafe(t *testing.T) {
	base := NewErrorCode(1000, "基准提示")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = base.WithMessage(fmt.Sprintf("并发消息 %d", n))
		}(i)
	}
	wg.Wait()

	assert.Equal(t, "基准提示", base.Message())
}

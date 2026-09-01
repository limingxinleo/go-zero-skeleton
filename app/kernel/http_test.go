package kernel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/limingxinleo/go-zero-skeleton/app/constants"
	"github.com/limingxinleo/go-zero-skeleton/app/types"
	"github.com/stretchr/testify/assert"
)

func TestSend_Success(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	Send(w, r, "hello", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var body types.Response[string]
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.Equal(t, 0, body.Code)
	assert.Equal(t, "hello", body.Data)
	assert.Equal(t, "", body.Message)
}

func TestSend_Error(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	Send(w, r, nil, constants.ServerError.WithMessage("查询失败"))

	assert.Equal(t, http.StatusOK, w.Code)

	var body types.Response[any]
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.Equal(t, 500, body.Code)
	assert.Equal(t, "查询失败", body.Message)
	assert.Nil(t, body.Data)
}

func TestSend_ErrorWithUnderlying(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	// 底层错误只进日志，不暴露给客户端
	Send(w, r, nil, (&constants.ErrorCode{Code: 1001, Message: "参数错误"}))

	var body types.Response[any]
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.Equal(t, 1001, body.Code)
	assert.Equal(t, "参数错误", body.Message)
	assert.NotContains(t, w.Body.String(), "boom")
}

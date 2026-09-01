package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/limingxinleo/go-zero-skeleton/app/config"
	"github.com/limingxinleo/go-zero-skeleton/app/svc"
	"github.com/limingxinleo/go-zero-skeleton/app/types"
	"github.com/stretchr/testify/assert"
)

// 手工构造 ServiceContext，不 import app 包，因此无需 MySQL/Redis 即可测试 Handler 全链路
func newTestServiceContext() *svc.ServiceContext {
	cfg := config.Config{}
	cfg.Name = "main-api"

	return svc.NewServiceContext(cfg)
}

func TestIndexHandler(t *testing.T) {
	handler := IndexHandler(newTestServiceContext())

	r := httptest.NewRequest(http.MethodGet, "/?name=test", nil)
	w := httptest.NewRecorder()
	handler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var body types.Response[string]
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.Equal(t, 0, body.Code)
	assert.Equal(t, "Hi test, welcome to main-api", body.Data)
	assert.Equal(t, "", body.Message)
}

func TestIndexHandler_DefaultName(t *testing.T) {
	// form 标签声明 name,optional,default=world，缺省时应取默认值
	handler := IndexHandler(newTestServiceContext())

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler(w, r)

	var body types.Response[string]
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.Equal(t, 0, body.Code)
	assert.Equal(t, "Hi world, welcome to main-api", body.Data)
}

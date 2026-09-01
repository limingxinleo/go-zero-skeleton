package kernel

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/limingxinleo/go-zero-skeleton/app/kernel/ctx"
	"github.com/stretchr/testify/assert"
)

func TestServerMiddleware(t *testing.T) {
	var (
		called       bool
		ctxHasLogger bool
	)

	next := func(w http.ResponseWriter, r *http.Request) {
		called = true
		ctxHasLogger = r.Context().Value(ctx.KeyLogger) != nil
		w.WriteHeader(http.StatusOK)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	ServerMiddleware(next)(w, r)

	assert.True(t, called, "next handler 应被执行")
	assert.True(t, ctxHasLogger, "请求 context 应注入 logger")
	assert.Equal(t, "go-zero", w.Header().Get("Server"))
}

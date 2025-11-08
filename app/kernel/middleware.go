package kernel

import (
	"github.com/limingxinleo/go-zero-skeleton/app/kernel/ctx"
	"net/http"
)

func ServerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(ctx.NewContext(r.Context()))
		next(w, r)
		w.Header().Add("Server", "go-zero")
	}
}

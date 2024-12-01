package kernel

import (
	"net/http"
)

func ServerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Server", "go-zero")
		next(w, r)
	}
}

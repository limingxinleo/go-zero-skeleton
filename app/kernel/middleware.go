package kernel

import (
	"net/http"
)

func ServerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
		w.Header().Add("Server", "go-zero")
	}
}

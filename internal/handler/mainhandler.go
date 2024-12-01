package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"main/internal/logic"
	"main/internal/svc"
	"main/internal/types"
)

func MainHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FromRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewMainLogic(r.Context(), svcCtx)
		resp, err := l.Main(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

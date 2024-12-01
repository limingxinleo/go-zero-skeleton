package handler

import (
	"main/app/kernel"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"main/app/logic"
	"main/app/svc"
	"main/app/types"
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
		kernel.Send(w, r, resp, err)
	}
}

package controller

import (
	"github.com/limingxinleo/go-zero-skeleton/app/kernel"
	"net/http"

	"github.com/limingxinleo/go-zero-skeleton/app/service"
	"github.com/limingxinleo/go-zero-skeleton/app/svc"
	"github.com/limingxinleo/go-zero-skeleton/app/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func IndexHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FromRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := service.NewIndexService(r.Context(), svcCtx)
		resp, err := l.Index(&req)
		kernel.Send(w, r, resp, err)
	}
}

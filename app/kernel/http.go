package kernel

import (
	"net/http"

	"github.com/limingxinleo/go-zero-skeleton/app/types"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func Send(w http.ResponseWriter, r *http.Request, resp any, err ErrorCode) {
	var body types.Response[any]
	if err != nil {
		if unwrapped := err.Unwrap(); unwrapped != nil {
			// 底层错误只进日志（携带 trace），不暴露给客户端
			logx.WithContext(r.Context()).Errorv(unwrapped)
		}

		body = types.Response[any]{
			Code:    err.Code(),
			Message: err.Message(),
			TraceId: trace.TraceIDFromContext(r.Context()),
		}
	} else {
		body = types.Response[any]{
			Code:    0,
			Data:    resp,
			TraceId: trace.TraceIDFromContext(r.Context()),
		}
	}

	httpx.OkJsonCtx(r.Context(), w, body)
}

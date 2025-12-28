package kernel

import (
	"net/http"

	"github.com/limingxinleo/go-zero-skeleton/app/types"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func Send(w http.ResponseWriter, r *http.Request, resp any, err ErrorCodeInterface) {
	var body types.Response[any]
	if err != nil {
		if err.Err() != nil {
			logx.ErrorStack(err.Err())
		}

		body = types.Response[any]{
			Code:    err.GetCode(),
			Message: err.GetMessage(),
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

package kernel

import (
	"github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/rest/httpx"
	"main/app/types"
	"net/http"
)

func Send(w http.ResponseWriter, r *http.Request, resp any, err ErrorCodeInterface) {
	var body types.Response[any]
	if err != nil {
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

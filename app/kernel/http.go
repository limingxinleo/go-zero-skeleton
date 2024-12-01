package kernel

import (
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
		}
	} else {
		body = types.Response[any]{
			Code: 0,
			Data: resp,
		}
	}

	httpx.OkJsonCtx(r.Context(), w, body)
}

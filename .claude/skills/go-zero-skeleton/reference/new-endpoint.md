# 新增 HTTP 接口

以新增 `GET /user/info`（查询参数 `id`）为例，共 5 步。所有代码**手写**（骨架布局与 goctl 标准布局不同，禁止用 goctl 生成覆盖）。以下模板与现有 `index` 接口风格完全一致，照抄替换即可。

## 1. 更新接口契约 `main.api`

```api
syntax = "v1"

type UserInfoRequest {
    Id uint64 `form:"id"`
}

type UserInfoResponse {
    NickName string `json:"nick_name"`
}

service main-api {
    @handler UserInfo
    get /user/info (UserInfoRequest) returns (UserInfoResponse)
}
```

`main.api` 仅作为契约文档，不参与编译，需与后续手写代码保持同步。

## 2. 定义 DTO `app/types/`

可追加到 `types.go`，或按模块新建 `app/types/user.go`（`package types`）：

```go
package types

type UserInfoRequest struct {
	Id uint64 `form:"id"`
}

type UserInfoResponse struct {
	NickName string `json:"nick_name"`
}
```

绑定 tag 规范（由 `httpx.Parse` 处理）：
- 查询参数用 `form:"name,optional,default=xxx"`；路径参数用 `path:"name"`；JSON body 用 `json:"name"`。
- `optional` 可选、`default=` 默认值、`options=a|b` 枚举校验。
- 注意：`types.go` 中的 `Response[T]` 为统一响应包装，由 `kernel.Send` 使用，勿动。

## 3. 编写 Service `app/service/user.go`

```go
package service

import (
	"context"

	"github.com/limingxinleo/go-zero-skeleton/app/constants"
	"github.com/limingxinleo/go-zero-skeleton/app/kernel"
	"github.com/limingxinleo/go-zero-skeleton/app/svc"
	"github.com/limingxinleo/go-zero-skeleton/app/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserService struct {
	log    logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserService(ctx context.Context, svcCtx *svc.ServiceContext) *UserService {
	return &UserService{
		log:    logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserService) Info(req *types.UserInfoRequest) (result *types.UserInfoResponse, err kernel.ErrorCodeInterface) {
	if req.Id == 0 {
		return nil, constants.ParamsError.WithMessage("id 不能为空")
	}

	// TODO: 查询数据库填充数据
	result = &types.UserInfoResponse{NickName: "xxx"}
	return result, nil
}
```

要点：
- 返回值错误类型固定为 `kernel.ErrorCodeInterface`，业务失败返回 `app/constants` 中定义的错误码。
- 日志一律用 `l.log`（已携带 trace 的 context），不要直接调 `logx.Info`。

若需新增错误码，在 `app/constants/error_code.go` 追加：

```go
var ParamsError = &ErrorCode{Code: 1001, Message: "参数错误"}
```

## 4. 编写 Handler `app/controller/user_controller.go`

```go
package controller

import (
	"net/http"

	"github.com/limingxinleo/go-zero-skeleton/app/kernel"
	"github.com/limingxinleo/go-zero-skeleton/app/service"
	"github.com/limingxinleo/go-zero-skeleton/app/svc"
	"github.com/limingxinleo/go-zero-skeleton/app/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserInfoRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := service.NewUserService(r.Context(), svcCtx)
		resp, err := l.Info(&req)
		kernel.Send(w, r, resp, err)
	}
}
```

参数解析失败统一走 `httpx.ErrorCtx`；业务结果统一走 `kernel.Send`。

## 5. 注册路由 `app/controller/routes.go`

在 `RegisterHandlers` 的 `[]rest.Route` 中追加：

```go
{
	Method:  http.MethodGet,
	Path:    "/user/info",
	Handler: UserInfoHandler(serverCtx),
},
```

多组路由可用 `server.AddRoutes(...)` 添加新的代码块（支持按 `rest.WithJwt()` 等挂中间件选项）。

## 6.（可选）挂载依赖到 ServiceContext

若 Service 需要数据库 dao 等全局依赖，将其加入 `app/svc/servicecontext.go`，详见 `reference/database.md`。

完成后编写单元测试，见 `reference/testing-local.md`。

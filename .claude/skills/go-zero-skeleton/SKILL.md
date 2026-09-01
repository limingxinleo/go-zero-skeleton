---
name: go-zero-skeleton
description: 基于 go-zero + Gorm + Redis 的 API 骨架开发指南。在此项目中新增 HTTP 接口、编写 service 业务逻辑、生成或使用数据库模型、新增 GRPC 服务、添加配置项或错误码、编写单元测试、本地运行调试、使用 go-gen 脚手架生成代码文件，以及从骨架初始化新项目时使用。
---

# go-zero-skeleton 开发指南

基于 go-zero（REST 框架）+ Gorm + Redis 的单体 API 骨架。业务代码全部在 `app/` 下分层组织。

## 架构总览

```
.
├── main.go              # HTTP 服务入口（rest.Server + 全局中间件）
├── main.api             # 接口契约文档（goctl api 格式，手工维护）
├── cmd/main.go          # CLI 入口（cobra，gen:model 等命令）
├── etc/                 # 配置：main-api.yaml（默认）、unit-api.yaml（CI 单测）
│                        # 环境变量 ROOT_PATH / CONFIG_PATH 可覆盖路径
└── app/
    ├── bootstrap.go     # init() 初始化 Application —— import app 包即触发
    ├── application.go   # Application 单例：Config / ServiceContext / MySQL / Gorm / Redis
    ├── config/          # 配置结构体（Config 组合 rest.RestConf + Redis + MySQL DSN）
    ├── controller/      # HTTP Handler 与路由注册 routes.go
    ├── types/           # 请求/响应 DTO
    ├── service/         # 业务逻辑层（单元测试也放在此，与被测文件同目录）
    ├── svc/             # ServiceContext 依赖容器
    ├── constants/       # 错误码等业务常量
    ├── dao/             # gen:model 生成产物：model/ + query/（勿手改）
    ├── model/           # goctl model 原生方式生成产物（可选）
    └── kernel/          # 框架支撑：Send 统一响应 / 错误码接口 / 中间件 / ctx / gorm logger
```

请求流程：`routes.go` 分发 → `Handler`（`httpx.Parse` 解析参数）→ `Service`（业务逻辑）→ `dao/model`（数据访问），Handler 最终通过 `kernel.Send` 统一输出。

## 关键机制（所有任务必读）

### 包名以项目实际为准
本文档及各 reference 示例代码中的 import 路径均以骨架原始 module 名 `github.com/limingxinleo/go-zero-skeleton` 为例。实际使用方可能已改名为自己的包名：**凡示例中的 import 前缀与项目 `go.mod` 的 `module` 行不符时，一律以 `go.mod` 为准**，将 `github.com/limingxinleo/go-zero-skeleton` 整体替换为项目实际 module 名。从骨架初始化新项目的改名步骤见 `reference/init-project.md`。

### Application 单例与 import 副作用
`app/bootstrap.go` 的 `init()` 会依次：定位根目录 → 加载配置 → 创建 ServiceContext → 连接 MySQL（sqlx 与 Gorm 双实例）→ 连接 Redis。
因此**任何 import `app` 包的代码（包括单元测试）运行时都要求 MySQL/Redis 可达**，连接失败会直接 `log.Fatalf`。

全局获取方式：`app.GetApplication()`，返回 `.Config` / `.ServiceContext` / `.MySQL`（sqlx）/ `.Gorm` / `.Redis`。

### 统一响应格式（强制）
所有接口必须通过 `kernel.Send(w, r, resp, err)` 输出（见 `app/kernel/http.go`），固定结构：

```json
{"code": 0, "data": {}, "message": "", "trace_id": "xxx"}
```

成功 `code=0`；禁止在 Handler 中手写 `httpx.OkJson` / `w.Write`。

### 错误处理（强制）
- Service 方法签名统一返回 `(result T, err kernel.ErrorCode)`，不返回裸 `error`。`kernel.ErrorCode` 实现了标准 `error` 接口（含 `Unwrap`/`Is`），可直接用 `errors.Is(err, constants.XxxError)` 按错误码判断。
- 业务错误码集中定义在 `app/constants/error_code.go`：
  `var XxxError = NewErrorCode(1001, "描述")`。
- `WithError(err)` 附加底层错误（仅写入日志，不暴露给客户端）；`WithMessage(msg)` 覆盖对外提示。**两者均返回新副本、不修改原错误码**——错误码是全局单例，原地修改会在并发请求间互相覆盖（数据竞争），严禁改回「返回自身」的写法。
- 新错误码按业务模块分段编号，不要复用已有 Code。

### 分层约束
| 目录 | 职责 | 禁止 |
|---|---|---|
| `app/controller` | 解析参数、调 Service、`kernel.Send` | 写业务逻辑、直接访问数据库 |
| `app/service` | 业务逻辑、日志（`logx.WithContext(ctx)`）、返回错误码 | 操作 `http.ResponseWriter` |
| `app/svc` | ServiceContext 挂载全局依赖（如生成的 dao） | — |
| `app/kernel` | 框架支撑代码 | 一般不修改 |
| `app/types` | 请求/响应 DTO，手写维护 | — |

注意：`main.api` 只是接口契约文档，与代码**手工保持同步**。骨架目录布局与 goctl 标准布局不同（service 而非 logic、types 平铺等），**不要用 goctl 直接生成覆盖现有代码**，新增接口按模板手写（见 `reference/new-endpoint.md`）。

### 代码风格：方法挂在 struct 上（类似 PHP 的类）
业务代码（service、dao、各类 logic）**不写散落的包级函数**：一个业务单元定义一个 struct，所有方法都挂在该 struct 下面。固定模式 = struct + `NewXxx` 构造器，字段持有 `ctx`、`log`（`logx.WithContext(ctx)`）及依赖（`svcCtx`、`*gorm.DB` 等），方法写成 `func (s *XxxService) Method(...)`（接收者名与类型语义一致：service 用 `s`、dao 可用 `d`，不要沿用 go-zero logic 生态的 `l`）。新建文件可用 go-gen 生成骨架，或按下节模板手写——**无论哪种方式，结构必须一致**，再往 struct 上追加方法。仅中间件、`NewContext` 等确需函数形态的入口除外；controller 的 Handler 函数不受此约束。

> 取舍说明：struct 持有 `ctx` 字段是 go-zero logic 层的框架惯例（每请求在 Handler 中 `NewXxx` 新建实例），偏离 Go 官方「context 作为方法首参传递」的建议。代价是**此类 struct 严禁跨请求缓存复用**，否则 ctx 与 trace 会串请求。

### 最小文件策略：一个文件一个职责
**新逻辑新建专属文件，不往职责不符的现有文件里追加**；文件名即职责名。适用于所有目录（service、types、kernel 等）。例如新增用户登录态（存储 token 与 user_id）：不要写进 `app/kernel/ctx/context.go`（该文件只处理基础上下文 + 日志），而应新建 `app/kernel/ctx/user_auth.go`，并同样遵循 struct 风格：

```go
// app/kernel/ctx/user_auth.go —— 用户登录态（token + user_id）
package ctx

import "context"

type UserAuth struct {
	ctx context.Context
}

func NewUserAuth(ctx context.Context) *UserAuth {
	return &UserAuth{ctx: ctx}
}

// 方法按需追加（如 Token() / UserId()），全部挂在 UserAuth 上
```

此文件可用 `go-gen gen file name=UserAuth path=app/kernel/ctx` 生成（package 自动取 path 末段 `ctx`）。同一 package 下多文件共存，import 不变。例外：错误码（`app/constants/error_code.go`）、路由注册（`app/controller/routes.go`）等集中式定义，按既有约定继续在原文件追加。

### 脚手架 go-gen：生成代码文件（可选）
创建新的 service / dao / struct 文件时，若已安装 [go-gen](https://github.com/limingxinleo/go-gen) 则优先用它生成骨架（须在项目根目录执行）。**未安装时不必安装，直接按本节末尾的模板手写**，风格保持一致：

```bash
go install github.com/limingxinleo/go-gen@v1.3.4   # 安装（可选，仅一次）
```

常用命令：

```bash
go-gen gen service name=UserService             # → app/service/user_service.go（含 svcCtx；import 自动使用 go.mod 的 module 名）
go-gen gen gorm_dao name=UserDao                # → app/dao/user_dao.go（手写 Gorm DAO 骨架，持有 *gorm.DB）
go-gen gen dao name=UserDao                     # → app/dao/user_dao.go（手写 sqlx DAO 骨架，持有 sqlx.SqlConn）
go-gen gen file name=OrderLogic path=app/logic  # → app/logic/order_logic.go（通用 struct 骨架，package 取 path 末段）
```

行为要点：
- 文件名由 `name` 自动转 snake_case（`UserService` → `user_service.go`）。
- 目标文件已存在时直接报错拒绝覆盖；确认要覆盖加 `-f`。
- stub 模板缩进不完全规范，生成后执行 `gofmt -w <文件>` 统一格式。
- `dao` / `gorm_dao` 生成的是手写 DAO（package dao），与 `gen:model` 生成的 `app/dao/model` + `app/dao/query`（主推方式，见 `reference/database.md`）互不冲突，按需选用。
- `go-gen json2struct` 可将 JSON 转为 Go struct（`-s 'JSON串'`、`-f 文件`、`-c` 读剪贴板并写回），编写 `app/types` DTO 时可用。
- `go-gen config:create` 在项目下生成 `.go-gen/config.json` 与 stub 模板，可自定义生成规则；配置查找顺序：`./.go-gen/` → `~/.go-gen/` → 工具内置默认。

**未安装 go-gen 时，直接按以下模板手写**（与 go-gen 生成产物完全一致；`service` 模板见 `reference/new-endpoint.md` 第 3 步）：

```go
// app/dao/user_dao.go —— gorm_dao 模板（Gorm 手写 DAO）
package dao

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UserDao struct {
	db  *gorm.DB
	ctx context.Context
	log logx.Logger
}

func NewUserDao(db *gorm.DB, ctx context.Context) *UserDao {
	return &UserDao{
		db:  db,
		ctx: ctx,
		log: logx.WithContext(ctx),
	}
}
```

```go
// app/dao/user_dao.go —— dao 模板（sqlx 手写 DAO，差异：字段与构造参数为 sqlx.SqlConn）
package dao

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserDao struct {
	sqlConn sqlx.SqlConn
	ctx     context.Context
	log     logx.Logger
}

func NewUserDao(sqlConn sqlx.SqlConn, ctx context.Context) *UserDao {
	return &UserDao{
		sqlConn: sqlConn,
		ctx:     ctx,
		log:     logx.WithContext(ctx),
	}
}
```

`file` 模板最简：目标 package 下仅含 `ctx` 字段的 struct + `NewXxx(ctx context.Context)` 构造器，按需补依赖字段即可。

## 任务路由

| 任务 | 参考文档 |
|---|---|
| 从骨架初始化新项目（module/服务名/CI 改名清单） | `reference/init-project.md` |
| 新增 HTTP 接口（api/types/service/controller/routes 全流程模板） | `reference/new-endpoint.md` |
| 数据库：建表、gen:model 生成 dao、Gorm/sqlx 查询、事务；Redis 读写（`redis.Nil` 处理） | `reference/database.md` |
| 新增 GRPC 服务端 / 客户端调用 | `reference/grpc.md` |
| 单元测试编写、本地运行与调试 | `reference/testing-local.md` |

## 常用命令速查

```bash
go run main.go                            # 启动 HTTP 服务（需本地 MySQL/Redis 可达）
go run cmd/main.go gen:model              # 从数据库生成全部表的 Gorm dao
go run cmd/main.go gen:model user         # 仅生成指定表（可多个表名）
go-gen gen service name=UserService       # go-gen（可选脚手架）生成 service 骨架，另有 gorm_dao / dao / file
ROOT_PATH=$PWD go test ./... -v           # 本地跑单测（需 MySQL/Redis 可达，ROOT_PATH 必加，见下）
docker compose up -d --build              # 完整栈：MySQL + Redis + 应用（部署镜像）
DOCKERFILE=unit.Dockerfile docker compose up -d --build   # 单测环境镜像（含 Go 工具链）
```

### 本地跑单测必须带 ROOT_PATH（易踩坑）

`go test` 运行被测包的二进制时，工作目录是**包目录**（如 `app/service/`）而非仓库根目录；bootstrap 恰好以工作目录定位 `etc/main-api.yaml`（见「Application 单例与 import 副作用」）。因此本地裸跑 `go test ./...` 会报：

```
error: config file .../app/service/etc/main-api.yaml ... no such file or directory
```

在仓库根目录执行时显式指定 `ROOT_PATH=$PWD` 即可；CI 容器内由 `unit.Dockerfile` 内置的 `ROOT_PATH=/app` 解决，无需额外设置。更多测试细节见 `reference/testing-local.md`。

两个 Dockerfile 的区别：`Dockerfile` 为多阶段构建、最终 scratch 精简镜像（生产部署用）；`unit.Dockerfile` 基于golang:alpine、带完整源码与工具链，供 CI 在容器内执行 `go test`。

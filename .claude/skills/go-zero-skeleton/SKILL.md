---
name: go-zero-skeleton
description: 基于 go-zero + Gorm + Redis 的 API 骨架开发指南。在此项目中新增 HTTP 接口、编写 service 业务逻辑、生成或使用数据库模型、新增 GRPC 服务、添加配置项或错误码、编写单元测试、本地运行调试，以及从骨架初始化新项目时使用。
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
- Service 方法签名统一返回 `(result T, err kernel.ErrorCodeInterface)`，不返回裸 `error`。
- 业务错误码集中定义在 `app/constants/error_code.go`：
  `var XxxError = &ErrorCode{Code: 1001, Message: "描述"}`。
- `WithError(err)` 附加底层错误（仅写入日志，不暴露给客户端）；`WithMessage(msg)` 覆盖对外提示。
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

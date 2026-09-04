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

## 常用的组件库：samber/lo

[github.com/samber/lo](https://github.com/samber/lo) 是基于 Go 泛型的函数式工具库（`Map`/`Filter`/指针操作/集合运算等），一行泛型调用即可替代手写循环与样板代码，减少 service 层噪音。**它是可选依赖、不是骨架标配**——当前仅以 `// indirect` 形式出现在 `go.mod`（由其它依赖间接带入），骨架自身的 `app/` 代码并不 import 它。

### 安装判断：先判断值不值得用，再决定是否引入

**除非命中下表的场景且一次开发中会多处用到，否则不要引入**；单点、一次的简单转换直接用原生写法（`&x`、`for range`）即可，避免为「套库」而引入没必要的新依赖。口径如下：

- ✅ 值得引入：多处列表 DTO 转换、可空字段批量转换、集合去重/分组/求交、批量 `IN` 前清理参数；
- ❌ 不值得引入：一次性取一个值转指针、单层 for 循环加条件——手写几行更直白；
- 确认要用的执行（把 indirect 依赖转为直接依赖）：

```bash
go get github.com/samber/lo@v1.53.0 && go mod tidy
```

之后在任意文件 `import "github.com/samber/lo"`，以 `lo.xxx` 调用。

### 场景快查表

| 场景 | 推荐函数 | 典型用途 |
|---|---|---|
| 值类型 → Gorm 可空字段（零值→NULL） | `lo.EmptyableToPtr(x)` | `string`/`int64`/`time.Time` → `*string` 等 |
| 值类型 → 指针（无条件） | `lo.ToPtr(x)` | 空串也要落库为 `''`（而非 NULL）时 |
| 可空字段 → 值类型（nil→零值） | `lo.FromPtr(p)` | `*string` → 响应 DTO 的 `string` |
| 可空字段 → 值类型（nil→兜底值） | `lo.FromPtrOr(p, fallback)` | 昵称为空时兜底「未设置」 |
| 列表 DTO 转换 | `lo.Map` / `lo.FilterMap` / `lo.FlatMap` | DAO `[]model.*` → `[]*types.XxxResp` |
| 集合去重 / 交并 / 包含 | `lo.Uniq` / `lo.UniqBy` / `lo.Contains` / `lo.Every` | 参数去重后再查库 |
| 按字段分组 / 切块 | `lo.GroupBy` / `lo.Chunk` | 类目分组、批量 INSERT 分批 |
| 列表 → map 索引 | `lo.KeyBy` | O(1) 查找、批量填充 |
| map 取键值 / 转换 / 筛选 | `lo.Keys` / `lo.Values` / `lo.MapValues` / `lo.PickBy` | 稳定遍历、类型转换 |
| 内联判断（替代三目运算） | `lo.Ternary` / `lo.If(cond, a).Else(b)` | 参数上限、状态文案 |
| 第一个非零值 | `lo.CoalesceOrEmpty(a, b, c)` | 多来源取兜底 |
| 安全取首/末元素 | `lo.First` / `lo.FirstOr` / `lo.Last` / `lo.LastOr` | 空集合不 panic |
| 聚合 / 随机码 | `lo.SumBy` / `lo.MaxBy` / `lo.MinBy` / `lo.CountBy` / `lo.RandomString` | 金额求和、验证码/短码 |

### 场景一：Gorm 可空字段 ↔ 值类型（最常用，务必掌握）

`gen:model` 对**允许为 NULL** 的字段生成的模型类型是**指针**，而请求参数通常是非指针值类型，双向转换是最高频使用点：

```go
// 写：请求 string → 模型 *string。空串转 nil（落库 NULL），非空才转指针
user.UnionID = lo.EmptyableToPtr(req.UnionID)

// 若业务要求空串也必须落库为 ''（而非 NULL），用无条件转指针
user.UnionID = lo.ToPtr(req.UnionID)

// 读：模型 *string → 响应 DTO 的 string（nil 自动归一为零值，不会给前端写 null）
resp.UnionID = lo.FromPtr(user.UnionID)
// 读且 nil 时要兜底值
resp.Nickname = lo.FromPtrOr(user.Nickname, "未设置")
// 批量：[]*T → []T（nil 元素取零值）
tags := lo.FromSlicePtrOr(model.Tags, "")
```

> ⚠️ `EmptyableToPtr` 依据**零值**判定（空串/0/零时间/空结构体 → nil，其余 → 指针），由泛型 + reflect 实现、纯函数无副作用。若零值有业务语义、必须保留为值落库，请改用 `ToPtr`。

### 场景二：列表接口的 DTO 转换（service 层高频）

分页列表 DAO 返回 `[]model.*`，不能直接返回给前端，需转 `[]*types.XxxResp`，`lo.Map` 一行完成：

```go
list := lo.Map(models, func(m *model.User, _ int) *types.UserResp {
    return &types.UserResp{
        Id:       m.ID,
        Nickname: lo.FromPtr(m.Nickname), // 可空字段顺带解引用
        // ...
    }
})
```

转换同时过滤保留项用 `lo.FilterMap`（callback 返回 `(R, bool)`）；把嵌套切片结果拍平用 `lo.FlatMap`。

### 场景三：ID / 标签集合处理

```go
// 请求参数去重后再批量查库
ids := lo.Uniq(req.Ids)
db.Where("id IN ?", ids).Find(&users)

// IN 查询前按白名单筛掉不合法 id
ids = lo.Filter(ids, func(id int64, _ int) bool { return lo.Contains(allowedIDs, id) })

// 按字段分组（如商品按类目分组）
categorized := lo.GroupBy(goods, func(g *model.Goods) int64 { return g.CategoryID })

// 列表 → map 索引，O(1) 取用（如按用户ID批量填充昵称）
userMap := lo.KeyBy(users, func(u *model.User) int64 { return u.ID })

// 大批量 INSERT/UPSERT 会拼超长 SQL，分批执行
for _, batch := range lo.Chunk(items, 500) {
    db.CreateInBatches(batch, 500)
}
```

### 场景四：map 的稳定遍历与转换

Go map 原生遍历无序，需要稳定顺序时先取 key 再排序；map 值转换、筛选也用 lo：

```go
keys := lo.Keys(m)                                  // 全部 key；需要有序时自行 sort
vals := lo.Values(m)                                // 全部 value
m2 := lo.MapValues(m, func(v string, k string) int { return len(v) })
sub := lo.PickBy(m, func(k string, v any) bool { return strings.HasPrefix(k, "official_") })
```

### 场景五：内联判断与兜底（替代三目运算）

Go 无三目表达式，短判断用 `lo.If` / `lo.Ternary` 可避免临时变量：

```go
sex := lo.If(u.Gender == 1, "男").Else("女")
limit := lo.Ternary(pageSize > 100, 100, pageSize)            // 参数上限兜底
name := lo.CoalesceOrEmpty(profile.Nickname, req.Name, "游客") // 第一个非零值兜底
first := lo.FirstOr(items, zero)                              // 空集合安全取首
```

### 场景六：聚合与随机码

```go
total := lo.SumBy(orders, func(o *model.Order) int64 { return o.Amount })             // 金额求和
best := lo.MaxBy(orders, func(a, b *model.Order) bool { return a.Amount > b.Amount }) // 最大金额订单
done := lo.CountBy(orders, func(o *model.Order) bool { return o.Status == 2 })        // 满足条件计数
code := lo.RandomString(6, []rune("0123456789"))                                      // 随机验证码/短码
```

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

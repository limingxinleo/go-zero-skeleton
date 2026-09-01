# 数据库开发指南

连接在 bootstrap 阶段自动建立（见 SKILL.md「import 副作用」），通过 `app.GetApplication()` 获取：

| 字段 | 说明 |
|---|---|
| `.MySQL` | go-zero `sqlx.SqlConn`，原生 SQL + go-zero 内建缓存约定 |
| `.Gorm` | Gorm 实例，已接入 logx 日志（`app/kernel/logger`）、连接池（空闲 10 / 最大 100 / 生命周期 1h） |

## 建表流程

1. DDL 写入 `.github/init.sql` —— docker compose 首次启动 MySQL 时自动执行（`docker-compose.yml` 将其挂载到 `/docker-entrypoint-initdb.d/init.sql`）。
2. 对已运行的库，也可直接在 MySQL 中执行 DDL，或手动重建：`docker compose down -v && docker compose up -d --build`（`-v` 会清空数据卷，谨慎使用）。
3. 执行 `go run cmd/main.go gen:model` 生成代码。

## 方式一：Gorm + gen 生成 dao（骨架主推）

```bash
go run cmd/main.go gen:model            # 生成全部表
go run cmd/main.go gen:model user order # 仅生成指定表（可多个）
```

生成产物（由 `cmd/cmd/gen_model.go` 定义，勿手改）：
- `app/dao/model/` —— 表结构体（可空字段为指针类型，带 index/type 标签）
- `app/dao/query/` —— 类型安全查询代码

在 Service 中使用（具体 API 以生成的 `app/dao/query` 代码为准）：

```go
import (
	"github.com/limingxinleo/go-zero-skeleton/app"
	"github.com/limingxinleo/go-zero-skeleton/app/dao/query"
)

q := query.Use(app.GetApplication().Gorm)

// 查询单条
u, err := q.User.WithContext(s.ctx).Where(q.User.ID.Eq(req.Id)).First()
// 条件查询列表
users, err := q.User.WithContext(s.ctx).Where(q.User.Status.Eq(1)).Find()
// 创建
err = q.User.WithContext(s.ctx).Create(&model.User{NickName: "xxx"})
```

注意 `ErrRecordNotFound`（`gorm.ErrRecordNotFound`）需按业务转换为错误码返回，不要透传给 `kernel.Send`。

### 事务

```go
db := app.GetApplication().Gorm
err := db.WithContext(s.ctx).Transaction(func(tx *gorm.DB) error {
	// 在 tx 上操作；或 q := query.Use(tx) 使用生成的查询
	return nil
})
```

事务内的 error 为普通 `error`，出事务后按需包装为 `constants.ServerError.WithError(err)` 返回。

## 方式二：go-zero 原生 model

将 DDL 放到 `app/model/ddl/mysql.sql`，执行：

```bash
goctl model mysql ddl --src ./app/model/ddl/mysql.sql --dir ./app/model
```

产物自带 `FindOne` 等方法与 Redis 缓存约定（按需在方法调用处传入 `app.GetApplication().Redis`）。

## 方式三：直接写 SQL

```go
// sqlx
var nickName string
err := app.GetApplication().MySQL.QueryRowCtx(ctx, &nickName, "SELECT nick_name FROM user WHERE id = ?", id)

// Gorm 原生
var nickName string
err := app.GetApplication().Gorm.WithContext(ctx).
	Raw("SELECT nick_name FROM user WHERE id = ?", id).
	Scan(&nickName).Error
```

## Redis

`app.GetApplication().Redis` 为 go-zero `*redis.Redis`，连接在 bootstrap 阶段自动建立（见 SKILL.md「import 副作用」）。

go-zero 的 Redis 为双轨 API：本项目统一使用**带 ctx 的 `XxxCtx` 方法**（与 goctl 生成产物、Go 主流「所有 IO 调用传递 ctx」规范一致），保证 trace 链路完整、调用可随 ctx 超时取消。无 ctx 变体仅在确实拿不到 context 的极少数场景使用。

```go
rds := app.GetApplication().Redis

val, err := rds.GetCtx(ctx, "key")             // key 不存在时返回 redis.Nil
err = rds.SetCtx(ctx, "key", "value")          // 无过期时间
err = rds.SetexCtx(ctx, "key", "value", 3600)  // 带过期时间（秒）
n, err := rds.DelCtx(ctx, "key1", "key2")
ok, err := rds.ExistsCtx(ctx, "key")
n, err := rds.IncrCtx(ctx, "counter")
```

**key 不存在的处理（必读）**：`Get` 在 key 不存在时返回 `("", redis.Nil)`。`redis.Nil` 是预期内的哨兵错误而非服务故障，必须单独排除，否则接口会对空 key 误报 500：

```go
import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

val, err := app.GetApplication().Redis.GetCtx(ctx, "test")
if err != nil && !errors.Is(err, redis.Nil) {
	return "", constants.ServerError.WithError(err)
}
// redis.Nil 时 val 为空字符串，按业务语义处理（默认空值返回即可）
```
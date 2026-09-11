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

## 软删除（强制）

**禁止物理删除业务数据**。有删除需求的表一律增加 `is_deleted` 字段：删除 = 将该行 `is_deleted` 置 `1`（软删），查询活跃数据默认先过滤 `is_deleted = 0`。严禁 Gorm `Delete` / `Unscoped().Delete` / 手写 `DELETE FROM ...` / `TRUNCATE`（相关强制约定与常量模板见 SKILL.md「MySQL 软删除」，本节给具体写法）。

先按 SKILL.md 模板在项目内落地 `app/constants/soft_delete.go`，取值一律走常量、不写魔法值：

```go
package constants

const (
    No  = 0 // 未删除：新建默认值，所有查询的默认过滤条件
    Yes = 1 // 已删除：软删除时置此值
)
```

### 建表

DDL 统一写入 `.github/init.sql`（见「建表流程」），`is_deleted` 列建议带索引（几乎所有查询都过滤它）：

```sql
CREATE TABLE `user` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `nick_name`   VARCHAR(64)  NOT NULL DEFAULT '',
  `is_deleted`  TINYINT      NOT NULL DEFAULT 0 COMMENT '是否已删除：0-未删除，1-已删除',
  PRIMARY KEY (`id`),
  KEY `idx_is_deleted` (`is_deleted`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
```

不含删除需求的表可不用该字段；一旦加了，全项目按本节规范执行。`gen:model` 会把它生成为模型的 `IsDeleted` 字段（字节类型），与 `constants.No / constants.Yes` 比较即可。

### 查询：默认过滤未删除

**所有查询默认追加未删除条件。** 本章其余示例为保持简洁均省略了该条件，实际开发必须加上。

方式一（Gorm gen，骨架主推）：

```go
q := query.Use(app.GetApplication().Gorm)

// 单条
u, err := q.User.WithContext(s.ctx).
    Where(q.User.IsDeleted.Eq(constants.No)).
    First()

// 列表（与其它条件叠加）
users, err := q.User.WithContext(s.ctx).
    Where(q.User.Status.Eq(1), q.User.IsDeleted.Eq(constants.No)).
    Find()
```

方式二/三（sqlx 直接 SQL / Gorm 手写 DAO）：SQL 里补 `is_deleted = ?`，参数传 `constants.No`；或 Gorm 用 `Where(...).Where("is_deleted", constants.No)`：

```go
// sqlx
err := app.GetApplication().MySQL.QueryRowCtx(ctx, &nickName,
    "SELECT nick_name FROM user WHERE id = ? AND is_deleted = ?", id, constants.No)

// Gorm 原生
err = app.GetApplication().Gorm.WithContext(ctx).
    Where("status = ? AND is_deleted = ?", status, constants.No).
    Find(&users).Error
```

### 删除：一律软删

把目标行 `is_deleted` 置为 `constants.Yes`，**不允许** `db.Delete(&u)`、`Unscoped().Delete(&u)` 或手写 `DELETE FROM ...`：

```go
// Gorm gen：Update 字段置 1（先按未删除条件定位到行，避免对已删行重复标记）
info, err := q.User.WithContext(s.ctx).
    Where(q.User.ID.Eq(id), q.User.IsDeleted.Eq(constants.No)).
    Update(q.User.IsDeleted, constants.Yes)
if err != nil {
    return constants.ServerError.WithError(err)
}
if info.RowsAffected == 0 {
    // 目标不存在或已是软删状态，按业务处理
}

// Gorm 原生
err = app.GetApplication().Gorm.WithContext(ctx).
    Model(&model.User{}).
    Where("id = ? AND is_deleted = ?", id, constants.No).
    Update("is_deleted", constants.Yes).Error
```

事务内一致：`db.Transaction` 中用 `query.Use(tx)` 生成查询对象后，同样追加 `IsDeleted.Eq(constants.No)`/软删 Update，见「事务」。

> ⚠️ 唯一索引注意：软删后行仍占用唯一键，同自然键的新数据会插入失败。若业务允许「删除后同键重建」，不要把唯一索引建在自然键上（考虑去掉或用 `is_deleted` 参与联合唯一等方案）。

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

**key 全局前缀（强制）**：本项目 **禁用 Redis 物理 db 隔离业务**（go-zero 官方不推荐：多 db 仅单机 legacy 模式可用、Cluster 集群连接固定 db 0，当前版本的 `RedisConf` 也已移除 `DB` 配置项），连接一律 db 0。业务隔离统一走 **key 全局前缀**——按 SKILL.md「Redis key 全局前缀」模板在项目内提供 `RedisPrefix` 常量与 `RedisKey(key)` 统一拼接，**所有 key 均形如 `{RedisPrefix}:{业务key}`**、经 `RedisKey` 取 key，禁止手写拼 key：

```go
import "your-module/app/constants" // 项目内按模板实现的 constants 包，import 前缀以 go.mod 的 module 名为准
// 前缀由用户在初始化时手动输入一次写入 constants.RedisPrefix，此后全局沿用
rds := app.GetApplication().Redis
key := constants.RedisKey("user:test")                           // "{RedisPrefix}:user:test"

val, err := rds.GetCtx(ctx, key)             // key 不存在时返回 redis.Nil
err = rds.SetCtx(ctx, key, "value")          // 无过期时间
err = rds.SetexCtx(ctx, key, "value", 3600)  // 带过期时间（秒）
n, err := rds.DelCtx(ctx, key, constants.RedisKey("user:other"))
ok, err := rds.ExistsCtx(ctx, key)
n, err := rds.IncrCtx(ctx, constants.RedisKey("counter"))
```

**key 不存在的处理（必读）**：`Get` 在 key 不存在时返回 `("", redis.Nil)`。`redis.Nil` 是预期内的哨兵错误而非服务故障，必须单独排除，否则接口会对空 key 误报 500：

```go
import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

val, err := app.GetApplication().Redis.GetCtx(ctx, constants.RedisKey("test")) // 统一经 constants.RedisKey 取 key
if err != nil && !errors.Is(err, redis.Nil) {
	return "", constants.ServerError.WithError(err)
}
// redis.Nil 时 val 为空字符串，按业务语义处理（默认空值返回即可）
```
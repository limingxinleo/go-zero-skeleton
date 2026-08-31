# 单元测试与本地运行

## 前提：测试依赖真实 MySQL/Redis

`app/bootstrap.go` 的 `init()` 会在任何测试启动时加载配置并连接数据库，因此**本地直接 `go test` 前必须保证 `etc/main-api.yaml` 中的 MySQL/Redis 可达**。两种方式：

### 方式 A：docker compose 起单测环境（与 CI 一致）

```bash
DOCKERFILE=unit.Dockerfile docker compose up -d --remove-orphans --build
docker compose exec hyperf go test ./... -v      # 服务名以 docker-compose.yml 为准
docker compose down                              # 结束后清理
```

`unit.Dockerfile` 基于 golang:alpine、内含完整源码与工具链，容器内配置为 `etc/unit-api.yaml`（MySQL/Redis 指向 compose 内的 `mysql`/`redis` 主机名）。这也是 `.github/workflows/test.yml` 与 `.gitlab-ci.yml` 的执行方式。

### 方式 B：本地服务

本地已运行 MySQL（含 `hyperf` 库及 `init.sql` 建表）与 Redis 时，在仓库根目录执行：

```bash
ROOT_PATH=$PWD go test ./... -v
```

`ROOT_PATH` 必须显式指定：`go test` 的工作目录是被测包目录（如 `app/service/`），而 bootstrap 以工作目录定位 `etc/main-api.yaml`，缺少该变量时测试会因找不到配置文件直接失败。

## 单元测试模板

测试文件与被测 Service 同目录（参考 `app/service/index_test.go`），使用 testify：

```go
package service

import (
	"context"
	"testing"

	"github.com/limingxinleo/go-zero-skeleton/app"
	"github.com/limingxinleo/go-zero-skeleton/app/types"
	"github.com/stretchr/testify/assert"
)

func TestUserService_Info(t *testing.T) {
	svc := NewUserService(context.TODO(), app.GetApplication().ServiceContext)

	result, err := svc.Info(&types.UserInfoRequest{Id: 1})

	assert.Nil(t, err)
	assert.NotNil(t, result)
}
```

要点：
- `app.GetApplication().ServiceContext` 直接可用（import `app` 即完成初始化）。
- 测试数据尽量在测试内自建自清理；依赖固定种子数据时在 `.github/init.sql` 中补充。

## 本地启动 HTTP 服务

```bash
go run main.go
# Starting server at 0.0.0.0:8888 ...
curl 'http://127.0.0.1:8888/?name=test'
```

或用 docker compose 起完整栈（生产镜像）：

```bash
docker compose up -d --build
```

## 提交前检查

```bash
go build ./...
go vet ./...
ROOT_PATH=$PWD go test ./... -v
```

## CI 说明

- GitHub Actions（`.github/workflows/test.yml`）：push / PR / 每日定时，Go 1.21~1.25 矩阵，方式 A 执行测试。
- GitLab CI（`.gitlab-ci.yml`）：unit -> build -> deploy 三阶段，test 分支构建测试镜像并 `docker stack deploy`，tag 构建正式镜像。
- Release（`.github/workflows/release.yml`）：推送 `v*` tag 时创建 GitHub Release。

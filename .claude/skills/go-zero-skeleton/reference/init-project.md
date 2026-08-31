# 从骨架初始化新项目

克隆 go-zero-skeleton 后开发新项目时，按以下清单完成改名。骨架中残留历史命名 `hyperf`（服务名、CI 项目名），必须一并替换。

## 1. 替换 module 路径（影响所有 import）

```bash
# 假设新 module 为 github.com/yourname/new-project
git grep -l 'github.com/limingxinleo/go-zero-skeleton' | xargs sed -i '' 's|github.com/limingxinleo/go-zero-skeleton|github.com/yourname/new-project|g'
```

涉及 `go.mod` 及 `main.go`、`cmd/`、`app/` 下所有 Go 文件的 import。

## 2. 服务名与配置

- `etc/main-api.yaml` 与 `etc/unit-api.yaml`：`Name` 改为新服务名（该值会被 CLI 根命令、日志等引用）。
- 按需调整 `Port`、`RedisConf`、`MySqlConf.Dsn`（默认库名 `hyperf` 也建议改掉，并同步 `docker-compose.yml` 的 `MYSQL_DATABASE`）。

## 3. Docker Compose

`docker-compose.yml`：
- 服务名 `hyperf` 改为新名。注意 CI 通过容器名 `$(basename $(pwd))-hyperf-1` exec 进容器执行测试，改服务名后 `.github/workflows/test.yml` 与 `.gitlab-ci.yml` 中的 `-hyperf-1` 必须同步改。
- 镜像名 `hyperf/go-zero-skeleton:latest` 改为自己的镜像仓库地址。

## 4. CI/CD

- `.gitlab-ci.yml`：`PROJECT_NAME: hyperf`、`REGISTRY_URL: registry-docker.org` 按实际修改；三个 stage（unit/build/deploy）的 runner tag 按实际环境调整。
- `deploy.test.yml`：随 `PROJECT_NAME` 一起使用，一般无需改动。
- 不使用 GitLab CI 可删除 `.gitlab-ci.yml`；GitHub Actions（`.github/workflows/`）同理按需保留。

## 5. 数据库初始化

建表 SQL 写入 `.github/init.sql`（docker compose 首次启动 MySQL 时自动执行），随后用 `go run cmd/main.go gen:model` 生成 dao，详见 `reference/database.md`。

## 6. 其他

- 重写 `README.md` 为新项目文档。
- `main.api` 是接口契约文档，保留格式、替换为新项目接口。
- 清理示例代码：`app/controller/index_controller.go`、`app/service/index*.go`、`app/types/types.go` 中的 `FromRequest` 可删除，但 `Response[T]` 结构是 `kernel.Send` 的依赖，必须保留。

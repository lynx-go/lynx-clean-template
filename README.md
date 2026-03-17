# Lynx Clean Architecture Template / Lynx Clean 架构模板

一个基于 Lynx + DDD/Clean Architecture 的 Go 服务模板，内置 gRPC、grpc-gateway、任务调度、Pub/Sub 与 Wire 依赖注入。

A Go service template built with Lynx + DDD/Clean Architecture, including gRPC, grpc-gateway, scheduler, Pub/Sub, and Wire-based dependency injection.

## 1) Architecture / 架构说明

- `internal/api`: 传输层（gRPC handlers、event handlers、cronjobs）/ Transport layer.
- `internal/app`: 用例层（应用服务，编排业务流程）/ Use-case layer.
- `internal/domain`: 领域层（实体、领域服务、ports）/ Domain layer.
- `internal/infra`: 基础设施适配层（数据库、服务端组件、外部客户端）/ Infrastructure adapters.
- `cmd/server`: 服务启动与组件装配（Wire providers + bootstrap）/ Server bootstrap and composition.
- `cmd/cli`: CLI 启动入口与命令 / CLI entry and commands.

请求路径示例 / Typical request flow:

`internal/api/grpc/*.go` -> `internal/app/*.go` -> `internal/domain/*/repo` -> `internal/infra/bun/bunrepo/*`

## 2) Prerequisites / 环境准备

- Go（建议与 `go.mod` 保持一致版本）
- Docker + Docker Compose（用于本地 Postgres/Redis）
- Task (`go-task`)
- `protoc`, `buf`, `wire`, `migrate`（可通过 Task 安装部分工具）

## 3) Quick Start / 快速开始

### 3.1 Start local infra / 启动本地依赖

```pwsh
task up
```

### 3.2 Configure environment / 配置环境变量

项目会在启动时尝试加载 `.env`（开发方便，生产可不依赖该文件）。

The server opportunistically loads `.env` on startup (dev convenience).

推荐做法 / Recommended:

1. 参考 `configs/config.yaml.template`。
2. 使用 `LYNX_` 前缀环境变量覆盖敏感配置。

示例映射 / Example mapping:

- `security.jwt.secret` -> `LYNX_SECURITY_JWT_SECRET`
- `data.database.source` -> `LYNX_DATA_DATABASE_SOURCE`
- `data.redis.password` -> `LYNX_DATA_REDIS_PASSWORD`

### 3.3 Run migrations / 执行数据库迁移

```pwsh
task migrate
```

### 3.4 Run server / 启动服务

```pwsh
task dev
```

默认配置模板端口 / Default template ports:

- gRPC: `:8088`
- HTTP (grpc-gateway): `:9099`

## 4) Common Tasks / 常用任务

```pwsh
task up
task down
task down:clean
task migrate
task generate
task wire
task test
task build
task build-cli
```

说明 / Notes:

- `task generate` = `buf generate` + config proto generation + Wire generation.
- 修改 provider 后请执行 `task wire`。
- `genproto/` 下为生成文件，不要手改（Do not edit generated files manually）。

## 5) API & Code Generation / API 与代码生成

- Proto 定义位于 `proto/api/v1/*.proto`。
- 生成产物位于 `genproto/api/v1/*`。
- gRPC 实现位于 `internal/api/grpc/*.go`。

重新生成 / regenerate:

```pwsh
task generate:proto
task wire
```

## 6) CLI / 命令行

当前模板包含示例命令（例如 `hello`, `print-config`）。

Current template includes sample CLI commands (e.g., `hello`, `print-config`).

运行示例 / example:

```pwsh
go run ./cmd/cli hello user-123 -t user-123
go run ./cmd/cli print-config
```

## 7) Testing / 测试

```pwsh
task test
```

## 8) Project Layout / 目录结构

```text
cmd/            # server / cli bootstrap
internal/api    # transport layer
internal/app    # use cases
internal/domain # domain models + ports
internal/infra  # adapters (db, server, clients)
proto/          # protobuf contracts
genproto/       # generated protobuf code
db/migrations/  # database migrations
```

## 9) Known Template Notes / 模板注意事项

- `Taskfile.yml` 中个别历史命令描述可能与当前 CLI 示例命令不完全一致；请以 `cmd/cli/cmd/*` 实际实现为准。
- `docker/local/docker-compose.yml` 提供了 Postgres 与 Redis；Kafka 需按你的环境单独准备。

---

如需扩展业务模块，建议遵循现有分层：先定义 domain port，再在 infra 实现 adapter，最后由 app use-case 编排并在 api 层暴露。

For new modules, follow the same layering: define domain ports first, implement adapters in infra, orchestrate in app use-cases, and expose through api handlers.

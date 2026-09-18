# Lynx Clean Architecture Template / Lynx Clean 架构模板

一个基于 Lynx（v1.11.0）+ DDD / Clean Architecture 的 Go 服务模板，内置 gRPC、grpc-gateway、定时任务、事件总线（Pub/Sub）与 Wire 依赖注入。

A Go service template built with Lynx (v1.11.0) + DDD / Clean Architecture, including gRPC, grpc-gateway, scheduler, an event bus (Pub/Sub), and Wire-based dependency injection.

## 1) Architecture / 架构说明

- `internal/api`: 传输层（gRPC handlers、event handlers、cronjobs）/ Transport layer.
- `internal/app`: 用例层（应用服务，编排业务流程）/ Use-case layer.
- `internal/domain`: 领域层（实体、领域服务、ports）/ Domain layer.
- `internal/infra`: 基础设施适配层（数据库、服务端组件、外部客户端）/ Infrastructure adapters.
- `cmd/server`: 服务启动与组件装配（Wire providers + `boot.Bootstrap`）/ Server bootstrap and composition.
- `cmd/cli`: CLI 启动入口与命令 / CLI entry and commands.

请求路径示例 / Typical request flow:

`internal/api/grpc/*.go` -> `internal/app/*.go` -> `internal/domain/*/repo` -> `internal/infra/bun/bunrepo/*`

### 启动模型 / Bootstrap model

服务与 CLI 均基于 `lynx.NewRunner` 构建：

```go
runner := lynx.NewRunner(func(app lynx.App) error {
    app.SetLogger(zap.MustNewLogger(app))

    boot, cleanup, err := wireBootstrap(app)
    if err != nil {
        return err
    }
    app.OnPostStop(cleanup) // Wire cleanup 在所有服务停止后执行
    boot.Apply(app)         // 注册 hooks 与 services
    return nil
}, lynx.WithName("lynx-api"), lynx.WithBindConfigFunc(...))
runner.Run()
```

- 组件实现 `lynx.Service`（含 `Name/Init/Start/Stop`），通过 `app.Register(...)` 托管生命周期。
- 服务通过 `boot.Bootstrap`（由 Wire 生成）聚合 `PreStartHooks`/`DrainHooks`/`PreStopHooks`/`PostStopHooks` 与 `[]lynx.Service`，再调用 `boot.Apply(app)` 统一注册。
- 配置通过 `app.Config()`（`lynx.Config` 接口）读取；事件总线通过 `app.Bus()`（`eventbus.Bus`）获取。

## 2) Prerequisites / 环境准备

- Go ≥ 1.26.5（模块 `go` 指令已声明；`mise install` 会按 `mise.toml` 装齐对应工具链）。
- Docker + Docker Compose（用于本地 Postgres / Redis）。
- [mise](https://mise.jdx.dev/)（`curl https://mise.run | sh`）。
- `buf` / `protoc` / `protoc-gen-go` / `migrate` 由 `mise install` 按 `mise.toml` 钉版安装，无需手动逐个安装。

## 3) Quick Start / 快速开始

### 3.1 Start local infra / 启动本地依赖

```pwsh
mise run up
```

### 3.2 Configure environment / 配置环境变量

项目会在启动时尝试加载 `.env`（开发方便，生产可不依赖该文件）；mise 任务也会自动加载 `.env`（`[env] _.file`）。

The server opportunistically loads `.env` on startup (dev convenience).

推荐做法 / Recommended:

1. 参考 `configs/config.yaml.template`。
2. 使用 `LYNX_` 前缀环境变量覆盖配置：任意点分键都可覆盖（v1.8.0 起框架启用 env key replacer，点号映射为下划线），例如 `data.database.source` -> `LYNX_DATA_DATABASE_SOURCE`；敏感键另经显式 `BindEnv` 兜底（见 `internal/pkg/config/bind.go`）。

示例映射 / Example mapping:

- `security.jwt.secret` -> `LYNX_SECURITY_JWT_SECRET`
- `data.database.source` -> `LYNX_DATA_DATABASE_SOURCE`
- `data.redis.password` -> `LYNX_DATA_REDIS_PASSWORD`

### 3.3 Run migrations / 执行数据库迁移

```pwsh
mise run migrate
```

（一次性覆盖迁移 DSN 用 `MIGRATE_DSN=... mise run migrate`；mise 的 `[env]` 会覆盖同名 shell 变量。）

### 3.4 Run server / 启动服务

```pwsh
mise run dev
```

默认配置模板端口 / Default template ports:

- gRPC: `:8088`
- HTTP (grpc-gateway): `:9099`

## 4) Common Tasks / 常用任务

```pwsh
mise install      # 首次：安装 go / buf / protoc / protoc-gen-go / migrate
mise run up
mise run down
mise run down:clean
mise run migrate
mise run generate
mise run wire
mise run test
mise run build
mise run build-cli
```

说明 / Notes:

- `mise tasks` 可列出全部任务。
- `mise run generate` = `buf generate` + config proto 生成 + Wire 生成。
- 修改 provider 后请执行 `mise run wire`（会重新生成 `cmd/server`、`cmd/cli/cmd`、`tests` 下的 `wire_gen.go`）。
- `genproto/` 下为生成文件，不要手改（Do not edit generated files manually）。
- `mise run build-all` = `build` + `build-cli`；`mise run build-docker` 构建并推送镜像。

## 5) API & Code Generation / API 与代码生成

- Proto 定义位于 `proto/api/v1/*.proto`。
- 生成产物位于 `genproto/api/v1/*`。
- gRPC 实现位于 `internal/api/grpc/*.go`。

重新生成 / regenerate:

```pwsh
mise run generate:proto
mise run wire
```

## 6) Event Bus / 事件总线

消息总线基于 Lynx 核心包 `eventbus`（v1.5.2 起 `contrib/pubsub` 已移除），默认使用进程内内存总线，由框架在启动时自动注入并通过 `app.Bus()` 获取。

The message bus is built on Lynx's core `eventbus` package (the old `contrib/pubsub` was removed in v1.5.2). It defaults to an in-process memory bus, injected by the framework and reachable via `app.Bus()`.

- `pkg/pubsub.Broker` 包装 `eventbus.Bus`，并以 CloudEvents JSON 作为事件载荷格式。
- `pkg/pubsub.Router` 是一个 `lynx.Service`，在 `Start` 阶段把各 `Handler` 订阅到总线（基于 topic + event type 过滤）。
- 领域事件统一通过 `shared.EventPublisher` 端口发布，由 `internal/infra` 适配到 `pkg/pubsub`。

跨进程传输（Kafka）如需启用，可基于 `contrib/watermill` + `contrib/watermill-kafka` 构造 `eventbus.Transport` 并注入总线；本地模板默认仅使用内存总线（与本地 compose 不含 Kafka 的约定一致）。

Cross-process transport (Kafka) can be added via `contrib/watermill` + `contrib/watermill-kafka` by building an `eventbus.Transport`; the template defaults to the in-memory bus (matching the local compose that ships without Kafka).

## 7) CLI / 命令行

当前模板包含示例命令（例如 `hello`、`print-config`）。

Current template includes sample CLI commands (e.g., `hello`, `print-config`). CLI 基于 `lynx.NewRunner` + `app.Command` 构建，在命令执行期间启动所需组件并托管生命周期。

运行示例 / example:

```pwsh
go run ./cmd/cli hello -t user-123
go run ./cmd/cli print-config
```

## 8) Testing / 测试

```pwsh
mise run test
```

测试套件同样基于 `lynx.NewRunner` + `app.Command`，在 `tests` 包中通过 Wire 装配依赖。

## 9) Project Layout / 目录结构

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

## 10) New Module Template / 新模块开发模板

用于快速新增一个业务模块（例如 `orders`、`products`），保持与当前 DDD / Clean 分层一致。

Use this template to add a new business module (e.g., `orders`, `products`) while keeping the current DDD / Clean layering.

### 10.1 Layer Flow / 分层流转

`api handler` -> `app use-case` -> `domain port` -> `infra adapter`

### 10.2 Domain Port (interface) / 领域端口

`internal/domain/<module>/repo/<module>.go`

```go
package repo

import "context"

type Entity struct {
	ID string
}

type Repository interface {
	Create(ctx context.Context, e *Entity) error
	GetByID(ctx context.Context, id string) (*Entity, error)
}
```

### 10.3 Infra Adapter (implementation) / 基础设施适配器

`internal/infra/bun/bunrepo/<module>.go`

```go
package bunrepo

import (
	"context"

	domainrepo "github.com/lynx-go/lynx-clean-template/internal/domain/<module>/repo"
)

type ModuleRepo struct {
	// inject bun.DB or query builder here
}

func NewModuleRepo() domainrepo.Repository {
	return &ModuleRepo{}
}

func (r *ModuleRepo) Create(ctx context.Context, e *domainrepo.Entity) error {
	// persist with Bun
	return nil
}

func (r *ModuleRepo) GetByID(ctx context.Context, id string) (*domainrepo.Entity, error) {
	// query with Bun
	return &domainrepo.Entity{ID: id}, nil
}
```

### 10.4 App Use-case / 应用用例层

`internal/app/<module>.go`

```go
package app

import (
	"context"

	domainrepo "github.com/lynx-go/lynx-clean-template/internal/domain/<module>/repo"
)

type Module struct {
	repo domainrepo.Repository
}

func NewModule(repo domainrepo.Repository) *Module {
	return &Module{repo: repo}
}

func (m *Module) Create(ctx context.Context, id string) error {
	return m.repo.Create(ctx, &domainrepo.Entity{ID: id})
}
```

### 10.5 API Handler / 接口层

`internal/api/grpc/<module>.go`

```go
package grpc

import (
	"context"

	"github.com/lynx-go/lynx-clean-template/internal/app"
)

type ModuleService struct {
	uc *app.Module
}

func NewModuleService(uc *app.Module) *ModuleService {
	return &ModuleService{uc: uc}
}

func (s *ModuleService) Create(ctx context.Context /* req */) (/* resp */ any, error) {
	if err := s.uc.Create(ctx, "new-id"); err != nil {
		return nil, err
	}
	return nil, nil
}
```

### 10.6 Wiring Checklist / 注入与注册清单

- 在 `internal/domain/provides.go` 暴露 domain provider（如需要）。
- 在 `internal/infra/provides.go` 注入 adapter constructor（例如 `bunrepo.NewModuleRepo`）。
- 在 `internal/app/provides.go` 注入 use-case constructor（例如 `app.NewModule`）。
- 在 `internal/api/provides.go` 注入 handler constructor（例如 `grpc.NewModuleService`）。
- 在 server gRPC / gateway 注册新服务后，执行：

```pwsh
mise run wire
mise run generate:proto
```

## 11) Known Template Notes / 模板注意事项

- `mise.toml` 中个别历史命令描述可能与当前 CLI 示例命令不完全一致；请以 `cmd/cli/cmd/*` 实际实现为准。
- `docker/local/docker-compose.yml` 提供了 Postgres 与 Redis；Kafka 需按你的环境单独准备（见第 6 节）。
- Lynx v1.5.x：`contrib/kafka` 与 `contrib/pubsub` 已移除，发布/订阅统一走核心 `eventbus`；启动方式由旧 `lynx.New` / `lynx.CLI` 模型迁移到 `lynx.NewRunner` + `app.Register` / `app.Command`。
- Lynx v1.10.0：生命周期钩子更名为五阶段 `OnPreStart` / `OnPostStart` / `OnDrain` / `OnPreStop` / `OnPostStop`（Wire cleanup 等资源释放应挂 `app.OnPostStop`）；`boot.New` 参数顺序同步调整，provider 修改后需重新 `mise run wire`。
- Lynx v1.11.0：`boot.Bind(app)` 更名为 `boot.Apply(app)`。

---

如需扩展业务模块，建议遵循现有分层：先定义 domain port，再在 infra 实现 adapter，最后由 app use-case 编排并在 api 层暴露。

For new modules, follow the same layering: define domain ports first, implement adapters in infra, orchestrate in app use-cases, and expose through api handlers.

# AGENTS Guide

## Big Picture
- This repo is a Lynx-based clean architecture template: transport in `internal/api`, use-cases in `internal/app`, domain services/ports in `internal/domain`, adapters in `internal/infra`.
- Runtime composition is dependency-injected with Wire from `cmd/server/provides.go` -> `cmd/server/wire_gen.go`; CLI/test use separate provider sets in `cmd/cli/cmd/provides.go` and `tests/provides.go`.
- Main server components are started as Lynx components (`grpc`, `grpc-gateway`, scheduler, pubsub broker/router, kafka binder) in `cmd/server/provides.go`.

## Data and Request Flow
- gRPC APIs are defined in `proto/api/v1/*.proto`, generated to `genproto/api/v1/*`, and implemented in `internal/api/grpc/*.go`.
- Typical call path: gRPC handler (`internal/api/grpc/users.go`) -> app use-case (`internal/app/users.go`) -> domain repo port (`internal/domain/users/repo/users.go`) -> Bun adapter (`internal/infra/bun/bunrepo/users.go`).
- Auth middleware is centralized in `pkg/grpc/interceptor/jwt.go` and wired in `internal/infra/server/grpc.go`; JWT skip list currently includes `AuthService/Token`, `AuthService/VerifySignUpEmail`, and `AuthService/ResendSignUpEmailCode`.
- Sign-up now includes email OTP verification flow: `internal/api/grpc/auth.go` -> `internal/app/account.go` -> `internal/domain/users/repo/email_verification_codes.go` -> `internal/infra/bun/bunrepo/email_verification_codes.go`.
- Domain events are published from app layer (example: `internal/app/account.go` publishes `events.EventAccountCreated`) via `shared.EventPublisher` adapter in `internal/infra/domain_adapters.go`.

## Dev Workflows (Project-Specific)
- Use Task as the primary workflow runner (`Taskfile.yml`). Key tasks: `task up`, `task migrate`, `task generate`, `task wire`, `task dev`, `task test`, `task build`, `task build-cli`, `task build-all`.
- Proto generation is Buf-driven (`buf.yaml`, `buf.gen.yaml`) and writes into `genproto/`; do not hand-edit generated `*.pb.go` files.
- DI is Wire-driven; after provider changes, regenerate with `task wire` (server + cli + tests).
- Local infra comes from `docker/local/docker-compose.yml` (Postgres + Redis + Traefik). Kafka is expected by config but not provided by this compose file.

## Config and Environment Conventions
- Config schema is proto-defined in `internal/pkg/config/config.proto`; runtime binding is in `internal/pkg/config/bind.go`.
- Environment override prefix is `LYNX_`; keys map from dotted paths (for example `data.database.source` -> `LYNX_DATA_DATABASE_SOURCE`).
- `envBoundKeys` currently includes JWT secrets, DB/Redis credentials, and default file-bucket access keys (`internal/pkg/config/bind.go`); prefer env vars for these.
- `cmd/server/main.go` loads `.env` opportunistically via `godotenv`, then binds config from `./configs` by default.
- Use `configs/config.yaml.template` as the baseline and keep secrets in env vars for keys listed in `envBoundKeys` (`internal/pkg/config/bind.go`).

## Patterns to Follow When Editing
- Keep ports in domain and adapters in infra (example: `internal/domain/shared/event_publisher.go` + infra adapter in `internal/infra/domain_adapters.go`).
- Reuse `pkg/crud` list/filter/order abstractions in repo adapters (see `internal/infra/bun/bunrepo/users.go`) instead of ad-hoc SQL query parsing.
- Follow the existing email verification pattern in account flows: store only OTP hash, track status transitions (`active/used/expired/superseded/locked`), and enforce resend cooldown in app layer (`internal/app/account.go` + `internal/infra/bun/bunrepo/email_verification_codes.go`).
- JWT claims are custom-shaped via `pkg/jwtparser/jwt.go` (`tid`, `uid`, `usn`, `vrs`, plus unix-second `exp`/`iat`); keep claim keys and time semantics compatible when changing token generation/parsing.
- File URL resolution is consumed via `shared.FileURLResolver` (`internal/domain/shared/resource_url_resolver.go`) and implemented in `internal/domain/files/url_renderer.go`; support `http(s)://`, `bucket://`, and `catalog://id/...` forms.
- HTTP error shape for grpc-gateway is customized in `internal/infra/server/errors.go`; keep API errors compatible with that mapping.

## Known Template Drift / Checks
- `db/migrations/README.md` still shows a legacy example DSN/user (`skyline`) and `bash` snippets; for this template, prefer `task migrate` with the current env-based Postgres settings from `docker/local/docker-compose.yml`.
- Some implementation docs still reference `db/migrations/v2_email_verification_codes.sql`; current migrations are sequence-based (`db/migrations/000002_email_verification_codes.up.sql` / `.down.sql`).

## 对话和文档优先使用中文

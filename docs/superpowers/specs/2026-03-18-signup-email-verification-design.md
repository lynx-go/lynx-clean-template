# Sign-up Email Verification Design (OTP) / 注册邮箱验证码设计

Date: 2026-03-18
Status: Confirmed (Ready for planning)
Decision: Adopt Scheme 3 (dedicated verification table)

## 1. Goals / 目标

- 注册后用户先处于未验证状态，邮箱验证通过后才允许登录。
- 注册成功发送 6 位数字验证码。
- 规则固定：
  - 验证码有效期 10 分钟
  - 最多错误 5 次
  - 同一用户同一用途仅最新验证码有效
  - 重发冷却 60 秒
- 邮件发送能力抽象为通用端口：模板渲染 + 发送器可替换。
- 第一阶段只实现 mock 邮件服务商。

## 2. Non-goals / 非目标

- 第一阶段不接真实 SMTP/云邮件平台。
- 不包含前端页面与交互实现。
- 不包含短信等多渠道验证码。

## 3. Current State Assessment / 现状评估

当前 `users` 表已有 `email_confirmed_at`、`confirmation_token`、`confirmation_sent_at`、`status`，
但无法完整覆盖以下能力：失败次数、latest-only 强约束、一次性消费、重发审计。

结论：必须新增验证码专用表。

## 4. Final Architecture / 最终架构

### 4.1 Data Model / 数据模型

新增表：`email_verification_codes`

字段（最小集）：

- `id` varchar(64) pk
- `user_id` varchar(64) not null
- `email` varchar(256) not null
- `purpose` varchar(64) not null（首期：`signup_verify`）
- `code_hash` varchar(255) not null（不存明文）
- `status` int not null default 0
  - `0=active`, `1=used`, `2=expired`, `3=superseded`, `4=locked`
- `attempt_count` int not null default 0
- `max_attempts` int not null default 5
- `expires_at` timestamptz not null
- `sent_at` timestamptz not null
- `used_at` timestamptz null
- `created_at` timestamptz not null default now()
- `updated_at` timestamptz not null default now()
- `created_by` varchar(64) not null
- `updated_by` varchar(64) not null

约束与索引：

- 外键：`user_id -> users(id)`（防止孤儿记录）
- 普通索引：`(user_id, purpose, status)`、`(email, purpose, status)`、`(expires_at)`
- **部分唯一索引**：`(user_id, purpose)` where `status=0`（active）
  - 用于 DB 层保证 latest-only，避免并发产生多条 active

用户表变更：

- `users.email` 增加唯一约束
- 注册逻辑补充 email 冲突校验（现有仅校验 username）

### 4.2 Domain Ports / 领域端口

- `internal/domain/users/repo/email_verification_codes.go`
  - `Create(ctx, code)`
  - `GetLatestActive(ctx, userID, purpose)`
  - `SupersedeActiveByUserPurpose(ctx, userID, purpose)`
  - `IncrementAttempt(ctx, id)`
  - `MarkUsed(ctx, id, usedAt)`
  - `MarkStatus(ctx, id, status)`

- `internal/domain/shared/email_sender.go`
  - `Send(ctx, Message) error`

- `internal/domain/shared/email_template_renderer.go`
  - `Render(templateID string, vars map[string]string) (subject string, body string, err error)`

### 4.3 App Use-case Changes / 应用层改造

`internal/app/account.go`：

- `SignUp` 新流程：
  1. 校验 username/email 唯一
  2. 创建用户，状态置为未验证（`status=0`）
  3. 生成 6 位验证码并哈希
  4. 作废旧 active 码
  5. 写入新 active 码（`expires_at=now+10m`）
  6. 事务提交后发送邮件（失败仅记录日志，不回滚用户创建）

- 新增方法：
  - `VerifySignUpEmailCode(ctx, email, code)`
  - `ResendSignUpEmailCode(ctx, email)`

- `AuthorizeByPassword` 增加闸门：
  - `EmailConfirmedAt` 为空或 `Status != 1` 时返回 `email_not_verified`

校验流程（Verify）：

1. 按 email 查用户（email 已唯一）
2. 取最新 active 码（`user_id + purpose`）
3. 过期则置 expired 并返回 `verification_code_expired`
4. 比对哈希
5. 失败则 `attempt_count + 1`，到 5 次置 locked，返回 `verification_code_invalid`/`verification_code_locked`
6. 成功则在单事务内：`code -> used`，`user.email_confirmed_at/confirmed_at/status=1`

重发流程（Resend）：

1. 取最新 active 码
2. 若 `now - sent_at < 60s`，返回 `verification_code_rate_limited` + 剩余秒数
3. 作废 active 码（superseded）
4. 生成并保存新 active 码
5. 事务提交后发送邮件

说明：`locked` 作用于“当前验证码”，并不永久封禁账号；可在冷却后重发新码继续验证。

### 4.4 Transaction & Consistency / 事务与一致性

- 对状态更新采用单事务：
  - Verify 成功：`code used + user verified`
  - Resend：`supersede old + create new`
- 发邮件动作在事务提交后执行，避免发信失败导致事务一致性破坏。
- 通过 DB 唯一索引与事务组合，保证并发条件下 latest-only。

### 4.5 API Contract / 接口契约

在 `proto/api/v1/auth.proto` 新增：

- `rpc VerifySignUpEmail(VerifySignUpEmailRequest) returns (VerifySignUpEmailResponse)`
- `rpc ResendSignUpEmailCode(ResendSignUpEmailCodeRequest) returns (ResendSignUpEmailCodeResponse)`

消息：

- `VerifySignUpEmailRequest { string email; string code; }`
- `VerifySignUpEmailResponse { bool verified; }`
- `ResendSignUpEmailCodeRequest { string email; }`
- `ResendSignUpEmailCodeResponse { int64 next_retry_after_sec; }`

鉴权放行：

- 在 `internal/infra/server/grpc.go` 的 JWT 免鉴权名单中新增：
  - `AuthService/VerifySignUpEmail`
  - `AuthService/ResendSignUpEmailCode`

### 4.6 Infra Adapters / 基础设施适配

- Bun model/repo：`email_verification_codes`
- 邮件模板渲染：代码内置模板（`text/template`）
- 邮件发送器：mock provider（日志/内存模拟）

模板：

- `template_id = signup_email_code`
- vars: `code`, `expires_minutes`, `product_name`

### 4.7 Security / 安全

- OTP 仅存哈希，不存明文。
- OTP 生成采用安全随机源。
- 日志中禁止输出明文验证码。
- 对外错误语义稳定，避免泄漏敏感内部状态。

## 5. Error Contract / 错误契约

业务错误 key：

- `email_not_verified`
- `verification_code_invalid`
- `verification_code_expired`
- `verification_code_locked`
- `verification_code_rate_limited`
- `email_already_registered`

gRPC 映射建议（与 `internal/infra/server/errors.go` 保持兼容）：

- `email_not_verified` -> `FailedPrecondition`
- `verification_code_invalid` -> `InvalidArgument`
- `verification_code_expired` -> `FailedPrecondition`
- `verification_code_locked` -> `PermissionDenied`
- `verification_code_rate_limited` -> `ResourceExhausted`
- `email_already_registered` -> `AlreadyExists`

## 6. Testing Strategy / 测试策略

单元测试（先红后绿）：

- SignUp 生成未验证用户 + 发送验证码
- Token 在未验证邮箱时被拒绝
- Verify 成功后用户状态变为已验证
- Verify 错误码累计尝试次数，到 5 次锁定
- Verify 过期码失败
- Resend 冷却期内返回限流与剩余秒数

集成测试：

- Bun repo latest-only 与状态流转
- gRPC 链路：`SignUp -> VerifySignUpEmail -> Token`

## 7. Rollout Plan / 实施步骤

1. 新增 migration：`email_verification_codes` + `users.email` 唯一约束
2. 新增 domain ports/models
3. 新增 Bun model/repo + providers
4. 实现 mock sender + template renderer
5. 改造 `internal/app/account.go`
6. 更新 `proto/api/v1/auth.proto` 并 `task generate:proto`
7. 更新 gRPC handler + JWT 放行配置
8. 更新 Wire providers 并执行 `task wire`
9. 回归测试 `task test`

## 8. Acceptance Criteria / 验收标准

- 未验证邮箱用户无法获取 access token。
- 验证成功后可正常登录。
- 验证码 one-time use + latest-only 生效。
- 10 分钟有效、5 次错误锁定、60 秒重发冷却均生效。
- 邮件发送可通过端口替换 provider，应用层无需改动。


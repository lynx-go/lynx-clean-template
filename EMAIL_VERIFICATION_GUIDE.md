# Email Verification Feature - Quick Start Guide

完整的邮箱验证功能已实现。本指南说明如何在本地测试和使用。

## 前置条件

1. Docker + Docker Compose 已安装
2. Go 已安装
3. 项目已克隆

## 本地环境启动

### 步骤 1：启动本地基础设施（Postgres + Redis）

```bash
task up
```

此命令启动本地 Docker Compose 环境：
- PostgreSQL（端口 5432）
- Redis（端口 6379）

### 步骤 2：执行数据库迁移

```bash
# 设置必要的环境变量（如未设置）
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=password
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_DB=lynx

# 执行迁移（包括新的 email_verification_codes 表）
task migrate
```

### 步骤 3：启动服务器

```bash
task dev
```

服务器将在以下端点启动：
- gRPC：`localhost:8088`
- HTTP Gateway：`localhost:9099`

## API 使用说明

### 1. Sign Up（注册）

**请求**:
```bash
curl -X POST http://localhost:9099/v1/signUp \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "password123"
  }'
```

**响应**:
```json
{
  "user_info": {
    "id": "...",
    "username": "alice",
    "email": "alice@example.com",
    "status": 0,
    "...": "..."
  }
}
```

**重要**：注册后，用户状态为 `0`（未验证），并自动发送邮箱验证码。

**验证码日志**：检查服务器日志获取验证码（mock 邮件发送器输出）。
```
mock email sent to=alice@example.com subject="Your verification code" template_id=signup_email_code
```

### 2. Verify Email（验证邮箱）

从上一步的日志中提取验证码，或者运行：

```bash
# 注意：当前 mock 邮件发送器只输出日志，所以需要手动提取 6 位数字码
```

**请求**:
```bash
curl -X POST http://localhost:9099/v1/verify-email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "code": "123456"
  }'
```

**成功响应**:
```json
{
  "verified": true
}
```

**失败响应示例**:
- 验证码错误：`{"error": {"type": "invalid_request_error", "code": "InvalidArgument", "message": "verification_code_invalid"}}`
- 验证码过期：`{"error": {"type": "invalid_request_error", "code": "FailedPrecondition", "message": "verification_code_expired"}}`
- 验证码被锁定（5 次错误）：`{"error": {"type": "permission_error", "code": "PermissionDenied", "message": "verification_code_locked"}}`

### 3. Resend Email Code（重新发送验证码）

如果验证码已过期或丢失，可重新申请（60 秒冷却）。

**请求**:
```bash
curl -X POST http://localhost:9099/v1/resend-email-code \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com"
  }'
```

**成功响应**:
```json
{
  "next_retry_after_sec": 0
}
```

**冷却中响应**:
```json
{
  "error": {
    "type": "rate_limit_error",
    "code": "ResourceExhausted",
    "message": "verification_code_rate_limited",
    "next_retry_after_sec": 45
  }
}
```

### 4. Login（登录）

验证完成后，用户状态变为 `1`，可正常登录。

**验证前尝试登录** (应失败):
```bash
curl -X POST http://localhost:9099/v1/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "password",
    "email": "alice@example.com",
    "password": "password123"
  }'
```

**失败响应**:
```json
{
  "error": {
    "type": "authentication_error",
    "code": "Unauthenticated",
    "message": "email_not_verified"
  }
}
```

**验证后登录** (应成功):
```bash
# 先验证邮箱
curl -X POST http://localhost:9099/v1/verify-email \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "code": "123456"}'

# 然后登录
curl -X POST http://localhost:9099/v1/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "password",
    "email": "alice@example.com",
    "password": "password123"
  }'
```

**成功响应**:
```json
{
  "token": "eyJhbGc...",
  "expires_at": 1710768000,
  "refresh_token": "eyJhbGc...",
  "refresh_token_expires_at": 1710854400,
  "user_info": {
    "id": "...",
    "username": "alice",
    "email": "alice@example.com",
    "status": 1,
    "...": "..."
  }
}
```

## 规则总结

| 规则 | 实现 |
|------|------|
| 验证码格式 | 6 位数字 |
| 有效期 | 10 分钟 |
| 最大错误次数 | 5 次 |
| 重发冷却 | 60 秒 |
| 同用户仅最新有效 | ✅ DB 层强制 |
| OTP 存储方式 | SHA256 哈希 |
| 未验证用户状态 | 0 |
| 已验证用户状态 | 1 |

## 故障排查

### 验证码收不到
- 当前使用 mock 邮件发送器，只输出到日志
- 检查服务器控制台日志，搜索 `mock email sent`

### 验证失败显示"code not found"
- 确认验证码未过期（10 分钟）
- 确认输入的 email 与注册时相同

### 验证 5 次失败后无法重发
- 旧验证码已被锁定，调用 `ResendSignUpEmailCode` 获得新码
- 需等待 60 秒冷却后才能重发

### 登录仍然显示 "email_not_verified"
- 检查用户 `status` 字段是否为 1（不是 0）
- 检查 `email_confirmed_at` 字段是否已被设置

## 后续开发

### 替换邮件服务商

当需要集成真实邮件服务（如 SendGrid、AWS SES 等）时：

1. 实现新的 `EmailSender` adapter
2. 在 `internal/infra/domain_adapters.go` 或新建 adapter 文件
3. 更新 provider 注册
4. 无需修改应用层或 domain 层代码

示例：
```go
// internal/infra/mail/sendgrid_sender.go
type SendGridEmailSender struct {
    client *mail.Client
}

func (s *SendGridEmailSender) Send(ctx context.Context, msg shared.EmailMessage) error {
    // 调用 SendGrid API
    return nil
}
```

### 添加邮件模板

在 `internal/infra/domain_adapters.go` 的 `Render()` 方法添加新模板：

```go
case "password_reset":
    subject := "Reset your password"
    bodyTpl := "Click here to reset: {{.reset_link}}"
    body, err := renderTextTemplate(bodyTpl, vars)
    return subject, body, err
```

## 需要帮助？

查看设计文档：
- `docs/superpowers/specs/2026-03-18-signup-email-verification-design.md`
- `IMPLEMENTATION_SUMMARY.md`


# Email Verification Implementation Complete

## Summary of Changes

### 1. Domain Layer (新增端口)

- ✅ `internal/domain/shared/email_sender.go` - 邮件发送端口接口
- ✅ `internal/domain/shared/email_template_renderer.go` - 邮件模板渲染端口接口
- ✅ `internal/domain/users/repo/email_verification_codes.go` - 验证码仓储端口与数据模型

### 2. Infrastructure Layer (实现与适配)

**Bun ORM**:
- ✅ `internal/infra/bun/model/email_verification_code.go` - 数据库模型
- ✅ `internal/infra/bun/bunrepo/email_verification_codes.go` - Bun 仓储实现
- ✅ `internal/infra/bun/provides.go` - 注册新仓储到 Wire

**Adapters**:
- ✅ `internal/infra/domain_adapters.go` - 新增 mock 邮件发送器、内存模板渲染器
- ✅ `internal/infra/provides.go` - 注册邮件端口实现

**Database**:
- ✅ `db/migrations/v2_email_verification_codes.sql` - 新表 + 约束 + 索引 + email 唯一约束

### 3. Application Layer (用例与业务逻辑)

- ✅ `internal/domain/users/service.go`
  - 用户创建时状态设为 `0`（未验证）
  - 增加 email 冲突检查

- ✅ `internal/app/account.go` - 主要改动
  - 新增 `sendSignUpEmailCode()` - 生成、储存、发送验证码
  - 新增 `VerifySignUpEmailCode()` - 验证码验证与用户激活
  - 新增 `ResendSignUpEmailCode()` - 重发与冷却控制
  - 修改 `AuthorizeByPassword()` - 邮箱验证闸门
  - 修改 `CreateUser()` - 注册后发送验证码
  - 新增 helper 函数：`generateSixDigitCode()`, `hashCode()`, `verifyCodeHash()`

### 4. API Layer (gRPC 接口)

**Proto**:
- ✅ `proto/api/v1/auth.proto` - 新增两个 RPC + 消息定义
  - `VerifySignUpEmail`
  - `ResendSignUpEmailCode`

**Handler**:
- ✅ `internal/api/grpc/auth.go` - 新增两个处理方法
- ✅ `internal/infra/server/grpc.go` - 添加两个端点到 JWT 免鉴权名单

### 5. Code Generation & DI

- ✅ `proto/api/v1/auth.proto` 已生成 → `genproto/api/v1/*`
- ✅ `cmd/server/wire_gen.go` 已重新生成
- ✅ `cmd/cli/cmd/wire_gen.go` 已重新生成
- ✅ `tests/wire_gen.go` 已重新生成

## Core Business Rules Implemented

| Rule | Impl Status |
|------|------------|
| 6 位数字验证码 | ✅ |
| 10 分钟有效期 | ✅ |
| 5 次错误上限 | ✅ |
| 同用户仅最新码有效（DB 部分唯一索引）| ✅ |
| 重发 60 秒冷却 | ✅ |
| 未验证不能登录 | ✅ |
| 邮箱唯一性 | ✅ |
| OTP 哈希存储（不存明文）| ✅ |
| 模板渲染 + 发送端口可替换 | ✅ |
| Mock 邮件服务商 | ✅ |

## Files Changed (总计 15 文件)

**新增**: 7 文件
- `internal/domain/shared/email_sender.go`
- `internal/domain/shared/email_template_renderer.go`
- `internal/domain/users/repo/email_verification_codes.go`
- `internal/infra/bun/model/email_verification_code.go`
- `internal/infra/bun/bunrepo/email_verification_codes.go`
- `db/migrations/v2_email_verification_codes.sql`
- `docs/superpowers/specs/2026-03-18-signup-email-verification-design.md`

**修改**: 8 文件
- `internal/domain/users/service.go`
- `internal/app/account.go`
- `internal/api/grpc/auth.go`
- `internal/infra/domain_adapters.go`
- `internal/infra/provides.go`
- `internal/infra/bun/provides.go`
- `internal/infra/server/grpc.go`
- `proto/api/v1/auth.proto`

**重新生成**: 3 文件
- `cmd/server/wire_gen.go`
- `cmd/cli/cmd/wire_gen.go`
- `tests/wire_gen.go`

## Next Steps for Verification

1. **启动本地环境**
   ```bash
   task up
   ```

2. **执行数据库迁移**
   ```bash
   task migrate
   ```

3. **启动服务器**
   ```bash
   task dev
   ```

4. **测试 Sign-up 流程（生成验证码）**
   ```bash
   grpcurl -plaintext -d '{"username":"test","email":"test@example.com","password":"password123"}' \
     localhost:8088 lynx.api.v1.AuthService/SignUp
   ```

5. **测试 Verify Email（验证码验证）**
   - 检查 mock sender 日志获取验证码
   - 调用 VerifySignUpEmail 端点验证

6. **测试 Token 端点（登录验证）**
   - 未验证前：应返回 `email_not_verified`
   - 验证后：应正常返回 token

## Known Implementation Details

- **Mock 邮件发送**：当前只输出到日志，不真实发送
- **验证码 OTP**：使用 math/rand/v2 生成，SHA256 哈希存储
- **部分唯一索引**：确保同 user+purpose 仅一条 active 记录，DB 层强制一致性
- **错误语义**：gRPC 返回稳定 error key（如 `email_not_verified`）供前端按 key 处理

## Architecture Validation

✅ 遵循 Clean Architecture 分层
- Domain: 纯业务逻辑，无框架依赖
- App: 用例编排，依赖 domain ports
- Infra: 适配实现，依赖 domain ports（电话依赖指向内）
- API: gRPC handler 仅委托给 app use-case

✅ Wire DI 链路
- infra providers ← domain adapters + repos
- app providers ← domain services + repos + ports
- api providers ← app use-cases
- server bootstrap ← all above

✅ 数据一致性
- DB 层通过部分唯一索引 + 事务保证 latest-only
- 邮件发送在事务外（失败仅记日志，不回滚数据）


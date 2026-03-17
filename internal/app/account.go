package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	sharedpb "github.com/lynx-go/lynx-clean-template/genproto/shared"
	"github.com/lynx-go/lynx-clean-template/internal/domain/events"
	groupsrepo "github.com/lynx-go/lynx-clean-template/internal/domain/groups/repo"
	"github.com/lynx-go/lynx-clean-template/internal/domain/shared"
	"github.com/lynx-go/lynx-clean-template/internal/domain/users"
	usersrepo "github.com/lynx-go/lynx-clean-template/internal/domain/users/repo"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/contexts"
	apierrors "github.com/lynx-go/lynx-clean-template/pkg/errors"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen/uuid"
	"github.com/lynx-go/lynx-clean-template/pkg/jwtparser"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Account struct {
	usersRepo                  usersrepo.UsersRepo
	refreshTokensRepo          usersrepo.RefreshTokensRepo
	emailVerificationCodesRepo usersrepo.EmailVerificationCodesRepo
	groupsRepo                 groupsrepo.GroupsRepo
	config                     *config.AppConfig
	userSvc                    *users.UserService
	publisher                  shared.EventPublisher
	hasher                     shared.PasswordHasher
	logger                     shared.Logger
	emailTemplateRenderer      shared.EmailTemplateRenderer
	emailSender                shared.EmailSender
}

const (
	GrantTypePassword     = "password"
	GrantTypeRefreshToken = "refresh_token"
)

func NewAccount(
	usersRepo usersrepo.UsersRepo,
	refreshTokensRepo usersrepo.RefreshTokensRepo,
	emailVerificationCodesRepo usersrepo.EmailVerificationCodesRepo,
	config *config.AppConfig,
	groupsRepo groupsrepo.GroupsRepo,
	userSvc *users.UserService,
	publisher shared.EventPublisher,
	hasher shared.PasswordHasher,
	logger shared.Logger,
	emailTemplateRenderer shared.EmailTemplateRenderer,
	emailSender shared.EmailSender,
) *Account {
	return &Account{
		usersRepo:                  usersRepo,
		refreshTokensRepo:          refreshTokensRepo,
		emailVerificationCodesRepo: emailVerificationCodesRepo,
		config:                     config,
		groupsRepo:                 groupsRepo,
		userSvc:                    userSvc,
		publisher:                  publisher,
		hasher:                     hasher,
		logger:                     logger,
		emailTemplateRenderer:      emailTemplateRenderer,
		emailSender:                emailSender,
	}
}

func authFailed(msg string) *apierrors.APIError {
	return apierrors.New(401, msg)
}

func (uc *Account) AuthorizeByPassword(ctx context.Context, req *apipb.TokenRequest) (*apipb.TokenResponse, error) {
	email := req.GetEmail()
	username := req.GetUsername()
	password := req.GetPassword()
	if len(password) == 0 {
		return nil, authFailed("invalid username or password")
	}
	var user *usersrepo.User
	var err error
	if username != "" {
		user, err = uc.usersRepo.GetByUsername(ctx, username)
		if err != nil {
			return nil, errors.Wrap(err, "failed to query user")
		}
		if user == nil {
			return nil, authFailed("invalid username or password")
		}
	} else if email != "" {
		user, err = uc.usersRepo.GetByEmail(ctx, email)
		if err != nil {
			return nil, errors.Wrap(err, "failed to query user")
		}
		if user == nil {
			return nil, authFailed("invalid email or password")
		}
	}
	if user == nil {
		return nil, authFailed("invalid email or password")
	}
	if len(user.PasswordHash) == 0 {
		return nil, authFailed("invalid username or password")
	}

	// 检查邮箱是否已验证
	if user.Status != 1 || user.EmailConfirmedAt.IsZero() {
		return nil, apierrors.NewStatusError(http.StatusUnauthorized, "email_not_verified")
	}

	err = uc.hasher.Compare(user.PasswordHash, password)
	if err != nil {
		return nil, authFailed("invalid username or password")
	}
	now := time.Now()

	tokenId := uuid.NewString()
	tokenIssuedAt := now.Unix()
	meta := map[string]string{}
	token, exp, err := generateToken(uc.config, tokenId, tokenIssuedAt, user.ID.String(), username, meta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate access token")
	}
	refreshToken, refreshExp, err := generateRefreshToken(uc.config, tokenId, tokenIssuedAt, user.ID.String(), username, meta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate refresh token")
	}
	// 写入 refresh_token
	if _, err := uc.refreshTokensRepo.Create(ctx, usersrepo.RefreshTokenCreate{
		ID:        tokenId,
		UserID:    user.ID,
		Token:     refreshToken,
		CreatedAt: now,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to save refresh token")
	}

	// 更新用户最后登录时间
	user.LastSignInAt = now
	user.UpdatedAt = now
	if err := uc.usersRepo.Update(ctx, user, "last_sign_in_at", "updated_at"); err != nil {
		return nil, errors.Wrap(err, "failed to update user")
	}

	return &apipb.TokenResponse{
		Token:                 token,
		ExpiresAt:             exp,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExp,
		UserInfo:              NewProtoUserFromUser(user),
	}, nil
}

func generateToken(config *config.AppConfig, tokenID string, tokenIssuedAt int64, userID, username string, vars map[string]string) (string, int64, error) {
	exp := time.Now().UTC().Add(time.Duration(config.GetSecurity().GetJwt().TokenExpirySec) * time.Second).Unix()
	return generateTokenWithExpiry(config.GetSecurity().GetJwt().GetSecret(), tokenID, tokenIssuedAt, userID, username, vars, exp)
}

func generateRefreshToken(config *config.AppConfig, tokenID string, tokenIssuedAt int64, userID string, username string, vars map[string]string) (string, int64, error) {
	exp := time.Now().UTC().Add(time.Duration(config.GetSecurity().GetJwt().RefreshTokenExpirySec) * time.Second).Unix()
	return generateTokenWithExpiry(config.GetSecurity().GetJwt().GetRefreshTokenSecret(), tokenID, tokenIssuedAt, userID, username, vars, exp)
}

func generateTokenWithExpiry(signingKey, tokenID string, tokenIssuedAt int64, userID, username string, vars map[string]string, exp int64) (string, int64, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtparser.SessionTokenClaims{
		TokenId:   tokenID,
		UserId:    userID,
		Username:  username,
		Vars:      vars,
		ExpiresAt: exp,
		IssuedAt:  tokenIssuedAt,
	})
	signedToken, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to sign JWT token")
	}
	return signedToken, exp, nil
}

func (uc *Account) CreateUser(ctx context.Context, req *apipb.SignUpRequest, isSuperAdmin bool) (*apipb.SignUpResponse, error) {
	createReq := &users.UserCreateRequest{
		Username:     req.Username,
		DisplayName:  req.Username,
		Email:        req.Email,
		Password:     req.Password,
		IsSuperAdmin: isSuperAdmin,
	}
	id, err := uc.userSvc.Create(ctx, createReq)
	if err != nil {
		return nil, err
	}

	user, err := uc.usersRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// 生成并发送验证码（不阻塞主流程）
	if err := uc.sendSignUpEmailCode(ctx, user); err != nil {
		uc.logger.ErrorContext(ctx, "failed to send signup email code", err, "user_id", id)
	}

	// Publish event from app layer after successful domain operation
	if err := uc.publisher.Publish(ctx, events.TopicUserEvents.String(), events.EventAccountCreated.String(), &events.AccountCreatedEvent{
		UserID:      id,
		Username:    createReq.Username,
		DisplayName: createReq.DisplayName,
	}); err != nil {
		uc.logger.ErrorContext(ctx, "failed to publish users:created event", err)
	}


	return &apipb.SignUpResponse{
		UserInfo: NewProtoUserFromUser(user),
	}, nil
}

func (uc *Account) SignUp(ctx context.Context, req *apipb.SignUpRequest) (*apipb.SignUpResponse, error) {
	return uc.CreateUser(ctx, req, false)
}

func (uc *Account) RefreshToken(ctx context.Context, req *apipb.TokenRequest) (*apipb.TokenResponse, error) {
	reqToken := req.GetRefreshToken()
	userId, _, _, exp, tokenId, _, ok := jwtparser.ParseToken([]byte(uc.config.Security.Jwt.RefreshTokenSecret), reqToken)
	if !ok {
		return nil, apierrors.NewStatusError(http.StatusUnauthorized, "Refresh token invalid or expired.")
	}
	refreshToken, err := uc.refreshTokensRepo.Get(ctx, tokenId)
	if err != nil {
		return nil, err
	}
	if refreshToken == nil {
		return nil, apierrors.NewStatusError(http.StatusUnauthorized, "Refresh token invalid or expired.")
	}
	if refreshToken.Revoked {
		return nil, apierrors.NewStatusError(http.StatusUnauthorized, "Refresh token invalid or expired.")
	}
	expiresAt := time.Unix(exp, 0)
	now := time.Now()
	nowUTC := now.UTC()
	if nowUTC.After(expiresAt) {
		return nil, apierrors.NewStatusError(http.StatusUnauthorized, "Refresh token invalid or expired.")
	}
	if refreshToken.UserID.String() != userId {
		return nil, apierrors.NewStatusError(http.StatusUnauthorized, "Refresh token invalid or expired.")
	}
	user, err := uc.usersRepo.Get(ctx, refreshToken.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apierrors.NewStatusError(http.StatusUnauthorized, "User status invalid or expired.")
	}
	if user.Status != 1 {
		return nil, apierrors.NewStatusError(http.StatusUnauthorized, "User status invalid or expired.")
	}
	tokenIssuedAt := now.Unix()
	meta := map[string]string{}
	token, exp, err := generateToken(uc.config, tokenId, tokenIssuedAt, user.ID.String(), user.Username, meta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate access token")
	}
	return &apipb.TokenResponse{
		Token:     token,
		ExpiresAt: exp,
		UserInfo:  NewProtoUserFromUser(user),
	}, nil
}

func (uc *Account) Logout(ctx context.Context) error {
	// 从 context 中获取当前用户 ID
	uid, ok := contexts.UserID(ctx)
	if !ok || uid == "" {
		return apierrors.New(http.StatusUnauthorized, "not logged in")
	}

	// 注意：由于 JWT 无状态特性，我们主要撤销 refresh token
	// access token 会在过期后自动失效
	// 这里我们撤销该用户的所有 refresh tokens，实现完全登出
	// 如果只想撤销当前 token，需要从请求中获取 token_id

	// 简化实现：撤销用户的所有 refresh tokens
	// TODO: 如果需要只撤销当前 session，可以从 Authorization header 中解析 token_id
	now := time.Now()
	tokens, err := uc.refreshTokensRepo.ListByUser(ctx, uid)
	if err != nil {
		return errors.Wrap(err, "failed to query refresh tokens")
	}

	for _, token := range tokens {
		if !token.Revoked {
			if err := uc.refreshTokensRepo.Update(ctx, usersrepo.RefreshTokenUpdate{
				ID:        token.ID,
				Revoked:   lo.ToPtr(true),
				UpdatedAt: now,
			}); err != nil {
				uc.logger.ErrorContext(ctx, "failed to revoke refresh token", err, "token_id", token.ID)
			}
		}
	}

	return nil
}

// sendSignUpEmailCode generates a 6-digit OTP, supersedes old codes, and sends it via email.
func (uc *Account) sendSignUpEmailCode(ctx context.Context, user *usersrepo.User) error {
	now := time.Now()

	// Supersede any existing active codes for this user+purpose
	if err := uc.emailVerificationCodesRepo.SupersedeActiveByUserPurpose(
		ctx, user.ID, usersrepo.VerificationPurposeSignUp, now); err != nil {
		return errors.Wrap(err, "failed to supersede old codes")
	}

	// Generate 6-digit code
	code := generateSixDigitCode()
	codeHash := hashCode(code)

	// Create verification record
	verifyCode := &usersrepo.EmailVerificationCode{
		ID:          idgen.BigID(),
		UserID:      user.ID,
		Email:       user.Email,
		Purpose:     usersrepo.VerificationPurposeSignUp,
		CodeHash:    codeHash,
		Status:      usersrepo.VerificationCodeStatusActive,
		AttemptCount: 0,
		MaxAttempts: 5,
		ExpiresAt:   now.Add(10 * time.Minute),
		SentAt:      now,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   user.ID,
		UpdatedBy:   user.ID,
	}

	if err := uc.emailVerificationCodesRepo.Create(ctx, verifyCode); err != nil {
		return errors.Wrap(err, "failed to create verification code record")
	}

	// Render email template
	subject, body, err := uc.emailTemplateRenderer.Render("signup_email_code", map[string]string{
		"code":             code,
		"expires_minutes":  "10",
		"product_name":     "Lynx",
	})
	if err != nil {
		return errors.Wrap(err, "failed to render email template")
	}

	// Send email (async to not block signup)
	if err := uc.emailSender.Send(ctx, shared.EmailMessage{
		To:         user.Email,
		Subject:    subject,
		Body:       body,
		TemplateID: "signup_email_code",
	}); err != nil {
		return errors.Wrap(err, "failed to send email")
	}

	return nil
}

// VerifySignUpEmailCode verifies a code and marks user as verified if correct.
func (uc *Account) VerifySignUpEmailCode(ctx context.Context, email, code string) error {
	now := time.Now()

	// Lookup user by email
	user, err := uc.usersRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.Wrap(err, "failed to query user")
	}
	if user == nil {
		return apierrors.NewStatusError(http.StatusNotFound, "user_not_found")
	}

	// Get latest active code
	activeCode, err := uc.emailVerificationCodesRepo.GetLatestActive(
		ctx, user.ID, usersrepo.VerificationPurposeSignUp)
	if err != nil {
		return errors.Wrap(err, "failed to query verification code")
	}
	if activeCode == nil {
		return apierrors.NewStatusError(http.StatusBadRequest, "verification_code_not_found")
	}

	// Check expiry
	if now.After(activeCode.ExpiresAt) {
		if err := uc.emailVerificationCodesRepo.MarkStatus(
			ctx, activeCode.ID, usersrepo.VerificationCodeStatusExpired, now); err != nil {
			uc.logger.ErrorContext(ctx, "failed to mark code expired", err)
		}
		return apierrors.NewStatusError(http.StatusBadRequest, "verification_code_expired")
	}

	// Validate code
	if !verifyCodeHash(code, activeCode.CodeHash) {
		// Increment attempt
		attemptCount, err := uc.emailVerificationCodesRepo.IncrementAttempt(ctx, activeCode.ID, now)
		if err != nil {
			uc.logger.ErrorContext(ctx, "failed to increment attempt", err)
		}
		if attemptCount >= activeCode.MaxAttempts {
			if err := uc.emailVerificationCodesRepo.MarkStatus(
				ctx, activeCode.ID, usersrepo.VerificationCodeStatusLocked, now); err != nil {
				uc.logger.ErrorContext(ctx, "failed to mark code locked", err)
			}
			return apierrors.NewStatusError(http.StatusBadRequest, "verification_code_locked")
		}
		return apierrors.NewStatusError(http.StatusBadRequest, "verification_code_invalid")
	}

	// Mark code as used
	if err := uc.emailVerificationCodesRepo.MarkUsed(ctx, activeCode.ID, now); err != nil {
		return errors.Wrap(err, "failed to mark code used")
	}

	// Update user as verified
	user.EmailConfirmedAt = now
	user.ConfirmedAt = now
	user.Status = 1
	user.UpdatedAt = now
	if err := uc.usersRepo.Update(ctx, user, "email_confirmed_at", "confirmed_at", "status", "updated_at"); err != nil {
		return errors.Wrap(err, "failed to update user verification status")
	}

	return nil
}

// ResendSignUpEmailCode enforces cooldown and resends a new code.
func (uc *Account) ResendSignUpEmailCode(ctx context.Context, email string) (int64, error) {
	now := time.Now()

	// Lookup user by email
	user, err := uc.usersRepo.GetByEmail(ctx, email)
	if err != nil {
		return 0, errors.Wrap(err, "failed to query user")
	}
	if user == nil {
		return 0, apierrors.NewStatusError(http.StatusNotFound, "user_not_found")
	}

	// Check existing active code and cooldown
	activeCode, err := uc.emailVerificationCodesRepo.GetLatestActive(
		ctx, user.ID, usersrepo.VerificationPurposeSignUp)
	if err != nil {
		return 0, errors.Wrap(err, "failed to query verification code")
	}

	if activeCode != nil {
		elapsed := now.Sub(activeCode.SentAt).Seconds()
		if elapsed < 60 {
			remaining := int64(60 - int(elapsed))
			return remaining, apierrors.NewStatusError(http.StatusTooManyRequests, "verification_code_rate_limited")
		}
	}

	// Send new code
	if err := uc.sendSignUpEmailCode(ctx, user); err != nil {
		return 0, err
	}

	return 0, nil
}

// generateSixDigitCode returns a random 6-digit string.
func generateSixDigitCode() string {
	code := rand.IntN(1000000)
	return fmt.Sprintf("%06d", code)
}

// hashCode returns SHA256 hash of the code.
func hashCode(code string) string {
	h := sha256.New()
	h.Write([]byte(code))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// verifyCodeHash compares plaintext code against its hash.
func verifyCodeHash(code, hash string) bool {
	return hashCode(code) == hash
}

func NewProtoUserFromUser(v *usersrepo.User) *sharedpb.User {
	var userMeta *sharedpb.UserMetadata
	if md := v.UserMetadata; md != nil {
		userMeta = &sharedpb.UserMetadata{
			Bio:      md.Bio,
			Urls:     md.Urls,
			Language: md.Language,
		}
	} else {
		userMeta = &sharedpb.UserMetadata{}
	}
	if userMeta.Language == "" {
		userMeta.Language = "中文"
	}
	var appMeta *sharedpb.AppMetadata
	return &sharedpb.User{
		Id:             v.ID.String(),
		Username:       v.Username,
		DisplayName:    v.DisplayName,
		AvatarUrl:      v.AvatarURL,
		Phone:          v.Phone,
		Email:          v.Email,
		Gender:         int32(v.Gender),
		Status:         int32(v.Status),
		CreatedAtMs:    v.CreatedAt.UnixMilli(),
		UpdatedAtMs:    v.UpdatedAt.UnixMilli(),
		BannedUntilMs:  v.BannedUntil.UnixMilli(),
		LastSignInAtMs: v.LastSignInAt.UnixMilli(),
		IsSuperAdmin:   v.IsSuperAdmin,
		UserMetadata:   userMeta,
		AppMetadata:    appMeta,
	}
}

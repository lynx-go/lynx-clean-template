package app

import (
	"context"
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
	"github.com/lynx-go/lynx-clean-template/pkg/idgen/uuid"
	"github.com/lynx-go/lynx-clean-template/pkg/jwtparser"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Account struct {
	usersRepo         usersrepo.UsersRepo
	refreshTokensRepo usersrepo.RefreshTokensRepo
	groupsRepo        groupsrepo.GroupsRepo
	config            *config.AppConfig
	userSvc           *users.UserService
	publisher         shared.EventPublisher
	hasher            shared.PasswordHasher
	logger            shared.Logger
}

const (
	GrantTypePassword     = "password"
	GrantTypeRefreshToken = "refresh_token"
)

func NewAccount(
	usersRepo usersrepo.UsersRepo,
	refreshTokensRepo usersrepo.RefreshTokensRepo,
	config *config.AppConfig,
	groupsRepo groupsrepo.GroupsRepo,
	userSvc *users.UserService,
	publisher shared.EventPublisher,
	hasher shared.PasswordHasher,
	logger shared.Logger,
) *Account {
	return &Account{
		usersRepo:         usersRepo,
		refreshTokensRepo: refreshTokensRepo,
		config:            config,
		groupsRepo:        groupsRepo,
		userSvc:           userSvc,
		publisher:         publisher,
		hasher:            hasher,
		logger:            logger,
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

	// Publish event from app layer after successful domain operation
	if err := uc.publisher.Publish(ctx, events.TopicUserEvents.String(), events.EventAccountCreated.String(), &events.AccountCreatedEvent{
		UserID:      id,
		Username:    createReq.Username,
		DisplayName: createReq.DisplayName,
	}); err != nil {
		uc.logger.ErrorContext(ctx, "failed to publish users:created event", err)
	}

	user, err := uc.usersRepo.Get(ctx, id)
	if err != nil {
		return nil, err
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

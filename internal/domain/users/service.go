package users

import (
	"context"
	"math/rand/v2"
	"regexp"
	"time"

	"github.com/lynx-go/lynx-clean-template/internal/domain/shared"
	"github.com/lynx-go/lynx-clean-template/internal/domain/users/repo"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/lynx-go/lynx-clean-template/pkg/errors"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
)

type UserService struct {
	usersRepo repo.UsersRepo
	hasher    shared.PasswordHasher
	config    *config.AppConfig
}

var (
	invalidUsernameRegex = regexp.MustCompilePOSIX("([[:cntrl:]]|[[\t\n\r\f\v]])+")
	invalidCharsRegex    = regexp.MustCompilePOSIX("([[:cntrl:]]|[[:space:]])+")
	emailRegex           = regexp.MustCompile(`^.+@.+\..+$`)
)

func NewUserService(
	usersRepo repo.UsersRepo,
	hasher shared.PasswordHasher,
	config *config.AppConfig,
) *UserService {
	return &UserService{
		usersRepo: usersRepo,
		hasher:    hasher,
		config:    config,
	}
}

type UserCreateRequest struct {
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	//CreateDefaultTeamDisabled bool   `json:"create_default_team_disabled"`
}

func (req *UserCreateRequest) Validate() error {
	if req.Username == "" {
		return errors.Cause("username must not be empty")
	}
	if invalidUsernameRegex.Match([]byte(req.Username)) {
		return errors.Cause("invalid username")
	}
	if invalidCharsRegex.Match([]byte(req.Username)) {
		return errors.Cause("invalid username")
	}
	if !emailRegex.Match([]byte(req.Email)) {
		return errors.Cause("invalid email")
	}
	if len(req.Password) < 7 {
		return errors.Cause("password must be at least 7 characters")
	}
	return nil
}

func (svc *UserService) Create(ctx context.Context, req *UserCreateRequest) (idgen.ID, error) {
	if err := req.Validate(); err != nil {
		return "", err
	}
	email := req.Email
	username := req.Username
	user, err := svc.usersRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", errors.Wrap(err, "failed to query user")
	}
	if user != nil {
		return "", errors.Cause("username already registered")
	}
	existingByEmail, err := svc.usersRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.Wrap(err, "failed to query user")
	}
	if existingByEmail != nil {
		return "", errors.Cause("email already registered")
	}

	passwordHash, err := svc.hasher.Hash(req.Password)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate password hash")
	}
	now := time.Now()
	role := "user"
	if req.IsSuperAdmin {
		role = "admin"
	}

	// 生成用户 ID
	userID := idgen.BigID()

	// 创建用户
	newUser := &repo.User{
		ID:           userID,
		Username:     username,
		DisplayName:  req.DisplayName,
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         role,
		Status:       0,
		IsSuperAdmin: req.IsSuperAdmin,
		AvatarURL:    svc.randAvatarUrl(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if _, err := svc.usersRepo.CreateWithFirstUserSuperAdmin(ctx, newUser); err != nil {
		return "", errors.Wrap(err, "failed to create user")
	}

	return userID, nil
}

type UserChangePasswordRequest struct {
	UserID             idgen.ID `json:"user_id"`
	OldPassword        string   `json:"old_password"`
	NewPassword        string   `json:"new_password"`
	RequireOldPassword bool     `json:"require_old_password"`
	UpdatedBy          idgen.ID `json:"updated_by"`
}

func (req *UserChangePasswordRequest) Validate() error {
	if req.UserID == "" {
		return errors.Cause("user_id must not be empty")
	}
	if req.RequireOldPassword && req.OldPassword == "" {
		return errors.Cause("old_password is required")
	}
	if len(req.NewPassword) < 7 {
		return errors.Cause("password must be at least 7 characters")
	}
	return nil
}

func (svc *UserService) ChangePassword(ctx context.Context, req *UserChangePasswordRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	user, err := svc.usersRepo.Get(ctx, req.UserID)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}
	if user == nil {
		return errors.Cause("user not found")
	}

	// Verify old password if required
	if req.RequireOldPassword {
		if err := svc.hasher.Compare(user.PasswordHash, req.OldPassword); err != nil {
			return errors.Cause("old password is incorrect")
		}
	}

	passwordHash, err := svc.hasher.Hash(req.NewPassword)
	if err != nil {
		return errors.Wrap(err, "failed to generate password hash")
	}

	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()
	if req.UpdatedBy != "" {
		user.UpdatedBy = req.UpdatedBy
	}

	if err := svc.usersRepo.Update(ctx, user, "password_hash", "updated_at", "updated_by"); err != nil {
		return errors.Wrap(err, "failed to update password")
	}

	return nil
}

func (svc *UserService) randAvatarUrl() string {
	n := rand.IntN(len(svc.config.
		App.User.DefaultAvatarUrls))
	return svc.config.App.User.DefaultAvatarUrls[n]
}

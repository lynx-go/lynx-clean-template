package app

import (
	"context"
	"time"

	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	sharedpb "github.com/lynx-go/lynx-clean-template/genproto/shared"
	"github.com/lynx-go/lynx-clean-template/internal/domain/files"
	"github.com/lynx-go/lynx-clean-template/internal/domain/users/repo"
	apierrors "github.com/lynx-go/lynx-clean-template/pkg/errors"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// protobuf field name to database column name mapping for user updates
var userFieldMapping = map[string]string{
	"display_name":       "display_name",
	"avatar_url":         "avatar_url",
	"phone":              "phone",
	"email":              "email",
	"gender":             "gender",
	"status":             "status",
	"role":               "role",
	"is_super_admin":     "is_super_admin",
	"password_hash":      "password_hash",
	"app_metadata":       "app_metadata",
	"user_metadata":      "user_metadata",
	"updated_at":         "updated_at",
	"phone_confirmed_at": "phone_confirmed_at",
	"email_confirmed_at": "email_confirmed_at",
	"confirmed_at":       "confirmed_at",
	"last_sign_in_at":    "last_sign_in_at",
	"banned_until":       "banned_until",
}

// toColumnNames converts protobuf field names to database column names
func toColumnNames(paths []string) []string {
	columns := make([]string, 0, len(paths))
	for _, path := range paths {
		if col, ok := userFieldMapping[path]; ok {
			columns = append(columns, col)
		}
	}
	return columns
}

func NewUsers(
	usersRepo repo.UsersRepo,
	urlRenderer files.URLRenderer,
) *Users {
	return &Users{
		usersRepo:   usersRepo,
		urlRenderer: urlRenderer,
	}
}

type Users struct {
	usersRepo   repo.UsersRepo
	urlRenderer files.URLRenderer
}

// GetUserProfile returns a user's profile by ID
// When user_id is "-", returns the authenticated user's profile
// Only allows viewing own profile or requires admin/superadmin role
func (uc *Users) GetUserProfile(ctx context.Context, req *apipb.GetUserProfileRequest) (*apipb.UserProfile, error) {
	var userID idgen.ID
	currentUser := session.LoginUser(ctx)

	// When user_id is "-", return the current user's profile
	if req.UserId == "-" {
		userID = currentUser
	} else {
		userID = idgen.ID(req.UserId)
	}

	// Security check: Only allow viewing own profile or require admin role
	// Check if current user is viewing their own profile
	isOwnProfile := userID == currentUser
	if !isOwnProfile {
		// Check if user is superadmin - superadmin can view any profile
		isSuperAdmin, err := uc.usersRepo.IsSuperAdmin(ctx, currentUser)
		if err != nil {
			return nil, apierrors.Wrap(err, "failed to check admin status")
		}
		if !isSuperAdmin {
			return nil, apierrors.New(403, "You can only view your own profile")
		}
	}

	user, err := uc.usersRepo.Get(ctx, userID)
	if err != nil {
		return nil, apierrors.Wrap(err, "failed to query user info")
	}
	if user == nil {
		return nil, apierrors.Cause("user not found")
	}
	return toUserProfile(ctx, user, uc.urlRenderer), nil
}

func (uc *Users) UpdateMyProfile(ctx context.Context, req *apipb.UpdateMyProfileRequest) (*apipb.UserProfile, error) {
	uid := session.LoginUser(ctx)

	if req.User == nil {
		return nil, apierrors.New(400, "User data is required")
	}

	// Get current user to merge with updates
	currentUser, err := uc.usersRepo.Get(ctx, uid)
	if err != nil {
		return nil, apierrors.Wrap(err, "failed to query user info")
	}
	if currentUser == nil {
		return nil, apierrors.Cause("user not found")
	}

	// Update fields based on FieldMask or update all non-empty fields
	now := time.Now()
	currentUser.UpdatedAt = now

	var updateColumns []string

	// If update_mask is provided, only update specified fields
	// Otherwise, update all non-empty fields
	if req.UpdateMask != nil && len(req.UpdateMask.Paths) > 0 {
		for _, path := range req.UpdateMask.Paths {
			switch path {
			case "display_name":
				if req.User.DisplayName != "" {
					currentUser.DisplayName = req.User.DisplayName
				}
			case "avatar_url":
				if req.User.AvatarUrl != "" {
					currentUser.AvatarURL = req.User.AvatarUrl
				}
			case "phone":
				if req.User.Phone != "" {
					currentUser.Phone = req.User.Phone
				}
			case "email":
				if req.User.Email != "" {
					currentUser.Email = req.User.Email
				}
			case "user_metadata":
				if req.User.UserMetadata != nil {
					currentUser.UserMetadata = toRepoUserMetadata(req.User.UserMetadata)
				}
			}
		}
		updateColumns = toColumnNames(req.UpdateMask.Paths)
		// Always include updated_at
		updateColumns = append(updateColumns, "updated_at")
	} else {
		// Update all non-empty fields
		if req.User.DisplayName != "" {
			currentUser.DisplayName = req.User.DisplayName
		}
		if req.User.AvatarUrl != "" {
			currentUser.AvatarURL = req.User.AvatarUrl
		}
		if req.User.UserMetadata != nil {
			currentUser.UserMetadata = toRepoUserMetadata(req.User.UserMetadata)
		}
		// No updateColumns specified - repo will update all columns
	}

	if err := uc.usersRepo.Update(ctx, currentUser, updateColumns...); err != nil {
		return nil, apierrors.Wrap(err, "failed to update user info")
	}

	user, err := uc.usersRepo.Get(ctx, uid)
	if err != nil {
		return nil, apierrors.Wrap(err, "failed to query user info")
	}
	if user == nil {
		return nil, apierrors.Cause("user not found")
	}

	return toUserProfile(ctx, user, uc.urlRenderer), nil
}

// toUserProfile converts repo.User to apipb.UserProfile
func toUserProfile(ctx context.Context, v *repo.User, urlRenderer files.URLRenderer) *apipb.UserProfile {
	// Render avatar URL through URLRenderer
	avatarURL := urlRenderer.Render(ctx, v.AvatarURL)

	user := &apipb.UserProfile{
		Id:           v.ID.String(),
		Username:     v.Username,
		DisplayName:  v.DisplayName,
		AvatarUrl:    avatarURL,
		Phone:        v.Phone,
		Email:        v.Email,
		Gender:       int32(v.Gender),
		Status:       int32(v.Status),
		IsSuperAdmin: v.IsSuperAdmin,
	}

	// Handle time fields - check for zero values to avoid nil timestamp issues
	if !v.CreatedAt.IsZero() {
		user.CreatedAt = timestamppb.New(v.CreatedAt)
	}
	if !v.UpdatedAt.IsZero() {
		user.UpdatedAt = timestamppb.New(v.UpdatedAt)
	}
	if !v.LastSignInAt.IsZero() {
		user.LastSignInAt = timestamppb.New(v.LastSignInAt)
	}
	if !v.BannedUntil.IsZero() {
		user.BannedUntil = timestamppb.New(v.BannedUntil)
	}

	if v.UserMetadata != nil {
		user.UserMetadata = &sharedpb.UserMetadata{
			Bio:      v.UserMetadata.Bio,
			Urls:     v.UserMetadata.Urls,
			Language: v.UserMetadata.Language,
		}
	}

	return user
}

// toRepoUserMetadata converts sharedpb.UserMetadata to repo.UserMetadata
func toRepoUserMetadata(meta *sharedpb.UserMetadata) *repo.UserMetadata {
	if meta == nil {
		return nil
	}
	return &repo.UserMetadata{
		Bio:      meta.Bio,
		Urls:     meta.Urls,
		Language: meta.Language,
	}
}

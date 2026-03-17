package mapper

import (
	"github.com/lynx-go/lynx-clean-template/internal/domain/users/repo"
	"github.com/lynx-go/lynx-clean-template/internal/infra/bun/model"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"github.com/samber/lo"
)

// ToDomainUser converts a Bun model.User to a domain repo.User.
func ToDomainUser(u *model.User) *repo.User {
	user := &repo.User{
		ID:           idgen.ID(u.ID),
		Username:     u.Username,
		DisplayName:  u.DisplayName,
		PasswordHash: u.PasswordHash,
		AvatarURL:    u.AvatarURL,
		Phone:        u.Phone,
		Email:        u.Email,
		Status:       u.Status,
		Gender:       u.Gender,
		Role:         u.Role,
		IsSuperAdmin: u.IsSuperAdmin,
		AppMetadata:  ToAppMetadata(u.AppMetadata),
		UserMetadata: ToUserMetadata(u.UserMetadata),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		CreatedBy:    idgen.ID(u.CreatedBy),
		UpdatedBy:    idgen.ID(u.UpdatedBy),
	}

	if !u.PhoneConfirmedAt.IsZero() {
		user.PhoneConfirmedAt = u.PhoneConfirmedAt
	}
	if !u.EmailConfirmedAt.IsZero() {
		user.EmailConfirmedAt = u.EmailConfirmedAt
	}
	if !u.ConfirmedAt.IsZero() {
		user.ConfirmedAt = u.ConfirmedAt
	}
	if !u.LastSignInAt.IsZero() {
		user.LastSignInAt = u.LastSignInAt
	}
	if !u.BannedUntil.IsZero() {
		user.BannedUntil = u.BannedUntil
	}

	return user
}

// FromAppMetadata converts a domain AppMetadata to a map for Bun JSONB storage.
func FromAppMetadata(m *repo.AppMetadata) map[string]any {
	if m == nil {
		return nil
	}
	return map[string]any{}
}

// ToAppMetadata converts a Bun JSONB map to a domain AppMetadata.
func ToAppMetadata(m map[string]any) *repo.AppMetadata {
	if len(m) == 0 {
		return &repo.AppMetadata{}
	}
	return &repo.AppMetadata{}
}

// FromUserMetadata converts a domain UserMetadata to a map for Bun JSONB storage.
func FromUserMetadata(m *repo.UserMetadata) map[string]any {
	if m == nil {
		return nil
	}
	return map[string]any{
		"bio":      m.Bio,
		"urls":     m.Urls,
		"language": m.Language,
	}
}

// ToUserMetadata converts a Bun JSONB map to a domain UserMetadata.
func ToUserMetadata(m map[string]any) *repo.UserMetadata {
	if len(m) == 0 {
		return &repo.UserMetadata{}
	}

	result := &repo.UserMetadata{}
	if v, ok := m["bio"].(string); ok {
		result.Bio = v
	}
	if v, ok := m["urls"].([]any); ok {
		result.Urls = lo.Map(v, func(s any, _ int) string {
			if str, ok := s.(string); ok {
				return str
			}
			return ""
		})
	}
	if v, ok := m["language"].(string); ok {
		result.Language = v
	}
	return result
}

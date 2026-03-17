package bunrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lynx-go/lynx-clean-template/internal/domain/users/repo"
	"github.com/lynx-go/lynx-clean-template/internal/infra/bun/model"
	"github.com/lynx-go/lynx-clean-template/internal/infra/clients"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"github.com/uptrace/bun"
)

type EmailVerificationCodesRepo struct {
	db *bun.DB
}

func NewEmailVerificationCodesRepo(data *clients.DataClients) repo.EmailVerificationCodesRepo {
	return &EmailVerificationCodesRepo{db: data.GetBunDB()}
}

func (r *EmailVerificationCodesRepo) Create(ctx context.Context, code *repo.EmailVerificationCode) error {
	m := &model.EmailVerificationCode{
		ID:           code.ID.String(),
		UserID:       code.UserID.String(),
		Email:        code.Email,
		Purpose:      string(code.Purpose),
		CodeHash:     code.CodeHash,
		Status:       int8(code.Status),
		AttemptCount: code.AttemptCount,
		MaxAttempts:  code.MaxAttempts,
		ExpiresAt:    code.ExpiresAt,
		SentAt:       code.SentAt,
		CreatedAt:    code.CreatedAt,
		UpdatedAt:    code.UpdatedAt,
		CreatedBy:    code.CreatedBy.String(),
		UpdatedBy:    code.UpdatedBy.String(),
	}
	if !code.UsedAt.IsZero() {
		m.UsedAt = code.UsedAt
	}
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	return err
}

func (r *EmailVerificationCodesRepo) GetLatestActive(ctx context.Context, userID idgen.ID, purpose repo.VerificationPurpose) (*repo.EmailVerificationCode, error) {
	var m model.EmailVerificationCode
	err := r.db.NewSelect().
		Model(&m).
		Where("evc.user_id = ?", userID.String()).
		Where("evc.purpose = ?", string(purpose)).
		Where("evc.status = ?", int8(repo.VerificationCodeStatusActive)).
		Order("evc.created_at DESC").
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainEmailVerificationCode(&m), nil
}

func (r *EmailVerificationCodesRepo) SupersedeActiveByUserPurpose(ctx context.Context, userID idgen.ID, purpose repo.VerificationPurpose, now time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*model.EmailVerificationCode)(nil)).
		Set("status = ?", int8(repo.VerificationCodeStatusSuperseded)).
		Set("updated_at = ?", now).
		Where("user_id = ?", userID.String()).
		Where("purpose = ?", string(purpose)).
		Where("status = ?", int8(repo.VerificationCodeStatusActive)).
		Exec(ctx)
	return err
}

func (r *EmailVerificationCodesRepo) IncrementAttempt(ctx context.Context, id idgen.ID, now time.Time) (int, error) {
	var m model.EmailVerificationCode
	err := r.db.NewSelect().
		Model(&m).
		Where("evc.id = ?", id.String()).
		Scan(ctx)
	if err != nil {
		return 0, err
	}

	attemptCount := m.AttemptCount + 1
	_, err = r.db.NewUpdate().
		Model((*model.EmailVerificationCode)(nil)).
		Set("attempt_count = ?", attemptCount).
		Set("updated_at = ?", now).
		Where("id = ?", id.String()).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return attemptCount, nil
}

func (r *EmailVerificationCodesRepo) MarkUsed(ctx context.Context, id idgen.ID, usedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*model.EmailVerificationCode)(nil)).
		Set("status = ?", int8(repo.VerificationCodeStatusUsed)).
		Set("used_at = ?", usedAt).
		Set("updated_at = ?", usedAt).
		Where("id = ?", id.String()).
		Exec(ctx)
	return err
}

func (r *EmailVerificationCodesRepo) MarkStatus(ctx context.Context, id idgen.ID, status repo.VerificationCodeStatus, now time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*model.EmailVerificationCode)(nil)).
		Set("status = ?", int8(status)).
		Set("updated_at = ?", now).
		Where("id = ?", id.String()).
		Exec(ctx)
	return err
}

func toDomainEmailVerificationCode(m *model.EmailVerificationCode) *repo.EmailVerificationCode {
	v := &repo.EmailVerificationCode{
		ID:           idgen.ID(m.ID),
		UserID:       idgen.ID(m.UserID),
		Email:        m.Email,
		Purpose:      repo.VerificationPurpose(m.Purpose),
		CodeHash:     m.CodeHash,
		Status:       repo.VerificationCodeStatus(m.Status),
		AttemptCount: m.AttemptCount,
		MaxAttempts:  m.MaxAttempts,
		ExpiresAt:    m.ExpiresAt,
		SentAt:       m.SentAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		CreatedBy:    idgen.ID(m.CreatedBy),
		UpdatedBy:    idgen.ID(m.UpdatedBy),
	}
	if !m.UsedAt.IsZero() {
		v.UsedAt = m.UsedAt
	}
	return v
}


package repo

import (
	"context"
	"time"

	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
)

type VerificationPurpose string

const (
	VerificationPurposeSignUp VerificationPurpose = "signup_verify"
)

type VerificationCodeStatus int8

const (
	VerificationCodeStatusActive VerificationCodeStatus = iota
	VerificationCodeStatusUsed
	VerificationCodeStatusExpired
	VerificationCodeStatusSuperseded
	VerificationCodeStatusLocked
)

type EmailVerificationCode struct {
	ID          idgen.ID
	UserID      idgen.ID
	Email       string
	Purpose     VerificationPurpose
	CodeHash    string
	Status      VerificationCodeStatus
	AttemptCount int
	MaxAttempts int
	ExpiresAt   time.Time
	SentAt      time.Time
	UsedAt      time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   idgen.ID
	UpdatedBy   idgen.ID
}

type EmailVerificationCodesRepo interface {
	Create(ctx context.Context, code *EmailVerificationCode) error
	GetLatestActive(ctx context.Context, userID idgen.ID, purpose VerificationPurpose) (*EmailVerificationCode, error)
	SupersedeActiveByUserPurpose(ctx context.Context, userID idgen.ID, purpose VerificationPurpose, now time.Time) error
	IncrementAttempt(ctx context.Context, id idgen.ID, now time.Time) (attemptCount int, err error)
	MarkUsed(ctx context.Context, id idgen.ID, usedAt time.Time) error
	MarkStatus(ctx context.Context, id idgen.ID, status VerificationCodeStatus, now time.Time) error
}


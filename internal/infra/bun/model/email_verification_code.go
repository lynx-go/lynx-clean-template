package model

import (
	"time"

	"github.com/uptrace/bun"
)

type EmailVerificationCode struct {
	bun.BaseModel `bun:"table:email_verification_codes,alias:evc"`

	ID           string    `bun:"id,pk"`
	UserID       string    `bun:"user_id"`
	Email        string    `bun:"email"`
	Purpose      string    `bun:"purpose"`
	CodeHash     string    `bun:"code_hash"`
	Status       int8      `bun:"status"`
	AttemptCount int       `bun:"attempt_count"`
	MaxAttempts  int       `bun:"max_attempts"`
	ExpiresAt    time.Time `bun:"expires_at"`
	SentAt       time.Time `bun:"sent_at"`
	UsedAt       time.Time `bun:"used_at"`
	CreatedAt    time.Time `bun:"created_at"`
	UpdatedAt    time.Time `bun:"updated_at"`
	CreatedBy    string    `bun:"created_by"`
	UpdatedBy    string    `bun:"updated_by"`
}


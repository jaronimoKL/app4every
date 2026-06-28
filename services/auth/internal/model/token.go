package model

import "time"

const (
	PurposeEmailVerify   = "email_verify"
	PurposePasswordReset = "password_reset"
)

type VerificationToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	Purpose   string    `json:"purpose"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

package model

import "time"

// User — основная структура пользователя.
// json:"-" означает что поле никогда не попадёт в JSON-ответ (хэш пароля).
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// --- Запросы Auth ---

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`   // опционально — можно регистрироваться только с логином
	Password string `json:"password"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"` // email ИЛИ username — сервер определяет сам
	Password   string `json:"password"`
}

type AuthResponse struct {
	User        User   `json:"user"`
	AccessToken string `json:"access_token"`
}

// --- Запросы профиля ---

type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// --- Запросы сброса пароля ---

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

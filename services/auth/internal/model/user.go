package model

import "time"

// User — основная структура пользователя.
// json:"-" означает что поле никогда не попадёт в JSON-ответ (хэш пароля).
type User struct {
	ID           int64     `json:"id"`
	Username              string    `json:"username"`
	Email                 string    `json:"email"`
	EmailVerified         bool      `json:"email_verified"`
	PasswordHash          string    `json:"-"`
	ShikimoriAccessToken  *string   `json:"shikimori_access_token,omitempty"`
	ShikimoriRefreshToken *string   `json:"-"`
	ShikimoriUserID       *int64    `json:"shikimori_user_id,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type Friendship struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	FriendID  int64     `json:"friend_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Запросы Auth ---

type RegisterRequest struct {
	Username   string `json:"username"`
	Email      string `json:"email"`   // опционально — можно регистрироваться только с логином
	Password   string `json:"password"`
	InviteCode string `json:"invite_code"` // Required for registration
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
	Identifier string `json:"identifier"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

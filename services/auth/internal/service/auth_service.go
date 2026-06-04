package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"app4every/services/auth/internal/config"
	"app4every/services/auth/internal/model"
	"app4every/services/auth/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidToken        = errors.New("invalid refresh token")
	ErrWrongPassword       = errors.New("current password is incorrect")
	ErrUserAlreadyExists   = errors.New("email or username already taken")
)

type AuthService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.User, string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error

	// Профиль
	UpdateProfile(ctx context.Context, userID int64, req model.UpdateProfileRequest) (*model.User, error)
	ChangePassword(ctx context.Context, userID int64, req model.ChangePasswordRequest) error

	// Сброс пароля (заглушки — реальная отправка email будет позже)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type authService struct {
	cfg         *config.Config
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewAuthService(cfg *config.Config, userRepo repository.UserRepository, sessionRepo repository.SessionRepository) AuthService {
	return &authService{
		cfg:         cfg,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// ── Auth ──

func (s *authService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, req.Username, req.Email, string(hashedPassword))
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(ctx context.Context, req model.LoginRequest) (*model.User, string, string, error) {
	// GetByIdentifier ищет по email ИЛИ по username — одним запросом
	user, err := s.userRepo.GetByIdentifier(ctx, req.Identifier)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", "", ErrInvalidCredentials
		}
		return nil, "", "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, "", "", err
	}

	// Храним Refresh Token в Redis на 30 дней
	if err = s.sessionRepo.Set(ctx, refreshToken, user.ID, 30*24*time.Hour); err != nil {
		return nil, "", "", fmt.Errorf("failed to save session: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	userID, err := s.sessionRepo.Get(ctx, refreshToken)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	_ = s.sessionRepo.Delete(ctx, refreshToken)

	newAccessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	if err = s.sessionRepo.Set(ctx, newRefreshToken, userID, 30*24*time.Hour); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.sessionRepo.Delete(ctx, refreshToken)
}

// ── Профиль ──

// UpdateProfile обновляет username и email.
// Проверка на уникальность — на уровне БД (UNIQUE constraint).
func (s *authService) UpdateProfile(ctx context.Context, userID int64, req model.UpdateProfileRequest) (*model.User, error) {
	user, err := s.userRepo.UpdateProfile(ctx, userID, req.Username, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	return user, nil
}

// ChangePassword проверяет текущий пароль, затем устанавливает новый.
func (s *authService) ChangePassword(ctx context.Context, userID int64, req model.ChangePasswordRequest) error {
	// Получаем пользователя чтобы сравнить пароль
	// GetByID возвращает PasswordHash
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Проверяем что текущий пароль верный
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return ErrWrongPassword
	}

	// Хэшируем новый пароль
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	return s.userRepo.UpdatePassword(ctx, userID, string(newHash))
}

// ── Сброс пароля (заглушки) ──

// ForgotPassword — заглушка. В продакшне здесь будет отправка письма.
// Намеренно не говорим клиенту, существует ли email — защита от перебора.
func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	// TODO: интеграция с SMTP / SendGrid / Resend
	// 1. Сгенерировать токен: token := generateSecureToken()
	// 2. Сохранить в Redis с TTL 1 час: redis.Set("reset:"+token, userID, 1h)
	// 3. Отправить письмо с ссылкой: /reset-password?token=<token>
	log.Printf("[ForgotPassword STUB] Запрос сброса пароля для: %s", email)
	return nil
}

// ResetPassword — заглушка. В продакшне проверяет токен из Redis.
func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// TODO:
	// 1. Получить userID из Redis: userID, err := redis.Get("reset:"+token)
	// 2. Удалить токен: redis.Del("reset:"+token)
	// 3. Захэшировать и сохранить новый пароль
	log.Printf("[ResetPassword STUB] Попытка сброса пароля с токеном: %s", token)
	return nil
}

// ── Вспомогательные функции ──

func (s *authService) generateAccessToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *authService) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

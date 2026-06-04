package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"app4every/services/auth/internal/config"
	"app4every/services/auth/internal/model"
	"app4every/services/auth/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid refresh token")
)

type AuthService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.User, string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
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

func (s *authService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return s.userRepo.Create(ctx, req.Email, string(hashedPassword))
}

func (s *authService) Login(ctx context.Context, req model.LoginRequest) (*model.User, string, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", "", ErrInvalidCredentials
		}
		return nil, "", "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
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

	// Храним Refresh Token в Redis 30 дней
	err = s.sessionRepo.Set(ctx, refreshToken, user.ID, 30*24*time.Hour)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to save session: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	userID, err := s.sessionRepo.Get(ctx, refreshToken)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	// Удаляем старый Refresh Token
	_ = s.sessionRepo.Delete(ctx, refreshToken)

	newAccessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Сохраняем новый Refresh Token
	err = s.sessionRepo.Set(ctx, newRefreshToken, userID, 30*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.sessionRepo.Delete(ctx, refreshToken)
}

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
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

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
	ErrInvalidCredentials       = errors.New("invalid email or password")
	ErrInvalidToken              = errors.New("invalid refresh token")
	ErrWrongPassword             = errors.New("current password is incorrect")
	ErrUserAlreadyExists         = errors.New("email or username already taken")
	ErrCannotFriendSelf          = errors.New("cannot friend yourself")
	ErrFriendshipAlreadyExists   = errors.New("friendship already exists or pending")
	ErrUserNotFound              = errors.New("user not found")
)

type AuthService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.User, string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	GetConfig() *config.Config

	// Профиль
	UpdateProfile(ctx context.Context, userID int64, req model.UpdateProfileRequest) (*model.User, error)
	ChangePassword(ctx context.Context, userID int64, req model.ChangePasswordRequest) error

	// Сброс пароля (заглушки — реальная отправка email будет позже)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error

	// Дружба
	GetFriends(ctx context.Context, userID int64) ([]*model.User, error)
	GetRequests(ctx context.Context, userID int64) ([]*model.User, error)
	SendRequest(ctx context.Context, userID int64, identifier string) error
	AcceptRequest(ctx context.Context, userID, targetID int64) error
	DeclineRequest(ctx context.Context, userID, targetID int64) error
	DeleteFriend(ctx context.Context, userID, targetID int64) error
	SearchUsers(ctx context.Context, q string, excludeID int64) ([]*model.User, error)
	
	// Invite Codes
	GenerateInvite(ctx context.Context, userID int64) (*model.InviteCode, error)
	ListUserInvites(ctx context.Context, userID int64) ([]*model.InviteCode, error)

	// Shikimori
	ShikimoriCallback(ctx context.Context, userID int64, code string) error
	GetShikimoriRates(ctx context.Context, userID int64) ([]byte, error)
	SyncShikimoriRate(ctx context.Context, userID int64, payload []byte) ([]byte, error)
	
	SetNotificationService(svc NotificationService)
}

type authService struct {
	cfg         *config.Config
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	inviteRepo  repository.InviteRepository
	tokenRepo   repository.VerificationTokenRepository
	mailerSvc   MailerService
	notifSvc    NotificationService
}

func NewAuthService(cfg *config.Config, userRepo repository.UserRepository, sessionRepo repository.SessionRepository, inviteRepo repository.InviteRepository, tokenRepo repository.VerificationTokenRepository, mailerSvc MailerService) AuthService {
	return &authService{
		cfg:         cfg,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		inviteRepo:  inviteRepo,
		tokenRepo:   tokenRepo,
		mailerSvc:   mailerSvc,
	}
}

func (s *authService) SetNotificationService(svc NotificationService) {
	s.notifSvc = svc
}

func (s *authService) GetConfig() *config.Config {
	return s.cfg
}

// ── Auth ──

func (s *authService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	if req.InviteCode == "" {
		return nil, errors.New("invite code is required")
	}

	// 1. Verify and reserve the invite code (checking if it exists and is unused)
	var invite *model.InviteCode
	var isMasterCode bool

	if s.cfg.MasterInviteCode != "" && req.InviteCode == s.cfg.MasterInviteCode {
		isMasterCode = true
	} else {
		var err error
		invite, err = s.inviteRepo.GetByCode(ctx, req.InviteCode)
		if err != nil {
			return nil, errors.New("invalid or used invite code")
		}
		if invite.UsedBy != nil {
			return nil, errors.New("invite code already used")
		}
	}

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

	// 2. Mark invite code as used
	if !isMasterCode && invite != nil {
		if err := s.inviteRepo.MarkAsUsed(ctx, invite.ID, user.ID); err != nil {
			// Log error but don't fail registration
			log.Printf("Failed to mark invite code %d as used by user %d: %v", invite.ID, user.ID, err)
		}

		// 3. Create bidirectional friendship
		if err := s.userRepo.CreateFriendship(ctx, invite.CreatedBy, user.ID, "accepted"); err != nil {
			log.Printf("Failed to create friendship between %d and %d: %v", invite.CreatedBy, user.ID, err)
		}
	}

	return user, nil
}

// ── Invite Codes ──

func (s *authService) GenerateInvite(ctx context.Context, userID int64) (*model.InviteCode, error) {
	// Generate random 8-character code
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	code := hex.EncodeToString(b)
	
	return s.inviteRepo.Create(ctx, code, userID)
}

func (s *authService) ListUserInvites(ctx context.Context, userID int64) ([]*model.InviteCode, error) {
	return s.inviteRepo.GetByCreatedBy(ctx, userID)
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

func (s *authService) generateVerificationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *authService) ForgotPassword(ctx context.Context, identifier string) error {
	user, err := s.userRepo.GetByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// Намеренно не возвращаем ошибку, чтобы не сливать email
			return nil
		}
		return err
	}

	// Удаляем старые токены
	_ = s.tokenRepo.DeleteByUserIDAndPurpose(ctx, user.ID, model.PurposePasswordReset)

	tokenStr, err := s.generateVerificationToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(1 * time.Hour).Unix()
	_, err = s.tokenRepo.Create(ctx, user.ID, tokenStr, model.PurposePasswordReset, expiresAt)
	if err != nil {
		return err
	}

	// Отправляем email асинхронно
	go func() {
		if err := s.mailerSvc.SendPasswordResetEmail(user.Email, tokenStr); err != nil {
			log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
		}
	}()

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	token, err := s.tokenRepo.GetByToken(ctx, tokenStr, model.PurposePasswordReset)
	if err != nil {
		return err // ErrTokenNotFound
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, token.UserID, string(newHash)); err != nil {
		return err
	}

	_ = s.tokenRepo.DeleteByToken(ctx, tokenStr)

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

func (s *authService) GetFriends(ctx context.Context, userID int64) ([]*model.User, error) {
	return s.userRepo.GetFriends(ctx, userID)
}

func (s *authService) GetRequests(ctx context.Context, userID int64) ([]*model.User, error) {
	return s.userRepo.GetRequests(ctx, userID)
}

func (s *authService) SendRequest(ctx context.Context, userID int64, identifier string) error {
	var targetUser *model.User
	var err error

	var idVal int64
	if _, scanErr := fmt.Sscanf(identifier, "%d", &idVal); scanErr == nil {
		targetUser, err = s.userRepo.GetByID(ctx, idVal)
	} else {
		targetUser, err = s.userRepo.GetByIdentifier(ctx, identifier)
	}

	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if targetUser.ID == userID {
		return ErrCannotFriendSelf
	}

	existing, err := s.userRepo.GetFriendship(ctx, userID, targetUser.ID)
	if err != nil {
		return err
	}

	if existing != nil {
		switch existing.Status {
		case "accepted":
			return ErrFriendshipAlreadyExists
		case "pending":
			if existing.UserID == userID {
				return ErrFriendshipAlreadyExists
			}
			return s.userRepo.UpdateFriendshipStatus(ctx, userID, targetUser.ID, "accepted")
		case "declined":
			return s.userRepo.UpdateFriendshipStatus(ctx, userID, targetUser.ID, "pending")
		}
	}

	err = s.userRepo.CreateFriendship(ctx, userID, targetUser.ID, "pending")
	if err == nil && s.notifSvc != nil {
		sender, _ := s.userRepo.GetByID(ctx, userID)
		senderName := "Пользователь"
		if sender != nil {
			senderName = sender.Username
		}
		
		s.notifSvc.SendNotification(ctx, targetUser.ID, "friend_request", senderName+" отправил вам заявку в друзья", map[string]interface{}{
			"sender_id":   userID,
			"sender_name": senderName,
		}, false)
	}
	return err
}

func (s *authService) AcceptRequest(ctx context.Context, userID, targetID int64) error {
	existing, err := s.userRepo.GetFriendship(ctx, userID, targetID)
	if err != nil {
		return err
	}
	if existing == nil || existing.Status != "pending" || existing.FriendID != userID {
		return errors.New("no pending request to accept")
	}
	return s.userRepo.UpdateFriendshipStatus(ctx, userID, targetID, "accepted")
}

func (s *authService) DeclineRequest(ctx context.Context, userID, targetID int64) error {
	existing, err := s.userRepo.GetFriendship(ctx, userID, targetID)
	if err != nil {
		return err
	}
	if existing == nil || existing.Status != "pending" || existing.FriendID != userID {
		return errors.New("no pending request to decline")
	}
	return s.userRepo.UpdateFriendshipStatus(ctx, userID, targetID, "declined")
}

func (s *authService) DeleteFriend(ctx context.Context, userID, targetID int64) error {
	return s.userRepo.DeleteFriendship(ctx, userID, targetID)
}

func (s *authService) SearchUsers(ctx context.Context, q string, excludeID int64) ([]*model.User, error) {
	return s.userRepo.SearchUsers(ctx, q, excludeID)
}

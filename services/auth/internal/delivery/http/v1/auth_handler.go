package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"app4every/services/auth/internal/model"
	"app4every/services/auth/internal/service"
	delivery "app4every/services/auth/internal/delivery/http"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// ── Вспомогательные функции ──

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": errCode, "message": message})
}

// ── Auth ──

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if req.Username == "" || req.Password == "" || req.Email == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "username, email and password are required")
		return
	}

	user, err := h.authService.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			writeError(w, http.StatusConflict, "conflict", "email or username already taken")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "registration failed")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if req.Identifier == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "identifier and password are required")
		return
	}

	user, accessToken, refreshToken, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "invalid email or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "")
		return
	}

	// Refresh Token — в HttpOnly-куку (недоступна JS, защита от XSS)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // true в продакшне с HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, model.AuthResponse{User: *user, AccessToken: accessToken})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing refresh token")
		return
	}

	accessToken, refreshToken, err := h.authService.Refresh(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired refresh token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{"access_token": accessToken})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	if cookie, err := r.Cookie("refresh_token"); err == nil {
		_ = h.authService.Logout(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

// ── Профиль (защищённые маршруты) ──

// UpdateProfile обновляет username и email текущего пользователя.
// PUT /api/v1/users/profile
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	userID := r.Context().Value(delivery.UserIDKey).(int64)

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if req.Email == "" || req.Username == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "username and email are required")
		return
	}

	user, err := h.authService.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			writeError(w, http.StatusConflict, "conflict", "email or username already taken")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update profile")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// ChangePassword меняет пароль текущего пользователя, требуя текущий пароль.
// POST /api/v1/users/password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	userID := r.Context().Value(delivery.UserIDKey).(int64)

	var req model.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "current_password and new_password are required")
		return
	}

	if err := h.authService.ChangePassword(r.Context(), userID, req); err != nil {
		if errors.Is(err, service.ErrWrongPassword) {
			writeError(w, http.StatusUnauthorized, "wrong_password", "current password is incorrect")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to change password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password changed successfully"})
}

// ── Сброс пароля (публичные маршруты) ──

// ForgotPassword — заглушка, всегда отвечает успехом.
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	var req model.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	// Вызываем сервис (он логирует запрос), но клиенту всегда говорим "ок".
	// Это защита от перебора: злоумышленник не узнает, существует ли email.
	_ = h.authService.ForgotPassword(r.Context(), req.Identifier)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "If an account with that email/username exists, we sent a reset link to the linked email.",
	})
}

// ResetPassword — заглушка.
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	var req model.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if req.Token == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "token and new_password are required")
		return
	}

	if err := h.authService.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token", "invalid or expired reset token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Password reset successfully."})
}

func (h *AuthHandler) HandleFriends(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(delivery.UserIDKey).(int64)

	// DELETE /api/v1/users/friends/{id}
	if r.Method == http.MethodDelete {
		p := strings.TrimPrefix(r.URL.Path, "/api/v1/users/friends/")
		targetID, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid user ID")
			return
		}
		if err := h.authService.DeleteFriend(r.Context(), userID, targetID); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	// GET /api/v1/users/friends/requests
	if strings.HasSuffix(r.URL.Path, "/requests") {
		reqs, err := h.authService.GetRequests(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, reqs)
		return
	}

	// GET /api/v1/users/friends
	friends, err := h.authService.GetFriends(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, friends)
}

func (h *AuthHandler) FriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}
	userID := r.Context().Value(delivery.UserIDKey).(int64)

	var req struct {
		Identifier string `json:"identifier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if req.Identifier == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "identifier is required")
		return
	}

	if err := h.authService.SendRequest(r.Context(), userID, req.Identifier); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		if errors.Is(err, service.ErrCannotFriendSelf) {
			writeError(w, http.StatusBadRequest, "bad_request", "cannot friend yourself")
			return
		}
		if errors.Is(err, service.ErrFriendshipAlreadyExists) {
			writeError(w, http.StatusConflict, "conflict", "friendship already exists or pending")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "request processed successfully"})
}

func (h *AuthHandler) AcceptFriend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}
	userID := r.Context().Value(delivery.UserIDKey).(int64)

	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	if err := h.authService.AcceptRequest(r.Context(), userID, req.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "request accepted"})
}

func (h *AuthHandler) DeclineFriend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}
	userID := r.Context().Value(delivery.UserIDKey).(int64)

	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	if err := h.authService.DeclineRequest(r.Context(), userID, req.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "request declined"})
}

func (h *AuthHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}
	userID := r.Context().Value(delivery.UserIDKey).(int64)
	q := r.URL.Query().Get("q")

	users, err := h.authService.SearchUsers(r.Context(), q, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// ── Shikimori OAuth ──

// GET /api/v1/auth/shikimori/login
func (h *AuthHandler) ShikimoriLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	cfg := h.authService.GetConfig()
	
	encodedRedirectURI := url.QueryEscape(cfg.ShikimoriRedirectURI)
	authUrl := fmt.Sprintf("https://shikimori.io/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=user_rates",
		cfg.ShikimoriClientID, encodedRedirectURI)
	
	// Передаем токен через параметр state, чтобы Shikimori вернул его нам в callback
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if tokenString != "" {
		authUrl += "&state=" + tokenString
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"url": authUrl})
}

// GET /api/v1/auth/shikimori/callback
func (h *AuthHandler) ShikimoriCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "missing code")
		return
	}

	userID := r.Context().Value(delivery.UserIDKey).(int64)

	if err := h.authService.ShikimoriCallback(r.Context(), userID, code); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Редирект обратно на клиент в профиль
	http.Redirect(w, r, "/profile?shikimori_linked=true", http.StatusFound)
}

// GET /api/v1/auth/shikimori/rates
func (h *AuthHandler) GetShikimoriRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	userID := r.Context().Value(delivery.UserIDKey).(int64)
	data, err := h.authService.GetShikimoriRates(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// POST /api/v1/auth/shikimori/rates
func (h *AuthHandler) SyncShikimoriRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	userID := r.Context().Value(delivery.UserIDKey).(int64)
	
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "failed to read body")
		return
	}
	defer r.Body.Close()

	data, err := h.authService.SyncShikimoriRate(r.Context(), userID, bodyBytes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ── Invite Codes ──

// POST /api/v1/auth/invites
func (h *AuthHandler) GenerateInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	userID := r.Context().Value(delivery.UserIDKey).(int64)

	invite, err := h.authService.GenerateInvite(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, invite)
}

// GET /api/v1/auth/invites
func (h *AuthHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	userID := r.Context().Value(delivery.UserIDKey).(int64)

	invites, err := h.authService.ListUserInvites(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, invites)
}

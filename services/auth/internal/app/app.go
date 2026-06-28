package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"app4every/services/auth/internal/config"
	"app4every/services/auth/internal/database"
	delivery "app4every/services/auth/internal/delivery/http"
	v1 "app4every/services/auth/internal/delivery/http/v1"
	"app4every/services/auth/internal/hub"
	"app4every/services/auth/internal/repository"
	"app4every/services/auth/internal/service"
)

func Run() error {
	cfg := config.LoadConfig()

	// 1. Инициализация БД
	dbPool, err := database.NewPostgresPool(cfg)
	if err != nil {
		return fmt.Errorf("failed to init pg: %w", err)
	}
	defer dbPool.Close()

	// 2. Инициализация Redis
	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to init redis: %w", err)
	}
	defer redisClient.Close()

	// 3. Репозитории
	userRepo := repository.NewUserRepository(dbPool)
	sessionRepo := repository.NewSessionRepository(redisClient)
	notificationRepo := repository.NewNotificationRepo(dbPool)
	inviteRepo := repository.NewInviteRepository(dbPool)
	tokenRepo := repository.NewVerificationTokenRepository(dbPool)

	// 4. Сервисы
	mailerSvc := service.NewMailerService(cfg)
	authService := service.NewAuthService(cfg, userRepo, sessionRepo, inviteRepo, tokenRepo, mailerSvc)
	
	notificationHub := hub.NewNotificationHub()
	notificationService := service.NewNotificationService(notificationRepo, notificationHub)
	authService.SetNotificationService(notificationService) // We will add this method to send friend requests

	// 5. Хэндлеры
	authHandler := v1.NewAuthHandler(authService)
	notificationHandler := v1.NewNotificationHandler(notificationService, notificationHub)

	// 6. Роутер
	mux := http.NewServeMux()

	// ── Публичные маршруты ──
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","message":"App4Every backend is running"}`))
	})
	mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("/api/v1/auth/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("/api/v1/auth/reset-password", authHandler.ResetPassword)

	// ── Внутренние маршруты (только для docker network) ──
	mux.HandleFunc("/internal/users/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/internal/users/"), "/")
		if len(parts) == 2 && parts[1] == "friends" {
			userID, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				http.Error(w, "invalid user id", http.StatusBadRequest)
				return
			}
			friends, err := authService.GetFriends(r.Context(), userID)
			if err != nil {
				http.Error(w, "internal_error", http.StatusInternalServerError)
				return
			}
			var friendIDs []int64
			for _, f := range friends {
				friendIDs = append(friendIDs, f.ID)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(friendIDs)
			return
		}
		if len(parts) == 2 && parts[1] == "shikimori" {
			userID, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				http.Error(w, "invalid user id", http.StatusBadRequest)
				return
			}
			user, err := userRepo.GetByID(r.Context(), userID)
			if err != nil {
				http.Error(w, "user_not_found", http.StatusNotFound)
				return
			}
			if user.ShikimoriAccessToken == nil || user.ShikimoriUserID == nil {
				http.Error(w, "no_shikimori_tokens", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":      *user.ShikimoriAccessToken,
				"refresh_token":     user.ShikimoriRefreshToken,
				"shikimori_user_id": *user.ShikimoriUserID,
			})
			return
		}
		http.NotFound(w, r)
	})
	
	mux.HandleFunc("/internal/notifications", notificationHandler.InternalSendNotification)

	// ── Защищённые маршруты ──
	// Все маршруты в protectedMux проходят через AuthMiddleware.
	authMiddleware := delivery.AuthMiddleware(cfg)
	protectedMux := http.NewServeMux()

	// GET /api/v1/auth/me — данные текущего пользователя
	protectedMux.HandleFunc("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(delivery.UserIDKey).(int64)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		user, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			http.Error(w, `{"error":"user_not_found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	// PUT /api/v1/users/profile — обновление username и email
	protectedMux.HandleFunc("/api/v1/users/profile", authHandler.UpdateProfile)

	// POST /api/v1/users/password — смена пароля
	protectedMux.HandleFunc("/api/v1/users/password", authHandler.ChangePassword)

	// Shikimori OAuth routes
	protectedMux.HandleFunc("/api/v1/auth/shikimori/login", authHandler.ShikimoriLogin)
	protectedMux.HandleFunc("/api/v1/auth/shikimori/callback", authHandler.ShikimoriCallback)

	// Invites
	protectedMux.HandleFunc("/api/v1/auth/invites", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.GenerateInvite(w, r)
		} else if r.Method == http.MethodGet {
			authHandler.ListInvites(w, r)
		} else {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	// Друзья и поиск
	protectedMux.HandleFunc("/api/v1/users/friends", authHandler.HandleFriends)
	protectedMux.HandleFunc("/api/v1/users/friends/", authHandler.HandleFriends) // trailing slash matches DELETE and /requests
	protectedMux.HandleFunc("/api/v1/users/friends/request", authHandler.FriendRequest)
	protectedMux.HandleFunc("/api/v1/users/friends/accept", authHandler.AcceptFriend)
	protectedMux.HandleFunc("/api/v1/users/friends/decline", authHandler.DeclineFriend)
	protectedMux.HandleFunc("/api/v1/users/search", authHandler.SearchUsers)

	// Уведомления
	protectedMux.HandleFunc("/api/v1/auth/notifications", notificationHandler.GetNotifications)
	protectedMux.HandleFunc("/api/v1/auth/notifications/read", notificationHandler.MarkAsRead)
	protectedMux.HandleFunc("/api/v1/auth/notifications/", notificationHandler.DeleteNotification)
	protectedMux.HandleFunc("/api/v1/auth/ws/notifications", notificationHandler.ServeWS)

	// Регистрируем защищённые префиксы в основном мультиплексоре
	mux.Handle("/api/v1/auth/me", authMiddleware(protectedMux))
	mux.Handle("/api/v1/auth/notifications", authMiddleware(protectedMux))
	mux.Handle("/api/v1/auth/notifications/", authMiddleware(protectedMux))
	mux.Handle("/api/v1/auth/ws/notifications", authMiddleware(protectedMux))
	mux.Handle("/api/v1/auth/shikimori/", authMiddleware(protectedMux))
	mux.Handle("/api/v1/auth/invites", authMiddleware(protectedMux))
	mux.Handle("/api/v1/users/", authMiddleware(protectedMux)) // /users/ — суффикс "/" = префикс-матчинг

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	fmt.Printf("Server is starting on port %s...\n", cfg.Port)
	return server.ListenAndServe()
}

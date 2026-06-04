package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"app4every/services/auth/internal/config"
	"app4every/services/auth/internal/database"
	delivery "app4every/services/auth/internal/delivery/http"
	v1 "app4every/services/auth/internal/delivery/http/v1"
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

	// 4. Сервисы
	authService := service.NewAuthService(cfg, userRepo, sessionRepo)

	// 5. Хэндлеры
	authHandler := v1.NewAuthHandler(authService)

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

	// Регистрируем защищённые префиксы в основном мультиплексоре
	mux.Handle("/api/v1/auth/me", authMiddleware(protectedMux))
	mux.Handle("/api/v1/users/", authMiddleware(protectedMux)) // /users/ — суффикс "/" = префикс-матчинг

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	fmt.Printf("Server is starting on port %s...\n", cfg.Port)
	return server.ListenAndServe()
}

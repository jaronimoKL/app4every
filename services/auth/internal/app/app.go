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

	// Healthcheck
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","message":"Super-App backend is running"}`))
	})

	// Auth роуты
	mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)

	// Пример защищенного роута
	protectedMux := http.NewServeMux()
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

	// Применяем AuthMiddleware на защищенные роуты
	authMiddleware := delivery.AuthMiddleware(cfg)
	mux.Handle("/api/v1/auth/me", authMiddleware(protectedMux))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	fmt.Printf("Server is starting on port %s...\n", cfg.Port)
	return server.ListenAndServe()
}

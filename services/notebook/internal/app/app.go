package app

import (
	"fmt"
	"net/http"

	"app4every/services/notebook/internal/config"
	"app4every/services/notebook/internal/database"
	delivery "app4every/services/notebook/internal/delivery/http"
	v1 "app4every/services/notebook/internal/delivery/http/v1"
	"app4every/services/notebook/internal/repository"
	"app4every/services/notebook/internal/service"
)

func Run() error {
	cfg := config.LoadConfig()

	// 1. БД
	dbPool, err := database.NewPostgresPool(cfg)
	if err != nil {
		return fmt.Errorf("failed to init pg: %w", err)
	}
	defer dbPool.Close()

	// 2. Слои
	noteRepo := repository.NewNoteRepository(dbPool)
	noteSvc := service.NewNoteService(noteRepo)
	noteHandler := v1.NewNoteHandler(noteSvc)

	// 3. Middleware
	authMiddleware := delivery.AuthMiddleware(cfg)

	// 4. Роутер
	// Все маршруты /api/v1/notes/* защищены JWT.
	// Go 1.21 ServeMux:
	//   - "/api/v1/notes"  → точное совпадение → list + create
	//   - "/api/v1/notes/" → суффикс "/" = суб-дерево → get/update/delete по ID
	mux := http.NewServeMux()

	mux.Handle("/api/v1/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"notebook"}`))
	}))

	// Оборачиваем в middleware вручную (нет фреймворка)
	mux.Handle("/api/v1/notes", authMiddleware(http.HandlerFunc(noteHandler.HandleNotes)))
	mux.Handle("/api/v1/notes/", authMiddleware(http.HandlerFunc(noteHandler.HandleNoteByID)))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	fmt.Printf("[notebook] Server starting on port %s...\n", cfg.Port)
	return server.ListenAndServe()
}

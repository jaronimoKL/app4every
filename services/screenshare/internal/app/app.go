package app

import (
	"fmt"
	"net/http"

	"app4every/services/screenshare/internal/config"
	"app4every/services/screenshare/internal/database"
	delivery "app4every/services/screenshare/internal/delivery/http"
	"app4every/services/screenshare/internal/hub"
)

func Run() error {
	cfg := config.LoadConfig()

	dbPool, err := database.NewPostgresPool(cfg)
	if err != nil {
		return fmt.Errorf("failed to init postgres: %w", err)
	}
	defer dbPool.Close()

	h := hub.NewHub()
	handler := delivery.NewScreenshareHandler(h)
	authMiddleware := delivery.AuthMiddleware(cfg, dbPool)

	mux := http.NewServeMux()

	mux.Handle("/api/v1/screenshare/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"screenshare"}`))
	}))

	mux.Handle("/api/v1/screenshare/ws", authMiddleware(http.HandlerFunc(handler.HandleWS)))
	mux.Handle("/api/v1/screenshare/rooms/", authMiddleware(http.HandlerFunc(handler.HandleRoomInfo)))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	fmt.Printf("[screenshare] Server starting on port %s...\n", cfg.Port)
	return server.ListenAndServe()
}

package app

import (
	"fmt"
	"net/http"

	"app4every/services/watchparty/internal/config"
	delivery "app4every/services/watchparty/internal/delivery/http"
	"app4every/services/watchparty/internal/hub"
)

func Run() error {
	cfg := config.LoadConfig()

	// Initialize the WebSocket hub
	h := hub.NewHub()

	// Handlers
	handler := delivery.NewHandler(h, cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/watchparty/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"watchparty-service"}`))
	})

	mux.HandleFunc("/api/v1/watchparty/ws", handler.ServeWS)
	mux.HandleFunc("/api/v1/watchparty/rooms/", handler.GetRoomState)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	fmt.Printf("WatchParty Service is starting on port %s...\n", cfg.Port)
	return server.ListenAndServe()
}

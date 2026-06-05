package app

import (
	"fmt"
	"net/http"

	"app4every/services/reviews/internal/config"
	"app4every/services/reviews/internal/database"
	delivery "app4every/services/reviews/internal/delivery/http"
	v1 "app4every/services/reviews/internal/delivery/http/v1"
	"app4every/services/reviews/internal/repository"
	"app4every/services/reviews/internal/service"
)

func Run() error {
	cfg := config.LoadConfig()

	dbPool, err := database.NewPostgresPool(cfg)
	if err != nil {
		return fmt.Errorf("failed to init postgres: %w", err)
	}
	defer dbPool.Close()

	repo    := repository.NewReviewRepository(dbPool)
	svc     := service.NewReviewService(repo)
	handler := v1.NewReviewHandler(svc)

	groupRepo := repository.NewGroupRepository(dbPool)
	groupSvc  := service.NewGroupService(groupRepo, repo)
	hub       := v1.NewHub()
	groupHandler := v1.NewGroupHandler(groupSvc, hub)

	auth    := delivery.AuthMiddleware(cfg)

	mux := http.NewServeMux()

	mux.Handle("/api/v1/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"reviews"}`))
	}))

	// /api/v1/reviews      → list + create
	// /api/v1/reviews/     → by ID, links management
	mux.Handle("/api/v1/reviews", auth(http.HandlerFunc(handler.HandleReviews)))
	mux.Handle("/api/v1/reviews/", auth(http.HandlerFunc(handler.HandleReviewsByID)))

	// /api/v1/groups       → list + create groups
	// /api/v1/groups/      → group details, invites, items, websocket
	mux.Handle("/api/v1/groups", auth(http.HandlerFunc(groupHandler.HandleGroups)))
	mux.Handle("/api/v1/groups/", auth(http.HandlerFunc(groupHandler.HandleGroupsByID)))

	server := &http.Server{Addr: ":" + cfg.Port, Handler: mux}
	fmt.Printf("[reviews] Server starting on port %s...\n", cfg.Port)
	return server.ListenAndServe()
}

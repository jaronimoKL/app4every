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

	"github.com/redis/go-redis/v9"
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

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
	})

	integrationHandler := v1.NewIntegrationHandler(redisClient)

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

	// Integrations
	mux.Handle("/api/v1/reviews/integrations/shikimori/search", auth(http.HandlerFunc(integrationHandler.ShikimoriSearch)))
	mux.Handle("/api/v1/reviews/integrations/aniliberty/search", auth(http.HandlerFunc(integrationHandler.AnilibertySearch)))
	mux.Handle("/api/v1/reviews/integrations/aniliberty/episodes/", auth(http.HandlerFunc(integrationHandler.AnilibertyEpisodes)))
	mux.Handle("/api/v1/reviews/integrations/aniliberty/proxy", http.HandlerFunc(integrationHandler.AnilibertyProxy))
	mux.Handle("/api/v1/reviews/integrations/kodik/search", auth(http.HandlerFunc(integrationHandler.KodikSearch)))


	server := &http.Server{Addr: ":" + cfg.Port, Handler: mux}
	fmt.Printf("[reviews] Server starting on port %s...\n", cfg.Port)
	return server.ListenAndServe()
}

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"app4every/services/reviews/internal/model"
)

type ShikimoriTokens struct {
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	ShikimoriUserID int64  `json:"shikimori_user_id"`
}

func (s *groupService) syncShikimori(ctx context.Context, userID int64, item *model.GroupItem) {
	if item.ShikimoriID == nil || *item.ShikimoriID == 0 {
		return
	}

	// 1. Fetch tokens from auth-service
	authURL := fmt.Sprintf("http://auth-service:8080/internal/users/%d/shikimori", userID)
	req, _ := http.NewRequestWithContext(ctx, "GET", authURL, nil)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return // Ignore if user has no shikimori linked
	}

	var tokens ShikimoriTokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		resp.Body.Close()
		return
	}
	resp.Body.Close()

	// 2. Map status to Shikimori status
	var targetStatus string
	switch item.Status {
	case model.StatusPlanned:
		targetStatus = "planned"
	case model.StatusWatching:
		targetStatus = "watching"
	case model.StatusCompleted:
		targetStatus = "completed"
	case model.StatusDropped:
		targetStatus = "dropped"
	default:
		return
	}

	// 3. Call Shikimori API to sync user_rate
	// It's tricky to know if user already has this anime in list (so we need PUT) or not (POST)
	// Shikimori accepts POST /api/v2/user_rates with { "user_rate": { "user_id": ..., "target_id": ..., "target_type": "Anime", "status": ... } }
	
	payload := map[string]interface{}{
		"user_rate": map[string]interface{}{
			"user_id":     tokens.ShikimoriUserID,
			"target_id":   *item.ShikimoriID,
			"target_type": "Anime",
			"status":      targetStatus,
		},
	}
	
	if item.Status == model.StatusCompleted && item.ShikimoriScore != nil && *item.ShikimoriScore > 0 {
		payload["user_rate"].(map[string]interface{})["score"] = *item.ShikimoriScore
	}
	if item.EpisodesTotal != nil && item.Status == model.StatusCompleted {
		payload["user_rate"].(map[string]interface{})["episodes"] = *item.EpisodesTotal
	} else if item.CurrentEpisode > 0 {
		payload["user_rate"].(map[string]interface{})["episodes"] = item.CurrentEpisode
	}

	bodyBytes, _ := json.Marshal(payload)
	
	sReq, _ := http.NewRequestWithContext(ctx, "POST", "https://shikimori.io/api/v2/user_rates", bytes.NewBuffer(bodyBytes))
	sReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	sReq.Header.Set("Content-Type", "application/json")
	sReq.Header.Set("User-Agent", "App4Every")

	sResp, err := client.Do(sReq)
	if err != nil {
		log.Printf("syncShikimori request failed: %v", err)
		return
	}
	defer sResp.Body.Close()

	if sResp.StatusCode != http.StatusOK && sResp.StatusCode != http.StatusCreated {
		log.Printf("syncShikimori failed with status: %d", sResp.StatusCode)
	}
}

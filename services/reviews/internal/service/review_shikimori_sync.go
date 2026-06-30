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

type ShikimoriGraphQLResponse struct {
	Data struct {
		UserRates []struct {
			ID       string `json:"id"`
			Status   string `json:"status"`
			Score    int    `json:"score"`
			Episodes int    `json:"episodes"`
			Anime    *struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Russian string `json:"russian"`
				Poster  *struct {
					OriginalUrl string `json:"originalUrl"`
				} `json:"poster"`
				Score    float64 `json:"score"`
				Episodes int     `json:"episodes"`
				Genres   []struct {
					Name    string `json:"name"`
					Russian string `json:"russian"`
				} `json:"genres"`
			} `json:"anime"`
		} `json:"userRates"`
	} `json:"data"`
}

func (s *reviewService) SyncShikimori(ctx context.Context, userID int64) error {
	// 1. Get tokens from auth service
	authURL := fmt.Sprintf("http://auth-service:8080/internal/users/%d/shikimori", userID)
	req, _ := http.NewRequestWithContext(ctx, "GET", authURL, nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch auth tokens: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("shikimori not linked or auth error: status %d", resp.StatusCode)
	}

	var tokens ShikimoriTokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return fmt.Errorf("failed to decode tokens: %w", err)
	}

	if tokens.ShikimoriUserID == 0 {
		return fmt.Errorf("shikimori user id is 0")
	}

	// 2. Query Shikimori GraphQL API
	query := `query {
		userRates(userId: %d, targetType: Anime, limit: 500) {
			id
			status
			score
			episodes
			anime {
				id
				name
				russian
				poster { originalUrl }
				score
				episodes
				genres { name russian }
			}
		}
	}`
	q := fmt.Sprintf(query, tokens.ShikimoriUserID)
	
	payload := map[string]string{"query": q}
	bodyBytes, _ := json.Marshal(payload)

	sReq, _ := http.NewRequestWithContext(ctx, "POST", "https://shikimori.io/api/graphql", bytes.NewBuffer(bodyBytes))
	sReq.Header.Set("Content-Type", "application/json")
	sReq.Header.Set("User-Agent", "App4Every")

	sResp, err := client.Do(sReq)
	if err != nil {
		return fmt.Errorf("shikimori graphql error: %w", err)
	}
	defer sResp.Body.Close()

	var gqlResp ShikimoriGraphQLResponse
	if err := json.NewDecoder(sResp.Body).Decode(&gqlResp); err != nil {
		return fmt.Errorf("failed to decode graphql: %w", err)
	}

	// 3. Fetch user's existing reviews
	existingReviews, err := s.repo.GetAllByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing reviews: %w", err)
	}
	existingMap := make(map[int]*model.Review)
	for _, r := range existingReviews {
		if r.ShikimoriID != nil {
			existingMap[*r.ShikimoriID] = r
		}
	}

	// 4. Sync each item
	for _, rate := range gqlResp.Data.UserRates {
		if rate.Anime == nil {
			continue
		}

		shikiID := 0
		fmt.Sscanf(rate.Anime.ID, "%d", &shikiID)
		if shikiID == 0 {
			continue
		}

		status := model.StatusPlanned
		switch rate.Status {
		case "watching":
			status = model.StatusWatching
		case "completed":
			status = model.StatusCompleted
		case "dropped":
			status = model.StatusDropped
		case "on_hold":
			status = model.StatusOnHold
		}

		title := rate.Anime.Russian
		if title == "" {
			title = rate.Anime.Name
		}
		
		poster := ""
		if rate.Anime.Poster != nil {
			poster = rate.Anime.Poster.OriginalUrl
		}

		var rating *int16
		if rate.Score > 0 {
			r := int16(rate.Score)
			rating = &r
		}

		var epsTotal *int
		if rate.Anime.Episodes > 0 {
			epsTotal = &rate.Anime.Episodes
		}
		
		score := rate.Anime.Score

		existing, found := existingMap[shikiID]
		var revID int64
		if !found {
			// Create
			req := model.CreateReviewRequest{
				Title:          title,
				ContentType:    model.TypeAnime,
				Status:         status,
				Rating:         rating,
				PosterURL:      poster,
				ShikimoriID:    &shikiID,
				EpisodesTotal:  epsTotal,
				ShikimoriScore: &score,
			}
			rev, err := s.repo.Create(ctx, userID, req)
			if err != nil {
				log.Printf("failed to sync create review for %d: %v", shikiID, err)
				continue
			}
			revID = rev.ID
		} else {
			// Update existing to sync status and score
			revID = existing.ID
			req := model.UpdateReviewRequest{
				Title:          existing.Title,
				ContentType:    existing.ContentType,
				Status:         status,
				Rating:         rating,
				PosterURL:      existing.PosterURL,
				ShikimoriID:    existing.ShikimoriID,
				EpisodesTotal:  existing.EpisodesTotal,
				AnilibertyAlias: existing.AnilibertyAlias,
				ShikimoriScore: &score,
				Description:    existing.Description,
				Notes:          existing.Notes,
			}
			_, err := s.repo.Update(ctx, revID, userID, req)
			if err != nil {
				log.Printf("failed to sync update review for %d: %v", shikiID, err)
			}
		}

		// 5. Sync genres
		// Get existing genres for this review
		extGenres, _ := s.repo.GetGenres(ctx, revID)
		gMap := make(map[string]bool)
		for _, g := range extGenres {
			gMap[g.Name] = true
		}

		for _, g := range rate.Anime.Genres {
			name := g.Russian
			if name == "" {
				name = g.Name
			}
			if !gMap[name] {
				s.repo.AddGenre(ctx, revID, userID, name)
				gMap[name] = true
			}
		}
	}

	return nil
}

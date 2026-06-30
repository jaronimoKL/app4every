package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (s *authService) ShikimoriCallback(ctx context.Context, userID int64, code string) error {
	// 1. Обменять code на access_token и refresh_token
	tokenURL := "https://shikimori.io/oauth/token"
	payload := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     s.cfg.ShikimoriClientID,
		"client_secret": s.cfg.ShikimoriClientSecret,
		"code":          code,
		"redirect_uri":  s.cfg.ShikimoriRedirectURI,
	}
	bodyBytes, _ := json.Marshal(payload)
	
	req, _ := http.NewRequest("POST", tokenURL, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "App4Every")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request to Shikimori token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shikimori returned non-200 status %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode shikimori token response: %w", err)
	}

	// 2. Получить ID пользователя в Shikimori (/api/users/whoami)
	whoamiURL := "https://shikimori.io/api/users/whoami"
	wReq, _ := http.NewRequest("GET", whoamiURL, nil)
	wReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	wReq.Header.Set("User-Agent", "App4Every")

	wResp, err := client.Do(wReq)
	if err != nil {
		return fmt.Errorf("failed to get shikimori user info: %w", err)
	}
	defer wResp.Body.Close()

	if wResp.StatusCode != http.StatusOK {
		return fmt.Errorf("shikimori whoami returned non-200 status %d", wResp.StatusCode)
	}

	var whoamiResp struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(wResp.Body).Decode(&whoamiResp); err != nil {
		return fmt.Errorf("failed to decode shikimori whoami response: %w", err)
	}

	// 3. Сохранить токены и shikimori_user_id в БД
	if err := s.userRepo.UpdateShikimoriTokens(ctx, userID, tokenResp.AccessToken, tokenResp.RefreshToken, whoamiResp.ID); err != nil {
		return fmt.Errorf("failed to save shikimori tokens to db: %w", err)
	}

	return nil
}

func (s *authService) GetShikimoriRates(ctx context.Context, userID int64) ([]byte, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.ShikimoriAccessToken == nil || user.ShikimoriUserID == nil {
		return nil, fmt.Errorf("shikimori account not linked")
	}

	url := fmt.Sprintf("https://shikimori.io/api/users/%d/anime_rates?limit=5000", *user.ShikimoriUserID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+*user.ShikimoriAccessToken)
	req.Header.Set("User-Agent", "App4Every")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Token expired, need to refresh (simplified for now, ideally we should call a refresh logic)
		return nil, fmt.Errorf("shikimori token expired")
	}

	return io.ReadAll(resp.Body)
}

func (s *authService) SyncShikimoriRate(ctx context.Context, userID int64, payload []byte) ([]byte, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.ShikimoriAccessToken == nil {
		return nil, fmt.Errorf("shikimori account not linked")
	}

	url := "https://shikimori.io/api/v2/user_rates"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+*user.ShikimoriAccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "App4Every")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *authService) GetShikimoriWhoami(ctx context.Context, userID int64) ([]byte, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.ShikimoriAccessToken == nil {
		return nil, fmt.Errorf("shikimori account not linked")
	}

	url := "https://shikimori.io/api/users/whoami"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+*user.ShikimoriAccessToken)
	req.Header.Set("User-Agent", "App4Every")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get whoami: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (s *authService) UnlinkShikimori(ctx context.Context, userID int64) error {
	// Очищаем токены и ID в БД
	return s.userRepo.UpdateShikimoriTokens(ctx, userID, "", "", 0)
}

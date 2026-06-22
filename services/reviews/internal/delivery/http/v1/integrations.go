package v1

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type IntegrationHandler struct {
	redis *redis.Client
}

func NewIntegrationHandler(r *redis.Client) *IntegrationHandler {
	return &IntegrationHandler{redis: r}
}

func (h *IntegrationHandler) ShikimoriSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "10"
	}

	if query == "" {
		http.Error(w, `{"error":"missing query"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	cacheKey := fmt.Sprintf("shikimori:search:%s:%s", query, limit)

	// Try cache
	if cached, err := h.redis.Get(ctx, cacheKey).Result(); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	searchURL := fmt.Sprintf("https://shikimori.one/api/animes?search=%s&limit=%s&censored=false", url.QueryEscape(query), url.QueryEscape(limit))
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		http.Error(w, `{"error":"request creation failed"}`, http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "app4every/1.0 (contact@jaronimo.work.gd)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, `{"error":"request failed"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, `{"error":"read body failed"}`, http.StatusInternalServerError)
		return
	}

	if resp.StatusCode == http.StatusOK {
		h.redis.Set(ctx, cacheKey, string(body), time.Hour)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (h *IntegrationHandler) AnilibertySearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"missing query parameter"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	// Истинный эндпоинт поиска
	searchURL := fmt.Sprintf("https://aniliberty.top/api/v1/app/search/releases?query=%s", url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		http.Error(w, `{"error":"request creation failed"}`, http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, `{"error":"request failed"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *IntegrationHandler) AnilibertyEpisodes(w http.ResponseWriter, r *http.Request) {
	alias := strings.TrimPrefix(r.URL.Path, "/api/v1/reviews/integrations/aniliberty/episodes/")
	if alias == "" {
		http.Error(w, `{"error":"missing alias"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	releaseURL := fmt.Sprintf("https://aniliberty.top/api/v1/anime/releases/%s", url.PathEscape(alias))
	req, err := http.NewRequestWithContext(ctx, "GET", releaseURL, nil)
	if err != nil {
		http.Error(w, `{"error":"request creation failed"}`, http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, `{"error":"request failed"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *IntegrationHandler) AnilibertyProxy(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, `{"error":"missing url"}`, http.StatusBadRequest)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), "GET", targetURL, nil)
	if err != nil {
		http.Error(w, `{"error":"request creation failed"}`, http.StatusInternalServerError)
		return
	}
	// Mimic a generic mobile browser to avoid some hotlink protections
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Referer", "https://anilibria.top/")
	req.Header.Set("Origin", "https://anilibria.top")

	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			for k, v := range via[0].Header {
				r.Header[k] = v
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, `{"error":"request failed"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Forward response headers
	for k, vv := range resp.Header {
		// Skip CORS related headers and Content-Length
		lowerK := strings.ToLower(k)
		if lowerK == "access-control-allow-origin" || lowerK == "content-length" {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(resp.StatusCode)

	if strings.Contains(targetURL, ".m3u8") {
		// Read m3u8 body and rewrite absolute .ts urls
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		
		lines := strings.Split(bodyStr, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "http") {
				lines[i] = fmt.Sprintf("/api/v1/reviews/integrations/aniliberty/proxy?url=%s", url.QueryEscape(trimmed))
			}
		}
		w.Write([]byte(strings.Join(lines, "\n")))
	} else {
		io.Copy(w, resp.Body)
	}
}


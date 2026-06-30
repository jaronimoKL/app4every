package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	delivery "app4every/services/reviews/internal/delivery/http"
	"app4every/services/reviews/internal/model"
	"app4every/services/reviews/internal/service"
)

type ReviewHandler struct {
	svc service.ReviewService
}

func NewReviewHandler(svc service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

// ── Утилиты ──

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": code, "message": msg})
}

func userID(r *http.Request) int64 {
	return r.Context().Value(delivery.UserIDKey).(int64)
}

// parsedPath описывает распознанный URL вида:
// /api/v1/reviews/{id}
// /api/v1/reviews/{id}/links
// /api/v1/reviews/{id}/links/{linkId}
// /api/v1/reviews/{id}/genres
// /api/v1/reviews/{id}/genres/{genreId}
type parsedPath struct {
	reviewID int64
	action   string // "" | "links" | "genres"
	subID    int64  // задан только для links/{linkId} или genres/{genreId}
	hasSub   bool
}

func parsePath(rawPath string) (*parsedPath, error) {
	p := strings.TrimPrefix(rawPath, "/api/v1/reviews/")
	p = strings.TrimSuffix(p, "/")
	parts := strings.Split(p, "/")

	if len(parts) == 0 || parts[0] == "" {
		return nil, fmt.Errorf("missing id")
	}
	reviewID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid review id")
	}
	res := &parsedPath{reviewID: reviewID}
	if len(parts) >= 2 {
		res.action = parts[1] // "links" или "genres"
	}
	if len(parts) >= 3 {
		res.subID, err = strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid sub id")
		}
		res.hasSub = true
	}
	return res, nil
}

// ── Маршруты ──

// HandleReviews — /api/v1/reviews (без ID)
func (h *ReviewHandler) HandleReviews(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	switch r.Method {
	case http.MethodGet:
		reviews, err := h.svc.ListReviews(r.Context(), uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, reviews)

	case http.MethodPost:
		var req model.CreateReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
			return
		}
		if req.Title == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "title is required")
			return
		}
		// Дефолты
		if req.ContentType == "" {
			req.ContentType = model.TypeMovie
		}
		if req.Status == "" {
			req.Status = model.StatusPlanned
		}
		rev, err := h.svc.CreateReview(r.Context(), uid, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, rev)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
	}
}

// HandleReviewsByID — /api/v1/reviews/{id}[/links[/{linkId}]]
func (h *ReviewHandler) HandleReviewsByID(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)

	pp, err := parsePath(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	// ── /api/v1/reviews/{id}/links ──
	if pp.action == "links" {
		switch {
		case r.Method == http.MethodPost && !pp.hasSub:
			var req model.AddLinkRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
				return
			}
			if req.URL == "" {
				writeError(w, http.StatusBadRequest, "bad_request", "url is required")
				return
			}
			link, err := h.svc.AddLink(r.Context(), pp.reviewID, uid, req)
			if err != nil {
				if errors.Is(err, service.ErrReviewNotFound) {
					writeError(w, http.StatusNotFound, "not_found", "review not found")
					return
				}
				writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, link)

		case r.Method == http.MethodDelete && pp.hasSub:
			if err := h.svc.DeleteLink(r.Context(), pp.subID, pp.reviewID, uid); err != nil {
				if errors.Is(err, service.ErrLinkNotFound) {
					writeError(w, http.StatusNotFound, "not_found", "link not found")
					return
				}
				writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		}
		return
	}

	// ── /api/v1/reviews/{id}/genres ──
	if pp.action == "genres" {
		switch {
		case r.Method == http.MethodPost && !pp.hasSub:
			var req struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
				return
			}
			if req.Name == "" {
				writeError(w, http.StatusBadRequest, "bad_request", "name is required")
				return
			}
			genre, err := h.svc.AddGenre(r.Context(), pp.reviewID, uid, req.Name)
			if err != nil {
				if errors.Is(err, service.ErrReviewNotFound) {
					writeError(w, http.StatusNotFound, "not_found", "review not found")
					return
				}
				writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, genre)

		case r.Method == http.MethodDelete && pp.hasSub:
			if err := h.svc.DeleteGenre(r.Context(), pp.subID, pp.reviewID, uid); err != nil {
				if errors.Is(err, service.ErrGenreNotFound) {
					writeError(w, http.StatusNotFound, "not_found", "genre not found")
					return
				}
				writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		}
		return
	}

	// ── /api/v1/reviews/{id} ──
	switch r.Method {
	case http.MethodGet:
		rev, err := h.svc.GetReview(r.Context(), pp.reviewID, uid)
		if err != nil {
			if errors.Is(err, service.ErrReviewNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "review not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, rev)

	case http.MethodPut:
		var req model.UpdateReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
			return
		}
		rev, err := h.svc.UpdateReview(r.Context(), pp.reviewID, uid, req)
		if err != nil {
			if errors.Is(err, service.ErrReviewNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "review not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, rev)

	case http.MethodDelete:
		if err := h.svc.DeleteReview(r.Context(), pp.reviewID, uid); err != nil {
			if errors.Is(err, service.ErrReviewNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "review not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
	}
}

// POST /api/v1/reviews/sync-shikimori
func (h *ReviewHandler) SyncShikimori(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	uid := userID(r)
	if err := h.svc.SyncShikimori(r.Context(), uid); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

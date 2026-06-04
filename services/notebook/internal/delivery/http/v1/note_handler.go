package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	delivery "app4every/services/notebook/internal/delivery/http"
	"app4every/services/notebook/internal/model"
	"app4every/services/notebook/internal/service"
)

type NoteHandler struct {
	svc service.NoteService
}

func NewNoteHandler(svc service.NoteService) *NoteHandler {
	return &NoteHandler{svc: svc}
}

// ── Вспомогательные функции ──

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

func userIDFromCtx(r *http.Request) int64 {
	return r.Context().Value(delivery.UserIDKey).(int64)
}

// extractID получает числовой ID из конца URL-пути.
// Например /api/v1/notes/42 → 42
func extractID(path string) (int64, error) {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) == 0 {
		return 0, errors.New("missing id")
	}
	return strconv.ParseInt(parts[len(parts)-1], 10, 64)
}

// ── Маршруты ──

// HandleNotes обрабатывает /api/v1/notes (без ID)
// GET  → список заметок пользователя
// POST → создать заметку
func (h *NoteHandler) HandleNotes(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)

	switch r.Method {
	case http.MethodGet:
		notes, err := h.svc.ListNotes(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, notes)

	case http.MethodPost:
		var req model.CreateNoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
			return
		}
		note, err := h.svc.CreateNote(r.Context(), userID, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, note)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
	}
}

// HandleNoteByID обрабатывает /api/v1/notes/{id}
// GET    → получить одну заметку
// PUT    → обновить заметку
// DELETE → удалить заметку
func (h *NoteHandler) HandleNoteByID(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)

	id, err := extractID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid note id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		note, err := h.svc.GetNote(r.Context(), id, userID)
		if err != nil {
			if errors.Is(err, service.ErrNoteNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "note not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, note)

	case http.MethodPut:
		var req model.UpdateNoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
			return
		}
		note, err := h.svc.UpdateNote(r.Context(), id, userID, req)
		if err != nil {
			if errors.Is(err, service.ErrNoteNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "note not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, note)

	case http.MethodDelete:
		if err := h.svc.DeleteNote(r.Context(), id, userID); err != nil {
			if errors.Is(err, service.ErrNoteNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "note not found")
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

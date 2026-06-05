package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"app4every/services/reviews/internal/model"
	"app4every/services/reviews/internal/service"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type GroupHandler struct {
	svc service.GroupService
	hub *Hub
}

func NewGroupHandler(svc service.GroupService, hub *Hub) *GroupHandler {
	return &GroupHandler{
		svc: svc,
		hub: hub,
	}
}

type parsedGroupPath struct {
	IsInvites bool
	InviteID  int64
	InviteAct string // "accept" | "decline" | ""
	GroupID   int64
	GroupAct  string // "" | "invite" | "leave" | "ws" | "items"
	ItemID    int64
	ItemAct   string // "" | "rating" | "links"
	LinkID    int64
}

func parseGroupPath(rawPath string) (*parsedGroupPath, error) {
	p := strings.TrimPrefix(rawPath, "/api/v1/groups/")
	p = strings.TrimSuffix(p, "/")
	parts := strings.Split(p, "/")

	if len(parts) == 0 || parts[0] == "" {
		return nil, fmt.Errorf("invalid path")
	}

	res := &parsedGroupPath{}

	// Case 1: Invites
	if parts[0] == "invites" {
		res.IsInvites = true
		if len(parts) == 1 {
			return res, nil
		}
		inviteID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid invite id")
		}
		res.InviteID = inviteID
		if len(parts) >= 3 {
			res.InviteAct = parts[2]
		}
		return res, nil
	}

	// Case 2: Group ID
	groupID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid group id")
	}
	res.GroupID = groupID

	if len(parts) == 1 {
		return res, nil
	}

	res.GroupAct = parts[1]
	if res.GroupAct != "invite" && res.GroupAct != "leave" && res.GroupAct != "ws" && res.GroupAct != "items" {
		return nil, fmt.Errorf("invalid group action")
	}

	if len(parts) == 2 {
		return res, nil
	}

	if res.GroupAct == "items" {
		itemID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid item id")
		}
		res.ItemID = itemID

		if len(parts) == 3 {
			return res, nil
		}

		res.ItemAct = parts[3]
		if res.ItemAct != "rating" && res.ItemAct != "links" {
			return nil, fmt.Errorf("invalid item action")
		}

		if len(parts) == 4 {
			return res, nil
		}

		if res.ItemAct == "links" {
			linkID, err := strconv.ParseInt(parts[4], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid link id")
			}
			res.LinkID = linkID
		}
	}

	return res, nil
}

// HandleGroups — /api/v1/groups (без ID)
func (h *GroupHandler) HandleGroups(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	switch r.Method {
	case http.MethodGet:
		groups, err := h.svc.ListGroups(r.Context(), uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, groups)

	case http.MethodPost:
		var req model.CreateGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "name is required")
			return
		}
		if req.InviteIDs == nil {
			req.InviteIDs = []int64{}
		}

		group, err := h.svc.CreateGroup(r.Context(), uid, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, group)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
	}
}

// HandleGroupsByID — /api/v1/groups/...
func (h *GroupHandler) HandleGroupsByID(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	pp, err := parseGroupPath(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	// ── Входящие приглашения ──
	if pp.IsInvites {
		switch {
		case r.Method == http.MethodGet && pp.InviteID == 0:
			invites, err := h.svc.ListInvites(r.Context(), uid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}
			writeJSON(w, http.StatusOK, invites)

		case r.Method == http.MethodPost && pp.InviteID != 0 && pp.InviteAct == "accept":
			groupID, err := h.svc.AcceptInvite(r.Context(), pp.InviteID, uid)
			if err != nil {
				if errors.Is(err, service.ErrInviteNotFound) {
					writeError(w, http.StatusNotFound, "not_found", "invite not found")
					return
				}
				writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]int64{"group_id": groupID})

		case r.Method == http.MethodPost && pp.InviteID != 0 && pp.InviteAct == "decline":
			err := h.svc.DeclineInvite(r.Context(), pp.InviteID, uid)
			if err != nil {
				if errors.Is(err, service.ErrInviteNotFound) {
					writeError(w, http.StatusNotFound, "not_found", "invite not found")
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

	// ── WebSocket апгрейд ──
	if pp.GroupAct == "ws" {
		isMember, err := h.svc.IsGroupMember(r.Context(), pp.GroupID, uid)
		if err != nil || !isMember {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not member of this group")
			return
		}
		h.handleWS(w, r, pp.GroupID, uid)
		return
	}

	// ── Операции над самой группой ──
	if pp.GroupAct == "" {
		switch r.Method {
		case http.MethodGet:
			// Проверим членство
			isMember, err := h.svc.IsGroupMember(r.Context(), pp.GroupID, uid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}
			if !isMember {
				writeError(w, http.StatusForbidden, "forbidden", "not member of this group")
				return
			}

			group, err := h.svc.GetGroupByID(r.Context(), pp.GroupID)
			if err != nil {
				if errors.Is(err, service.ErrGroupNotFound) {
					writeError(w, http.StatusNotFound, "not_found", "group not found")
					return
				}
				writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}

			// Подгружаем участников и записи
			members, _ := h.svc.GetGroupMembers(r.Context(), pp.GroupID)
			items, _ := h.svc.GetGroupItems(r.Context(), pp.GroupID)
			group.Members = members
			group.Items = items

			writeJSON(w, http.StatusOK, group)

		case http.MethodDelete:
			err := h.svc.DeleteGroup(r.Context(), pp.GroupID, uid)
			if err != nil {
				if errors.Is(err, service.ErrGroupNotFound) {
					writeError(w, http.StatusNotFound, "not_found", "group not found or not owner")
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

	// ── Пригласить участника ──
	if pp.GroupAct == "invite" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
			return
		}

		var req model.InviteUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
			return
		}
		if req.Identifier == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "identifier is required")
			return
		}

		err := h.svc.InviteUser(r.Context(), pp.GroupID, uid, req.Identifier)
		if err != nil {
			if errors.Is(err, service.ErrUnauthorized) {
				writeError(w, http.StatusForbidden, "forbidden", "not authorized to invite")
				return
			}
			if errors.Is(err, service.ErrUserNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "user not found")
				return
			}
			if errors.Is(err, service.ErrAlreadyMember) {
				writeError(w, http.StatusConflict, "conflict", "user already member")
				return
			}
			if errors.Is(err, service.ErrAlreadyInvited) {
				writeError(w, http.StatusConflict, "conflict", "user already invited")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "invited"})
		return
	}

	// ── Выйти из группы ──
	if pp.GroupAct == "leave" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
			return
		}

		err := h.svc.LeaveGroup(r.Context(), pp.GroupID, uid)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// ── Записи (Group Items) ──
	if pp.GroupAct == "items" {
		// Проверим членство
		isMember, err := h.svc.IsGroupMember(r.Context(), pp.GroupID, uid)
		if err != nil || !isMember {
			writeError(w, http.StatusForbidden, "forbidden", "not member of this group")
			return
		}

		// Без ID конкретной записи (список или добавление)
		if pp.ItemID == 0 {
			switch r.Method {
			case http.MethodGet:
				items, err := h.svc.GetGroupItems(r.Context(), pp.GroupID)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
					return
				}
				writeJSON(w, http.StatusOK, items)

			case http.MethodPost:
				var req model.CreateGroupItemRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
					return
				}
				if req.Title == "" {
					writeError(w, http.StatusBadRequest, "bad_request", "title is required")
					return
				}
				if req.ContentType == "" {
					req.ContentType = model.TypeMovie
				}
				if req.Status == "" {
					req.Status = model.StatusPlanned
				}
				if req.Genres == nil {
					req.Genres = []string{}
				}

				item, err := h.svc.AddGroupItem(r.Context(), pp.GroupID, uid, req)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
					return
				}

				// Broadcast
				h.hub.Broadcast(pp.GroupID, "item_added", item)

				writeJSON(w, http.StatusCreated, item)

			default:
				writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
			}
			return
		}

		// С ID конкретной записи
		if pp.ItemID != 0 && pp.ItemAct == "" {
			switch r.Method {
			case http.MethodPut:
				var req model.UpdateGroupItemRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
					return
				}
				if req.Title == "" {
					writeError(w, http.StatusBadRequest, "bad_request", "title is required")
					return
				}
				if req.ContentType == "" {
					req.ContentType = model.TypeMovie
				}
				if req.Status == "" {
					req.Status = model.StatusPlanned
				}
				if req.Genres == nil {
					req.Genres = []string{}
				}

				item, err := h.svc.UpdateGroupItem(r.Context(), pp.GroupID, pp.ItemID, uid, req)
				if err != nil {
					if errors.Is(err, service.ErrItemNotFound) {
						writeError(w, http.StatusNotFound, "not_found", "item not found or not creator")
						return
					}
					writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
					return
				}

				// Broadcast
				h.hub.Broadcast(pp.GroupID, "item_updated", item)

				writeJSON(w, http.StatusOK, item)

			case http.MethodDelete:
				err := h.svc.DeleteGroupItem(r.Context(), pp.GroupID, pp.ItemID, uid)
				if err != nil {
					if errors.Is(err, service.ErrItemNotFound) {
						writeError(w, http.StatusNotFound, "not_found", "item not found or not creator")
						return
					}
					writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
					return
				}

				// Broadcast
				h.hub.Broadcast(pp.GroupID, "item_deleted", map[string]int64{"id": pp.ItemID})

				w.WriteHeader(http.StatusNoContent)

			default:
				writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
			}
			return
		}

		// ── Оценка записи ──
		if pp.ItemID != 0 && pp.ItemAct == "rating" {
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
				return
			}

			var req model.GroupItemRatingRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
				return
			}
			if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 10) {
				writeError(w, http.StatusBadRequest, "bad_request", "rating must be between 1 and 10")
				return
			}

			_, err := h.svc.UpdateItemRating(r.Context(), pp.GroupID, pp.ItemID, uid, req.Rating)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}

			// Broadcast updated item & return it
			item, err := h.svc.GetGroupItemByID(r.Context(), pp.ItemID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
				return
			}
			h.hub.Broadcast(pp.GroupID, "item_updated", item)
			writeJSON(w, http.StatusOK, item)
			return
		}

		// ── Ссылки к записи ──
		if pp.ItemID != 0 && pp.ItemAct == "links" {
			switch {
			case r.Method == http.MethodPost && pp.LinkID == 0:
				var req model.AddGroupItemLinkRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
					return
				}
				if req.Label == "" || req.URL == "" {
					writeError(w, http.StatusBadRequest, "bad_request", "label and url are required")
					return
				}

				_, err := h.svc.AddGroupItemLink(r.Context(), pp.GroupID, pp.ItemID, uid, req.Label, req.URL)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
					return
				}

				// Broadcast updated item & return it
				item, err := h.svc.GetGroupItemByID(r.Context(), pp.ItemID)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
					return
				}
				h.hub.Broadcast(pp.GroupID, "item_updated", item)
				writeJSON(w, http.StatusCreated, item)

			case r.Method == http.MethodDelete && pp.LinkID != 0:
				err := h.svc.DeleteGroupItemLink(r.Context(), pp.GroupID, pp.LinkID, pp.ItemID, uid)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
					return
				}

				// Broadcast updated item & return it
				item, err := h.svc.GetGroupItemByID(r.Context(), pp.ItemID)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
					return
				}
				h.hub.Broadcast(pp.GroupID, "item_updated", item)
				writeJSON(w, http.StatusOK, item)

			default:
				writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
			}
			return
		}
	}
}

func (h *GroupHandler) handleWS(w http.ResponseWriter, r *http.Request, groupID int64, userID int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		Conn:    conn,
		UserID:  userID,
		GroupID: groupID,
	}

	h.hub.Register(client)

	defer func() {
		h.hub.Unregister(client)
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

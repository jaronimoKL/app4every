package model

import "time"

// GroupList — группа (список)
type GroupList struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	OwnerID   int64         `json:"owner_id"`
	CreatedAt time.Time     `json:"created_at"`
	Members   []GroupMember `json:"members,omitempty"`
	Items     []GroupItem   `json:"items,omitempty"`
}

// GroupMember — участник группы
type GroupMember struct {
	ID       int64     `json:"id"`
	GroupID  int64     `json:"group_id"`
	UserID   int64     `json:"user_id"`
	Username string    `json:"username"` // JOIN c users
	JoinedAt time.Time `json:"joined_at"`
}

// GroupInvite — приглашение в группу
type GroupInvite struct {
	ID              int64     `json:"id"`
	GroupID         int64     `json:"group_id"`
	GroupName       string    `json:"group_name,omitempty"`       // JOIN c group_lists
	InviterID       int64     `json:"inviter_id"`
	InviterUsername string    `json:"inviter_username,omitempty"` // JOIN c users
	InviteeeID      int64     `json:"invitee_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

// GroupItem — запись в групповом списке
type GroupItem struct {
	ID              int64             `json:"id"`
	GroupID         int64             `json:"group_id"`
	AddedBy         int64             `json:"added_by"`
	AddedByUsername string            `json:"added_by_username"` // JOIN c users
	Title           string            `json:"title"`
	ContentType     ContentType       `json:"content_type"`
	Status          ReviewStatus      `json:"status"`
	CurrentEpisode  int               `json:"current_episode"`
	MaxEpisodes     int               `json:"max_episodes"`
	Notes           string            `json:"notes"`
	PosterURL       string            `json:"poster_url"`
	Genres          []string          `json:"genres"` // массив строк (TEXT[])
	ShikimoriID     *int              `json:"shikimori_id,omitempty"`
	TmdbID          *int              `json:"tmdb_id,omitempty"`
	Description     string            `json:"description"`
	EpisodesTotal   *int              `json:"episodes_total,omitempty"`
	AnilibertyAlias string            `json:"aniliberty_alias"`
	ShikimoriScore  *float64          `json:"shikimori_score,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Ratings         []GroupItemRating `json:"ratings"`
	Links           []GroupItemLink   `json:"links"`
	AverageRating   float64           `json:"average_rating"` // Вычисляемое
}

// GroupItemRating — персональная оценка участника к записи
type GroupItemRating struct {
	ID        int64     `json:"id"`
	ItemID    int64     `json:"item_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"` // JOIN c users
	Rating    *int16    `json:"rating"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GroupItemLink — ссылка прикреплённая к записи участником
type GroupItemLink struct {
	ID        int64     `json:"id"`
	ItemID    int64     `json:"item_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"` // JOIN c users
	Label     string    `json:"label"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// WSEvent WebSocket JSON события
type WSEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// ── Запросы к REST API ──

type CreateGroupRequest struct {
	Name      string  `json:"name"`
	InviteIDs []int64 `json:"invite_ids"`
}

type InviteUserRequest struct {
	Identifier string `json:"identifier"`
}

type CreateGroupItemRequest struct {
	Title           string       `json:"title"`
	ContentType     ContentType  `json:"content_type"`
	Status          ReviewStatus `json:"status"`
	CurrentEpisode  int          `json:"current_episode"`
	MaxEpisodes     int          `json:"max_episodes"`
	Notes           string       `json:"notes"`
	PosterURL       string       `json:"poster_url"`
	Genres          []string     `json:"genres"`
	ShikimoriID     *int         `json:"shikimori_id,omitempty"`
	TmdbID          *int         `json:"tmdb_id,omitempty"`
	Description     string       `json:"description"`
	EpisodesTotal   *int         `json:"episodes_total,omitempty"`
	AnilibertyAlias string       `json:"aniliberty_alias"`
	ShikimoriScore  *float64     `json:"shikimori_score,omitempty"`
}

type UpdateGroupItemRequest struct {
	Title           string       `json:"title"`
	ContentType     ContentType  `json:"content_type"`
	Status          ReviewStatus `json:"status"`
	CurrentEpisode  int          `json:"current_episode"`
	MaxEpisodes     int          `json:"max_episodes"`
	Notes           string       `json:"notes"`
	PosterURL       string       `json:"poster_url"`
	Genres          []string     `json:"genres"`
	ShikimoriID     *int         `json:"shikimori_id,omitempty"`
	TmdbID          *int         `json:"tmdb_id,omitempty"`
	Description     string       `json:"description"`
	EpisodesTotal   *int         `json:"episodes_total,omitempty"`
	AnilibertyAlias string       `json:"aniliberty_alias"`
	ShikimoriScore  *float64     `json:"shikimori_score,omitempty"`
}

type GroupItemRatingRequest struct {
	Rating *int16 `json:"rating"`
}

type AddGroupItemLinkRequest struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

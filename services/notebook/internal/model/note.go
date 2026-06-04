package model

import "time"

// Note — одна заметка пользователя.
type Note struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateNoteRequest — тело запроса при создании заметки.
type CreateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// UpdateNoteRequest — тело запроса при обновлении.
type UpdateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

package service

import (
	"context"
	"errors"

	"app4every/services/notebook/internal/model"
	"app4every/services/notebook/internal/repository"
)

var ErrNoteNotFound = errors.New("note not found")

type NoteService interface {
	CreateNote(ctx context.Context, userID int64, req model.CreateNoteRequest) (*model.Note, error)
	ListNotes(ctx context.Context, userID int64) ([]*model.Note, error)
	GetNote(ctx context.Context, id, userID int64) (*model.Note, error)
	UpdateNote(ctx context.Context, id, userID int64, req model.UpdateNoteRequest) (*model.Note, error)
	DeleteNote(ctx context.Context, id, userID int64) error
}

type noteService struct {
	repo repository.NoteRepository
}

func NewNoteService(repo repository.NoteRepository) NoteService {
	return &noteService{repo: repo}
}

func (s *noteService) CreateNote(ctx context.Context, userID int64, req model.CreateNoteRequest) (*model.Note, error) {
	return s.repo.Create(ctx, userID, req.Title, req.Content)
}

func (s *noteService) ListNotes(ctx context.Context, userID int64) ([]*model.Note, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *noteService) GetNote(ctx context.Context, id, userID int64) (*model.Note, error) {
	note, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoteNotFound) {
			return nil, ErrNoteNotFound
		}
		return nil, err
	}
	return note, nil
}

func (s *noteService) UpdateNote(ctx context.Context, id, userID int64, req model.UpdateNoteRequest) (*model.Note, error) {
	note, err := s.repo.Update(ctx, id, userID, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, repository.ErrNoteNotFound) {
			return nil, ErrNoteNotFound
		}
		return nil, err
	}
	return note, nil
}

func (s *noteService) DeleteNote(ctx context.Context, id, userID int64) error {
	err := s.repo.Delete(ctx, id, userID)
	if errors.Is(err, repository.ErrNoteNotFound) {
		return ErrNoteNotFound
	}
	return err
}

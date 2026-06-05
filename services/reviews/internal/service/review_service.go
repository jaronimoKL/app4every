package service

import (
	"context"
	"errors"

	"app4every/services/reviews/internal/model"
	"app4every/services/reviews/internal/repository"
)

var (
	ErrReviewNotFound = errors.New("review not found")
	ErrLinkNotFound   = errors.New("link not found")
)

type ReviewService interface {
	CreateReview(ctx context.Context, userID int64, req model.CreateReviewRequest) (*model.Review, error)
	ListReviews(ctx context.Context, userID int64) ([]*model.Review, error)
	GetReview(ctx context.Context, id, userID int64) (*model.Review, error)
	UpdateReview(ctx context.Context, id, userID int64, req model.UpdateReviewRequest) (*model.Review, error)
	DeleteReview(ctx context.Context, id, userID int64) error
	AddLink(ctx context.Context, reviewID, userID int64, req model.AddLinkRequest) (*model.ReviewLink, error)
	DeleteLink(ctx context.Context, linkID, reviewID, userID int64) error
}

type reviewService struct {
	repo repository.ReviewRepository
}

func NewReviewService(repo repository.ReviewRepository) ReviewService {
	return &reviewService{repo: repo}
}

func (s *reviewService) CreateReview(ctx context.Context, userID int64, req model.CreateReviewRequest) (*model.Review, error) {
	return s.repo.Create(ctx, userID, req)
}

func (s *reviewService) ListReviews(ctx context.Context, userID int64) ([]*model.Review, error) {
	return s.repo.GetAllByUserID(ctx, userID)
}

func (s *reviewService) GetReview(ctx context.Context, id, userID int64) (*model.Review, error) {
	rev, err := s.repo.GetByID(ctx, id, userID)
	if errors.Is(err, repository.ErrReviewNotFound) {
		return nil, ErrReviewNotFound
	}
	return rev, err
}

func (s *reviewService) UpdateReview(ctx context.Context, id, userID int64, req model.UpdateReviewRequest) (*model.Review, error) {
	rev, err := s.repo.Update(ctx, id, userID, req)
	if errors.Is(err, repository.ErrReviewNotFound) {
		return nil, ErrReviewNotFound
	}
	return rev, err
}

func (s *reviewService) DeleteReview(ctx context.Context, id, userID int64) error {
	err := s.repo.Delete(ctx, id, userID)
	if errors.Is(err, repository.ErrReviewNotFound) {
		return ErrReviewNotFound
	}
	return err
}

func (s *reviewService) AddLink(ctx context.Context, reviewID, userID int64, req model.AddLinkRequest) (*model.ReviewLink, error) {
	link, err := s.repo.AddLink(ctx, reviewID, userID, req)
	if errors.Is(err, repository.ErrReviewNotFound) {
		return nil, ErrReviewNotFound
	}
	return link, err
}

func (s *reviewService) DeleteLink(ctx context.Context, linkID, reviewID, userID int64) error {
	err := s.repo.DeleteLink(ctx, linkID, reviewID, userID)
	if errors.Is(err, repository.ErrLinkNotFound) {
		return ErrLinkNotFound
	}
	return err
}

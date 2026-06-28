package service

import (
	"context"

	"app4every/services/auth/internal/hub"
	"app4every/services/auth/internal/repository"
)

type NotificationService interface {
	SendNotification(ctx context.Context, userID int64, notifType, message string, metadata map[string]interface{}, transient bool) error
	GetUserNotifications(ctx context.Context, userID int64) ([]*repository.Notification, error)
	MarkAsRead(ctx context.Context, userID int64, notificationIDs []int64) error
	DeleteNotification(ctx context.Context, userID int64, notificationID int64) error
}

type notificationService struct {
	repo repository.NotificationRepo
	hub  *hub.NotificationHub
}

func NewNotificationService(repo repository.NotificationRepo, notificationHub *hub.NotificationHub) NotificationService {
	return &notificationService{
		repo: repo,
		hub:  notificationHub,
	}
}

func (s *notificationService) SendNotification(ctx context.Context, userID int64, notifType, message string, metadata map[string]interface{}, transient bool) error {
	n := &repository.Notification{
		UserID:   userID,
		Type:     notifType,
		Message:  message,
		Metadata: metadata,
	}

	// 1. Save to DB if not transient
	if !transient {
		if err := s.repo.Create(ctx, n); err != nil {
			return err
		}
	}

	// 2. Broadcast via WS
	s.hub.SendToUser(userID, map[string]interface{}{
		"event":        "new_notification",
		"notification": n,
	})

	return nil
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userID int64) ([]*repository.Notification, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, userID int64, notificationIDs []int64) error {
	return s.repo.MarkAsRead(ctx, userID, notificationIDs)
}

func (s *notificationService) DeleteNotification(ctx context.Context, userID int64, notificationID int64) error {
	return s.repo.Delete(ctx, userID, notificationID)
}

package service

import (
	"context"
	"errors"
	"fmt"

	"app4every/services/reviews/internal/model"
	"app4every/services/reviews/internal/repository"
)

var (
	ErrGroupNotFound   = errors.New("group not found")
	ErrItemNotFound    = errors.New("item not found")
	ErrInviteNotFound  = errors.New("invite not found")
	ErrUnauthorized    = errors.New("unauthorized action")
	ErrAlreadyMember   = errors.New("user is already a member")
	ErrAlreadyInvited  = errors.New("user has already been invited")
	ErrUserNotFound    = errors.New("user not found")
)

type GroupService interface {
	ListGroups(ctx context.Context, userID int64) ([]*model.GroupList, error)
	CreateGroup(ctx context.Context, ownerID int64, req model.CreateGroupRequest) (*model.GroupList, error)
	GetGroupByID(ctx context.Context, id int64) (*model.GroupList, error)
	DeleteGroup(ctx context.Context, id, ownerID int64) error
	IsGroupMember(ctx context.Context, groupID, userID int64) (bool, error)

	InviteUser(ctx context.Context, groupID, inviterID int64, identifier string) error
	ListInvites(ctx context.Context, userID int64) ([]*model.GroupInvite, error)
	AcceptInvite(ctx context.Context, inviteID, userID int64) (int64, error)
	DeclineInvite(ctx context.Context, inviteID, userID int64) error
	LeaveGroup(ctx context.Context, groupID, userID int64) error
	GetGroupMembers(ctx context.Context, groupID int64) ([]model.GroupMember, error)

	AddGroupItem(ctx context.Context, groupID, userID int64, req model.CreateGroupItemRequest) (*model.GroupItem, error)
	UpdateGroupItem(ctx context.Context, groupID, itemID, userID int64, req model.UpdateGroupItemRequest) (*model.GroupItem, error)
	DeleteGroupItem(ctx context.Context, groupID, itemID, userID int64) error
	GetGroupItems(ctx context.Context, groupID int64) ([]model.GroupItem, error)
	GetGroupItemByID(ctx context.Context, itemID int64) (*model.GroupItem, error)

	UpdateItemRating(ctx context.Context, groupID, itemID, userID int64, rating *int16) (*model.GroupItemRating, error)
	AddGroupItemLink(ctx context.Context, groupID, itemID, userID int64, label, url string) (*model.GroupItemLink, error)
	DeleteGroupItemLink(ctx context.Context, groupID, linkID, itemID, userID int64) error
}

type groupService struct {
	groupRepo  repository.GroupRepository
	reviewRepo repository.ReviewRepository
}

func NewGroupService(groupRepo repository.GroupRepository, reviewRepo repository.ReviewRepository) GroupService {
	return &groupService{
		groupRepo:  groupRepo,
		reviewRepo: reviewRepo,
	}
}

func (s *groupService) ListGroups(ctx context.Context, userID int64) ([]*model.GroupList, error) {
	return s.groupRepo.ListGroups(ctx, userID)
}

func (s *groupService) CreateGroup(ctx context.Context, ownerID int64, req model.CreateGroupRequest) (*model.GroupList, error) {
	return s.groupRepo.CreateGroup(ctx, ownerID, req)
}

func (s *groupService) GetGroupByID(ctx context.Context, id int64) (*model.GroupList, error) {
	g, err := s.groupRepo.GetGroupByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrGroupNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return g, nil
}

func (s *groupService) DeleteGroup(ctx context.Context, id, ownerID int64) error {
	err := s.groupRepo.DeleteGroup(ctx, id, ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrGroupNotFound) {
			return ErrGroupNotFound
		}
		return err
	}
	return nil
}

func (s *groupService) IsGroupMember(ctx context.Context, groupID, userID int64) (bool, error) {
	return s.groupRepo.IsGroupMember(ctx, groupID, userID)
}

func (s *groupService) InviteUser(ctx context.Context, groupID, inviterID int64, identifier string) error {
	// Проверим, что приглашающий - член или владелец группы
	isMember, err := s.groupRepo.IsGroupMember(ctx, groupID, inviterID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrUnauthorized
	}

	err = s.groupRepo.InviteUser(ctx, groupID, inviterID, identifier)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		if errors.Is(err, repository.ErrAlreadyMember) {
			return ErrAlreadyMember
		}
		if errors.Is(err, repository.ErrAlreadyInvited) {
			return ErrAlreadyInvited
		}
		return err
	}
	return nil
}

func (s *groupService) ListInvites(ctx context.Context, userID int64) ([]*model.GroupInvite, error) {
	return s.groupRepo.ListInvites(ctx, userID)
}

func (s *groupService) AcceptInvite(ctx context.Context, inviteID, userID int64) (int64, error) {
	groupID, err := s.groupRepo.AcceptInvite(ctx, inviteID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrInviteNotFound) {
			return 0, ErrInviteNotFound
		}
		return 0, err
	}
	return groupID, nil
}

func (s *groupService) DeclineInvite(ctx context.Context, inviteID, userID int64) error {
	err := s.groupRepo.DeclineInvite(ctx, inviteID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrInviteNotFound) {
			return ErrInviteNotFound
		}
		return err
	}
	return nil
}

func (s *groupService) LeaveGroup(ctx context.Context, groupID, userID int64) error {
	return s.groupRepo.LeaveGroup(ctx, groupID, userID)
}

func (s *groupService) GetGroupMembers(ctx context.Context, groupID int64) ([]model.GroupMember, error) {
	return s.groupRepo.GetGroupMembers(ctx, groupID)
}

func (s *groupService) AddGroupItem(ctx context.Context, groupID, userID int64, req model.CreateGroupItemRequest) (*model.GroupItem, error) {
	isMember, err := s.groupRepo.IsGroupMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	item, err := s.groupRepo.AddGroupItem(ctx, groupID, userID, req)
	if err != nil {
		return nil, err
	}

	// Получаем имя группы для заметок
	groupName := ""
	g, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err == nil && g != nil {
		groupName = g.Name
	}

	// Правило 2: Когда запись добавляется автором -> авто-создание рецензии в его личном списке
	s.syncPersonalReview(ctx, userID, item.Title, item.ContentType, item.Status, item.PosterURL, item.Notes, item.Genres, groupName)

	// Правило 1: Если статус сразу completed -> синхронизируем completed рецензии всем участникам группы
	if item.Status == model.StatusCompleted {
		s.syncCompletedToAllMembers(ctx, groupID, item, groupName)
	}

	return item, nil
}

func (s *groupService) UpdateGroupItem(ctx context.Context, groupID, itemID, userID int64, req model.UpdateGroupItemRequest) (*model.GroupItem, error) {
	isMember, err := s.groupRepo.IsGroupMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	item, err := s.groupRepo.UpdateGroupItem(ctx, groupID, itemID, userID, req)
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}

	groupName := ""
	g, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err == nil && g != nil {
		groupName = g.Name
	}

	// Правило 1: При смене статуса на completed -> синхронизируем completed рецензии всем участникам
	if item.Status == model.StatusCompleted {
		s.syncCompletedToAllMembers(ctx, groupID, item, groupName)
	}

	return item, nil
}

func (s *groupService) DeleteGroupItem(ctx context.Context, groupID, itemID, userID int64) error {
	isMember, err := s.groupRepo.IsGroupMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrUnauthorized
	}

	err = s.groupRepo.DeleteGroupItem(ctx, groupID, itemID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			return ErrItemNotFound
		}
		return err
	}
	return nil
}

func (s *groupService) GetGroupItems(ctx context.Context, groupID int64) ([]model.GroupItem, error) {
	return s.groupRepo.GetGroupItems(ctx, groupID)
}

func (s *groupService) GetGroupItemByID(ctx context.Context, itemID int64) (*model.GroupItem, error) {
	item, err := s.groupRepo.GetGroupItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *groupService) UpdateItemRating(ctx context.Context, groupID, itemID, userID int64, rating *int16) (*model.GroupItemRating, error) {
	isMember, err := s.groupRepo.IsGroupMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	rat, err := s.groupRepo.UpdateItemRating(ctx, itemID, userID, rating)
	if err != nil {
		return nil, err
	}

	// Если статус записи не completed, при выставлении оценки мы не меняем статус. 
	// Но если оценка выставляется, то средняя оценка пересчитается на стороне репозитория.
	return rat, nil
}

func (s *groupService) AddGroupItemLink(ctx context.Context, groupID, itemID, userID int64, label, url string) (*model.GroupItemLink, error) {
	isMember, err := s.groupRepo.IsGroupMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	return s.groupRepo.AddGroupItemLink(ctx, itemID, userID, label, url)
}

func (s *groupService) DeleteGroupItemLink(ctx context.Context, groupID, linkID, itemID, userID int64) error {
	isMember, err := s.groupRepo.IsGroupMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrUnauthorized
	}

	return s.groupRepo.DeleteGroupItemLink(ctx, linkID, itemID, userID)
}

// ── Вспомогательные методы синхронизации ──

func (s *groupService) syncPersonalReview(ctx context.Context, userID int64, title string, contentType model.ContentType, status model.ReviewStatus, posterURL, notes string, genres []string, groupName string) {
	existingReviews, err := s.reviewRepo.GetAllByUserID(ctx, userID)
	if err != nil {
		return
	}

	var match *model.Review
	for _, r := range existingReviews {
		if r.Title == title && r.ContentType == contentType {
			match = r
			break
		}
	}

	if match == nil {
		// Создаем рецензию
		notesWithGroup := notes
		if groupName != "" {
			if notesWithGroup != "" {
				notesWithGroup += "\n\n"
			}
			notesWithGroup += fmt.Sprintf("Синхронизировано из группы: %s", groupName)
		}

		createReq := model.CreateReviewRequest{
			Title:       title,
			ContentType: contentType,
			Status:      status,
			PosterURL:   posterURL,
			Notes:       notesWithGroup,
		}
		newReview, err := s.reviewRepo.Create(ctx, userID, createReq)
		if err == nil && newReview != nil {
			for _, gName := range genres {
				_, _ = s.reviewRepo.AddGenre(ctx, newReview.ID, userID, gName)
			}
		}
	} else {
		// Если рецензия существует и статус перешел в completed, обновляем ее статус
		if status == model.StatusCompleted && match.Status != model.StatusCompleted {
			updateReq := model.UpdateReviewRequest{
				Title:       match.Title,
				ContentType: match.ContentType,
				Status:      model.StatusCompleted,
				Rating:      match.Rating,
				Notes:       match.Notes,
				PosterURL:   match.PosterURL,
			}
			_, _ = s.reviewRepo.Update(ctx, match.ID, userID, updateReq)
		}
	}
}

func (s *groupService) syncCompletedToAllMembers(ctx context.Context, groupID int64, item *model.GroupItem, groupName string) {
	members, err := s.groupRepo.GetGroupMembers(ctx, groupID)
	if err != nil {
		return
	}

	for _, m := range members {
		s.syncPersonalReview(ctx, m.UserID, item.Title, item.ContentType, model.StatusCompleted, item.PosterURL, item.Notes, item.Genres, groupName)
	}
}

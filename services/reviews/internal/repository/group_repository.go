package repository

import (
	"context"
	"errors"
	"fmt"

	"app4every/services/reviews/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrGroupNotFound   = errors.New("group not found")
	ErrItemNotFound    = errors.New("group item not found")
	ErrInviteNotFound  = errors.New("invite not found")
	ErrUnauthorized    = errors.New("unauthorized action")
	ErrAlreadyMember   = errors.New("user is already a member")
	ErrAlreadyInvited  = errors.New("user has already been invited")
	ErrUserNotFound    = errors.New("user not found")
)

type GroupRepository interface {
	// Группы
	ListGroups(ctx context.Context, userID int64) ([]*model.GroupList, error)
	CreateGroup(ctx context.Context, ownerID int64, req model.CreateGroupRequest) (*model.GroupList, error)
	GetGroupByID(ctx context.Context, id int64) (*model.GroupList, error)
	DeleteGroup(ctx context.Context, id, ownerID int64) error
	IsGroupMember(ctx context.Context, groupID, userID int64) (bool, error)
	IsGroupOwner(ctx context.Context, groupID, userID int64) (bool, error)

	// Члены и приглашения
	InviteUser(ctx context.Context, groupID, inviterID int64, identifier string) error
	ListInvites(ctx context.Context, userID int64) ([]*model.GroupInvite, error)
	AcceptInvite(ctx context.Context, inviteID, userID int64) (int64, error) // Возвращает groupID
	DeclineInvite(ctx context.Context, inviteID, userID int64) error
	LeaveGroup(ctx context.Context, groupID, userID int64) error
	GetGroupMembers(ctx context.Context, groupID int64) ([]model.GroupMember, error)

	// Записи
	AddGroupItem(ctx context.Context, groupID, userID int64, req model.CreateGroupItemRequest) (*model.GroupItem, error)
	UpdateGroupItem(ctx context.Context, groupID, itemID, userID int64, req model.UpdateGroupItemRequest) (*model.GroupItem, error)
	UpdateGroupItemStatusOnly(ctx context.Context, groupID, itemID int64, status model.ReviewStatus) (*model.GroupItem, error)
	DeleteGroupItem(ctx context.Context, groupID, itemID, userID int64) error
	GetGroupItemByID(ctx context.Context, itemID int64) (*model.GroupItem, error)
	GetGroupItems(ctx context.Context, groupID int64) ([]model.GroupItem, error)

	// Оценки
	UpdateItemRating(ctx context.Context, itemID, userID int64, rating *int16) (*model.GroupItemRating, error)
	GetItemRatings(ctx context.Context, itemID int64) ([]model.GroupItemRating, error)

	// Ссылки
	AddGroupItemLink(ctx context.Context, itemID, userID int64, label, url string) (*model.GroupItemLink, error)
	DeleteGroupItemLink(ctx context.Context, linkID, itemID, userID int64) error
	GetItemLinks(ctx context.Context, itemID int64) ([]model.GroupItemLink, error)
}

type postgresGroupRepository struct {
	db *pgxpool.Pool
}

func NewGroupRepository(db *pgxpool.Pool) GroupRepository {
	return &postgresGroupRepository{db: db}
}

// ── Группы ──

func (r *postgresGroupRepository) ListGroups(ctx context.Context, userID int64) ([]*model.GroupList, error) {
	query := `
		SELECT gl.id, gl.name, gl.owner_id, gl.created_at
		FROM group_lists gl
		LEFT JOIN group_members gm ON gl.id = gm.group_id
		WHERE gl.owner_id = $1 OR gm.user_id = $1
		GROUP BY gl.id
		ORDER BY gl.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*model.GroupList
	for rows.Next() {
		g := &model.GroupList{Members: []model.GroupMember{}, Items: []model.GroupItem{}}
		if err := rows.Scan(&g.ID, &g.Name, &g.OwnerID, &g.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	if groups == nil {
		groups = []*model.GroupList{}
	}
	return groups, nil
}

func (r *postgresGroupRepository) CreateGroup(ctx context.Context, ownerID int64, req model.CreateGroupRequest) (*model.GroupList, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Создаем группу
	g := &model.GroupList{Members: []model.GroupMember{}, Items: []model.GroupItem{}}
	err = tx.QueryRow(ctx, `
		INSERT INTO group_lists (name, owner_id, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id, name, owner_id, created_at
	`, req.Name, ownerID).Scan(&g.ID, &g.Name, &g.OwnerID, &g.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 2. Владелец автоматически становится участником
	_, err = tx.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id, joined_at)
		VALUES ($1, $2, NOW())
	`, g.ID, ownerID)
	if err != nil {
		return nil, err
	}

	// 3. Отправляем приглашения
	for _, inviteeID := range req.InviteIDs {
		if inviteeID == ownerID {
			continue // Не приглашаем самого себя
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO group_invites (group_id, inviter_id, invitee_id, status, created_at)
			VALUES ($1, $2, $3, 'pending', NOW())
			ON CONFLICT (group_id, invitee_id) DO NOTHING
		`, g.ID, ownerID, inviteeID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return g, nil
}

func (r *postgresGroupRepository) GetGroupByID(ctx context.Context, id int64) (*model.GroupList, error) {
	g := &model.GroupList{Members: []model.GroupMember{}, Items: []model.GroupItem{}}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, owner_id, created_at
		FROM group_lists WHERE id = $1
	`, id).Scan(&g.ID, &g.Name, &g.OwnerID, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return g, nil
}

func (r *postgresGroupRepository) DeleteGroup(ctx context.Context, id, ownerID int64) error {
	res, err := r.db.Exec(ctx, `DELETE FROM group_lists WHERE id = $1 AND owner_id = $2`, id, ownerID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrGroupNotFound
	}
	return nil
}

func (r *postgresGroupRepository) IsGroupMember(ctx context.Context, groupID, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2
		)
	`, groupID, userID).Scan(&exists)
	return exists, err
}

func (r *postgresGroupRepository) IsGroupOwner(ctx context.Context, groupID, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM group_lists WHERE id = $1 AND owner_id = $2
		)
	`, groupID, userID).Scan(&exists)
	return exists, err
}

// ── Члены и приглашения ──

func (r *postgresGroupRepository) InviteUser(ctx context.Context, groupID, inviterID int64, identifier string) error {
	var targetID int64
	var err error

	var idVal int64
	if _, scanErr := fmt.Sscanf(identifier, "%d", &idVal); scanErr == nil {
		err = r.db.QueryRow(ctx, "SELECT id FROM users WHERE id = $1", idVal).Scan(&targetID)
	} else {
		err = r.db.QueryRow(ctx, "SELECT id FROM users WHERE username = $1", identifier).Scan(&targetID)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	// Проверяем членство
	var isMember bool
	err = r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)", groupID, targetID).Scan(&isMember)
	if err != nil {
		return err
	}
	if isMember {
		return ErrAlreadyMember
	}

	// Проверяем наличие активного приглашения
	var isInvited bool
	err = r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM group_invites WHERE group_id = $1 AND invitee_id = $2 AND status = 'pending')", groupID, targetID).Scan(&isInvited)
	if err != nil {
		return err
	}
	if isInvited {
		return ErrAlreadyInvited
	}

	// Вставляем или обновляем статус приглашения
	_, err = r.db.Exec(ctx, `
		INSERT INTO group_invites (group_id, inviter_id, invitee_id, status, created_at)
		VALUES ($1, $2, $3, 'pending', NOW())
		ON CONFLICT (group_id, invitee_id) DO UPDATE SET status = 'pending', created_at = NOW()
	`, groupID, inviterID, targetID)
	return err
}

func (r *postgresGroupRepository) ListInvites(ctx context.Context, userID int64) ([]*model.GroupInvite, error) {
	query := `
		SELECT gi.id, gi.group_id, gl.name as group_name, gi.inviter_id, u.username as inviter_username, gi.invitee_id, gi.status, gi.created_at
		FROM group_invites gi
		JOIN group_lists gl ON gi.group_id = gl.id
		JOIN users u ON gi.inviter_id = u.id
		WHERE gi.invitee_id = $1 AND gi.status = 'pending'
		ORDER BY gi.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*model.GroupInvite
	for rows.Next() {
		inv := &model.GroupInvite{}
		if err := rows.Scan(&inv.ID, &inv.GroupID, &inv.GroupName, &inv.InviterID, &inv.InviterUsername, &inv.InviteeeID, &inv.Status, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}
	if invites == nil {
		invites = []*model.GroupInvite{}
	}
	return invites, nil
}

func (r *postgresGroupRepository) AcceptInvite(ctx context.Context, inviteID, userID int64) (int64, error) {
	var groupID int64
	var status string
	err := r.db.QueryRow(ctx, "SELECT group_id, status FROM group_invites WHERE id = $1 AND invitee_id = $2", inviteID, userID).Scan(&groupID, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInviteNotFound
		}
		return 0, err
	}
	if status != "pending" {
		return 0, errors.New("invite is not pending")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// 1. Принимаем инвайт
	_, err = tx.Exec(ctx, "UPDATE group_invites SET status = 'accepted' WHERE id = $1", inviteID)
	if err != nil {
		return 0, err
	}

	// 2. Вступаем в группу
	_, err = tx.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id, joined_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (group_id, user_id) DO NOTHING
	`, groupID, userID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return groupID, nil
}

func (r *postgresGroupRepository) DeclineInvite(ctx context.Context, inviteID, userID int64) error {
	res, err := r.db.Exec(ctx, "UPDATE group_invites SET status = 'declined' WHERE id = $1 AND invitee_id = $2 AND status = 'pending'", inviteID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrInviteNotFound
	}
	return nil
}

func (r *postgresGroupRepository) LeaveGroup(ctx context.Context, groupID, userID int64) error {
	// Проверим, что не владелец
	var ownerID int64
	err := r.db.QueryRow(ctx, "SELECT owner_id FROM group_lists WHERE id = $1", groupID).Scan(&ownerID)
	if err != nil {
		return err
	}
	if ownerID == userID {
		return errors.New("owner cannot leave their group list, they must delete it")
	}

	res, err := r.db.Exec(ctx, "DELETE FROM group_members WHERE group_id = $1 AND user_id = $2", groupID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("membership not found")
	}
	return nil
}

func (r *postgresGroupRepository) GetGroupMembers(ctx context.Context, groupID int64) ([]model.GroupMember, error) {
	query := `
		SELECT gm.id, gm.group_id, gm.user_id, u.username, gm.joined_at
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = $1
		ORDER BY gm.joined_at ASC
	`
	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.GroupMember
	for rows.Next() {
		m := model.GroupMember{}
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.Username, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if members == nil {
		members = []model.GroupMember{}
	}
	return members, nil
}

// ── Записи ──

func (r *postgresGroupRepository) AddGroupItem(ctx context.Context, groupID, userID int64, req model.CreateGroupItemRequest) (*model.GroupItem, error) {
	item := &model.GroupItem{Ratings: []model.GroupItemRating{}, Links: []model.GroupItemLink{}}
	err := r.db.QueryRow(ctx, `
		INSERT INTO group_items (group_id, added_by, title, content_type, status, current_episode, max_episodes, notes, poster_url, genres, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, group_id, added_by, title, content_type, status, current_episode, max_episodes, notes, poster_url, genres, created_at, updated_at
	`, groupID, userID, req.Title, req.ContentType, req.Status, req.CurrentEpisode, req.MaxEpisodes, req.Notes, req.PosterURL, req.Genres).
		Scan(&item.ID, &item.GroupID, &item.AddedBy, &item.Title, &item.ContentType, &item.Status, &item.CurrentEpisode, &item.MaxEpisodes,
			&item.Notes, &item.PosterURL, &item.Genres, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Добавим username автора
	_ = r.db.QueryRow(ctx, "SELECT username FROM users WHERE id = $1", userID).Scan(&item.AddedByUsername)
	return item, nil
}

func (r *postgresGroupRepository) UpdateGroupItem(ctx context.Context, groupID, itemID, userID int64, req model.UpdateGroupItemRequest) (*model.GroupItem, error) {
	item := &model.GroupItem{}
	err := r.db.QueryRow(ctx, `
		UPDATE group_items
		SET title = $1, content_type = $2, status = $3, current_episode = $4, max_episodes = $5, notes = $6, poster_url = $7, genres = $8, updated_at = NOW()
		WHERE id = $9 AND group_id = $10 AND added_by = $11
		RETURNING id, group_id, added_by, title, content_type, status, current_episode, max_episodes, notes, poster_url, genres, created_at, updated_at
	`, req.Title, req.ContentType, req.Status, req.CurrentEpisode, req.MaxEpisodes, req.Notes, req.PosterURL, req.Genres, itemID, groupID, userID).
		Scan(&item.ID, &item.GroupID, &item.AddedBy, &item.Title, &item.ContentType, &item.Status, &item.CurrentEpisode, &item.MaxEpisodes,
			&item.Notes, &item.PosterURL, &item.Genres, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}

	// Добавим username автора
	_ = r.db.QueryRow(ctx, "SELECT username FROM users WHERE id = $1", item.AddedBy).Scan(&item.AddedByUsername)

	// Загружаем оценки и ссылки
	ratings, _ := r.GetItemRatings(ctx, itemID)
	links, _ := r.GetItemLinks(ctx, itemID)
	item.Ratings = ratings
	item.Links = links
	item.AverageRating = calculateAverage(ratings)

	return item, nil
}

func (r *postgresGroupRepository) UpdateGroupItemStatusOnly(ctx context.Context, groupID, itemID int64, status model.ReviewStatus) (*model.GroupItem, error) {
	item := &model.GroupItem{}
	err := r.db.QueryRow(ctx, `
		UPDATE group_items
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND group_id = $3
		RETURNING id, group_id, added_by, title, content_type, status, current_episode, max_episodes, notes, poster_url, genres, created_at, updated_at
	`, status, itemID, groupID).
		Scan(&item.ID, &item.GroupID, &item.AddedBy, &item.Title, &item.ContentType, &item.Status, &item.CurrentEpisode, &item.MaxEpisodes,
			&item.Notes, &item.PosterURL, &item.Genres, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}

	_ = r.db.QueryRow(ctx, "SELECT username FROM users WHERE id = $1", item.AddedBy).Scan(&item.AddedByUsername)

	ratings, _ := r.GetItemRatings(ctx, itemID)
	links, _ := r.GetItemLinks(ctx, itemID)
	item.Ratings = ratings
	item.Links = links
	item.AverageRating = calculateAverage(ratings)

	return item, nil
}

func (r *postgresGroupRepository) DeleteGroupItem(ctx context.Context, groupID, itemID, userID int64) error {
	res, err := r.db.Exec(ctx, "DELETE FROM group_items WHERE id = $1 AND group_id = $2 AND added_by = $3", itemID, groupID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *postgresGroupRepository) GetGroupItemByID(ctx context.Context, itemID int64) (*model.GroupItem, error) {
	item := &model.GroupItem{}
	err := r.db.QueryRow(ctx, `
		SELECT id, group_id, added_by, title, content_type, status, current_episode, max_episodes, notes, poster_url, genres, created_at, updated_at
		FROM group_items WHERE id = $1
	`, itemID).Scan(&item.ID, &item.GroupID, &item.AddedBy, &item.Title, &item.ContentType, &item.Status, &item.CurrentEpisode, &item.MaxEpisodes,
		&item.Notes, &item.PosterURL, &item.Genres, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}

	_ = r.db.QueryRow(ctx, "SELECT username FROM users WHERE id = $1", item.AddedBy).Scan(&item.AddedByUsername)

	ratings, err := r.GetItemRatings(ctx, itemID)
	if err != nil {
		return nil, err
	}
	links, err := r.GetItemLinks(ctx, itemID)
	if err != nil {
		return nil, err
	}

	item.Ratings = ratings
	item.Links = links
	item.AverageRating = calculateAverage(ratings)

	return item, nil
}

func (r *postgresGroupRepository) GetGroupItems(ctx context.Context, groupID int64) ([]model.GroupItem, error) {
	query := `
		SELECT gi.id, gi.group_id, gi.added_by, u.username as added_by_username, gi.title, gi.content_type, gi.status, gi.current_episode, gi.max_episodes, gi.notes, gi.poster_url, gi.genres, gi.created_at, gi.updated_at
		FROM group_items gi
		JOIN users u ON gi.added_by = u.id
		WHERE gi.group_id = $1
		ORDER BY gi.updated_at DESC
	`
	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.GroupItem
	var itemIDs []int64
	itemMap := make(map[int64]*model.GroupItem)

	for rows.Next() {
		it := model.GroupItem{Ratings: []model.GroupItemRating{}, Links: []model.GroupItemLink{}}
		err := rows.Scan(&it.ID, &it.GroupID, &it.AddedBy, &it.AddedByUsername, &it.Title, &it.ContentType, &it.Status, &it.CurrentEpisode, &it.MaxEpisodes,
			&it.Notes, &it.PosterURL, &it.Genres, &it.CreatedAt, &it.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
		itemIDs = append(itemIDs, it.ID)
	}

	if len(items) == 0 {
		return []model.GroupItem{}, nil
	}

	// Для удобства маппинга
	for i := range items {
		itemMap[items[i].ID] = &items[i]
	}

	// Загружаем все оценки одним запросом
	ratingRows, err := r.db.Query(ctx, `
		SELECT gir.id, gir.item_id, gir.user_id, u.username, gir.rating, gir.updated_at
		FROM group_item_ratings gir
		JOIN users u ON gir.user_id = u.id
		WHERE gir.item_id = ANY($1)
	`, itemIDs)
	if err == nil {
		defer ratingRows.Close()
		for ratingRows.Next() {
			rat := model.GroupItemRating{}
			if err := ratingRows.Scan(&rat.ID, &rat.ItemID, &rat.UserID, &rat.Username, &rat.Rating, &rat.UpdatedAt); err == nil {
				if item, ok := itemMap[rat.ItemID]; ok {
					item.Ratings = append(item.Ratings, rat)
				}
			}
		}
	}

	// Загружаем все ссылки одним запросом
	linkRows, err := r.db.Query(ctx, `
		SELECT gil.id, gil.item_id, gil.user_id, u.username, gil.label, gil.url, gil.created_at
		FROM group_item_links gil
		JOIN users u ON gil.user_id = u.id
		WHERE gil.item_id = ANY($1)
		ORDER BY gil.created_at ASC
	`, itemIDs)
	if err == nil {
		defer linkRows.Close()
		for linkRows.Next() {
			lnk := model.GroupItemLink{}
			if err := linkRows.Scan(&lnk.ID, &lnk.ItemID, &lnk.UserID, &lnk.Username, &lnk.Label, &lnk.URL, &lnk.CreatedAt); err == nil {
				if item, ok := itemMap[lnk.ItemID]; ok {
					item.Links = append(item.Links, lnk)
				}
			}
		}
	}

	// Вычисляем среднюю оценку для каждой записи
	for i := range items {
		items[i].AverageRating = calculateAverage(items[i].Ratings)
	}

	return items, nil
}

// ── Оценки ──

func (r *postgresGroupRepository) UpdateItemRating(ctx context.Context, itemID, userID int64, rating *int16) (*model.GroupItemRating, error) {
	var err error
	if rating == nil {
		_, err = r.db.Exec(ctx, "DELETE FROM group_item_ratings WHERE item_id = $1 AND user_id = $2", itemID, userID)
		if err != nil {
			return nil, err
		}
		return &model.GroupItemRating{ItemID: itemID, UserID: userID, Rating: nil}, nil
	}

	rat := &model.GroupItemRating{}
	err = r.db.QueryRow(ctx, `
		INSERT INTO group_item_ratings (item_id, user_id, rating, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (item_id, user_id) DO UPDATE SET rating = EXCLUDED.rating, updated_at = NOW()
		RETURNING id, item_id, user_id, rating, updated_at
	`, itemID, userID, rating).Scan(&rat.ID, &rat.ItemID, &rat.UserID, &rat.Rating, &rat.UpdatedAt)
	if err != nil {
		return nil, err
	}

	_ = r.db.QueryRow(ctx, "SELECT username FROM users WHERE id = $1", userID).Scan(&rat.Username)
	return rat, nil
}

func (r *postgresGroupRepository) GetItemRatings(ctx context.Context, itemID int64) ([]model.GroupItemRating, error) {
	query := `
		SELECT gir.id, gir.item_id, gir.user_id, u.username, gir.rating, gir.updated_at
		FROM group_item_ratings gir
		JOIN users u ON gir.user_id = u.id
		WHERE gir.item_id = $1
	`
	rows, err := r.db.Query(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []model.GroupItemRating
	for rows.Next() {
		rat := model.GroupItemRating{}
		if err := rows.Scan(&rat.ID, &rat.ItemID, &rat.UserID, &rat.Username, &rat.Rating, &rat.UpdatedAt); err != nil {
			return nil, err
		}
		ratings = append(ratings, rat)
	}
	if ratings == nil {
		ratings = []model.GroupItemRating{}
	}
	return ratings, nil
}

// ── Ссылки ──

func (r *postgresGroupRepository) AddGroupItemLink(ctx context.Context, itemID, userID int64, label, url string) (*model.GroupItemLink, error) {
	link := &model.GroupItemLink{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO group_item_links (item_id, user_id, label, url, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, item_id, user_id, label, url, created_at
	`, itemID, userID, label, url).Scan(&link.ID, &link.ItemID, &link.UserID, &link.Label, &link.URL, &link.CreatedAt)
	if err != nil {
		return nil, err
	}

	_ = r.db.QueryRow(ctx, "SELECT username FROM users WHERE id = $1", userID).Scan(&link.Username)
	return link, nil
}

func (r *postgresGroupRepository) DeleteGroupItemLink(ctx context.Context, linkID, itemID, userID int64) error {
	res, err := r.db.Exec(ctx, `
		DELETE FROM group_item_links
		WHERE id = $1 AND item_id = $2 AND user_id = $3
	`, linkID, itemID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("link not found")
	}
	return nil
}

func (r *postgresGroupRepository) GetItemLinks(ctx context.Context, itemID int64) ([]model.GroupItemLink, error) {
	query := `
		SELECT gil.id, gil.item_id, gil.user_id, u.username, gil.label, gil.url, gil.created_at
		FROM group_item_links gil
		JOIN users u ON gil.user_id = u.id
		WHERE gil.item_id = $1
		ORDER BY gil.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.GroupItemLink
	for rows.Next() {
		lnk := model.GroupItemLink{}
		if err := rows.Scan(&lnk.ID, &lnk.ItemID, &lnk.UserID, &lnk.Username, &lnk.Label, &lnk.URL, &lnk.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, lnk)
	}
	if links == nil {
		links = []model.GroupItemLink{}
	}
	return links, nil
}

// ── Вспомогательные ──

func calculateAverage(ratings []model.GroupItemRating) float64 {
	if len(ratings) == 0 {
		return 0
	}
	var sum float64
	var count float64
	for _, r := range ratings {
		if r.Rating != nil {
			sum += float64(*r.Rating)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / count
}

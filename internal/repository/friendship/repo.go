package friendship

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KaoriEl/golang-boilerplate/internal/entity"
)

type Repository struct {
	db *pgxpool.Pool
	qb squirrel.StatementBuilderType
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
		qb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Create создает запрос на дружбу
func (r *Repository) Create(ctx context.Context, friendship *entity.Friendship) error {
	query, args, err := r.qb.
		Insert("friendships").
		Columns("id", "user_id", "friend_id", "status").
		Values(friendship.ID, friendship.UserID, friendship.FriendID, friendship.Status).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	err = r.db.QueryRow(ctx, query, args...).Scan(&friendship.CreatedAt, &friendship.UpdatedAt)
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	return nil
}

// UpdateStatus обновляет статус дружбы
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query, args, err := r.qb.
		Update("friendships").
		Set("status", status).
		Set("updated_at", squirrel.Expr("now()")).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	return nil
}

// Delete удаляет дружбу
func (r *Repository) Delete(ctx context.Context, userID, friendID uuid.UUID) error {
	query, args, err := r.qb.
		Delete("friendships").
		Where(squirrel.Or{
			squirrel.And{
				squirrel.Eq{"user_id": userID},
				squirrel.Eq{"friend_id": friendID},
			},
			squirrel.And{
				squirrel.Eq{"user_id": friendID},
				squirrel.Eq{"friend_id": userID},
			},
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	return nil
}

// GetFriends получает список друзей пользователя
func (r *Repository) GetFriends(ctx context.Context, userID uuid.UUID) ([]*entity.Friendship, error) {
	query, args, err := r.qb.
		Select("id", "user_id", "friend_id", "status", "created_at", "updated_at").
		From("friendships").
		Where(squirrel.Or{
			squirrel.Eq{"user_id": userID},
			squirrel.Eq{"friend_id": userID},
		}).
		Where(squirrel.Eq{"status": "accepted"}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	var friendships []*entity.Friendship
	for rows.Next() {
		f := &entity.Friendship{}
		err := rows.Scan(&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		friendships = append(friendships, f)
	}

	return friendships, nil
}

// GetPendingRequests получает входящие запросы на дружбу
func (r *Repository) GetPendingRequests(ctx context.Context, userID uuid.UUID) ([]*entity.Friendship, error) {
	query, args, err := r.qb.
		Select("id", "user_id", "friend_id", "status", "created_at", "updated_at").
		From("friendships").
		Where(squirrel.Eq{"friend_id": userID, "status": "pending"}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	var friendships []*entity.Friendship
	for rows.Next() {
		f := &entity.Friendship{}
		err := rows.Scan(&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		friendships = append(friendships, f)
	}

	return friendships, nil
}

// CheckFriendship проверяет существование активной дружбы (pending или accepted)
func (r *Repository) CheckFriendship(ctx context.Context, userID, friendID uuid.UUID) (*entity.Friendship, error) {
	query, args, err := r.qb.
		Select("id", "user_id", "friend_id", "status", "created_at", "updated_at").
		From("friendships").
		Where(squirrel.Or{
			squirrel.And{
				squirrel.Eq{"user_id": userID},
				squirrel.Eq{"friend_id": friendID},
			},
			squirrel.And{
				squirrel.Eq{"user_id": friendID},
				squirrel.Eq{"friend_id": userID},
			},
		}).
		Where(squirrel.Or{
			squirrel.Eq{"status": "pending"},
			squirrel.Eq{"status": "accepted"},
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	f := &entity.Friendship{}
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("execute query: %w", err)
	}

	return f, nil
}

// CheckRejectedFriendship проверяет существование отклоненной дружбы
func (r *Repository) CheckRejectedFriendship(ctx context.Context, userID, friendID uuid.UUID) (*entity.Friendship, error) {
	query, args, err := r.qb.
		Select("id", "user_id", "friend_id", "status", "created_at", "updated_at").
		From("friendships").
		Where(squirrel.Or{
			squirrel.And{
				squirrel.Eq{"user_id": userID},
				squirrel.Eq{"friend_id": friendID},
			},
			squirrel.And{
				squirrel.Eq{"user_id": friendID},
				squirrel.Eq{"friend_id": userID},
			},
		}).
		Where(squirrel.Eq{"status": "rejected"}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	f := &entity.Friendship{}
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("execute query: %w", err)
	}

	return f, nil
}

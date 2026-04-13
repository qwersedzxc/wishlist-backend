package friendship

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/qwersedzxc/wishlist-backend/internal/entity"
)

type FriendshipRepo interface {
	Create(ctx context.Context, f *entity.Friendship) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, userID, friendID uuid.UUID) error
	GetFriends(ctx context.Context, userID uuid.UUID) ([]*entity.Friendship, error)
	GetPendingRequests(ctx context.Context, userID uuid.UUID) ([]*entity.Friendship, error)
	CheckFriendship(ctx context.Context, userID, friendID uuid.UUID) (*entity.Friendship, error)
	CheckRejectedFriendship(ctx context.Context, userID, friendID uuid.UUID) (*entity.Friendship, error)
}

type UserRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]*entity.User, error)
	GetAllUsers(ctx context.Context) ([]*entity.User, error)
}

type UseCase struct {
	friendshipRepo FriendshipRepo
	userRepo       UserRepo
	log            *slog.Logger
}

func New(fr FriendshipRepo, ur UserRepo, log *slog.Logger) *UseCase {
	return &UseCase{friendshipRepo: fr, userRepo: ur, log: log}
}

// SendRequest отправляет запрос на дружбу
func (uc *UseCase) SendRequest(ctx context.Context, userID, friendID uuid.UUID) (*entity.Friendship, error) {
	if userID == friendID {
		return nil, fmt.Errorf("cannot add yourself as friend")
	}

	// Проверяем существующую дружбу (любой статус)
	existing, err := uc.friendshipRepo.CheckFriendship(ctx, userID, friendID)
	if err != nil {
		return nil, err
	}
	
	// Если есть активная дружба (pending или accepted), запрещаем
	if existing != nil {
		return nil, fmt.Errorf("friendship already exists")
	}

	// Проверяем, есть ли отклоненная дружба
	rejectedFriendship, err := uc.checkRejectedFriendship(ctx, userID, friendID)
	if err != nil {
		return nil, err
	}

	if rejectedFriendship != nil {
		// Обновляем существующую отклоненную дружбу на pending
		err = uc.friendshipRepo.UpdateStatus(ctx, rejectedFriendship.ID, "pending")
		if err != nil {
			return nil, err
		}
		rejectedFriendship.Status = "pending"
		return rejectedFriendship, nil
	}

	// Создаем новую дружбу
	f := &entity.Friendship{
		ID:       uuid.New(),
		UserID:   userID,
		FriendID: friendID,
		Status:   "pending",
	}
	if err := uc.friendshipRepo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

// checkRejectedFriendship проверяет наличие отклоненной дружбы
func (uc *UseCase) checkRejectedFriendship(ctx context.Context, userID, friendID uuid.UUID) (*entity.Friendship, error) {
	return uc.friendshipRepo.CheckRejectedFriendship(ctx, userID, friendID)
}

// AcceptRequest принимает запрос на дружбу
func (uc *UseCase) AcceptRequest(ctx context.Context, friendshipID uuid.UUID) error {
	return uc.friendshipRepo.UpdateStatus(ctx, friendshipID, "accepted")
}

// RejectRequest отклоняет запрос на дружбу
func (uc *UseCase) RejectRequest(ctx context.Context, friendshipID uuid.UUID) error {
	return uc.friendshipRepo.UpdateStatus(ctx, friendshipID, "rejected")
}

// RemoveFriend удаляет дружбу
func (uc *UseCase) RemoveFriend(ctx context.Context, userID, friendID uuid.UUID) error {
	return uc.friendshipRepo.Delete(ctx, userID, friendID)
}

// GetFriends возвращает список друзей с данными пользователей
func (uc *UseCase) GetFriends(ctx context.Context, userID uuid.UUID) ([]*entity.Friendship, error) {
	return uc.friendshipRepo.GetFriends(ctx, userID)
}

// GetPendingRequests возвращает входящие запросы
func (uc *UseCase) GetPendingRequests(ctx context.Context, userID uuid.UUID) ([]*entity.Friendship, error) {
	return uc.friendshipRepo.GetPendingRequests(ctx, userID)
}

// SearchUsers ищет пользователей
func (uc *UseCase) SearchUsers(ctx context.Context, query string) ([]*entity.User, error) {
	return uc.userRepo.SearchUsers(ctx, query, 20)
}

// GetUserByID возвращает пользователя по ID
func (uc *UseCase) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

// GetAllUsers возвращает всех пользователей (для проверки дней рождения)
func (uc *UseCase) GetAllUsers(ctx context.Context) ([]*entity.User, error) {
	return uc.userRepo.GetAllUsers(ctx)
}

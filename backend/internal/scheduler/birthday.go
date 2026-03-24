package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/KaoriEl/golang-boilerplate/internal/entity"
	"github.com/KaoriEl/golang-boilerplate/internal/types"
)

// BirthdayScheduler проверяет дни рождения друзей и отправляет уведомления
type BirthdayScheduler struct {
	friendshipUC FriendshipUseCase
	emailService EmailService
	log          *slog.Logger
}

// FriendshipUseCase интерфейс для работы с друзьями
type FriendshipUseCase interface {
	GetFriends(ctx context.Context, userID uuid.UUID) ([]*entity.Friendship, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetAllUsers(ctx context.Context) ([]*entity.User, error)
}

// EmailService интерфейс для отправки email
type EmailService interface {
	SendBirthdayReminder(to string, data types.BirthdayReminderData) error
}

// New создает новый планировщик
func New(friendshipUC FriendshipUseCase, emailService EmailService, log *slog.Logger) *BirthdayScheduler {
	return &BirthdayScheduler{
		friendshipUC: friendshipUC,
		emailService: emailService,
		log:          log,
	}
}

// Start запускает планировщик проверки дней рождения
func (s *BirthdayScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Проверяем раз в день
	defer ticker.Stop()

	// Запускаем первую проверку сразу
	s.checkBirthdays(ctx)

	for {
		select {
		case <-ctx.Done():
			s.log.Info("birthday scheduler stopped")
			return
		case <-ticker.C:
			s.checkBirthdays(ctx)
		}
	}
}

// checkBirthdays проверяет дни рождения и отправляет уведомления
func (s *BirthdayScheduler) checkBirthdays(ctx context.Context) {
	s.log.Info("checking birthdays...")

	users, err := s.friendshipUC.GetAllUsers(ctx)
	if err != nil {
		s.log.Error("failed to get users for birthday check", "error", err)
		return
	}

	today := time.Now()
	
	for _, user := range users {
		if user.BirthDate == nil {
			continue // Пропускаем пользователей без даты рождения
		}

		// Проверяем, есть ли день рождения в ближайшие 3 дня
		daysUntilBirthday := s.calculateDaysUntilBirthday(*user.BirthDate, today)
		
		if daysUntilBirthday >= 0 && daysUntilBirthday <= 3 {
			s.notifyFriendsAboutBirthday(ctx, user, daysUntilBirthday)
		}
	}
}

// calculateDaysUntilBirthday вычисляет количество дней до дня рождения
func (s *BirthdayScheduler) calculateDaysUntilBirthday(birthDate, today time.Time) int {
	// Получаем день рождения в этом году
	thisYear := today.Year()
	birthday := time.Date(thisYear, birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, today.Location())
	
	// Если день рождения уже прошел в этом году, берем следующий год
	if birthday.Before(today) {
		birthday = time.Date(thisYear+1, birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, today.Location())
	}
	
	duration := birthday.Sub(today)
	return int(duration.Hours() / 24)
}

// notifyFriendsAboutBirthday уведомляет друзей о дне рождения
func (s *BirthdayScheduler) notifyFriendsAboutBirthday(ctx context.Context, birthdayUser *entity.User, daysLeft int) {
	friends, err := s.friendshipUC.GetFriends(ctx, birthdayUser.ID)
	if err != nil {
		s.log.Error("failed to get friends for birthday notification", "error", err, "userID", birthdayUser.ID)
		return
	}

	for _, friendship := range friends {
		// Определяем ID друга
		friendID := friendship.FriendID
		if friendship.UserID != birthdayUser.ID {
			friendID = friendship.UserID
		}

		friend, err := s.friendshipUC.GetUserByID(ctx, friendID)
		if err != nil {
			s.log.Error("failed to get friend for birthday notification", "error", err, "friendID", friendID)
			continue
		}

		// Отправляем уведомление в горутине
		go func(friend *entity.User, birthdayUser *entity.User, daysLeft int) {
			emailData := types.BirthdayReminderData{
				RecipientName: friend.Username,
				FriendName:    birthdayUser.Username,
				BirthDate:     birthdayUser.BirthDate.Format("02.01"),
				DaysLeft:      daysLeft,
				AppURL:        "http://localhost:3000", // TODO: получать из конфига
			}

			if err := s.emailService.SendBirthdayReminder(friend.Email, emailData); err != nil {
				s.log.Error("failed to send birthday reminder email", 
					"error", err, 
					"to", friend.Email, 
					"birthdayUser", birthdayUser.Username)
			} else {
				s.log.Info("birthday reminder email sent", 
					"to", friend.Email, 
					"birthdayUser", birthdayUser.Username,
					"daysLeft", daysLeft)
			}
		}(friend, birthdayUser, daysLeft)
	}
}
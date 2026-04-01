package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/KaoriEl/golang-boilerplate/internal/dto"
	"github.com/KaoriEl/golang-boilerplate/internal/entity"
)

type UseCase struct {
	userRepo UserRepository
	roleRepo RoleRepository
	jwtSecret string
	log      *slog.Logger
}

func New(userRepo UserRepository, roleRepo RoleRepository, jwtSecret string, log *slog.Logger) *UseCase {
	return &UseCase{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		jwtSecret: jwtSecret,
		log:       log,
	}
}

func userToOutput(user *entity.User) *dto.UserOutput {
	return &dto.UserOutput{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		AvatarURL: user.AvatarURL,
		FullName:  user.FullName,
		BirthDate: user.BirthDate,
		Bio:       user.Bio,
		Phone:     user.Phone,
		City:      user.City,
		CreatedAt: user.CreatedAt,
	}
}

// Register регистрирует нового пользователя
func (uc *UseCase) Register(ctx context.Context, input dto.UserRegisterInput) (*dto.AuthResponse, error) {
	// Проверяем, не существует ли пользователь с таким email
	existingUser, _ := uc.userRepo.GetByEmail(ctx, input.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Проверяем, не существует ли пользователь с таким username
	existingUser, _ = uc.userRepo.GetByUsername(ctx, input.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("user with this username already exists")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	hashedPasswordStr := string(hashedPassword)

	// Создаем пользователя
	user := &entity.User{
		ID:           uuid.New(),
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: &hashedPasswordStr,
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Назначаем роль "user" новому пользователю
	if uc.roleRepo != nil {
		userRole, err := uc.roleRepo.GetRoleByName(ctx, "user")
		if err == nil && userRole != nil {
			_ = uc.roleRepo.AssignRoleToUser(ctx, user.ID, userRole.ID, nil)
		}
	}

	// Генерируем JWT токен
	token, err := uc.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &dto.AuthResponse{User: *userToOutput(user), Token: token}, nil
}

// Login авторизует пользователя
func (uc *UseCase) Login(ctx context.Context, input dto.UserLoginInput) (*dto.AuthResponse, error) {
	// Получаем пользователя по email
	user, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Проверяем пароль
	if user.PasswordHash == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Генерируем JWT токен
	token, err := uc.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &dto.AuthResponse{User: *userToOutput(user), Token: token}, nil
}

// FindOrCreateByOAuth находит или создаёт пользователя по данным OAuth провайдера
func (uc *UseCase) FindOrCreateByOAuth(ctx context.Context, provider, providerID, email, name, avatarURL string) (*dto.AuthResponse, error) {
	// Ищем по provider + providerID
	user, err := uc.userRepo.GetByProviderID(ctx, provider, providerID)
	if err != nil {
		// Не нашли — ищем по email
		user, err = uc.userRepo.GetByEmail(ctx, email)
		if err != nil {
			// Создаём нового пользователя
			username := name
			if username == "" {
				username = email
			}
			// Убираем пробелы из username
			for i, c := range username {
				if c == ' ' {
					username = username[:i] + "_" + username[i+1:]
				}
			}

			p := provider
			pid := providerID
			av := avatarURL
			user = &entity.User{
				ID:         uuid.New(),
				Email:      email,
				Username:   username,
				Provider:   &p,
				ProviderID: &pid,
				AvatarURL:  &av,
			}
			if err := uc.userRepo.Create(ctx, user); err != nil {
				return nil, fmt.Errorf("create user: %w", err)
			}

			uc.log.Info("new user created via OAuth", "userID", user.ID, "email", email, "provider", provider)

			// Назначаем роль "user" новому пользователю
			if uc.roleRepo != nil {
				userRole, err := uc.roleRepo.GetRoleByName(ctx, "user")
				if err != nil {
					uc.log.Error("failed to get user role", "error", err)
				} else if userRole != nil {
					err = uc.roleRepo.AssignRoleToUser(ctx, user.ID, userRole.ID, nil)
					if err != nil {
						uc.log.Error("failed to assign role", "error", err, "userID", user.ID, "roleID", userRole.ID)
					} else {
						uc.log.Info("role assigned to user", "userID", user.ID, "roleID", userRole.ID, "roleName", userRole.Name)
					}
				} else {
					uc.log.Warn("user role not found in database")
				}
			} else {
				uc.log.Warn("roleRepo is nil")
			}
		}
	}

	token, err := uc.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &dto.AuthResponse{User: *userToOutput(user), Token: token}, nil
}

// UpdateProfile обновляет профиль пользователя
func (uc *UseCase) UpdateProfile(ctx context.Context, id uuid.UUID, input dto.UpdateProfileInput) (*dto.UserOutput, error) {
	user, err := uc.userRepo.UpdateProfile(ctx, id, input.FullName, input.Bio, input.Phone, input.City, input.BirthDate)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return userToOutput(user), nil
}

// GetUserByID получает пользователя по ID
func (uc *UseCase) GetUserByID(ctx context.Context, id uuid.UUID) (*dto.UserOutput, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return userToOutput(user), nil
}

// ValidateToken проверяет JWT токен и возвращает ID пользователя
func (uc *UseCase) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(uc.jwtSecret), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return uuid.Nil, fmt.Errorf("invalid token claims")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, fmt.Errorf("parse user id: %w", err)
		}

		return userID, nil
	}

	return uuid.Nil, fmt.Errorf("invalid token")
}

// generateToken генерирует JWT токен для пользователя
func (uc *UseCase) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 дней
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return tokenString, nil
}

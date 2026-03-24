package oauth

import (
	"context"
	"fmt"

	"github.com/KaoriEl/golang-boilerplate/internal/config"
	"golang.org/x/oauth2"
)

// UserInfo данные пользователя, полученные от OAuth2 провайдера.
type UserInfo struct {
	ExternalID string
	Login      string
	Name       string
	Email      string
	AvatarURL  string
}

// Provider контракт OAuth2-провайдера.
type Provider interface {
	Config() *oauth2.Config
	GetUserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error)
}

// New создаёт провайдер по имени из конфига.
// Чтобы добавить новый провайдер:
//  1. Создай файл internal/oauth/<name>.go
//  2. Реализуй интерфейс Provider
//  3. Добавь case в этот switch
func New(cfg config.OAuthCfg) (Provider, error) {
	switch cfg.Provider {
	case "github":
		return newGitHubProvider(cfg), nil
	case "google":
		return newGoogleProvider(cfg), nil
	case "vk":
		return newVKProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unknown oauth provider %q", cfg.Provider)
	}
}

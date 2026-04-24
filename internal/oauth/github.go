package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/qwersedzxc/wishlist-backend/internal/config"
	"golang.org/x/oauth2"
)

// gitHubProvider реализует Provider для GitHub OAuth2.
type gitHubProvider struct {
	cfg config.OAuthCfg
}

func newGitHubProvider(cfg config.OAuthCfg) *gitHubProvider {
	return &gitHubProvider{cfg: cfg}
}

// Config возвращает oauth2.Config для GitHub.
func (p *gitHubProvider) Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		RedirectURL:  p.cfg.RedirectURL,
		Scopes:       []string{"user:email"},
		Endpoint: oauth2.Endpoint{ //nolint:gosec
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}
}

// GetUserInfo запрашивает данные пользователя у GitHub API.
func (p *gitHubProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error) {
	client := p.Config().Client(ctx, token)
	
	// Получаем основные данные пользователя
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return UserInfo{}, fmt.Errorf("get user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UserInfo{}, fmt.Errorf("github api status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UserInfo{}, fmt.Errorf("read body: %w", err)
	}

	var user struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return UserInfo{}, fmt.Errorf("unmarshal: %w", err)
	}

	// Если email не публичный, запрашиваем отдельно
	email := user.Email
	if email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			emailBody, _ := io.ReadAll(emailResp.Body)
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.Unmarshal(emailBody, &emails) == nil {
				for _, e := range emails {
					if e.Primary {
						email = e.Email
						break
					}
				}
				if email == "" && len(emails) > 0 {
					email = emails[0].Email
				}
			}
		}
	}

	name := user.Name
	if name == "" {
		name = user.Login
	}

	return UserInfo{
		ExternalID: fmt.Sprintf("%d", user.ID),
		Login:      user.Login,
		Email:      email,
		Name:       name,
		AvatarURL:  user.AvatarURL,
	}, nil
}

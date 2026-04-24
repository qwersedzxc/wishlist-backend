package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/qwersedzxc/wishlist-backend/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleProvider struct {
	cfg config.OAuthCfg
}

func newGoogleProvider(cfg config.OAuthCfg) *googleProvider {
	return &googleProvider{cfg: cfg}
}

func (p *googleProvider) Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		RedirectURL:  p.cfg.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func (p *googleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error) {
	client := p.Config().Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return UserInfo{}, fmt.Errorf("get userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UserInfo{}, fmt.Errorf("userinfo status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UserInfo{}, fmt.Errorf("read body: %w", err)
	}

	var data struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return UserInfo{}, fmt.Errorf("unmarshal: %w", err)
	}

	return UserInfo{
		ExternalID: data.Sub,
		Email:      data.Email,
		Name:       data.Name,
		AvatarURL:  data.Picture,
	}, nil
}

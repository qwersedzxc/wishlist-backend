package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/KaoriEl/golang-boilerplate/internal/config"
	"golang.org/x/oauth2"
)

type vkProvider struct {
	cfg config.OAuthCfg
}

func newVKProvider(cfg config.OAuthCfg) *vkProvider {
	return &vkProvider{cfg: cfg}
}

func (p *vkProvider) Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		RedirectURL:  p.cfg.RedirectURL,
		Scopes:       []string{"email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.vk.com/authorize",
			TokenURL: "https://oauth.vk.com/access_token",
		},
	}
}

func (p *vkProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error) {
	// VK возвращает email прямо в токене (extra field)
	email, _ := token.Extra("email").(string)
	userID, _ := token.Extra("user_id").(float64)

	// Запрашиваем профиль через API
	params := url.Values{}
	params.Set("access_token", token.AccessToken)
	params.Set("fields", "photo_200,screen_name")
	params.Set("v", "5.131")

	apiURL := "https://api.vk.com/method/users.get?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return UserInfo{}, fmt.Errorf("build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UserInfo{}, fmt.Errorf("get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UserInfo{}, fmt.Errorf("read body: %w", err)
	}

	var result struct {
		Response []struct {
			ID         int    `json:"id"`
			FirstName  string `json:"first_name"`
			LastName   string `json:"last_name"`
			ScreenName string `json:"screen_name"`
			Photo      string `json:"photo_200"`
		} `json:"response"`
	}
	if err := json.Unmarshal(body, &result); err != nil || len(result.Response) == 0 {
		return UserInfo{}, fmt.Errorf("parse response: %w", err)
	}

	u := result.Response[0]
	name := u.FirstName + " " + u.LastName
	if u.ScreenName != "" {
		name = u.ScreenName
	}

	externalID := fmt.Sprintf("%d", u.ID)
	if userID > 0 {
		externalID = fmt.Sprintf("%.0f", userID)
	}

	return UserInfo{
		ExternalID: externalID,
		Email:      email,
		Name:       name,
		AvatarURL:  u.Photo,
	}, nil
}
